package tasks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrTaskNotFound    = errors.New("task not found")
)

type ProjectRef struct {
	ID      string
	OwnerID string
}

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ProjectID   string    `json:"project_id"`
	AssigneeID  *string   `json:"assignee_id"`
	CreatorID   string    `json:"creator_id"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskRecord struct {
	Task
	ProjectOwnerID string
}

type ListFilters struct {
	Status     *string
	AssigneeID *string
	Page       int
	Limit      int
}

type CreateTaskParams struct {
	ID          string
	Title       string
	Description *string
	ProjectID   string
	AssigneeID  *string
	CreatorID   string
	DueDate     *string
}

type UpdateTaskParams struct {
	ID          string
	Title       string
	Description *string
	Status      string
	Priority    string
	AssigneeID  *string
	DueDate     *string
}

type Repository interface {
	GetProjectByID(ctx context.Context, projectID string) (ProjectRef, error)
	UserHasProjectTaskAccess(ctx context.Context, projectID string, userID string) (bool, error)
	ListByProjectID(ctx context.Context, projectID string, filters ListFilters) ([]Task, error)
	UserExists(ctx context.Context, userID string) (bool, error)
	Create(ctx context.Context, params CreateTaskParams) (Task, error)
	GetByID(ctx context.Context, taskID string) (TaskRecord, error)
	Update(ctx context.Context, params UpdateTaskParams) (Task, error)
	Delete(ctx context.Context, taskID string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetProjectByID(ctx context.Context, projectID string) (ProjectRef, error) {
	const query = `
		SELECT id, owner_id
		FROM projects
		WHERE id = $1
	`

	var project ProjectRef
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&project.ID, &project.OwnerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProjectRef{}, ErrProjectNotFound
		}

		return ProjectRef{}, fmt.Errorf("get project by id: %w", err)
	}

	return project, nil
}

func (r *PostgresRepository) UserHasProjectTaskAccess(ctx context.Context, projectID string, userID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM tasks
			WHERE
				project_id = $1
				AND assignee_id = $2
		)
	`

	var hasAccess bool
	if err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(&hasAccess); err != nil {
		return false, fmt.Errorf("check task project access: %w", err)
	}

	return hasAccess, nil
}

func (r *PostgresRepository) ListByProjectID(ctx context.Context, projectID string, filters ListFilters) ([]Task, error) {
	const query = `
		SELECT
			id,
			title,
			description,
			status,
			priority,
			project_id,
			assignee_id,
			creator_id,
			CASE
				WHEN due_date IS NULL THEN NULL
				ELSE TO_CHAR(due_date, 'YYYY-MM-DD')
			END AS due_date,
			created_at,
			updated_at
		FROM tasks
		WHERE
			project_id = $1
			AND ($2::text IS NULL OR status = $2)
			AND ($3::uuid IS NULL OR assignee_id = $3::uuid)
		ORDER BY created_at ASC
		LIMIT $4 OFFSET $5
	`

	offset := (filters.Page - 1) * filters.Limit
	rows, err := r.pool.Query(ctx, query, projectID, filters.Status, filters.AssigneeID, filters.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list tasks by project id: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.ProjectID,
			&task.AssigneeID,
			&task.CreatorID,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate task rows: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}

	return exists, nil
}

func (r *PostgresRepository) Create(ctx context.Context, params CreateTaskParams) (Task, error) {
	const query = `
		INSERT INTO tasks (
			id,
			title,
			description,
			status,
			priority,
			project_id,
			assignee_id,
			creator_id,
			due_date
		)
		VALUES ($1, $2, $3, 'todo', 'medium', $4, $5, $6, $7)
		RETURNING
			id,
			title,
			description,
			status,
			priority,
			project_id,
			assignee_id,
			creator_id,
			CASE
				WHEN due_date IS NULL THEN NULL
				ELSE TO_CHAR(due_date, 'YYYY-MM-DD')
			END AS due_date,
			created_at,
			updated_at
	`

	var task Task
	err := r.pool.QueryRow(
		ctx,
		query,
		params.ID,
		params.Title,
		params.Description,
		params.ProjectID,
		params.AssigneeID,
		params.CreatorID,
		params.DueDate,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&task.AssigneeID,
		&task.CreatorID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, taskID string) (TaskRecord, error) {
	const query = `
		SELECT
			t.id,
			t.title,
			t.description,
			t.status,
			t.priority,
			t.project_id,
			t.assignee_id,
			t.creator_id,
			CASE
				WHEN t.due_date IS NULL THEN NULL
				ELSE TO_CHAR(t.due_date, 'YYYY-MM-DD')
			END AS due_date,
			t.created_at,
			t.updated_at,
			p.owner_id
		FROM tasks t
		INNER JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1
	`

	var task TaskRecord
	err := r.pool.QueryRow(ctx, query, taskID).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&task.AssigneeID,
		&task.CreatorID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.ProjectOwnerID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TaskRecord{}, ErrTaskNotFound
		}

		return TaskRecord{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (r *PostgresRepository) Update(ctx context.Context, params UpdateTaskParams) (Task, error) {
	const query = `
		UPDATE tasks
		SET
			title = $2,
			description = $3,
			status = $4,
			priority = $5,
			assignee_id = $6,
			due_date = $7,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			title,
			description,
			status,
			priority,
			project_id,
			assignee_id,
			creator_id,
			CASE
				WHEN due_date IS NULL THEN NULL
				ELSE TO_CHAR(due_date, 'YYYY-MM-DD')
			END AS due_date,
			created_at,
			updated_at
	`

	var task Task
	err := r.pool.QueryRow(
		ctx,
		query,
		params.ID,
		params.Title,
		params.Description,
		params.Status,
		params.Priority,
		params.AssigneeID,
		params.DueDate,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.ProjectID,
		&task.AssigneeID,
		&task.CreatorID,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, ErrTaskNotFound
		}

		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, taskID string) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}
