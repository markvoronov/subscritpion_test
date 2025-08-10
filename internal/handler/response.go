package handler

import (
	"encoding/json"
	"net/http"
)

const (
	StatusError = "Error"
)

type apiError struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		h.log.Error("marshal response error", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		h.log.Error("write response error", "err", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, apiError{Status: StatusError, Error: msg})
}
