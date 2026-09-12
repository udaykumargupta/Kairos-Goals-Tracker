package kairos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Neon talks to Postgres over Neon's SQL-over-HTTP endpoint instead of the wire
// protocol. That keeps this backend dependency-free (standard library only) and
// suits serverless well: every call is a stateless HTTPS request, so there are no
// pooled TCP connections to leak when function instances come and go.
type Neon struct {
	endpoint string // https://<host>/sql
	conn     string // the full postgres:// connection string, sent as a header
	client   *http.Client
}

// NewNeon derives the HTTP endpoint from a postgres:// connection string.
func NewNeon(dsn string) (*Neon, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid database URL: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		return nil, fmt.Errorf("database URL has no host")
	}
	return &Neon{
		endpoint: "https://" + host + "/sql",
		conn:     dsn,
		client:   &http.Client{Timeout: 20 * time.Second},
	}, nil
}

type neonRequest struct {
	Query  string `json:"query"`
	Params []any  `json:"params"`
}

type neonResponse struct {
	Rows     []Row  `json:"rows"`
	Command  string `json:"command"`
	RowCount int    `json:"rowCount"`
	// Error fields, present when the statement fails.
	Message string `json:"message"`
	Code    string `json:"code"`
	Detail  string `json:"detail"`
}

// Row is one result row: column name -> raw JSON value.
type Row map[string]json.RawMessage

// Query runs a parameterised statement. Always pass values via args ($1, $2, …)
// so they are bound server-side and never concatenated into SQL.
func (n *Neon) Query(ctx context.Context, sql string, args ...any) ([]Row, error) {
	if args == nil {
		args = []any{}
	}
	body, err := json.Marshal(neonRequest{Query: sql, Params: args})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Neon-Connection-String", n.conn)
	// Return arrays/objects as JSON rather than Postgres text where possible.
	req.Header.Set("Neon-Raw-Text-Output", "false")

	res, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("database request failed: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return nil, err
	}

	var parsed neonResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("database returned status %d with an unreadable body", res.StatusCode)
	}
	if res.StatusCode != http.StatusOK {
		msg := parsed.Message
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		return nil, fmt.Errorf("database error (%d): %s", res.StatusCode, msg)
	}
	return parsed.Rows, nil
}

// Exec runs a statement whose rows we don't care about.
func (n *Neon) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := n.Query(ctx, sql, args...)
	return err
}

// QueryOne returns the first row, or nil when the result set is empty.
func (n *Neon) QueryOne(ctx context.Context, sql string, args ...any) (Row, error) {
	rows, err := n.Query(ctx, sql, args...)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

// ---- Row accessors -------------------------------------------------------
//
// Neon encodes values by Postgres type: int8 arrives as a JSON *string*, int4 as a
// number, bool as a bool. To keep every query unambiguous we cast ids and
// timestamps to text/bigint-text in SQL, so these helpers stay simple.

// IsNull reports whether the column is SQL NULL (or absent).
func (r Row) IsNull(col string) bool {
	v, ok := r[col]
	return !ok || string(v) == "null"
}

// Str returns a text column, with SQL NULL flattened to "".
func (r Row) Str(col string) string {
	if r.IsNull(col) {
		return ""
	}
	var s string
	if err := json.Unmarshal(r[col], &s); err == nil {
		return s
	}
	// Unquoted scalar (number/bool) — use its literal form.
	return string(r[col])
}

// Int64 parses an integer column, accepting both JSON numbers and quoted strings.
func (r Row) Int64(col string) int64 {
	if r.IsNull(col) {
		return 0
	}
	n, err := strconv.ParseInt(strings.Trim(string(r[col]), `"`), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// Bool returns a boolean column.
func (r Row) Bool(col string) bool {
	if r.IsNull(col) {
		return false
	}
	var b bool
	if err := json.Unmarshal(r[col], &b); err == nil {
		return b
	}
	return strings.Trim(string(r[col]), `"`) == "t"
}

// JSON returns a jsonb column as raw JSON. Columns selected as ::text arrive as a
// JSON string containing JSON, so unwrap that case before handing it back.
func (r Row) JSON(col string) json.RawMessage {
	if r.IsNull(col) {
		return nil
	}
	raw := r[col]
	if len(raw) > 0 && raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return json.RawMessage(s)
		}
	}
	return raw
}
