// package httpapi

// import (
// 	"encoding/json"
// 	"net/http"
// )

// type KPIResponse struct {
// 	Customers struct {
// 		Total     int            `json:"total"`
// 		Active    int            `json:"active"`
// 		ByGender  map[string]int `json:"by_gender"`
// 		BySegment map[string]int `json:"by_segment"`
// 	} `json:"customers"`
// 	CreditApplications struct {
// 		Total    int            `json:"total"`
// 		ByStatus map[string]int `json:"by_status"`
// 	} `json:"credit_applications"`
// 	VehicleOwnership struct {
// 		Total int `json:"total"`
// 	} `json:"vehicle_ownership"`
// }

// func (h *Handlers) GetKPI(w http.ResponseWriter, r *http.Request) {
// 	var resp KPIResponse
// 	resp.Customers.ByGender = map[string]int{}
// 	resp.Customers.BySegment = map[string]int{}
// 	resp.CreditApplications.ByStatus = map[string]int{}

// 	// customers total
// 	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM customers`).Scan(&resp.Customers.Total)
// 	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM customers WHERE status = 'Active'`).Scan(&resp.Customers.Active)

// 	// by gender
// 	rows, err := h.DB.Query(`SELECT gender, COUNT(*) FROM customers GROUP BY gender`)
// 	if err == nil {
// 		defer rows.Close()
// 		for rows.Next() {
// 			var k string
// 			var v int
// 			_ = rows.Scan(&k, &v)
// 			resp.Customers.ByGender[k] = v
// 		}
// 	}

// 	// by segment
// 	rows2, err := h.DB.Query(`SELECT customer_segment, COUNT(*) FROM customers GROUP BY customer_segment`)
// 	if err == nil {
// 		defer rows2.Close()
// 		for rows2.Next() {
// 			var k string
// 			var v int
// 			_ = rows2.Scan(&k, &v)
// 			resp.Customers.BySegment[k] = v
// 		}
// 	}

// 	// credit applications
// 	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM credit_applications`).Scan(&resp.CreditApplications.Total)
// 	rows3, err := h.DB.Query(`SELECT application_status, COUNT(*) FROM credit_applications GROUP BY application_status`)
// 	if err == nil {
// 		defer rows3.Close()
// 		for rows3.Next() {
// 			var k string
// 			var v int
// 			_ = rows3.Scan(&k, &v)
// 			resp.CreditApplications.ByStatus[k] = v
// 		}
// 	}

// 	// vehicles
// 	_ = h.DB.QueryRow(`SELECT COUNT(*) FROM vehicle_ownership`).Scan(&resp.VehicleOwnership.Total)

// 	w.Header().Set("Content-Type", "application/json")
// 	enc := json.NewEncoder(w)
// 	enc.SetEscapeHTML(false)
// 	_ = enc.Encode(resp)
// }

// stats_kpi.go for postgres
// internal/httpapi/stats_kpi.go
package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

type KPIResponse struct {
	Customers struct {
		Total     int            `json:"total"`
		Active    int            `json:"active"`
		ByGender  map[string]int `json:"by_gender"`
		BySegment map[string]int `json:"by_segment"`
	} `json:"customers"`
	CreditApplications struct {
		Total    int            `json:"total"`
		ByStatus map[string]int `json:"by_status"`
	} `json:"credit_applications"`
	VehicleOwnership struct {
		Total int `json:"total"`
	} `json:"vehicle_ownership"`
}

func (h *Handlers) GetKPI(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var resp KPIResponse
	resp.Customers.ByGender = map[string]int{}
	resp.Customers.BySegment = map[string]int{}
	resp.CreditApplications.ByStatus = map[string]int{}

	// customers total & active
	if err := h.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM customers`).Scan(&resp.Customers.Total); err != nil {
		writeError(w, http.StatusInternalServerError, "count customers failed", err)
		return
	}
	if err := h.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM customers WHERE status = 'Active'`).Scan(&resp.Customers.Active); err != nil {
		writeError(w, http.StatusInternalServerError, "count active customers failed", err)
		return
	}

	// by gender
	if rows, err := h.DB.QueryContext(ctx, `SELECT gender, COUNT(*) FROM customers GROUP BY gender`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var k sql.NullString
			var v int
			if err := rows.Scan(&k, &v); err != nil {
				writeError(w, http.StatusInternalServerError, "scan customers by_gender failed", err)
				return
			}
			key := "Unknown"
			if k.Valid && k.String != "" {
				key = k.String
			}
			resp.Customers.ByGender[key] = v
		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, "iterate customers by_gender failed", err)
			return
		}
	} else {
		writeError(w, http.StatusInternalServerError, "query customers by_gender failed", err)
		return
	}

	// by segment
	if rows, err := h.DB.QueryContext(ctx, `SELECT customer_segment, COUNT(*) FROM customers GROUP BY customer_segment`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var k sql.NullString
			var v int
			if err := rows.Scan(&k, &v); err != nil {
				writeError(w, http.StatusInternalServerError, "scan customers by_segment failed", err)
				return
			}
			key := "Unknown"
			if k.Valid && k.String != "" {
				key = k.String
			}
			resp.Customers.BySegment[key] = v
		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, "iterate customers by_segment failed", err)
			return
		}
	} else {
		writeError(w, http.StatusInternalServerError, "query customers by_segment failed", err)
		return
	}

	// credit applications total
	if err := h.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM credit_applications`).Scan(&resp.CreditApplications.Total); err != nil {
		writeError(w, http.StatusInternalServerError, "count credit_applications failed", err)
		return
	}

	// credit applications by status
	if rows, err := h.DB.QueryContext(ctx, `SELECT application_status, COUNT(*) FROM credit_applications GROUP BY application_status`); err == nil {
		defer rows.Close()
		for rows.Next() {
			var k sql.NullString
			var v int
			if err := rows.Scan(&k, &v); err != nil {
				writeError(w, http.StatusInternalServerError, "scan credit_applications by_status failed", err)
				return
			}
			key := "Unknown"
			if k.Valid && k.String != "" {
				key = k.String
			}
			resp.CreditApplications.ByStatus[key] = v
		}
		if err := rows.Err(); err != nil {
			writeError(w, http.StatusInternalServerError, "iterate credit_applications by_status failed", err)
			return
		}
	} else {
		writeError(w, http.StatusInternalServerError, "query credit_applications by_status failed", err)
		return
	}

	// vehicles total
	if err := h.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_ownership`).Scan(&resp.VehicleOwnership.Total); err != nil {
		writeError(w, http.StatusInternalServerError, "count vehicle_ownership failed", err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
