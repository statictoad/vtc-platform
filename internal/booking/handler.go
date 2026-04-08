package booking

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/statictoad/vtc-platform/pkg/httperror"
)

// Handler wires HTTP routes to the booking Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a new booking Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches booking routes to a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/bookings", func(r chi.Router) {
		r.Get("/", h.listByClient)
		r.Post("/", h.create)
		r.Get("/{id}", h.get)
		r.Patch("/{id}/confirm", h.confirm)
		r.Patch("/{id}/cancel", h.cancel)
		r.Patch("/{id}/complete", h.complete)
	})
}

// =============================================================================
// Request / Response types
// =============================================================================

type createBookingRequest struct {
	ClientID          string    `json:"client_id"`
	VehicleID         string    `json:"vehicle_id"`
	ScheduledAt       time.Time `json:"scheduled_at"`
	PickupAddress     string    `json:"pickup_address"`
	DropoffAddress    string    `json:"dropoff_address"`
	TotalAmount       *int      `json:"total_amount"`
	EstimatedDistance *int      `json:"estimated_distance"`
	Co2Grams          *int      `json:"co2_grams"`
}

type bookingResponse struct {
	ID                 string        `json:"id"`
	ClientID           string        `json:"client_id"`
	VehicleID          string        `json:"vehicle_id"`
	ScheduledAt        time.Time     `json:"scheduled_at"`
	PickupAddress      string        `json:"pickup_address"`
	DropoffAddress     string        `json:"dropoff_address"`
	Status             BookingStatus `json:"status"`
	TotalAmount        *int          `json:"total_amount,omitempty"`
	TaxRateBasisPoints *int          `json:"tax_rate_basis_points,omitempty"`
	EstimatedDistance  *int          `json:"estimated_distance,omitempty"`
	Co2Grams           *int          `json:"co2_grams,omitempty"`
	ExternalInvoiceID  *string       `json:"external_invoice_id,omitempty"`
	InvoiceURL         *string       `json:"invoice_url,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`
}

// toResponse converts a domain Booking to a bookingResponse.
func toResponse(b *Booking) bookingResponse {
	return bookingResponse{
		ID:                 b.ID,
		ClientID:           b.ClientID,
		VehicleID:          b.VehicleID,
		ScheduledAt:        b.ScheduledAt,
		PickupAddress:      b.PickupAddress,
		DropoffAddress:     b.DropoffAddress,
		Status:             b.Status,
		TotalAmount:        b.TotalAmount,
		TaxRateBasisPoints: b.TaxRateBasisPoints,
		EstimatedDistance:  b.EstimatedDistance,
		Co2Grams:           b.Co2Grams,
		ExternalInvoiceID:  b.ExternalInvoiceID,
		InvoiceURL:         b.InvoiceURL,
		CreatedAt:          b.CreatedAt,
		UpdatedAt:          b.UpdatedAt,
	}
}

func toResponseList(bookings []Booking) []bookingResponse {
	result := make([]bookingResponse, len(bookings))
	for i, b := range bookings {
		result[i] = toResponse(&b)
	}
	return result
}

// =============================================================================
// Handlers
// =============================================================================

// GET /bookings?client_id=xxx
func (h *Handler) listByClient(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		httperror.BadRequest(w, "client_id query parameter is required")
		return
	}

	bookings, err := h.svc.ListByClient(r.Context(), clientID)
	if err != nil {
		slog.Error("listByClient failed", "error", err)
		httperror.Internal(w)
		return
	}

	respond(w, http.StatusOK, toResponseList(bookings))
}

// GET /bookings/{id}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	b, err := h.svc.GetBooking(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrBookingNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		httperror.Internal(w)
		return
	}

	respond(w, http.StatusOK, toResponse(b))
}

// POST /bookings
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.BadRequest(w, "invalid request body")
		return
	}

	input := CreateBookingInput{
		ClientID:          req.ClientID,
		VehicleID:         req.VehicleID,
		ScheduledAt:       req.ScheduledAt,
		PickupAddress:     req.PickupAddress,
		DropoffAddress:    req.DropoffAddress,
		TotalAmount:       req.TotalAmount,
		EstimatedDistance: req.EstimatedDistance,
		Co2Grams:          req.Co2Grams,
	}

	b, err := h.svc.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingClientID),
			errors.Is(err, ErrMissingVehicleID),
			errors.Is(err, ErrMissingScheduledAt),
			errors.Is(err, ErrMissingPickupAddress),
			errors.Is(err, ErrMissingDropoffAddress):
			httperror.BadRequest(w, err.Error())
		case errors.Is(err, ErrScheduledAtInPast):
			httperror.UnprocessableEntity(w, err.Error())
		default:
			slog.Error("create booking failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	respond(w, http.StatusCreated, toResponse(b))
}

// PATCH /bookings/{id}/confirm
func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Confirm(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, ErrBookingNotFound):
			httperror.NotFound(w, err.Error())
		case errors.Is(err, ErrInvalidTransition):
			httperror.UnprocessableEntity(w, err.Error())
		default:
			slog.Error("confirm booking failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /bookings/{id}/cancel
func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Cancel(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, ErrBookingNotFound):
			httperror.NotFound(w, err.Error())
		case errors.Is(err, ErrInvalidTransition):
			httperror.UnprocessableEntity(w, err.Error())
		default:
			slog.Error("cancel booking failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /bookings/{id}/complete
func (h *Handler) complete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Complete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, ErrBookingNotFound):
			httperror.NotFound(w, err.Error())
		case errors.Is(err, ErrInvalidTransition):
			httperror.UnprocessableEntity(w, err.Error())
		default:
			slog.Error("complete booking failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// =============================================================================
// Helpers
// =============================================================================

func respond(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
