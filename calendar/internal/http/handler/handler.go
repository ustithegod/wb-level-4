package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ustithegod/wb-level-4/calendar/internal/domain"
	"github.com/ustithegod/wb-level-4/calendar/internal/service"
)

type Handler struct {
	svc *service.Service
}

type createEventRequest struct {
	UserID   int64  `json:"user_id"`
	Date     string `json:"date"`
	Title    string `json:"event"`
	RemindAt string `json:"remind_at"`
}

type updateEventRequest struct {
	ID       int64   `json:"id"`
	UserID   int64   `json:"user_id"`
	Date     string  `json:"date"`
	Title    string  `json:"event"`
	RemindAt *string `json:"remind_at"`
}

type deleteEventRequest struct {
	ID int64 `json:"id"`
}

type response struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Post("/create_event", h.createEvent)
	r.Post("/update_event", h.updateEvent)
	r.Post("/delete_event", h.deleteEvent)
	r.Get("/events_for_day", h.eventsForDay)
	r.Get("/events_for_week", h.eventsForWeek)
	r.Get("/events_for_month", h.eventsForMonth)
	return r
}

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	event, err := h.svc.CreateEvent(r.Context(), service.CreateEventInput{
		UserID:   req.UserID,
		Date:     req.Date,
		Title:    req.Title,
		RemindAt: req.RemindAt,
	})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Result: event})
}

func (h *Handler) updateEvent(w http.ResponseWriter, r *http.Request) {
	var req updateEventRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	event, err := h.svc.UpdateEvent(r.Context(), service.UpdateEventInput{
		ID:       req.ID,
		UserID:   req.UserID,
		Date:     req.Date,
		Title:    req.Title,
		RemindAt: req.RemindAt,
	})
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Result: event})
}

func (h *Handler) deleteEvent(w http.ResponseWriter, r *http.Request) {
	var req deleteEventRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.svc.DeleteEvent(r.Context(), req.ID); err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Result: "event deleted"})
}

func (h *Handler) eventsForDay(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.EventsForDay(r.Context(), queryInt64(r, "user_id"), r.URL.Query().Get("date"))
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Result: events})
}

func (h *Handler) eventsForWeek(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.EventsForWeek(r.Context(), queryInt64(r, "user_id"), r.URL.Query().Get("date"))
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Result: events})
}

func (h *Handler) eventsForMonth(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.EventsForMonth(r.Context(), queryInt64(r, "user_id"), r.URL.Query().Get("date"))
	if err != nil {
		writeMappedError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{Result: events})
}

func decodeRequest(r *http.Request, dst any) error {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		return json.NewDecoder(r.Body).Decode(dst)
	}

	if err := r.ParseForm(); err != nil {
		return err
	}

	switch v := dst.(type) {
	case *createEventRequest:
		v.UserID = parseInt64(r.FormValue("user_id"))
		v.Date = r.FormValue("date")
		v.Title = r.FormValue("event")
		v.RemindAt = r.FormValue("remind_at")
	case *updateEventRequest:
		v.ID = parseInt64(r.FormValue("id"))
		v.UserID = parseInt64(r.FormValue("user_id"))
		v.Date = r.FormValue("date")
		v.Title = r.FormValue("event")
		if _, ok := r.Form["remind_at"]; ok {
			value := r.FormValue("remind_at")
			v.RemindAt = &value
		}
	case *deleteEventRequest:
		v.ID = parseInt64(r.FormValue("id"))
	default:
		return errors.New("unsupported request type")
	}

	return nil
}

func queryInt64(r *http.Request, key string) int64 {
	return parseInt64(r.URL.Query().Get(key))
}

func parseInt64(raw string) int64 {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

func writeMappedError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidDate),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrEmptyTitle),
		strings.Contains(err.Error(), "invalid remind_at"):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, domain.ErrEventNotFound):
		writeError(w, http.StatusServiceUnavailable, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, response{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, code int, payload response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
