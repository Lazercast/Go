package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go-task-api/internal/model"
	"go-task-api/internal/repository"
)

// ErrValidation is returned when the input provided by the client is
// invalid. Handlers translate this into a 400 Bad Request response.
var ErrValidation = errors.New("validation error")

// ErrTaskNotFound is re-exported from the repository so callers only
// need to depend on the service package.
var ErrTaskNotFound = repository.ErrTaskNotFound

const maxTitleLength = 255

// CreateTaskInput holds the fields accepted when creating a task.
type CreateTaskInput struct {
	Title       string
	Description string
	Status      string
}

// UpdateTaskInput holds the fields accepted when updating a task.
type UpdateTaskInput struct {
	Title       string
	Description string
	Status      string
}

// TaskService contains the business logic for managing tasks. It
// validates input and delegates persistence to a TaskRepository.
type TaskService struct {
	repo repository.TaskRepository
}

// NewTaskService creates a TaskService backed by the given repository.
func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

// CreateTask validates the input and stores a new task.
func (s *TaskService) CreateTask(ctx context.Context, input CreateTaskInput) (*model.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, validationErrorf("title is required")
	}
	if len(title) > maxTitleLength {
		return nil, validationErrorf("title must be at most %d characters", maxTitleLength)
	}

	status := input.Status
	if status == "" {
		status = model.StatusTodo
	}
	if !model.ValidStatuses[status] {
		return nil, validationErrorf("status must be one of: todo, in_progress, done")
	}

	task := &model.Task{
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask returns a single task by ID.
func (s *TaskService) GetTask(ctx context.Context, id int64) (*model.Task, error) {
	return s.repo.GetByID(ctx, id)
}

// ListTasks returns every task in the system.
func (s *TaskService) ListTasks(ctx context.Context) ([]model.Task, error) {
	return s.repo.List(ctx)
}

// UpdateTask validates the input and replaces an existing task's
// title, description and status.
func (s *TaskService) UpdateTask(ctx context.Context, id int64, input UpdateTaskInput) (*model.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, validationErrorf("title is required")
	}
	if len(title) > maxTitleLength {
		return nil, validationErrorf("title must be at most %d characters", maxTitleLength)
	}

	status := input.Status
	if status == "" {
		status = model.StatusTodo
	}
	if !model.ValidStatuses[status] {
		return nil, validationErrorf("status must be one of: todo, in_progress, done")
	}

	task := &model.Task{
		ID:          id,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Status:      status,
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Fetch the full row (including created_at) so the handler can
	// return a complete representation of the task.
	return s.repo.GetByID(ctx, id)
}

// DeleteTask removes a task by ID.
func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// validationErrorf builds an error that wraps ErrValidation, so callers
// can check the error kind with errors.Is(err, service.ErrValidation)
// while still getting a specific, user-facing message.
func validationErrorf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrValidation, fmt.Sprintf(format, args...))
}
