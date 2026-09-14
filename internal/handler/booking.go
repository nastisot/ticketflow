package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"ticketflow/internal/middleware"
	"ticketflow/internal/service"
)

type BookingHandler struct {
	service *service.BookingService
	logger  *slog.Logger
}

func NewBookingHandler(service *service.BookingService, logger *slog.Logger) *BookingHandler {
	return &BookingHandler{
		service: service,
		logger:  logger,
	}
}

type CreateBookingRequest struct {
	UserID int64 `json:"user_id"`
}

func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	seatIDStr := r.PathValue("id")
	seatID, err := strconv.ParseInt(seatIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid seat id", http.StatusBadRequest)
		return
	}
	if seatID <= 0 {
		http.Error(w, "invalid seat id", http.StatusBadRequest)
		return
	}

	var req CreateBookingRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	booking, err := h.service.Create(r.Context(), seatID, req.UserID)
	if err != nil {
		if errors.Is(err, service.ErrSeatAlreadyBooked) {
			http.Error(w, "seat already booked", http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrSeatNotFound) {
			http.Error(w, "seat not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		h.logger.Error(
			"failed to create booking",
			"error", err,
			"user_id", req.UserID,
			"seat_id", seatID,
			"request_id", middleware.GetRequestID(r.Context()),
		)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	started, err := writeJSON(w, http.StatusCreated, booking)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"user_id", req.UserID,
			"seat_id", seatID,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func (h *BookingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := r.PathValue("id")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}
	if bookingID <= 0 {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}
	booking, err := h.service.GetByID(r.Context(), bookingID)
	if err != nil {
		if errors.Is(err, service.ErrBookingNotFound) {
			http.Error(w, "booking not found", http.StatusNotFound)
			return
		}

		h.logger.Error(
			"failed to get booking",
			"error", err,
			"booking_id", bookingID,
			"request_id", middleware.GetRequestID(r.Context()),
		)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	started, err := writeJSON(w, http.StatusOK, booking)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"booking_id", bookingID,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}
}

func (h *BookingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	bookingIDStr := r.PathValue("id")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil || bookingID <= 0 {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}
	if err := h.service.Delete(r.Context(), bookingID); err != nil {
		if errors.Is(err, service.ErrBookingNotFound) {
			http.Error(w, "booking not found", http.StatusNotFound)
			return
		}

		h.logger.Error(
			"failed to delete booking",
			"error", err,
			"booking_id", bookingID,
			"request_id", middleware.GetRequestID(r.Context()),
		)

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
