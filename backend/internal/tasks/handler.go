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

type validationErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
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

type deleteTaskResponse struct {
	Message string `json:"message"`
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
		h.writeUnauthorized(w)
		return
	}

	projectID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	assignee := r.URL.Query().Get("assignee")
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	projectID, statusPtr, assigneePtr, pageValue, limitValue, validationErrors := validation.ValidateTaskListInputs(projectID, status, assignee, page, limit)
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
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
			h.writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		default:
			h.logger.Error("http_tasks_list_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, result); err != nil {
		h.logger.Error("http_tasks_list_response_failed", "error", err)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	projectID := chi.URLParam(r, "id")
	var req createTaskRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	projectID, assigneeID, dueDate, validationErrors := validation.ValidateCreateTaskInput(projectID, req.Title, req.AssigneeID, req.DueDate)
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
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
			h.writeError(w, http.StatusNotFound, "project not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAssigneeNotFound):
			h.writeValidationError(w, validation.Errors{"assignee_id": "assignee_id must reference an existing user"})
		default:
			h.logger.Error("http_tasks_create_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusCreated, task); err != nil {
		h.logger.Error("http_tasks_create_response_failed", "error", err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	taskID := chi.URLParam(r, "id")
	var req updateTaskRequest
	if err := decodeJSONBody(r.Body, &req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
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
		h.writeValidationError(w, validationErrors)
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
			h.writeError(w, http.StatusNotFound, "task not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, ErrAssigneeNotFound):
			h.writeValidationError(w, validation.Errors{"assignee_id": "assignee_id must reference an existing user"})
		default:
			h.logger.Error("http_tasks_update_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, task); err != nil {
		h.logger.Error("http_tasks_update_response_failed", "error", err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, ok := middleware.CurrentUserIDFromContext(r.Context())
	if !ok {
		h.writeUnauthorized(w)
		return
	}

	taskID, validationErrors := validation.ValidateTaskID(chi.URLParam(r, "id"))
	if validationErrors.HasAny() {
		h.writeValidationError(w, validationErrors)
		return
	}

	if err := h.service.Delete(r.Context(), currentUserID, taskID); err != nil {
		switch {
		case errors.Is(err, ErrTaskNotFound):
			h.writeError(w, http.StatusNotFound, "task not found")
		case errors.Is(err, ErrForbidden):
			h.writeError(w, http.StatusForbidden, "forbidden")
		default:
			h.logger.Error("http_tasks_delete_failed", "error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, deleteTaskResponse{Message: "task deleted"}); err != nil {
		h.logger.Error("http_tasks_delete_response_failed", "error", err)
	}
}

func (h *Handler) writeValidationError(w http.ResponseWriter, validationErrors validation.Errors) {
	if err := response.JSON(w, http.StatusBadRequest, validationErrorResponse{
		Error:  "validation failed",
		Fields: validationErrors,
	}); err != nil {
		h.logger.Error("http_tasks_validation_response_failed", "error", err)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	if err := response.Error(w, status, message); err != nil {
		h.logger.Error("http_tasks_error_response_failed", "error", err)
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
