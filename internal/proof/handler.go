package proof

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/statictoad/vtc-platform/pkg/httperror"
)

// Handler wires HTTP routes to the proof Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a new proof Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches proof routes to a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/proofs", func(r chi.Router) {
		r.Get("/booking/{bookingID}", h.getByBookingID)
	})
}

// =============================================================================
// Response types
// =============================================================================

type proofResponse struct {
	ID                 string    `json:"id"`
	BookedAt           time.Time `json:"booked_at"`
	OperatorName       string    `json:"operator_name"`
	OperatorSiret      string    `json:"operator_siret"`
	OperatorEvtcNumber string    `json:"operator_evtc_number"`
	VehicleNumberPlate string    `json:"vehicle_number_plate"`
	ClientFirstName    string    `json:"client_first_name"`
	ClientLastName     string    `json:"client_last_name"`
	ClientPhone        *string   `json:"client_phone,omitempty"`
	PickupAddress      string    `json:"pickup_address"`
	DropoffAddress     string    `json:"dropoff_address"`
	ScheduledAt        time.Time `json:"scheduled_at"`
	ProofHash          *string   `json:"proof_hash,omitempty"`
	BookingID          string    `json:"booking_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func toResponse(p *Proof) proofResponse {
	return proofResponse{
		ID:                 p.ID,
		BookedAt:           p.BookedAt,
		OperatorName:       p.OperatorName,
		OperatorSiret:      p.OperatorSiret,
		OperatorEvtcNumber: p.OperatorEvtcNumber,
		VehicleNumberPlate: p.VehicleNumberPlate,
		ClientFirstName:    p.ClientFirstName,
		ClientLastName:     p.ClientLastName,
		ClientPhone:        p.ClientPhone,
		PickupAddress:      p.PickupAddress,
		DropoffAddress:     p.DropoffAddress,
		ScheduledAt:        p.ScheduledAt,
		ProofHash:          p.ProofHash,
		BookingID:          p.BookingID,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

// =============================================================================
// Handlers
// =============================================================================

// GET /proofs/booking/{bookingID}
func (h *Handler) getByBookingID(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "bookingID")

	p, err := h.svc.GetByBookingID(r.Context(), bookingID)
	if err != nil {
		if errors.Is(err, ErrProofNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		slog.Error("get proof failed", "error", err)
		httperror.Internal(w)
		return
	}

	respond(w, http.StatusOK, toResponse(p))
}

// =============================================================================
// Helpers
// =============================================================================

func respond(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
