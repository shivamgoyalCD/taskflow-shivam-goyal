package projects

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("forbidden")

type ListResult struct {
	Projects []Project `json:"projects"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type Detail struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	Tasks       []Task    `json:"tasks"`
}

type CreateInput struct {
	CurrentUserID string
	Name          string
	Description   *string
}

type UpdateInput struct {
	ProjectID     string
	CurrentUserID string
	Name          *string
	Description   *string
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, currentUserID string, page int, limit int) (ListResult, error) {
	projects, err := s.repository.ListAccessible(ctx, currentUserID, page, limit)
	if err != nil {
		return ListResult{}, err
	}
	if projects == nil {
		projects = make([]Project, 0)
	}

	return ListResult{
		Projects: projects,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Project, error) {
	project, err := s.repository.Create(ctx, CreateProjectParams{
		ID:          uuid.NewString(),
		Name:        strings.TrimSpace(input.Name),
		Description: normalizeDescription(input.Description),
		OwnerID:     input.CurrentUserID,
	})
	if err != nil {
		return Project{}, err
	}

	return project, nil
}

func (s *Service) GetDetail(ctx context.Context, currentUserID string, projectID string) (Detail, error) {
	project, err := s.repository.GetByID(ctx, projectID)
	if err != nil {
		return Detail{}, err
	}

	if project.OwnerID != currentUserID {
		hasAccess, err := s.repository.UserHasTaskAccess(ctx, projectID, currentUserID)
		if err != nil {
			return Detail{}, fmt.Errorf("check project access: %w", err)
		}
		if !hasAccess {
			return Detail{}, ErrForbidden
		}
	}

	tasks, err := s.repository.ListTasksByProjectID(ctx, projectID)
	if err != nil {
		return Detail{}, err
	}
	if tasks == nil {
		tasks = make([]Task, 0)
	}

	return Detail{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
		Tasks:       tasks,
	}, nil
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (Project, error) {
	project, err := s.repository.GetByID(ctx, input.ProjectID)
	if err != nil {
		return Project{}, err
	}

	if project.OwnerID != input.CurrentUserID {
		return Project{}, ErrForbidden
	}

	name := project.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}

	description := project.Description
	if input.Description != nil {
		description = normalizeDescription(input.Description)
	}

	updatedProject, err := s.repository.Update(ctx, UpdateProjectParams{
		ID:          input.ProjectID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return Project{}, err
	}

	return updatedProject, nil
}

func (s *Service) Delete(ctx context.Context, currentUserID string, projectID string) error {
	project, err := s.repository.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if project.OwnerID != currentUserID {
		return ErrForbidden
	}

	return s.repository.Delete(ctx, projectID)
}

func normalizeDescription(description *string) *string {
	if description == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*description)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
