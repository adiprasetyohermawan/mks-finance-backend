package httpapi

import (
	"net/http"
	"strings"
)

func NewRouter(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	// existing
	mux.HandleFunc("/api/v1/health", h.Health)

	// customers list (upgrade filter/sort)
	mux.HandleFunc("/api/v1/customers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.ListCustomers(w, r)
	})

	// customer profile
	mux.HandleFunc("/api/v1/customers/", func(w http.ResponseWriter, r *http.Request) {
		// expected: /api/v1/customers/{id}/profile
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/customers/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 2 && parts[1] == "profile" {
			// Handler implementation is responsible for extracting customer_id from the request path.
			h.GetCustomerProfile(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// KPI
	mux.HandleFunc("/api/v1/stats/kpi", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.GetKPI(w, r)
	})

	// Sync health
	mux.HandleFunc("/api/v1/sync/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.GetSyncHealth(w, r)
	})

	return mux
}
