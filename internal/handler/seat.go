package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"ticketflow/internal/middleware"
	"ticketflow/internal/service"
)

type SeatHandler struct {
	service *service.SeatService
	logger  *slog.Logger
}

func NewSeatHandler(service *service.SeatService, logger *slog.Logger) *SeatHandler {
	return &SeatHandler{
		service: service,
		logger:  logger,
	}
}

type CreateSeatRequest struct {
	Number     string `json:"number"`
	PriceCents int64  `json:"price_cents"`
}

func (h *SeatHandler) Create(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id event", http.StatusBadRequest)
		return
	}
	if eventID <= 0 {
		http.Error(w, "invalid id event", http.StatusBadRequest)
		return
	}
	var req CreateSeatRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	number := strings.TrimSpace(req.Number)

	if number == "" {
		http.Error(w, "invalid number", http.StatusBadRequest)
		return
	}
	if req.PriceCents < 0 {
		http.Error(w, "invalid price", http.StatusBadRequest)
		return
	}
	createdSeat, err := h.service.Create(r.Context(), eventID, number, req.PriceCents)
	if errors.Is(err, service.ErrSeatAlreadyExists) {
		http.Error(w, "seat already exists", http.StatusConflict)
		return
	}
	if errors.Is(err, service.ErrEventNotFound) {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error(
			"failed to create seat",
			"error", err,
			"eventID", eventID,
			"number", number,
			"price_cents", req.PriceCents,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	started, err := writeJSON(w, http.StatusCreated, createdSeat)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"eventID", eventID,
			"number", number,
			"price_cents", req.PriceCents,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func (h *SeatHandler) GetByEventID(w http.ResponseWriter, r *http.Request) {
	eventIDStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id event", http.StatusBadRequest)
		return
	}
	if eventID <= 0 {
		http.Error(w, "invalid id event", http.StatusBadRequest)
		return
	}
	seats, err := h.service.GetByEventID(r.Context(), eventID)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		h.logger.Error(
			"failed to get seats",
			"error", err,
			"event_id", eventID,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	started, err := writeJSON(w, http.StatusOK, seats)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"event_id", eventID,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}
