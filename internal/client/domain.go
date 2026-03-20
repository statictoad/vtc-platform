package client

import (
	"strings"
	"time"
)

// Client is the core domain model for client-service.
// It is owned entirely by this service — no other service imports this struct.
type Client struct {
	ID          string
	ClerkUserID string
	FirstName   string
	LastName    string
	Email       string
	Phone       *string // nullable — set after signup via UpdateClientInput
	Notes       *string // operator only
	CanPayLater bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateClientInput carries the data needed to create a new client.
// Populated from the Clerk webhook payload — not from a user form.
// Phone is intentionally absent — it is set later via UpdateClientInput.
type CreateClientInput struct {
	ClerkUserID string // from Clerk webhook
	FirstName   string // from Clerk webhook
	LastName    string // from Clerk webhook
	Email       string // from Clerk webhook
}

// UpdateClientInput carries the data for a profile update.
// All fields are pointers — only provided fields are updated.
// ClerkUserID and Email are immutable — owned by Clerk, never updated here.
type UpdateClientInput struct {
	Phone       *string
	Notes       *string
	CanPayLater *bool
}

// WebhookClientCreated maps the relevant fields from a Clerk webhook payload.
// Clerk sends much more data — we only extract what we need.
// Used for "user.created" and "user.updated" event types.
type WebhookClientCreated struct {
	Data struct {
		ID             string `json:"id"` // clerk_user_id
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		EmailAddresses []struct {
			EmailAddress string `json:"email_address"`
			Primary      bool   `json:"primary"` // Clerk supports multiple emails
		} `json:"email_addresses"`
	} `json:"data"`
	Type string `json:"type"` // "user.created", "user.updated", "user.deleted"
}

// PrimaryEmail returns the primary email from the Clerk webhook payload.
// Falls back to the first email if no primary is found.
func (w WebhookClientCreated) PrimaryEmail() string {
	for _, e := range w.Data.EmailAddresses {
		if e.Primary {
			return e.EmailAddress
		}
	}
	if len(w.Data.EmailAddresses) > 0 {
		return w.Data.EmailAddresses[0].EmailAddress
	}
	return ""
}

// ToCreateInput maps a WebhookClientCreated to a CreateClientInput.
func (w WebhookClientCreated) ToCreateInput() CreateClientInput {
	return CreateClientInput{
		ClerkUserID: w.Data.ID,
		FirstName:   w.Data.FirstName,
		LastName:    w.Data.LastName,
		Email:       w.PrimaryEmail(),
	}
}

// =============================================================================
// Validation
// =============================================================================

func (i CreateClientInput) Validate() error {
	if i.ClerkUserID == "" {
		return ErrMissingClerkUserID
	}
	if i.FirstName == "" {
		return ErrMissingFirstName
	}
	if i.LastName == "" {
		return ErrMissingLastName
	}
	if i.Email == "" {
		return ErrMissingEmail
	}
	if !isValidEmail(i.Email) {
		return ErrInvalidEmail
	}
	return nil
}

func (i UpdateClientInput) Validate() error {
	if i.Phone != nil && *i.Phone == "" {
		return ErrMissingPhone
	}
	if i.Phone != nil && !isValidPhone(*i.Phone) {
		return ErrInvalidPhone
	}
	return nil
}

// isValidEmail is a basic sanity check — not RFC 5322 exhaustive.
// Clerk already validated the email at signup so this is a safety net only.
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// isValidPhone checks for a basic E.164-ish format.
// French numbers: +33612345678
func isValidPhone(phone string) bool {
	if len(phone) < 10 || len(phone) > 15 {
		return false
	}
	for i, c := range phone {
		if i == 0 && c == '+' {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// =============================================================================
// Domain errors
// =============================================================================

type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrClientNotFound       domainError = "client not found"
	ErrDuplicateClerkUserID domainError = "client already exists"
	ErrMissingClerkUserID   domainError = "clerk_user_id is required"
	ErrMissingFirstName     domainError = "first_name is required"
	ErrMissingLastName      domainError = "last_name is required"
	ErrMissingEmail         domainError = "email is required"
	ErrInvalidEmail         domainError = "email is invalid"
	ErrMissingPhone         domainError = "phone is required"
	ErrInvalidPhone         domainError = "phone must be a valid number (e.g. +33612345678)"
	ErrInvalidWebhookType   domainError = "invalid webhook event type"
)
