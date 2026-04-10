package tasks

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

type createTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

type optionalString struct {
	Set   bool
	Value *string
}

func (o *optionalString) UnmarshalJSON(data []byte) error {
	o.Set = true

	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.Value = &value
	return nil
}

type updateTaskRequest struct {
	Title       optionalString `json:"title"`
	Description optionalString `json:"description"`
	Status      optionalString `json:"status"`
	Priority    optionalString `json:"priority"`
	AssigneeID  optionalString `json:"assignee_id"`
	DueDate     optionalString `json:"due_date"`
}

func NewHandler(logger *slog.Logger, service *Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) ListByProject(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_tasks_list_unauthorized_response_failed", "error", err)
		}
		return
	}

	projectID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	assignee := r.URL.Query().Get("assignee")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	projectID, statusPtr, assigneePtr, pageValue, limitValue, validationErrors := validation.ValidateTaskListInputs(projectID, status, assignee, page, limit)
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_tasks_list_validation_response_failed", "error", err)
		}
		return
	}

	result, err := h.service.List(r.Context(), ListInput{
		ProjectID:     projectID,
		CurrentUserID: currentUserID,
		Status:        statusPtr,
		AssigneeID:    assigneePtr,
		Page:          pageValue,
		Limit:         limitValue,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			if writeErr := response.NotFound(w, "project not found"); writeErr != nil {
				h.logger.Error("http_tasks_list_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_tasks_list_forbidden_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_tasks_list_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_tasks_list_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.OK(w, result); err != nil {
		h.logger.Error("http_tasks_list_response_failed", "error", err)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_tasks_create_unauthorized_response_failed", "error", err)
		}
		return
	}

	projectID := chi.URLParam(r, "id")
	var req createTaskRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		if writeErr := response.BadRequest(w, "invalid request body"); writeErr != nil {
			h.logger.Error("http_tasks_create_bad_request_response_failed", "error", writeErr)
		}
		return
	}

	projectID, assigneeID, dueDate, validationErrors := validation.ValidateCreateTaskInput(projectID, req.Title, req.AssigneeID, req.DueDate)
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_tasks_create_validation_response_failed", "error", err)
		}
		return
	}

	task, err := h.service.Create(r.Context(), CreateInput{
		ProjectID:     projectID,
		CurrentUserID: currentUserID,
		Title:         req.Title,
		Description:   req.Description,
		AssigneeID:    assigneeID,
		DueDate:       dueDate,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			if writeErr := response.NotFound(w, "project not found"); writeErr != nil {
				h.logger.Error("http_tasks_create_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_tasks_create_forbidden_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrAssigneeNotFound):
			if writeErr := response.ValidationError(w, validation.Errors{"assignee_id": "assignee_id must reference an existing user"}); writeErr != nil {
				h.logger.Error("http_tasks_create_assignee_validation_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_tasks_create_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_tasks_create_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.Created(w, task); err != nil {
		h.logger.Error("http_tasks_create_response_failed", "error", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_tasks_update_unauthorized_response_failed", "error", err)
		}
		return
	}

	taskID := chi.URLParam(r, "id")
	var req updateTaskRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		if writeErr := response.BadRequest(w, "invalid request body"); writeErr != nil {
			h.logger.Error("http_tasks_update_bad_request_response_failed", "error", writeErr)
		}
		return
	}

	taskID, validationErrors := validation.ValidateUpdateTaskInput(
		taskID,
		req.Title.Set, req.Title.Value,
		req.Description.Set, req.Description.Value,
		req.Status.Set, req.Status.Value,
		req.Priority.Set, req.Priority.Value,
		req.AssigneeID.Set, req.AssigneeID.Value,
		req.DueDate.Set, req.DueDate.Value,
	)
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_tasks_update_validation_response_failed", "error", err)
		}
		return
	}

	task, err := h.service.Update(r.Context(), UpdateInput{
		TaskID:        taskID,
		CurrentUserID: currentUserID,
		Title:         OptionalString{Set: req.Title.Set, Value: req.Title.Value},
		Description:   OptionalString{Set: req.Description.Set, Value: req.Description.Value},
		Status:        OptionalString{Set: req.Status.Set, Value: req.Status.Value},
		Priority:      OptionalString{Set: req.Priority.Set, Value: req.Priority.Value},
		AssigneeID:    OptionalString{Set: req.AssigneeID.Set, Value: req.AssigneeID.Value},
		DueDate:       OptionalString{Set: req.DueDate.Set, Value: req.DueDate.Value},
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			if writeErr := response.NotFound(w, "task not found"); writeErr != nil {
				h.logger.Error("http_tasks_update_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_tasks_update_forbidden_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrAssigneeNotFound):
			if writeErr := response.ValidationError(w, validation.Errors{"assignee_id": "assignee_id must reference an existing user"}); writeErr != nil {
				h.logger.Error("http_tasks_update_assignee_validation_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_tasks_update_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_tasks_update_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.OK(w, task); err != nil {
		h.logger.Error("http_tasks_update_response_failed", "error", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		if err := response.Unauthorized(w); err != nil {
			h.logger.Error("http_tasks_delete_unauthorized_response_failed", "error", err)
		}
		return
	}

	taskID, validationErrors := validation.ValidateTaskID(chi.URLParam(r, "id"))
	if validationErrors.HasAny() {
		if err := response.ValidationError(w, validationErrors); err != nil {
			h.logger.Error("http_tasks_delete_validation_response_failed", "error", err)
		}
		return
	}

	if err := h.service.Delete(r.Context(), currentUserID, taskID); err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			if writeErr := response.NotFound(w, "task not found"); writeErr != nil {
				h.logger.Error("http_tasks_delete_not_found_response_failed", "error", writeErr)
			}
		case errors.Is(err, ErrForbidden):
			if writeErr := response.Forbidden(w); writeErr != nil {
				h.logger.Error("http_tasks_delete_forbidden_response_failed", "error", writeErr)
			}
		default:
			h.logger.Error("http_tasks_delete_failed", "error", err)
			if writeErr := response.InternalServerError(w); writeErr != nil {
				h.logger.Error("http_tasks_delete_error_response_failed", "error", writeErr)
			}
		}
		return
	}

	if err := response.NoContent(w); err != nil {
		h.logger.Error("http_tasks_delete_response_failed", "error", err)
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
