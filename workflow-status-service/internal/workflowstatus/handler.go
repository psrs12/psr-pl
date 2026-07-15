package workflowstatus

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const basePath = "/api/v1/workflow-status"

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/applications/{id}/status", h.status)
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	token := bearerToken(r)
	if token == "" {
		h.writeError(w, r, http.StatusUnauthorized, errors.New("missing session token"))
		return
	}

	status, err := h.service.GetStatus(r.Context(), token, id)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			h.writeError(w, r, http.StatusUnauthorized, err)
			return
		}
		h.writeError(w, r, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, status)
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

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimPrefix(auth, prefix)
}
