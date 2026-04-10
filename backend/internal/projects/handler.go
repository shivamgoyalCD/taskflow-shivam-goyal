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

type createProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type updateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
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
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_projects_list_unauthorized_response_failed", "error", err)
		}
		return
	}

	page, limit, validationErrors := validation.ValidateProjectPagination(r.URL.Query().Get("page"), r.URL.Query().Get("limit"))
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_projects_list_validation_response_failed", "error", err)
		}
		return
	}

	result, err := h.service.List(r.Context(), currentUserID, page, limit)
	if err != nil {
		h.logger.Error("http_projects_list_failed", "error", err)
		if writeErr := response.InternalServerError(w); writeErr != nil {
			h.logger.Error("http_projects_list_error_response_failed", "error", writeErr)
		}
		return
	}

	if err := response.OK(w, result); err != nil {
		h.logger.Error("http_projects_list_response_failed", "error", err)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_projects_create_unauthorized_response_failed", "error", err)
		}
		return
	}

	var req createProjectRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		if writeErr := response.BadRequest(w, "invalid request body"); writeErr != nil {
			h.logger.Error("http_projects_create_bad_request_response_failed", "error", writeErr)
		}
		return
	}

	validationErrors := validation.ValidateCreateProject(req.Name)
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_projects_create_validation_response_failed", "error", err)
		}
		return
	}

	project, err := h.service.Create(r.Context(), CreateInput{
		CurrentUserID: currentUserID,
		Name:          req.Name,
		Description:   req.Description,
	})
	if err != nil {
		h.logger.Error("http_projects_create_failed", "error", err)
		if writeErr := response.InternalServerError(w); writeErr != nil {
			h.logger.Error("http_projects_create_error_response_failed", "error", writeErr)
		}
		return
	}

	if err := response.Created(w, project); err != nil {
		h.logger.Error("http_projects_create_response_failed", "error", err)
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_projects_get_unauthorized_response_failed", "error", err)
		}
		return
	}

	projectID := chi.URLParam(r, "id")
	project, err := h.service.GetDetail(r.Context(), currentUserID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			if writeErr := response.NotFound(w, "project not found"); writeErr != nil {
				h.logger.Error("http_projects_get_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_projects_get_forbidden_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_projects_get_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_projects_get_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.OK(w, project); err != nil {
		h.logger.Error("http_projects_get_response_failed", "error", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_projects_update_unauthorized_response_failed", "error", err)
		}
		return
	}

	projectID := chi.URLParam(r, "id")
	var req updateProjectRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		if writeErr := response.BadRequest(w, "invalid request body"); writeErr != nil {
			h.logger.Error("http_projects_update_bad_request_response_failed", "error", writeErr)
		}
		return
	}

	validationErrors := validation.ValidateUpdateProject(req.Name, req.Description)
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_projects_update_validation_response_failed", "error", err)
		}
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
			if writeErr := response.NotFound(w, "project not found"); writeErr != nil {
				h.logger.Error("http_projects_update_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_projects_update_forbidden_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_projects_update_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_projects_update_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.OK(w, project); err != nil {
		h.logger.Error("http_projects_update_response_failed", "error", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_projects_delete_unauthorized_response_failed", "error", err)
		}
		return
	}

	projectID := chi.URLParam(r, "id")
	if err := h.service.Delete(r.Context(), currentUserID, projectID); err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			if writeErr := response.NotFound(w, "project not found"); writeErr != nil {
				h.logger.Error("http_projects_delete_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_projects_delete_forbidden_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_projects_delete_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_projects_delete_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.OK(w, response.MessageBody{Message: "project deleted"}); err != nil {
		h.logger.Error("http_projects_delete_response_failed", "error", err)
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
