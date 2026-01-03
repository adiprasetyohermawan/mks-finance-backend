package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CustomerSummary is the row shape for the list endpoint.
// JSON field names must match the existing frontend expectations.
type CustomerSummary struct {
	CustomerID       string    `json:"customer_id"`
	NIK              string    `json:"nik"`
	FullName         string    `json:"full_name"`
	Gender           string    `json:"gender"`
	City             string    `json:"city"`
	Province         string    `json:"province"`
	CustomerSegment  string    `json:"customer_segment"`
	Status           string    `json:"status"`
	RegistrationDate time.Time `json:"registration_date"`
	LastUpdated      time.Time `json:"last_updated"`
}

type ListCustomersResponse struct {
	Customers []CustomerSummary `json:"customers"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
	Total     int               `json:"total,omitempty"`
}

// ListCustomers serves:
//
//	GET /api/v1/customers?limit=20&offset=0
//
// plus optional filters:
//
//	q, status, segment, province, city, gender
//
// plus sort:
//
//	sort_by=last_updated|registration_date|full_name
//	sort_dir=asc|desc
func (h *Handlers) ListCustomers(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	segment := strings.TrimSpace(r.URL.Query().Get("segment"))
	province := strings.TrimSpace(r.URL.Query().Get("province"))
	city := strings.TrimSpace(r.URL.Query().Get("city"))
	gender := strings.TrimSpace(r.URL.Query().Get("gender"))

	sortBy := strings.TrimSpace(r.URL.Query().Get("sort_by"))
	sortDir := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort_dir")))
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	// whitelist order-by columns
	orderCol := "last_updated"
	switch sortBy {
	case "", "last_updated":
		orderCol = "last_updated"
	case "registration_date":
		orderCol = "registration_date"
	case "full_name":
		orderCol = "full_name"
	default:
		orderCol = "last_updated"
	}

	where := make([]string, 0, 8)
	args := make([]any, 0, 16)

	if q != "" {
		where = append(where, "(customer_id LIKE ? OR nik LIKE ? OR full_name LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if segment != "" {
		where = append(where, "customer_segment = ?")
		args = append(args, segment)
	}
	if province != "" {
		where = append(where, "province = ?")
		args = append(args, province)
	}
	if city != "" {
		where = append(where, "city = ?")
		args = append(args, city)
	}
	if gender != "" {
		where = append(where, "gender = ?")
		args = append(args, gender)
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	total, _ := h.countCustomers(ctx, whereSQL, args)

	query := fmt.Sprintf(`
SELECT
  customer_id,
  nik,
  full_name,
  gender,
  city,
  province,
  customer_segment,
  status,
  registration_date,
  last_updated
FROM customers
%s
ORDER BY %s %s
LIMIT ? OFFSET ?`, whereSQL, orderCol, strings.ToUpper(sortDir))

	args2 := append(append([]any{}, args...), limit, offset)
	rows, err := h.DB.QueryContext(ctx, query, args2...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query customers failed", err)
		return
	}
	defer rows.Close()

	out := make([]CustomerSummary, 0, limit)
	for rows.Next() {
		var c CustomerSummary
		if err := rows.Scan(
			&c.CustomerID,
			&c.NIK,
			&c.FullName,
			&c.Gender,
			&c.City,
			&c.Province,
			&c.CustomerSegment,
			&c.Status,
			&c.RegistrationDate,
			&c.LastUpdated,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan customers failed", err)
			return
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "iterate customers failed", err)
		return
	}

	writeJSON(w, http.StatusOK, ListCustomersResponse{
		Customers: out,
		Limit:     limit,
		Offset:    offset,
		Total:     total,
	})
}

func (h *Handlers) countCustomers(ctx context.Context, whereSQL string, args []any) (int, error) {
	q := fmt.Sprintf(`SELECT COUNT(1) FROM customers %s`, whereSQL)
	var n int
	if err := h.DB.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
