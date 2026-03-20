package fleet

import "time"

// Vehicle is the core domain model for fleet-service.
// It is owned entirely by this service — no other service imports this struct.
type Vehicle struct {
	ID          string
	NumberPlate string
	Brand       string
	Model       string
	Year        int64
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateVehicleInput carries the data needed to create a new vehicle.
// Received from the HTTP handler after decoding and validating the request body.
type CreateVehicleInput struct {
	NumberPlate string
	Brand       string
	Model       string
	Year        int64
}

// Validate checks that the input is complete and coherent.
func (i CreateVehicleInput) Validate() error {
	if i.NumberPlate == "" {
		return ErrMissingNumberPlate
	}
	if i.Brand == "" {
		return ErrMissingBrand
	}
	if i.Model == "" {
		return ErrMissingModel
	}
	currentYear := int64(time.Now().Year())
	if i.Year < currentYear-7 || i.Year > currentYear+1 {
		return ErrInvalidYear
	}
	return nil
}

// UpdateVehicleInput carries the data needed to update an existing vehicle.
// All fields are pointers — only provided fields are updated.
type UpdateVehicleInput struct {
	NumberPlate *string
	Brand       *string
	Model       *string
	Year        *int64
	Active      *bool
}

// Validate checks that the input is coherent (but not necessarily complete).
func (i UpdateVehicleInput) Validate() error {
	if i.NumberPlate != nil && *i.NumberPlate == "" {
		return ErrMissingNumberPlate
	}
	if i.Brand != nil && *i.Brand == "" {
		return ErrMissingBrand
	}
	if i.Model != nil && *i.Model == "" {
		return ErrMissingModel
	}
	if i.Year != nil {
		currentYear := int64(time.Now().Year())
		if *i.Year < currentYear-7 || *i.Year > currentYear+1 {
			return ErrInvalidYear
		}
	}
	return nil
}

// =============================================================================
// Domain errors
// =============================================================================

type domainError string

func (e domainError) Error() string { return string(e) }

const (
	ErrVehicleNotFound      domainError = "vehicle not found"
	ErrMissingNumberPlate   domainError = "number_plate is required"
	ErrMissingBrand         domainError = "brand is required"
	ErrMissingModel         domainError = "model is required"
	ErrInvalidYear          domainError = "year must be between 1990 and next year"
	ErrDuplicateNumberPlate domainError = "a vehicle with this number plate already exists"
)
