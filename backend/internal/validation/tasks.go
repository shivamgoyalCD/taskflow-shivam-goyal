package validation

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	validTaskStatuses   = map[string]struct{}{"todo": {}, "in_progress": {}, "done": {}}
	validTaskPriorities = map[string]struct{}{"low": {}, "medium": {}, "high": {}}
)

func ValidateTaskListInputs(projectID string, status string, assignee string, page string, limit string) (string, *string, *string, int, int, Errors) {
	errors := Errors{}

	projectID = strings.TrimSpace(projectID)
	if _, err := uuid.Parse(projectID); err != nil {
		errors.Add("project_id", "project_id must be a valid UUID")
	}

	var statusPtr *string
	status = strings.TrimSpace(status)
	if status != "" {
		if !isValidTaskStatus(status) {
			errors.Add("status", "status must be one of todo, in_progress, done")
		} else {
			statusPtr = &status
		}
	}

	var assigneePtr *string
	assignee = strings.TrimSpace(assignee)
	if assignee != "" {
		if _, err := uuid.Parse(assignee); err != nil {
			errors.Add("assignee", "assignee must be a valid UUID")
		} else {
			assigneePtr = &assignee
		}
	}

	pageValue, limitValue, paginationErrors := ValidateProjectPagination(page, limit)
	for field, message := range paginationErrors {
		errors.Add(field, message)
	}

	return projectID, statusPtr, assigneePtr, pageValue, limitValue, errors
}

func ValidateCreateTaskInput(projectID string, title string, assigneeID *string, dueDate *string) (string, *string, *string, Errors) {
	errors := Errors{}

	projectID = strings.TrimSpace(projectID)
	if _, err := uuid.Parse(projectID); err != nil {
		errors.Add("project_id", "project_id must be a valid UUID")
	}

	if strings.TrimSpace(title) == "" {
		errors.Add("title", "title is required")
	}

	assigneeID = validateOptionalUUID(errors, "assignee_id", assigneeID)
	dueDate = validateOptionalDate(errors, "due_date", dueDate)

	return projectID, assigneeID, dueDate, errors
}

func ValidateUpdateTaskInput(
	taskID string,
	titleSet bool, title *string,
	descriptionSet bool, description *string,
	statusSet bool, status *string,
	prioritySet bool, priority *string,
	assigneeSet bool, assignee *string,
	dueDateSet bool, dueDate *string,
) (string, Errors) {
	errors := Errors{}

	taskID = strings.TrimSpace(taskID)
	if _, err := uuid.Parse(taskID); err != nil {
		errors.Add("task_id", "task_id must be a valid UUID")
	}

	if !titleSet && !descriptionSet && !statusSet && !prioritySet && !assigneeSet && !dueDateSet {
		errors.Add("body", "at least one field must be provided")
		return taskID, errors
	}

	if titleSet {
		if title == nil || strings.TrimSpace(*title) == "" {
			errors.Add("title", "title cannot be empty")
		}
	}

	if statusSet {
		if status == nil || !isValidTaskStatus(strings.TrimSpace(*status)) {
			errors.Add("status", "status must be one of todo, in_progress, done")
		}
	}

	if prioritySet {
		if priority == nil || !isValidTaskPriority(strings.TrimSpace(*priority)) {
			errors.Add("priority", "priority must be one of low, medium, high")
		}
	}

	if assigneeSet {
		validateOptionalUUID(errors, "assignee_id", assignee)
	}

	if dueDateSet {
		validateOptionalDate(errors, "due_date", dueDate)
	}

	return taskID, errors
}

func ValidateTaskID(taskID string) (string, Errors) {
	errors := Errors{}

	taskID = strings.TrimSpace(taskID)
	if _, err := uuid.Parse(taskID); err != nil {
		errors.Add("task_id", "task_id must be a valid UUID")
	}

	return taskID, errors
}

func validateOptionalUUID(errors Errors, field string, value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		errors.Add(field, field+" must be a valid UUID")
		return value
	}

	if _, err := uuid.Parse(trimmed); err != nil {
		errors.Add(field, field+" must be a valid UUID")
		return value
	}

	return &trimmed
}

func validateOptionalDate(errors Errors, field string, value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		errors.Add(field, field+" must be in YYYY-MM-DD format")
		return value
	}

	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		errors.Add(field, field+" must be in YYYY-MM-DD format")
		return value
	}

	return &trimmed
}

func isValidTaskStatus(status string) bool {
	_, ok := validTaskStatuses[status]
	return ok
}

func isValidTaskPriority(priority string) bool {
	_, ok := validTaskPriorities[priority]
	return ok
}
