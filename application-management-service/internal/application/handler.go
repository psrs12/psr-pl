package application

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// Handler is HTTP-only: request parsing, response encoding, error
// mapping. No business rule lives here — every decision is delegated to
// Service. Routes and payload shapes match application-management-ui's
// existing api/client.js exactly.
type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// basePath matches application-management-ui's VITE_APP_MANAGEMENT_API_URL
// convention (.../api/v1/application-management), so the UI's existing
// fetch calls work unmodified.
const basePath = "/api/v1/application-management"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST "+basePath+"/invitations/validate", h.validateInvitation)
	mux.HandleFunc("POST "+basePath+"/applications", h.submit)
	mux.HandleFunc("POST "+basePath+"/applications/login", h.login)
	mux.HandleFunc("GET "+basePath+"/applications/{id}/status", h.status)

	// Internal endpoints: no session-token auth, secured at the
	// network/IAM layer instead. Called by pricing-orchestration-service's
	// Lambda steps and by workflow-status-service, never by the browser.
	mux.HandleFunc("PUT /internal/applications/{id}/status", h.updateStatus)
	mux.HandleFunc("GET /internal/applications/{id}/execution", h.executionARN)
	mux.HandleFunc("POST /internal/sessions/validate", h.validateSession)
}

type validateInvitationRequestBody struct {
	Token string `json:"token"`
}

type validateInvitationResponseBody struct {
	IntakeID            string  `json:"intakeId"`
	OfferID             string  `json:"offerId"`
	CampaignOfferID     string  `json:"campaignOfferId,omitempty"`
	FirstName           string  `json:"firstName"`
	LastName            string  `json:"lastName"`
	Address             Address `json:"address"`
	RequestedAmount     float64 `json:"requestedAmount"`
	RequestedTermMonths int     `json:"requestedTermMonths"`
}

func (h *Handler) validateInvitation(w http.ResponseWriter, r *http.Request) {
	var body validateInvitationRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	offer, prefill, degraded, err := h.service.ResolveInvitation(r.Context(), body.Token)
	if err != nil {
		h.writeError(w, http.StatusBadGateway, err)
		return
	}
	if degraded {
		h.writeError(w, http.StatusNotFound, errors.New("invitation is invalid, expired, or already used"))
		return
	}

	writeJSON(w, http.StatusOK, validateInvitationResponseBody{
		IntakeID:            offer.IntakeID,
		OfferID:             offer.OfferID,
		CampaignOfferID:     offer.CampaignOfferID,
		FirstName:           prefill.FirstName,
		LastName:            prefill.LastName,
		Address:             prefill.Address,
		RequestedAmount:     centsToDollars(offer.AmountCents),
		RequestedTermMonths: offer.TermMonths,
	})
}

type submitRequestBody struct {
	IntakeID             *string `json:"intakeId"`
	SSNVerificationToken *string `json:"ssnVerificationToken"`
	SSN                  string  `json:"ssn"`
	FirstName            string  `json:"firstName"`
	LastName             string  `json:"lastName"`
	DateOfBirth          string  `json:"dateOfBirth"`
	Citizenship          string  `json:"citizenship"`
	Email                string  `json:"email"`
	Phone                string  `json:"phone"`
	Street               string  `json:"street"`
	AddressLine2         string  `json:"addressLine2"`
	City                 string  `json:"city"`
	State                string  `json:"state"`
	Zip                  string  `json:"zip"`
	EmployerName         string  `json:"employerName"`
	EmploymentStatus     string  `json:"employmentStatus"`
	AnnualIncome         float64 `json:"annualIncome"`
	MonthlyObligations   float64 `json:"monthlyObligations"`
	RequestedAmount      float64 `json:"requestedAmount"`
	TermMonths           int     `json:"termMonths"`
	LoanPurpose          string  `json:"loanPurpose"`
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var body submitRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	entryPath := EntryPathDirect
	if r.Header.Get("X-Channel-ID") == "ITA" {
		entryPath = EntryPathInvitation
	}

	applicant := Applicant{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Address: Address{
			Line1:      body.Street,
			Line2:      body.AddressLine2,
			City:       body.City,
			State:      body.State,
			PostalCode: body.Zip,
		},
		Phone: body.Phone,
		Email: body.Email,
		Employment: Employment{
			EmployerName:     body.EmployerName,
			EmploymentStatus: body.EmploymentStatus,
		},
		AnnualIncomeCents:       dollarsToCents(body.AnnualIncome),
		MonthlyObligationsCents: dollarsToCents(body.MonthlyObligations),
		SSN:                     body.SSN,
		DateOfBirth:             body.DateOfBirth,
		CitizenshipStatus:       body.Citizenship,
		RequestedAmountCents:    dollarsToCents(body.RequestedAmount),
		RequestedTermMonths:     body.TermMonths,
		LoanPurpose:             body.LoanPurpose,
	}
	if body.SSNVerificationToken != nil {
		applicant.SSNVerificationToken = *body.SSNVerificationToken
	}

	req := SubmitRequest{EntryPath: entryPath, Applicant: applicant}
	if body.IntakeID != nil {
		req.IntakeID = *body.IntakeID
	}

	app, err := h.service.Submit(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrIncompleteApplication) {
			h.writeError(w, http.StatusUnprocessableEntity, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, app)
}

type loginRequestBody struct {
	ApplicationID string `json:"applicationId"`
	Last4SSN      string `json:"last4SSN"`
	DateOfBirth   string `json:"dateOfBirth"`
}

type loginResponseBody struct {
	SessionToken  string `json:"sessionToken"`
	ApplicationID string `json:"applicationId"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var body loginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	token, err := h.service.Login(r.Context(), body.ApplicationID, body.Last4SSN, body.DateOfBirth)
	if err != nil {
		if errors.Is(err, ErrLoginFailed) {
			h.writeError(w, http.StatusUnauthorized, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponseBody{SessionToken: token, ApplicationID: body.ApplicationID})
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	token := bearerToken(r)
	if token == "" {
		h.writeError(w, http.StatusUnauthorized, errors.New("missing session token"))
		return
	}
	if err := h.service.Authenticate(r.Context(), token, id); err != nil {
		h.writeError(w, http.StatusUnauthorized, err)
		return
	}

	app, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

type updateStatusRequestBody struct {
	Status Status `json:"status"`
}

// updateStatus is called by the workflow (Step Functions / expiry sweep),
// not by the applicant-facing UI — no session-token auth here; access is
// controlled at the network/IAM layer.
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body updateStatusRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	app, err := h.service.UpdateStatus(r.Context(), id, body.Status)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, app)
}

func (h *Handler) executionARN(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	arn, err := h.service.ExecutionARN(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"executionArn": arn})
}

type validateSessionRequestBody struct {
	Token         string `json:"token"`
	ApplicationID string `json:"applicationId"`
}

func (h *Handler) validateSession(w http.ResponseWriter, r *http.Request) {
	var body validateSessionRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	valid := h.service.Authenticate(r.Context(), body.Token, body.ApplicationID) == nil
	writeJSON(w, http.StatusOK, map[string]bool{"valid": valid})
}

func (h *Handler) writeError(w http.ResponseWriter, status int, err error) {
	h.logger.Error("request failed", "status", status, "error", err)
	writeJSON(w, status, map[string]string{"message": strings.TrimSpace(err.Error())})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimPrefix(auth, prefix)
}

func dollarsToCents(dollars float64) int64 {
	return int64(dollars*100 + 0.5)
}

func centsToDollars(cents int64) float64 {
	return float64(cents) / 100
}
