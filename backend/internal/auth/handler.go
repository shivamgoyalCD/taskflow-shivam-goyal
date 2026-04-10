package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"taskflow-shivam-goyal/backend/internal/response"
	"taskflow-shivam-goyal/backend/internal/validation"
)

type Handler struct {
	logger  *slog.Logger
	service *Service
}

type validationErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewHandler(logger *slog.Logger, service *Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := validation.ValidateRegisterAuth(req.Name, req.Email, req.Password)
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
		return
	}

	authResponse, err := h.service.Register(r.Context(), RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			h.writeError(w, http.StatusConflict, "email already exists")
		default:
			h.logger.Error("http_auth_register_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusCreated, authResponse); err != nil {
		h.logger.Error("http_auth_register_response_failed", "error", err)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := validation.ValidateLoginAuth(req.Email, req.Password)
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
		return
	}

	authResponse, err := h.service.Login(r.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			h.writeError(w, http.StatusUnauthorized, "invalid credentials")
		default:
			h.logger.Error("http_auth_login_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, authResponse); err != nil {
		h.logger.Error("http_auth_login_response_failed", "error", err)
	}
}

func (h *Handler) writeValidationError(w http.ResponseWriter, validationErrors validation.Errors) {
	if err := response.JSON(w, http.StatusBadRequest, validationErrorResponse{
		Error:  "validation failed",
		Fields: validationErrors,
	}); err != nil {
		h.logger.Error("http_validation_response_failed", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	if err := response.Error(w, status, message); err != nil {
		h.logger.Error("http_auth_error_response_failed", "error", err)
	}
}

func decodeJSONBody(body io.ReadCloser, destination any) error {
	defer body.Close()

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
