package offerselection

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// basePath matches application-management-ui's VITE_PRICING_API_URL
// convention (.../api/v1/pricing-orchestration).
const basePath = "/api/v1/pricing-orchestration"

// Handler is HTTP-only: request parsing, response encoding, error
// mapping. No business rule lives here — every decision is delegated to
// Service. This is the one REST surface pricing-orchestration-service
// exposes; everything else in this service is Step Functions-invoked
// Lambda/Fargate steps with no API of their own.
type Handler struct {
	service  *Service
	sessions SessionValidator
	logger   *slog.Logger
}

func NewHandler(service *Service, sessions SessionValidator, logger *slog.Logger) *Handler {
	return &Handler{service: service, sessions: sessions, logger: logger}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/applications/{id}/selected-offer", h.get)
	mux.HandleFunc("POST "+basePath+"/applications/{id}/selected-offer/confirm", h.confirm)

	// Internal: called only by present-offer-lambda, not the browser —
	// secured at the network/IAM layer, same pattern as
	// application-management-service's internal endpoints.
	mux.HandleFunc("POST /internal/applications/{id}/selected-offer", h.present)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if !h.authenticated(w, r, id) {
		return
	}

	offer, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.writeError(w, http.StatusNotFound, err)
			return
		}
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, offer)
}

type presentRequestBody struct {
	ApplicationID string  `json:"applicationId"`
	TaskToken     string  `json:"taskToken"`
	AmountCents   int64   `json:"amountCents"`
	TermMonths    int     `json:"termMonths"`
	APRPercentage float64 `json:"aprPercentage"`
}

func (h *Handler) present(w http.ResponseWriter, r *http.Request) {
	var body presentRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	offer := SelectedOffer{
		ApplicationID: body.ApplicationID,
		AmountCents:   body.AmountCents,
		TermMonths:    body.TermMonths,
		APRPercentage: body.APRPercentage,
		TaskToken:     body.TaskToken,
	}
	if err := h.service.Present(r.Context(), offer); err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"applicationId": body.ApplicationID})
}

type confirmRequestBody struct {
	ConsentGiven bool `json:"consentGiven"`
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if !h.authenticated(w, r, id) {
		return
	}

	var body confirmRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	offer, err := h.service.Confirm(r.Context(), id, body.ConsentGiven)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			h.writeError(w, http.StatusNotFound, err)
		case errors.Is(err, ErrAlreadyConfirmed):
			h.writeError(w, http.StatusConflict, err)
		default:
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, offer)
}

// authenticated validates the request's session token for the given
// application id, writing an error response and returning false if it's
// missing or invalid. Callers must return immediately when this is false.
func (h *Handler) authenticated(w http.ResponseWriter, r *http.Request, applicationID string) bool {
	token := bearerToken(r)
	if token == "" {
		h.writeError(w, http.StatusUnauthorized, errors.New("missing session token"))
		return false
	}
	valid, err := h.sessions.ValidateSession(r.Context(), token, applicationID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
		return false
	}
	if !valid {
		h.writeError(w, http.StatusUnauthorized, errors.New("invalid session"))
		return false
	}
	return true
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimPrefix(auth, prefix)
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
