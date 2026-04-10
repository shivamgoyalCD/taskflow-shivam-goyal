package response

import (
	"encoding/json"
	"net/http"

	"taskflow-shivam-goyal/backend/internal/validation"
)

type ErrorBody struct {
	Error string `json:"error"`
}

type ValidationErrorBody struct {
	Error  string            `json:"error"`
	Fields validation.Errors `json:"fields"`
}

func JSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(body)
}

func Success(w http.ResponseWriter, status int, body any) error {
	return JSON(w, status, body)
}

func OK(w http.ResponseWriter, body any) error {
	return Success(w, http.StatusOK, body)
}

func Created(w http.ResponseWriter, body any) error {
	return Success(w, http.StatusCreated, body)
}

func Error(w http.ResponseWriter, status int, message string) error {
	return JSON(w, status, ErrorBody{
		Error: message,
	})
}

func BadRequest(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusBadRequest, message)
}

func ValidationError(w http.ResponseWriter, fields validation.Errors) error {
	return JSON(w, http.StatusBadRequest, ValidationErrorBody{
		Error:  "validation failed",
		Fields: fields,
	})
}

func Unauthorized(w http.ResponseWriter) error {
	return Error(w, http.StatusUnauthorized, "unauthorized")
}

func Forbidden(w http.ResponseWriter) error {
	return Error(w, http.StatusForbidden, "forbidden")
}

func NotFound(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusNotFound, "not found")
}

func Conflict(w http.ResponseWriter, message string) error {
	return Error(w, http.StatusConflict, message)
}

func InternalServerError(w http.ResponseWriter) error {
	return Error(w, http.StatusInternalServerError, "internal server error")
}

func MethodNotAllowed(w http.ResponseWriter) error {
	return Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
