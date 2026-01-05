// package httpapi

// import (
// 	"encoding/json"
// 	"net/http"
// 	"time"
// )

// type SyncHealthResponse struct {
// 	Status        string     `json:"status"` // ok | warn | error
// 	ToolName      string     `json:"tool_name"`
// 	SourceName    string     `json:"source_name"`
// 	TargetName    string     `json:"target_name"`
// 	LastSourceTS  *time.Time `json:"last_source_ts"`
// 	LastTargetTS  *time.Time `json:"last_target_ts"`
// 	LagSeconds    *int       `json:"lag_seconds"`
// 	LastSuccessAt *time.Time `json:"last_success_at"`
// 	LastError     *string    `json:"last_error"`

// 	// untuk “success criteria” demo
// 	SLATargetSeconds int `json:"sla_target_seconds"`
// }

// func (h *Handlers) GetSyncHealth(w http.ResponseWriter, r *http.Request) {
// 	var resp SyncHealthResponse
// 	resp.SLATargetSeconds = 10

// 	var lag sqlNullInt
// 	var lastErr sqlNullString
// 	var ls, lt, lsa sqlNullTime

// 	err := h.DB.QueryRow(`
// 		SELECT tool_name, source_name, target_name, last_source_ts, last_target_ts,
// 		       lag_seconds, last_success_at, last_error
// 		FROM sync_audit
// 		ORDER BY created_at DESC
// 		LIMIT 1`).Scan(
// 		&resp.ToolName, &resp.SourceName, &resp.TargetName,
// 		&ls, &lt, &lag, &lsa, &lastErr,
// 	)

// 	if err != nil {
// 		// Kalau belum ada data audit, tetap return status warn agar UI bisa kasih instruksi
// 		resp.Status = "warn"
// 		w.Header().Set("Content-Type", "application/json")
// 		_ = json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	resp.LastSourceTS = ls.Ptr()
// 	resp.LastTargetTS = lt.Ptr()
// 	resp.LagSeconds = lag.Ptr()
// 	resp.LastSuccessAt = lsa.Ptr()
// 	resp.LastError = lastErr.Ptr()

// 	// status simple
// 	resp.Status = "ok"
// 	if resp.LastError != nil && *resp.LastError != "" {
// 		resp.Status = "error"
// 	} else if resp.LagSeconds != nil && *resp.LagSeconds > resp.SLATargetSeconds {
// 		resp.Status = "warn"
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	enc := json.NewEncoder(w)
// 	enc.SetEscapeHTML(false)
// 	_ = enc.Encode(resp)
// }

// // Helpers untuk nullable scan
// type sqlNullInt struct{ V *int }

// func (n *sqlNullInt) Scan(src any) error {
// 	if src == nil {
// 		n.V = nil
// 		return nil
// 	}
// 	i := int(src.(int64))
// 	n.V = &i
// 	return nil
// }
// func (n sqlNullInt) Ptr() *int { return n.V }

// type sqlNullString struct{ V *string }

// func (n *sqlNullString) Scan(src any) error {
// 	if src == nil {
// 		n.V = nil
// 		return nil
// 	}
// 	s := src.(string)
// 	n.V = &s
// 	return nil
// }
// func (n sqlNullString) Ptr() *string { return n.V }

// type sqlNullTime struct{ V *time.Time }

// func (n *sqlNullTime) Scan(src any) error {
// 	if src == nil {
// 		n.V = nil
// 		return nil
// 	}
// 	t := src.(time.Time)
// 	n.V = &t
// 	return nil
// }
// func (n sqlNullTime) Ptr() *time.Time { return n.V }

// New SyncHealth for PostgresSQL
// internal/httpapi/sync_health.go
package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type SyncHealthResponse struct {
	Status     string `json:"status"` // ok | warn | error
	ToolName   string `json:"tool_name"`
	SourceName string `json:"source_name"`
	TargetName string `json:"target_name"`

	LastSourceTS  *time.Time `json:"last_source_ts"`
	LastTargetTS  *time.Time `json:"last_target_ts"`
	LagSeconds    *int       `json:"lag_seconds"`
	LastSuccessAt *time.Time `json:"last_success_at"`
	LastError     *string    `json:"last_error"`

	// untuk “success criteria” demo
	SLATargetSeconds int `json:"sla_target_seconds"`
}

func (h *Handlers) GetSyncHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp := SyncHealthResponse{
		SLATargetSeconds: 10,
		Status:           "warn", // default: warn kalau audit belum ada / belum stabil
	}

	// gunakan sql.Null* agar aman untuk NULL dari Postgres
	var (
		lastSourceTS  sql.NullTime
		lastTargetTS  sql.NullTime
		lagSeconds    sql.NullInt64
		lastSuccessAt sql.NullTime
		lastError     sql.NullString
	)

	err := h.DB.QueryRowContext(ctx, `
		SELECT
			tool_name,
			source_name,
			target_name,
			last_source_ts,
			last_target_ts,
			lag_seconds,
			last_success_at,
			last_error
		FROM sync_audit
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(
		&resp.ToolName,
		&resp.SourceName,
		&resp.TargetName,
		&lastSourceTS,
		&lastTargetTS,
		&lagSeconds,
		&lastSuccessAt,
		&lastError,
	)

	// Kalau tabel kosong / belum ada data audit: return warn agar UI bisa kasih instruksi.
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		writeError(w, http.StatusInternalServerError, "query sync_audit failed", err)
		return
	}

	// convert nullable -> pointer (sesuai JSON)
	if lastSourceTS.Valid {
		resp.LastSourceTS = &lastSourceTS.Time
	}
	if lastTargetTS.Valid {
		resp.LastTargetTS = &lastTargetTS.Time
	}
	if lagSeconds.Valid {
		v := int(lagSeconds.Int64)
		resp.LagSeconds = &v
	}
	if lastSuccessAt.Valid {
		resp.LastSuccessAt = &lastSuccessAt.Time
	}
	if lastError.Valid {
		s := lastError.String
		resp.LastError = &s
	}

	// hitung status simple
	resp.Status = "ok"
	if resp.LastError != nil && *resp.LastError != "" {
		resp.Status = "error"
	} else if resp.LagSeconds != nil && *resp.LagSeconds > resp.SLATargetSeconds {
		resp.Status = "warn"
	}

	// response
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(resp)
}
