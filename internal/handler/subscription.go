package handler

import (
	"context"
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

// SubscriptionService — контракт СЕРВИСА для хендлеров.
// Его будет реализовывать service.SubscriptionSvc.
type SubscriptionService interface {
	CreateSubscription(ctx context.Context, sub model.Subscription) (model.Subscription, error)
	GetSubscription(ctx context.Context, id int) (model.Subscription, error)
	ListSubscriptions(ctx context.Context, userID, serviceName string) ([]*model.Subscription, error)
	UpdateSubscription(ctx context.Context, sub model.Subscription) error
	DeleteSubscription(ctx context.Context, id int) error
	Sum(ctx context.Context, userID, serviceName string, startPeriod, endPeriod time.Time) (int, error)
	Ping(ctx context.Context) error
}

type Handler struct {
	services SubscriptionService
	log      *slog.Logger
}

func NewHandler(services SubscriptionService, log *slog.Logger) *Handler {
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
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		h.log.Error("invalid start_date", "value", req.StartDate, "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid start_date format")
		return
	}

	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			h.log.Error("invalid end_date", "value", req.EndDate, "err", err)
			h.writeError(w, http.StatusBadRequest, "invalid end_date format")
			return
		}

		// Проверка: end_date >= start_date
		if t.Before(startDate) {
			h.log.Error("end_date before start_date", "start_date", startDate, "end_date", t)
			h.writeError(w, http.StatusBadRequest, "end_date cannot be before start_date")
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
		h.writeError(w, http.StatusInternalServerError, "could not create subscription")
		return
	}

	h.writeJSON(w, http.StatusCreated, s)

}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error("invalid id", "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	sub, err := h.services.GetSubscription(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.writeError(w, http.StatusNotFound, "not found")
		} else {
			h.log.Error("get error", "err", err)
			h.writeError(w, http.StatusInternalServerError, "server error")
		}
		return
	}

	h.writeJSON(w, http.StatusOK, sub)

}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {

	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")

	subs, err := h.services.ListSubscriptions(r.Context(), userID, serviceName)
	if err != nil {
		h.log.Error("list error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	h.writeJSON(w, http.StatusOK, subs)

}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error("invalid id", "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid id")
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
		h.writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		h.log.Error("invalid start_date", "value", req.StartDate, "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid start_date format")
		return
	}
	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse("01-2006", req.EndDate)
		if err != nil {
			h.log.Error("invalid end_date", "value", req.EndDate, "err", err)
			h.writeError(w, http.StatusBadRequest, "invalid end_date format")
			return
		}
		if t.Before(startDate) {
			h.log.Error("end_date before start_date", "start_date", startDate, "end_date", t)
			h.writeError(w, http.StatusBadRequest, "end_date cannot be before start_date")
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
		h.writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.log.Error("invalid id", "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.services.DeleteSubscription(r.Context(), id); err != nil {
		h.log.Error("delete error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "server error")
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
		h.writeError(w, http.StatusBadRequest, "from/to required")
		return
	}
	startPeriod, err := time.Parse("01-2006", fromStr)
	if err != nil {
		h.log.Error("invalid from", "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid from format")
		return
	}
	endPeriod, err := time.Parse("01-2006", toStr)
	if err != nil {
		h.log.Error("invalid to", "err", err)
		h.writeError(w, http.StatusBadRequest, "invalid to format")
		return
	}

	sum, err := h.services.Sum(r.Context(), userID, serviceName, startPeriod, endPeriod)
	if err != nil {
		h.log.Error("sum error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	resp := struct {
		Total int `json:"total"`
	}{Total: sum}

	h.writeJSON(w, http.StatusOK, resp)

}

// проверяем имплиментацию
var _ SubscriptionService = (*service.SubscriptionSvc)(nil)
