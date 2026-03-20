package httperror

import (
	"encoding/json"
	"net/http"
)

// Response is the standard error envelope returned by all services.
//
// Example:
//
//	{
//	  "code":    404,
//	  "status":  "Not Found",
//	  "message": "booking not found"
//	}
type Response struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func write(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Code:    code,
		Status:  http.StatusText(code),
		Message: message,
	})
}

// 400
func BadRequest(w http.ResponseWriter, message string) {
	write(w, http.StatusBadRequest, message)
}

// 401
func Unauthorized(w http.ResponseWriter) {
	write(w, http.StatusUnauthorized, "unauthorized")
}

// 403
func Forbidden(w http.ResponseWriter) {
	write(w, http.StatusForbidden, "forbidden")
}

// 404
func NotFound(w http.ResponseWriter, message string) {
	write(w, http.StatusNotFound, message)
}

// 409
func Conflict(w http.ResponseWriter, message string) {
	write(w, http.StatusConflict, message)
}

// 422
func UnprocessableEntity(w http.ResponseWriter, message string) {
	write(w, http.StatusUnprocessableEntity, message)
}

// 500
func Internal(w http.ResponseWriter) {
	write(w, http.StatusInternalServerError, "internal server error")
}
