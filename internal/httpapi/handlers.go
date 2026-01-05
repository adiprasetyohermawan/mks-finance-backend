// package httpapi

// import (
// 	"context"
// 	"database/sql"
// 	"encoding/json"
// 	"net/http"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// type Handlers struct {
// 	DB *sql.DB
// }

// func NewHandlers(db *sql.DB) *Handlers {
// 	return &Handlers{DB: db}
// }

// func writeJSON(w http.ResponseWriter, status int, v any) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)
// 	_ = json.NewEncoder(w).Encode(v)
// }

// func writeError(w http.ResponseWriter, status int, message string, err error) {
// 	payload := map[string]any{"error": message}
// 	if err != nil {
// 		payload["details"] = err.Error()
// 	}
// 	writeJSON(w, status, payload)
// }

// func parseIntDefault(s string, def int) int {
// 	s = strings.TrimSpace(s)
// 	if s == "" {
// 		return def
// 	}
// 	i, err := strconv.Atoi(s)
// 	if err != nil {
// 		return def
// 	}
// 	return i
// }

// func queryInt(r *http.Request, key string, def int) int {
// 	return parseIntDefault(r.URL.Query().Get(key), def)
// }

// // Health is used by GET /api/v1/health
// func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
// 	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
// 	defer cancel()

// 	if err := h.DB.PingContext(ctx); err != nil {
// 		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
// 			"status": "down",
// 			"db":     "not_ok",
// 			"error":  err.Error(),
// 		})
// 		return
// 	}

// 	writeJSON(w, http.StatusOK, map[string]any{
// 		"status": "ok",
// 		"db":     "ok",
// 	})
// }

// New Handlers for PostgresSQL
// internal/httpapi/handlers.go
package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handlers struct {
	DB *sql.DB
}

func NewHandlers(db *sql.DB) *Handlers {
	return &Handlers{DB: db}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Untuk API POC biasanya lebih aman non-cache agar hasil terbaru selalu tampil
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	// enc.SetIndent("", "  ") // uncomment kalau mau respons lebih “cantik” saat debug
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string, err error) {
	payload := map[string]any{
		"error": message,
	}
	if err != nil {
		payload["details"] = err.Error()
	}
	writeJSON(w, status, payload)
}

func parseIntDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return i
}

func queryInt(r *http.Request, key string, def int) int {
	return parseIntDefault(r.URL.Query().Get(key), def)
}

// Health is used by GET /api/v1/health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// PingContext cukup umum untuk semua driver.
	if err := h.DB.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "down",
			"db":     "not_ok",
			"error":  err.Error(),
		})
		return
	}

	// Optional (lebih meyakinkan utk Postgres): pastikan query bisa dieksekusi.
	// Ini membantu mendeteksi kasus koneksi ada tapi query gagal.
	var one int
	if err := h.DB.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "down",
			"db":     "not_ok",
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"db":     "ok",
	})
}
