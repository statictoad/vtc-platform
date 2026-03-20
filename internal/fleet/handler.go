package fleet

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/statictoad/vtc-platform/pkg/httperror"
)

// Handler wires HTTP routes to the fleet Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a new fleet Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches fleet routes to a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/vehicles", func(r chi.Router) {
		r.Get("/", h.listActive)
		r.Post("/", h.create)
		r.Get("/{id}", h.getByID)
		r.Patch("/{id}", h.update)
		r.Post("/{id}/activate", h.activate)
		r.Post("/{id}/deactivate", h.deactivate)
	})
}

// =============================================================================
// Request / Response types
// =============================================================================

type createVehicleRequest struct {
	NumberPlate string `json:"number_plate"`
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	Year        int64  `json:"year"`
}

type updateVehicleRequest struct {
	NumberPlate *string `json:"number_plate"`
	Brand       *string `json:"brand"`
	Model       *string `json:"model"`
	Year        *int64  `json:"year"`
	Active      *bool   `json:"active"`
}

type vehicleResponse struct {
	ID          string    `json:"id"`
	NumberPlate string    `json:"number_plate"`
	Brand       string    `json:"brand"`
	Model       string    `json:"model"`
	Year        int64     `json:"year"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toResponse(v *Vehicle) vehicleResponse {
	return vehicleResponse{
		ID:          v.ID,
		NumberPlate: v.NumberPlate,
		Brand:       v.Brand,
		Model:       v.Model,
		Year:        v.Year,
		Active:      v.Active,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func toResponseList(vehicles []Vehicle) []vehicleResponse {
	result := make([]vehicleResponse, len(vehicles))
	for i, v := range vehicles {
		result[i] = toResponse(&v)
	}
	return result
}

// =============================================================================
// Handlers
// =============================================================================

// GET /vehicles
func (h *Handler) listActive(w http.ResponseWriter, r *http.Request) {
	vehicles, err := h.svc.ListActive(r.Context())
	if err != nil {
		slog.Error("listActive failed", "error", err)
		httperror.Internal(w)
		return
	}
	respond(w, http.StatusOK, toResponseList(vehicles))
}

// GET /vehicles/{id}
func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	v, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		slog.Error("getByID failed", "error", err)
		httperror.Internal(w)
		return
	}

	respond(w, http.StatusOK, toResponse(v))
}

// POST /vehicles
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.BadRequest(w, "invalid request body")
		return
	}

	input := CreateVehicleInput{
		NumberPlate: req.NumberPlate,
		Brand:       req.Brand,
		Model:       req.Model,
		Year:        req.Year,
	}

	v, err := h.svc.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingNumberPlate),
			errors.Is(err, ErrMissingBrand),
			errors.Is(err, ErrMissingModel):
			httperror.BadRequest(w, err.Error())
		case errors.Is(err, ErrInvalidYear):
			httperror.UnprocessableEntity(w, err.Error())
		case errors.Is(err, ErrDuplicateNumberPlate):
			httperror.Conflict(w, err.Error())
		default:
			slog.Error("create vehicle failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	respond(w, http.StatusCreated, toResponse(v))
}

// PATCH /vehicles/{id}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.BadRequest(w, "invalid request body")
		return
	}

	input := UpdateVehicleInput{
		NumberPlate: req.NumberPlate,
		Brand:       req.Brand,
		Model:       req.Model,
		Year:        req.Year,
		Active:      req.Active,
	}

	if err := h.svc.Update(r.Context(), id, input); err != nil {
		switch {
		case errors.Is(err, ErrVehicleNotFound):
			httperror.NotFound(w, err.Error())
		case errors.Is(err, ErrMissingNumberPlate),
			errors.Is(err, ErrMissingBrand),
			errors.Is(err, ErrMissingModel):
			httperror.BadRequest(w, err.Error())
		case errors.Is(err, ErrInvalidYear):
			httperror.UnprocessableEntity(w, err.Error())
		case errors.Is(err, ErrDuplicateNumberPlate):
			httperror.Conflict(w, err.Error())
		default:
			slog.Error("update vehicle failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /vehicles/{id}/activate
func (h *Handler) activate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Activate(r.Context(), id); err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		slog.Error("activate vehicle failed", "error", err)
		httperror.Internal(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /vehicles/{id}/deactivate
func (h *Handler) deactivate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Deactivate(r.Context(), id); err != nil {
		if errors.Is(err, ErrVehicleNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		slog.Error("deactivate vehicle failed", "error", err)
		httperror.Internal(w)
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
