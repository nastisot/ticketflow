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
	"time"
)

type EventHandler struct {
	service *service.EventService
	logger  *slog.Logger
}

func NewEventHandler(service *service.EventService, logger *slog.Logger) *EventHandler {
	return &EventHandler{
		service: service,
		logger:  logger,
	}
}

func (h *EventHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.GetAll(r.Context())
	if err != nil {
		h.logger.Error(
			"failed to get events",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	started, err := writeJSON(w, http.StatusOK, events)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
}

func (h *EventHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	event, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrEventNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}
		h.logger.Error(
			"failed to get event",
			"error", err,
			"event_id", id,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	started, err := writeJSON(w, http.StatusOK, event)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"event_id", id,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
}

type CreateEventRequest struct {
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Date    time.Time `json:"date"`
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(req.Name)
	address := strings.TrimSpace(req.Address)

	if name == "" {
		http.Error(w, "empty name", http.StatusBadRequest)
		return
	}
	if address == "" {
		http.Error(w, "empty address", http.StatusBadRequest)
		return
	}
	if req.Date.IsZero() {
		http.Error(w, "empty date", http.StatusBadRequest)
		return
	}
	createdEvent, err := h.service.Create(r.Context(), name, address, req.Date)
	if err != nil {
		h.logger.Error(
			"failed to create event",
			"error", err,
			"name", req.Name,
			"address", req.Address,
			"date", req.Date,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	started, err := writeJSON(w, http.StatusCreated, createdEvent)
	if err != nil {
		h.logger.Error(
			"failed to write response",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		if !started {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}
}
