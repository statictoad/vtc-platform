package client

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/statictoad/vtc-platform/pkg/httperror"
)

// Handler wires HTTP routes to the client Service.
type Handler struct {
	svc *Service
}

// NewHandler returns a new client Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches client routes to a chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Clerk webhook — no JWT auth, secured by webhook signature instead
	r.Post("/webhooks/clerk", h.handleClerkWebhook)

	r.Route("/clients", func(r chi.Router) {
		r.Get("/{id}", h.getByID)
		r.Get("/clerk/{clerkUserID}", h.getByClerkUserID)
		r.Patch("/{id}", h.update)
	})
}

// =============================================================================
// Request / Response types
// =============================================================================

type updateClientRequest struct {
	Phone       *string `json:"phone"`
	Notes       *string `json:"notes"`
	CanPayLater *bool   `json:"can_pay_later"`
}

type clientResponse struct {
	ID          string    `json:"id"`
	ClerkUserID string    `json:"clerk_user_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Phone       *string   `json:"phone,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	CanPayLater bool      `json:"can_pay_later"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toResponse(c *Client) clientResponse {
	return clientResponse{
		ID:          c.ID,
		ClerkUserID: c.ClerkUserID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Email:       c.Email,
		Phone:       c.Phone,
		Notes:       c.Notes,
		CanPayLater: c.CanPayLater,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// =============================================================================
// Handlers
// =============================================================================

// POST /webhooks/clerk
// Called by Clerk when a user is created, updated, or deleted.
// Secured by Clerk webhook signature — not by JWT.
// TODO: verify Clerk webhook signature before processing.
func (h *Handler) handleClerkWebhook(w http.ResponseWriter, r *http.Request) {
	var payload WebhookClientCreated
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		httperror.BadRequest(w, "invalid webhook payload")
		return
	}

	if err := h.svc.HandleWebhook(r.Context(), payload); err != nil {
		switch {
		case errors.Is(err, ErrDuplicateClerkUserID):
			// Idempotent — Clerk may fire the webhook more than once.
			// Return 200 so Clerk does not retry.
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, ErrMissingClerkUserID),
			errors.Is(err, ErrMissingFirstName),
			errors.Is(err, ErrMissingLastName),
			errors.Is(err, ErrMissingEmail),
			errors.Is(err, ErrInvalidEmail):
			httperror.BadRequest(w, err.Error())
		default:
			slog.Error("handleClerkWebhook failed", "error", err)
			httperror.Internal(w)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /clients/{id}
func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		slog.Error("getByID failed", "error", err)
		httperror.Internal(w)
		return
	}

	respond(w, http.StatusOK, toResponse(c))
}

// GET /clients/clerk/{clerkUserID}
// Used by the gateway to resolve the current user from a Clerk JWT.
func (h *Handler) getByClerkUserID(w http.ResponseWriter, r *http.Request) {
	clerkUserID := chi.URLParam(r, "clerkUserID")

	c, err := h.svc.GetByClerkUserID(r.Context(), clerkUserID)
	if err != nil {
		if errors.Is(err, ErrClientNotFound) {
			httperror.NotFound(w, err.Error())
			return
		}
		slog.Error("getByClerkUserID failed", "error", err)
		httperror.Internal(w)
		return
	}

	respond(w, http.StatusOK, toResponse(c))
}

// PATCH /clients/{id}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperror.BadRequest(w, "invalid request body")
		return
	}

	input := UpdateClientInput{
		Phone:       req.Phone,
		Notes:       req.Notes,
		CanPayLater: req.CanPayLater,
	}

	if err := h.svc.Update(r.Context(), id, input); err != nil {
		switch {
		case errors.Is(err, ErrClientNotFound):
			httperror.NotFound(w, err.Error())
		case errors.Is(err, ErrInvalidPhone):
			httperror.UnprocessableEntity(w, err.Error())
		case errors.Is(err, ErrMissingPhone):
			httperror.BadRequest(w, err.Error())
		default:
			slog.Error("update failed", "error", err)
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
