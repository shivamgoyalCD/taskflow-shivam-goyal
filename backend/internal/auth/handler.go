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
		if writeErr := response.BadRequest(w, "invalid request body"); writeErr != nil {
			h.logger.Error("http_auth_register_bad_request_response_failed", "error", writeErr)
		}
		return
	}

	validationErrors := validation.ValidateRegisterAuth(req.Name, req.Email, req.Password)
	if validationErrors.HasAny() {
		if writeErr := response.ValidationError(w, validationErrors); writeErr != nil {
			h.logger.Error("http_auth_register_validation_response_failed", "error", writeErr)
		}
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
			if writeErr := response.Conflict(w, "email already exists"); writeErr != nil {
				h.logger.Error("http_auth_register_conflict_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_auth_register_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_auth_register_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.Created(w, authResponse); err != nil {
		h.logger.Error("http_auth_register_response_failed", "error", err)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		if writeErr := response.BadRequest(w, "invalid request body"); writeErr != nil {
			h.logger.Error("http_auth_login_bad_request_response_failed", "error", writeErr)
		}
		return
	}

	validationErrors := validation.ValidateLoginAuth(req.Email, req.Password)
	if validationErrors.HasAny() {
		if writeErr := response.ValidationError(w, validationErrors); writeErr != nil {
			h.logger.Error("http_auth_login_validation_response_failed", "error", writeErr)
		}
		return
	}

	authResponse, err := h.service.Login(r.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			if writeErr := response.Error(w, http.StatusUnauthorized, "invalid credentials"); writeErr != nil {
				h.logger.Error("http_auth_login_unauthorized_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_auth_login_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_auth_login_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.OK(w, authResponse); err != nil {
		h.logger.Error("http_auth_login_response_failed", "error", err)
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
