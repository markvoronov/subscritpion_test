package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"subscription/internal/model"
	"subscription/internal/service"
	"time"
)

type Handler struct {
	services *service.SubscriptionSvc
	log      *slog.Logger
}

func NewHandler(services *service.SubscriptionSvc, log *slog.Logger) *Handler {
	return &Handler{services: services, log: log}
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
		UserID      string `json:"user_id"`
		StartDate   string `json:"start_date"` // "MM-YYYY"
		EndDate     string `json:"end_date,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("invalid request", "err", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		h.log.Error("invalid start_date", "err", err)
		http.Error(w, "invalid start_date format", http.StatusBadRequest)
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			h.log.Error("invalid end_date", "err", err)
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		endDate = &t
	}

	s := model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	s, err = h.services.CreateSubscription(r.Context(), s)
	if err != nil {
		h.log.Error("create subscription error", "err", err)
		http.Error(w, "could not create subscription", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error("invalid id", "err", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	sub, err := h.services.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			h.log.Error("get error", "err", err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {

	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	subs, err := h.services.ListSubscriptions(r.Context(), userID, serviceName)
	if err != nil {
		h.log.Error("list error", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(subs)
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error("invalid id", "err", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req struct {
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
		UserID      string `json:"user_id"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("invalid request", "err", err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		h.log.Error("invalid start_date", "err", err)
		http.Error(w, "invalid start_date format", http.StatusBadRequest)
		return
	}
	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			h.log.Error("invalid end_date", "err", err)
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}
		endDate = &t
	}
	sub := model.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}
	if err := h.services.UpdateSubscription(r.Context(), sub); err != nil {
		h.log.Error("update error", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error("invalid id", "err", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.services.DeleteSubscription(r.Context(), id); err != nil {
		h.log.Error("delete error", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SumSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		http.Error(w, "from/to required", http.StatusBadRequest)
		return
	}
	startPeriod, err := time.Parse("01-2006", fromStr)
	if err != nil {
		h.log.Error("invalid from", "err", err)
		http.Error(w, "invalid from format", http.StatusBadRequest)
		return
	}
	endPeriod, err := time.Parse("01-2006", toStr)
	if err != nil {
		h.log.Error("invalid to", "err", err)
		http.Error(w, "invalid to format", http.StatusBadRequest)
		return
	}

	sum, err := h.services.Sum(r.Context(), userID, serviceName, startPeriod, endPeriod)
	if err != nil {
		h.log.Error("sum error", "err", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	resp := struct {
		Total int `json:"total"`
	}{Total: sum}
	json.NewEncoder(w).Encode(resp)
}
