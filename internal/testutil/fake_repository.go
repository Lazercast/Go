// Package testutil provides in-memory test doubles shared across the
// service and handler test suites. Keeping it as its own package (as
// opposed to a _test.go file) lets both packages import the same fake
// implementation instead of duplicating it.
package testutil

import (
	"context"
	"time"

	"go-task-api/internal/model"
	"go-task-api/internal/repository"
)

// FakeTaskRepository is an in-memory implementation of
// repository.TaskRepository. It is used by service and handler tests
// so they can run with `go test ./...` without needing a real
// PostgreSQL instance.
type FakeTaskRepository struct {
	tasks  map[int64]model.Task
	nextID int64
}

// NewFakeTaskRepository creates an empty in-memory repository.
func NewFakeTaskRepository() *FakeTaskRepository {
	return &FakeTaskRepository{
		tasks:  make(map[int64]model.Task),
		nextID: 1,
	}
}

func (f *FakeTaskRepository) Create(ctx context.Context, task *model.Task) error {
	now := time.Now()
	task.ID = f.nextID
	task.CreatedAt = now
	task.UpdatedAt = now
	f.tasks[task.ID] = *task
	f.nextID++
	return nil
}

func (f *FakeTaskRepository) GetByID(ctx context.Context, id int64) (*model.Task, error) {
	task, ok := f.tasks[id]
	if !ok {
		return nil, repository.ErrTaskNotFound
	}
	return &task, nil
}

func (f *FakeTaskRepository) List(ctx context.Context) ([]model.Task, error) {
	tasks := make([]model.Task, 0, len(f.tasks))
	for _, task := range f.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (f *FakeTaskRepository) Update(ctx context.Context, task *model.Task) error {
	existing, ok := f.tasks[task.ID]
	if !ok {
		return repository.ErrTaskNotFound
	}

	existing.Title = task.Title
	existing.Description = task.Description
	existing.Status = task.Status
	existing.UpdatedAt = time.Now()

	f.tasks[task.ID] = existing
	task.UpdatedAt = existing.UpdatedAt
	return nil
}

func (f *FakeTaskRepository) Delete(ctx context.Context, id int64) error {
	if _, ok := f.tasks[id]; !ok {
		return repository.ErrTaskNotFound
	}
	delete(f.tasks, id)
	return nil
}
