// package httpapi

// import (
// 	"net/http"
// 	"strings"
// )

// func NewRouter(h *Handlers) http.Handler {
// 	mux := http.NewServeMux()

// 	// existing
// 	mux.HandleFunc("/api/v1/health", h.Health)

// 	// customers list (upgrade filter/sort)
// 	mux.HandleFunc("/api/v1/customers", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet {
// 			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		h.ListCustomers(w, r)
// 	})

// 	// customer profile
// 	mux.HandleFunc("/api/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
// 		// expected: /api/v1/customers/{id}/profile
// 		if r.Method != http.MethodGet {
// 			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		path := strings.TrimPrefix(r.URL.Path, "/api/v1/customers/")
// 		parts := strings.Split(strings.Trim(path, "/"), "/")
// 		if len(parts) == 2 && parts[1] == "profile" {
// 			// Handler implementation is responsible for extracting customer_id from the request path.
// 			h.GetCustomerProfile(w, r)
// 			return
// 		}
// 		http.NotFound(w, r)
// 	})

// 	// KPI
// 	mux.HandleFunc("/api/v1/stats/kpi", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet {
// 			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		h.GetKPI(w, r)
// 	})

// 	// Sync health
// 	mux.HandleFunc("/api/v1/sync/health", func(w http.ResponseWriter, r *http.Request) {
// 		if r.Method != http.MethodGet {
// 			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
// 			return
// 		}
// 		h.GetSyncHealth(w, r)
// 	})

// 	return mux
// }

// New Router for PostgresSQL
// internal/httpapi/router.go
package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	// Basic middleware (aman untuk PoC, production-like)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	// Routes
	r.Get("/api/v1/health", h.Health)

	r.Get("/api/v1/customers", h.ListCustomers)
	// Ini akan membuat chi.URLParam(r, "customer_id") bekerja (di customer_profile_360.go)
	r.Get("/api/v1/customers/{customer_id}/profile", h.GetCustomerProfile)

	r.Get("/api/v1/stats/kpi", h.GetKPI)
	r.Get("/api/v1/sync/health", h.GetSyncHealth)

	// Optional: 404 handler custom (kalau mau)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	return r
}
