package repository

import (
	"context"
	"database/sql"
	"errors"

	"go-task-api/internal/model"
)

// ErrTaskNotFound is returned when a task cannot be found by its ID.
var ErrTaskNotFound = errors.New("task not found")

// TaskRepository defines the persistence operations for tasks.
// Using an interface here keeps the service layer decoupled from
// PostgreSQL, which also makes the service easy to unit test with a
// fake in-memory implementation.
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id int64) (*model.Task, error)
	List(ctx context.Context) ([]model.Task, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id int64) error
}

// postgresTaskRepository is the PostgreSQL-backed implementation of
// TaskRepository.
type postgresTaskRepository struct {
	db *sql.DB
}

// NewPostgresTaskRepository creates a TaskRepository backed by the
// given database connection.
func NewPostgresTaskRepository(db *sql.DB) TaskRepository {
	return &postgresTaskRepository{db: db}
}

func (r *postgresTaskRepository) Create(ctx context.Context, task *model.Task) error {
	const query = `
		INSERT INTO tasks (title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query, task.Title, task.Description, task.Status).
		Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *postgresTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task model.Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatedAt, &task.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *postgresTaskRepository) List(ctx context.Context) ([]model.Task, error) {
	const query = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status,
			&task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *postgresTaskRepository) Update(ctx context.Context, task *model.Task) error {
	const query = `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query, task.Title, task.Description, task.Status, task.ID).
		Scan(&task.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaskNotFound
	}
	return err
}

func (r *postgresTaskRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTaskNotFound
	}

	return nil
}
