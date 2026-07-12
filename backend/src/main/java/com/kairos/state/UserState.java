package com.kairos.state;

import com.fasterxml.jackson.databind.JsonNode;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

import java.time.Instant;

/**
 * A single user's entire goal set, stored as one JSON document (PostgreSQL {@code jsonb}).
 * The primary key is the owning user's id (one row per user).
 *
 * <p>The document is mapped as a Jackson {@link JsonNode} so Hibernate stores it as raw
 * {@code jsonb} (no double-encoding). Per the chosen design the backend treats this JSON
 * as opaque — the frontend owns its shape — so the API evolves without DB migrations.
 */
@Entity
@Table(name = "user_state")
public class UserState {

    @Id
    @Column(name = "user_id")
    private Long userId;

    @JdbcTypeCode(SqlTypes.JSON)
    @Column(columnDefinition = "jsonb", nullable = false)
    private JsonNode data;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    protected UserState() {
        // for JPA
    }

    public UserState(Long userId, JsonNode data, Instant updatedAt) {
        this.userId = userId;
        this.data = data;
        this.updatedAt = updatedAt;
    }

    public void update(JsonNode data, Instant updatedAt) {
        this.data = data;
        this.updatedAt = updatedAt;
    }

    public Long getUserId() {
        return userId;
    }

    public JsonNode getData() {
        return data;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }
}
