package document

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

// basePath matches application-management-ui's VITE_DOCUMENT_API_URL
// convention (.../api/v1/document).
const basePath = "/api/v1/document"

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/applications/{id}/required-documents", h.requiredDocuments)
	mux.HandleFunc("GET "+basePath+"/applications/{id}/documents", h.get)
	mux.HandleFunc("POST "+basePath+"/applications/{id}/documents", h.submit)
}

func (h *Handler) requiredDocuments(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.service.RequiredDocuments(r.Context()))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	record, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

type submitRequestBody struct {
	DocumentID string `json:"documentId"`
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var body submitRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeError(w, r, http.StatusBadRequest, err)
		return
	}

	record, err := h.service.Submit(r.Context(), id, body.DocumentID)
	if err != nil {
		if errors.Is(err, ErrUnknownDocument) {
			h.writeError(w, r, http.StatusBadRequest, err)
			return
		}
		h.writeError(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, record)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, status int, err error) {
	LoggerFrom(r.Context(), h.logger).Error("request failed", "status", status, "error", err)
	writeJSON(w, status, map[string]string{"message": strings.TrimSpace(err.Error())})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
