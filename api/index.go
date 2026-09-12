// Package handler is the Vercel serverless entrypoint. Vercel compiles each .go
// file under api/ into its own function, so this is deliberately the only one:
// every /api/** request is rewritten here (see vercel.json) and dispatched by the
// shared router, which keeps all routes inside a single warm function.
package handler

import (
	"net/http"
	"sync"

	"github.com/udaykumargupta/kairos/internal/kairos"
)

var (
	once    sync.Once
	routes  http.Handler
	initErr error
)

// Handler is the entrypoint Vercel invokes.
func Handler(w http.ResponseWriter, r *http.Request) {
	// Built once per process; serverless reuses the process across warm requests,
	// so configuration parsing happens on the first call only.
	once.Do(func() {
		cfg, err := kairos.LoadConfig()
		if err != nil {
			initErr = err
			return
		}
		app, err := kairos.New(cfg)
		if err != nil {
			initErr = err
			return
		}
		routes = app.Routes()
	})

	if initErr != nil {
		http.Error(w, `{"error":"Internal Server Error","message":"Server is misconfigured"}`,
			http.StatusInternalServerError)
		return
	}
	routes.ServeHTTP(w, r)
}
