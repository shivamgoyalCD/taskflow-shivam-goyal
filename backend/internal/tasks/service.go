package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	defaultStatus   = "todo"
	defaultPriority = "medium"
)

var (
	ErrForbidden        = errors.New("forbidden")
	ErrAssigneeNotFound = errors.New("assignee not found")
)

type ListResult struct {
	Tasks      []Task  `json:"tasks"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Status     *string `json:"status,omitempty"`
	AssigneeID *string `json:"assignee_id,omitempty"`
}

type CreateInput struct {
	ProjectID     string
	CurrentUserID string
	Title         string
	Description   *string
	AssigneeID    *string
	DueDate       *string
}

type ListInput struct {
	ProjectID     string
	CurrentUserID string
	Status        *string
	AssigneeID    *string
	Page          int
	Limit         int
}

type OptionalString struct {
	Set   bool
	Value *string
}

type UpdateInput struct {
	TaskID        string
	CurrentUserID string
	Title         OptionalString
	Description   OptionalString
	Status        OptionalString
	Priority      OptionalString
	AssigneeID    OptionalString
	DueDate       OptionalString
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, input ListInput) (ListResult, error) {
	project, err := s.repository.GetProjectByID(ctx, input.ProjectID)
	if err != nil {
		return ListResult{}, err
	}

	if err := s.ensureProjectAccess(ctx, project, input.CurrentUserID); err != nil {
		return ListResult{}, err
	}

	tasks, err := s.repository.ListByProjectID(ctx, input.ProjectID, ListFilters{
		Status:     input.Status,
		AssigneeID: input.AssigneeID,
		Page:       input.Page,
		Limit:      input.Limit,
	})
	if err != nil {
		return ListResult{}, err
	}
	if tasks == nil {
		tasks = make([]Task, 0)
	}

	return ListResult{
		Tasks:      tasks,
		Page:       input.Page,
		Limit:      input.Limit,
		Status:     input.Status,
		AssigneeID: input.AssigneeID,
	}, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Task, error) {
	project, err := s.repository.GetProjectByID(ctx, input.ProjectID)
	if err != nil {
		return Task{}, err
	}

	if err := s.ensureProjectAccess(ctx, project, input.CurrentUserID); err != nil {
		return Task{}, err
	}

	if input.AssigneeID != nil {
		exists, err := s.repository.UserExists(ctx, *input.AssigneeID)
		if err != nil {
			return Task{}, err
		}
		if !exists {
			return Task{}, ErrAssigneeNotFound
		}
	}

	task, err := s.repository.Create(ctx, CreateTaskParams{
		ID:          uuid.NewString(),
		Title:       strings.TrimSpace(input.Title),
		Description: normalizeNullableString(input.Description),
		ProjectID:   input.ProjectID,
		AssigneeID:  input.AssigneeID,
		CreatorID:   input.CurrentUserID,
		DueDate:     input.DueDate,
	})
	if err != nil {
		return Task{}, err
	}

	return task, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Task, error) {
	taskRecord, err := s.repository.GetByID(ctx, input.TaskID)
	if err != nil {
		return Task{}, err
	}

	if !canUpdateTask(taskRecord, input.CurrentUserID) {
		return Task{}, ErrForbidden
	}

	title := taskRecord.Title
	if input.Title.Set && input.Title.Value != nil {
		title = strings.TrimSpace(*input.Title.Value)
	}

	description := taskRecord.Description
	if input.Description.Set {
		description = normalizeNullableString(input.Description.Value)
	}

	status := taskRecord.Status
	if input.Status.Set && input.Status.Value != nil {
		status = strings.TrimSpace(*input.Status.Value)
	}

	priority := taskRecord.Priority
	if input.Priority.Set && input.Priority.Value != nil {
		priority = strings.TrimSpace(*input.Priority.Value)
	}

	assigneeID := taskRecord.AssigneeID
	if input.AssigneeID.Set {
		assigneeID = input.AssigneeID.Value
		if assigneeID != nil {
			exists, err := s.repository.UserExists(ctx, *assigneeID)
			if err != nil {
				return Task{}, err
			}
			if !exists {
				return Task{}, ErrAssigneeNotFound
			}
		}
	}

	dueDate := taskRecord.DueDate
	if input.DueDate.Set {
		dueDate = input.DueDate.Value
	}

	updatedTask, err := s.repository.Update(ctx, UpdateTaskParams{
		ID:          input.TaskID,
		Title:       title,
		Description: description,
		Status:      status,
		Priority:    priority,
		AssigneeID:  assigneeID,
		DueDate:     dueDate,
	})
	if err != nil {
		return Task{}, err
	}

	return updatedTask, nil
}

func (s *Service) Delete(ctx context.Context, currentUserID string, taskID string) error {
	taskRecord, err := s.repository.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if taskRecord.ProjectOwnerID != currentUserID && taskRecord.CreatorID != currentUserID {
		return ErrForbidden
	}

	return s.repository.Delete(ctx, taskID)
}

func (s *Service) ensureProjectAccess(ctx context.Context, project ProjectRef, currentUserID string) error {
	if project.OwnerID == currentUserID {
		return nil
	}

	hasAccess, err := s.repository.UserHasProjectTaskAccess(ctx, project.ID, currentUserID)
	if err != nil {
		return fmt.Errorf("check project access: %w", err)
	}
	if !hasAccess {
		return ErrForbidden
	}

	return nil
}

func canUpdateTask(task TaskRecord, currentUserID string) bool {
	if task.ProjectOwnerID == currentUserID {
		return true
	}
	if task.CreatorID == currentUserID {
		return true
	}
	if task.AssigneeID != nil && *task.AssigneeID == currentUserID {
		return true
	}

	return false
}

func normalizeNullableString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
