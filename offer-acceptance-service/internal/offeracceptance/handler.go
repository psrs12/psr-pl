package offeracceptance

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// basePath matches application-management-ui's VITE_OFFER_ACCEPTANCE_API_URL
// convention (.../api/v1/offer-acceptance).
const basePath = "/api/v1/offer-acceptance"

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/applications/{id}/declarations", h.declarations)
	mux.HandleFunc("POST "+basePath+"/applications/{id}/esign", h.esign)
}

func (h *Handler) declarations(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.Declarations(r.Context()))
}

type esignRequestBody struct {
	AcceptedDeclarationIDs []string `json:"acceptedDeclarationIds"`
}

func (h *Handler) esign(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body esignRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, http.StatusBadRequest, err)
		return
	}

	record, err := h.service.ESign(r.Context(), id, body.AcceptedDeclarationIDs)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingRequiredDeclarations):
			h.writeError(w, http.StatusUnprocessableEntity, err)
		case errors.Is(err, ErrAlreadySigned):
			h.writeError(w, http.StatusConflict, err)
		default:
			h.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusCreated, record)
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
