package projects

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taskflow-shivam-goyal/backend/internal/middleware"
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

type createProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type deleteProjectResponse struct {
	Message string `json:"message"`
}

func NewHandler(logger *slog.Logger, service *Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	page, limit, validationErrors := validation.ValidateProjectPagination(r.URL.Query().Get("page"), r.URL.Query().Get("limit"))
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
		return
	}

	result, err := h.service.List(r.Context(), currentUserID, page, limit)
	if err != nil {
		h.logger.Error("http_projects_list_failed", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := response.JSON(w, http.StatusOK, result); err != nil {
		h.logger.Error("http_projects_list_response_failed", "error", err)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	var req createProjectRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := validation.ValidateCreateProject(req.Name)
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
		return
	}

	project, err := h.service.Create(r.Context(), CreateInput{
		CurrentUserID: currentUserID,
		Name:          req.Name,
		Description:   req.Description,
	})
	if err != nil {
		h.logger.Error("http_projects_create_failed", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := response.JSON(w, http.StatusCreated, project); err != nil {
		h.logger.Error("http_projects_create_response_failed", "error", err)
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	projectID := chi.URLParam(r, "id")
	project, err := h.service.GetDetail(r.Context(), currentUserID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			h.writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		default:
			h.logger.Error("http_projects_get_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, project); err != nil {
		h.logger.Error("http_projects_get_response_failed", "error", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	projectID := chi.URLParam(r, "id")
	var req updateProjectRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	validationErrors := validation.ValidateUpdateProject(req.Name, req.Description)
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
		return
	}

	project, err := h.service.Update(r.Context(), UpdateInput{
		ProjectID:     projectID,
		CurrentUserID: currentUserID,
		Name:          req.Name,
		Description:   req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			h.writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		default:
			h.logger.Error("http_projects_update_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, project); err != nil {
		h.logger.Error("http_projects_update_response_failed", "error", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	projectID := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), currentUserID, projectID); err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			h.writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		default:
			h.logger.Error("http_projects_delete_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, deleteProjectResponse{Message: "project deleted"}); err != nil {
		h.logger.Error("http_projects_delete_response_failed", "error", err)
	}
}

func (h *Handler) writeValidationError(w http.ResponseWriter, validationErrors validation.Errors) {
	if err := response.JSON(w, http.StatusBadRequest, validationErrorResponse{
		Error:  "validation failed",
		Fields: validationErrors,
	}); err != nil {
		h.logger.Error("http_projects_validation_response_failed", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	if err := response.Error(w, status, message); err != nil {
		h.logger.Error("http_projects_error_response_failed", "error", err)
	}
}

func (h *Handler) writeUnauthorized(w http.ResponseWriter) {
	h.writeError(w, http.StatusUnauthorized, "unauthorized")
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
