package httpapi

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type CustomerDetail struct {
	CustomerID               string    `json:"customer_id"`
	NIK                      string    `json:"nik"`
	FullName                 string    `json:"full_name"`
	DateOfBirth              string    `json:"date_of_birth"`
	Gender                   string    `json:"gender"`
	MaritalStatus            string    `json:"marital_status"`
	PhoneNumber              string    `json:"phone_number"`
	Email                    *string   `json:"email"`
	Address                  string    `json:"address"`
	City                     string    `json:"city"`
	Province                 string    `json:"province"`
	PostalCode               string    `json:"postal_code"`
	Occupation               string    `json:"occupation"`
	EmployerName             *string   `json:"employer_name"`
	MonthlyIncome            string    `json:"monthly_income"`
	EmploymentStatus         string    `json:"employment_status"`
	YearsOfEmployment        *int      `json:"years_of_employment"`
	EducationLevel           string    `json:"education_level"`
	EmergencyContactName     string    `json:"emergency_contact_name"`
	EmergencyContactPhone    string    `json:"emergency_contact_phone"`
	EmergencyContactRelation string    `json:"emergency_contact_relation"`
	CreditScore              *int      `json:"credit_score"`
	CustomerSegment          string    `json:"customer_segment"`
	RegistrationDate         time.Time `json:"registration_date"`
	LastUpdated              time.Time `json:"last_updated"`
	Status                   string    `json:"status"`
}

type CreditApplication struct {
	ApplicationID        string     `json:"application_id"`
	CustomerID           string     `json:"customer_id"`
	ApplicationDate      time.Time  `json:"application_date"`
	VehicleType          string     `json:"vehicle_type"`
	VehicleBrand         string     `json:"vehicle_brand"`
	VehicleModel         string     `json:"vehicle_model"`
	VehicleYear          int        `json:"vehicle_year"`
	VehiclePrice         string     `json:"vehicle_price"`
	DownPayment          string     `json:"down_payment"`
	LoanAmount           string     `json:"loan_amount"`
	TenorMonths          int        `json:"tenor_months"`
	InterestRate         string     `json:"interest_rate"`
	MonthlyInstallment   string     `json:"monthly_installment"`
	ApplicationStatus    string     `json:"application_status"`
	ApprovalDate         *time.Time `json:"approval_date"`
	RejectionReason      *string    `json:"rejection_reason"`
	DisbursementDate     *string    `json:"disbursement_date"`
	FirstInstallmentDate *string    `json:"first_installment_date"`
	LastPaymentDate      *string    `json:"last_payment_date"`
	OutstandingAmount    *string    `json:"outstanding_amount"`
	PaymentStatus        *string    `json:"payment_status"`
	CollateralStatus     *string    `json:"collateral_status"`
	Notes                *string    `json:"notes"`
	ProcessedBy          *string    `json:"processed_by"`
	ApprovedBy           *string    `json:"approved_by"`
	CreatedDate          time.Time  `json:"created_date"`
}

type VehicleOwnership struct {
	OwnershipID        string    `json:"ownership_id"`
	CustomerID         string    `json:"customer_id"`
	VehicleType        string    `json:"vehicle_type"`
	Brand              string    `json:"brand"`
	Model              string    `json:"model"`
	Year               int       `json:"year"`
	VehiclePrice       string    `json:"vehicle_price"`
	PurchaseDate       string    `json:"purchase_date"`
	OwnershipStatus    string    `json:"ownership_status"`
	RegistrationNumber *string   `json:"registration_number"`
	ChassisNumber      *string   `json:"chassis_number"`
	EngineNumber       *string   `json:"engine_number"`
	CreatedDate        time.Time `json:"created_date"`
}

type ProfileSummary struct {
	TotalCreditApplications int        `json:"total_credit_applications"`
	LatestApplicationDate   *time.Time `json:"latest_application_date,omitempty"`
	LatestApplicationStatus *string    `json:"latest_application_status,omitempty"`
	TotalVehicleOwnership   int        `json:"total_vehicle_ownership"`
	SumLoanAmount           *string    `json:"sum_loan_amount,omitempty"`
	AvgInterestRate         *string    `json:"avg_interest_rate,omitempty"`
}

type CustomerProfile360Response struct {
	Customer           CustomerDetail      `json:"customer"`
	CreditApplications []CreditApplication `json:"credit_applications"`
	VehicleOwnership   []VehicleOwnership  `json:"vehicle_ownership"`
	Summary            ProfileSummary      `json:"summary"`
}

// GetCustomerProfile serves:
//
//	GET /api/v1/customers/{customer_id}/profile
func (h *Handlers) GetCustomerProfile(w http.ResponseWriter, r *http.Request) {
	customerID := readCustomerID(r)
	if customerID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "customer_id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	// 1) customer
	var c CustomerDetail
	err := h.DB.QueryRowContext(ctx, `
		SELECT customer_id, nik, full_name, date_of_birth, gender, marital_status, phone_number, email,
		       address, city, province, postal_code, occupation, employer_name, monthly_income,
		       employment_status, years_of_employment, education_level,
		       emergency_contact_name, emergency_contact_phone, emergency_contact_relation,
		       credit_score, customer_segment, registration_date, last_updated, status
		FROM customers
		WHERE customer_id = ?
	`, customerID).Scan(
		&c.CustomerID, &c.NIK, &c.FullName, &c.DateOfBirth, &c.Gender, &c.MaritalStatus,
		&c.PhoneNumber, &c.Email,
		&c.Address, &c.City, &c.Province, &c.PostalCode,
		&c.Occupation, &c.EmployerName, &c.MonthlyIncome,
		&c.EmploymentStatus, &c.YearsOfEmployment, &c.EducationLevel,
		&c.EmergencyContactName, &c.EmergencyContactPhone, &c.EmergencyContactRelation,
		&c.CreditScore, &c.CustomerSegment, &c.RegistrationDate, &c.LastUpdated, &c.Status,
	)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "customer not found"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query customer failed", err)
		return
	}

	// 2) credit_applications
	apps := make([]CreditApplication, 0)
	appRows, err := h.DB.QueryContext(ctx, `
		SELECT application_id, customer_id, application_date, vehicle_type, vehicle_brand, vehicle_model, vehicle_year,
		       vehicle_price, down_payment, loan_amount, tenor_months, interest_rate, monthly_installment,
		       application_status, approval_date, rejection_reason, disbursement_date, first_installment_date,
		       last_payment_date, outstanding_amount, payment_status, collateral_status, notes, processed_by, approved_by, created_date
		FROM credit_applications
		WHERE customer_id = ?
		ORDER BY application_date DESC
	`, customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query credit_applications failed", err)
		return
	}
	defer appRows.Close()

	for appRows.Next() {
		var a CreditApplication
		if err := appRows.Scan(
			&a.ApplicationID, &a.CustomerID, &a.ApplicationDate, &a.VehicleType, &a.VehicleBrand, &a.VehicleModel, &a.VehicleYear,
			&a.VehiclePrice, &a.DownPayment, &a.LoanAmount, &a.TenorMonths, &a.InterestRate, &a.MonthlyInstallment,
			&a.ApplicationStatus, &a.ApprovalDate, &a.RejectionReason, &a.DisbursementDate, &a.FirstInstallmentDate,
			&a.LastPaymentDate, &a.OutstandingAmount, &a.PaymentStatus, &a.CollateralStatus, &a.Notes, &a.ProcessedBy, &a.ApprovedBy, &a.CreatedDate,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan credit_applications failed", err)
			return
		}
		apps = append(apps, a)
	}
	if err := appRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "iterate credit_applications failed", err)
		return
	}

	// 3) vehicle_ownership
	vehicles := make([]VehicleOwnership, 0)
	vRows, err := h.DB.QueryContext(ctx, `
		SELECT ownership_id, customer_id, vehicle_type, brand, model, year, vehicle_price, purchase_date,
		       ownership_status, registration_number, chassis_number, engine_number, created_date
		FROM vehicle_ownership
		WHERE customer_id = ?
		ORDER BY created_date DESC
	`, customerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query vehicle_ownership failed", err)
		return
	}
	defer vRows.Close()

	for vRows.Next() {
		var v VehicleOwnership
		if err := vRows.Scan(
			&v.OwnershipID, &v.CustomerID, &v.VehicleType, &v.Brand, &v.Model, &v.Year,
			&v.VehiclePrice, &v.PurchaseDate, &v.OwnershipStatus, &v.RegistrationNumber,
			&v.ChassisNumber, &v.EngineNumber, &v.CreatedDate,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "scan vehicle_ownership failed", err)
			return
		}
		vehicles = append(vehicles, v)
	}
	if err := vRows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "iterate vehicle_ownership failed", err)
		return
	}

	// 4) summary (best-effort)
	var sum ProfileSummary
	sum.TotalCreditApplications = len(apps)
	sum.TotalVehicleOwnership = len(vehicles)

	// latest application
	var latestDate time.Time
	var latestStatus string
	if err := h.DB.QueryRowContext(ctx, `
		SELECT application_date, application_status
		FROM credit_applications
		WHERE customer_id = ?
		ORDER BY application_date DESC
		LIMIT 1
	`, customerID).Scan(&latestDate, &latestStatus); err == nil {
		sum.LatestApplicationDate = &latestDate
		sum.LatestApplicationStatus = &latestStatus
	}

	// sum loan + avg rate (kalau loan_amount/interest_rate numeric di DB, query ini aman)
	var sumLoan sql.NullString
	var avgRate sql.NullString
	_ = h.DB.QueryRowContext(ctx, `
		SELECT CAST(SUM(loan_amount) AS CHAR), CAST(AVG(interest_rate) AS CHAR)
		FROM credit_applications
		WHERE customer_id = ?
	`, customerID).Scan(&sumLoan, &avgRate)

	if sumLoan.Valid {
		sum.SumLoanAmount = &sumLoan.String
	}
	if avgRate.Valid {
		sum.AvgInterestRate = &avgRate.String
	}

	writeJSON(w, http.StatusOK, CustomerProfile360Response{
		Customer:           c,
		CreditApplications: apps,
		VehicleOwnership:   vehicles,
		Summary:            sum,
	})
}

// readCustomerID tries to read the customer id from common router param names.
// This makes the handler resilient if the route uses a different placeholder name
// (e.g. `{id}` vs `{customer_id}`) while keeping the API path unchanged.
func readCustomerID(r *http.Request) string {
	// 1) preferred param name
	if v := chi.URLParam(r, "customer_id"); v != "" {
		return v
	}

	// 2) alternate param names (common variants)
	for _, key := range []string{"id", "customerId", "customerID"} {
		if v := chi.URLParam(r, key); v != "" {
			return v
		}
	}

	// 3) optionally support query param (useful for debugging)
	if v := r.URL.Query().Get("customer_id"); v != "" {
		return v
	}

	// 4) last resort: parse from path `/.../customers/{customer_id}/profile`
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "customers" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}
