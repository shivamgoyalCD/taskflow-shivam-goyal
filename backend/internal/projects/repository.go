package projects

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
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

type CreateProjectParams struct {
	ID          string
	Name        string
	Description *string
	OwnerID     string
}

type UpdateProjectParams struct {
	ID          string
	Name        string
	Description *string
}

type Repository interface {
	ListAccessible(ctx context.Context, userID string, page int, limit int) ([]Project, error)
	Create(ctx context.Context, params CreateProjectParams) (Project, error)
	GetByID(ctx context.Context, projectID string) (Project, error)
	UserHasTaskAccess(ctx context.Context, projectID string, userID string) (bool, error)
	ListTasksByProjectID(ctx context.Context, projectID string) ([]Task, error)
	Update(ctx context.Context, params UpdateProjectParams) (Project, error)
	Delete(ctx context.Context, projectID string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) ListAccessible(ctx context.Context, userID string, page int, limit int) ([]Project, error) {
	const query = `
		SELECT
			p.id,
			p.name,
			p.description,
			p.owner_id,
			p.created_at
		FROM projects p
		WHERE
			p.owner_id = $1
			OR EXISTS (
				SELECT 1
				FROM tasks t
				WHERE
					t.project_id = p.id
					AND (t.assignee_id = $1 OR t.creator_id = $1)
			)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list accessible projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.OwnerID,
			&project.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project list row: %w", err)
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project list rows: %w", err)
	}

	return projects, nil
}

func (r *PostgresRepository) Create(ctx context.Context, params CreateProjectParams) (Project, error) {
	const query = `
		INSERT INTO projects (id, name, description, owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, owner_id, created_at
	`

	var project Project
	err := r.pool.QueryRow(ctx, query, params.ID, params.Name, params.Description, params.OwnerID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		return Project{}, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, projectID string) (Project, error) {
	const query = `
		SELECT id, name, description, owner_id, created_at
		FROM projects
		WHERE id = $1
	`

	var project Project
	err := r.pool.QueryRow(ctx, query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}

		return Project{}, fmt.Errorf("get project by id: %w", err)
	}

	return project, nil
}

func (r *PostgresRepository) UserHasTaskAccess(ctx context.Context, projectID string, userID string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM tasks
			WHERE
				project_id = $1
				AND (assignee_id = $2 OR creator_id = $2)
		)
	`

	var hasAccess bool
	if err := r.pool.QueryRow(ctx, query, projectID, userID).Scan(&hasAccess); err != nil {
		return false, fmt.Errorf("check project task access: %w", err)
	}

	return hasAccess, nil
}

func (r *PostgresRepository) ListTasksByProjectID(ctx context.Context, projectID string) ([]Task, error) {
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
		WHERE project_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project tasks: %w", err)
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
			return nil, fmt.Errorf("scan project task row: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project task rows: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepository) Update(ctx context.Context, params UpdateProjectParams) (Project, error) {
	const query = `
		UPDATE projects
		SET
			name = $2,
			description = $3
		WHERE id = $1
		RETURNING id, name, description, owner_id, created_at
	`

	var project Project
	err := r.pool.QueryRow(ctx, query, params.ID, params.Name, params.Description).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.OwnerID,
		&project.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Project{}, ErrProjectNotFound
		}

		return Project{}, fmt.Errorf("update project: %w", err)
	}

	return project, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, projectID string) error {
	const query = `DELETE FROM projects WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrProjectNotFound
	}

	return nil
}
