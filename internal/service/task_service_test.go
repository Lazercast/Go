package service_test

import (
	"context"
	"errors"
	"testing"

	"go-task-api/internal/model"
	"go-task-api/internal/service"
	"go-task-api/internal/testutil"
)

func newTestService() *service.TaskService {
	repo := testutil.NewFakeTaskRepository()
	return service.NewTaskService(repo)
}

func TestCreateTask_Success(t *testing.T) {
	svc := newTestService()

	task, err := svc.CreateTask(context.Background(), service.CreateTaskInput{
		Title:       "Write tests",
		Description: "Cover the service layer",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if task.ID == 0 {
		t.Error("expected task to be assigned an ID")
	}
	if task.Status != model.StatusTodo {
		t.Errorf("expected default status %q, got %q", model.StatusTodo, task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateTask_MissingTitle(t *testing.T) {
	svc := newTestService()

	_, err := svc.CreateTask(context.Background(), service.CreateTaskInput{Title: "  "})
	if !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	svc := newTestService()

	_, err := svc.CreateTask(context.Background(), service.CreateTaskInput{
		Title:  "Bad status",
		Status: "not-a-status",
	})
	if !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGetTask_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	created, err := svc.CreateTask(ctx, service.CreateTaskInput{Title: "Find me"})
	if err != nil {
		t.Fatalf("setup: unexpected error creating task: %v", err)
	}

	found, err := svc.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Title != "Find me" {
		t.Errorf("expected title %q, got %q", "Find me", found.Title)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetTask(context.Background(), 999)
	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestListTasks(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.CreateTask(ctx, service.CreateTaskInput{Title: "First"}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := svc.CreateTask(ctx, service.CreateTaskInput{Title: "Second"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	tasks, err := svc.ListTasks(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestUpdateTask_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	created, err := svc.CreateTask(ctx, service.CreateTaskInput{Title: "Original"})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	updated, err := svc.UpdateTask(ctx, created.ID, service.UpdateTaskInput{
		Title:  "Updated",
		Status: model.StatusInProgress,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Title != "Updated" {
		t.Errorf("expected title %q, got %q", "Updated", updated.Title)
	}
	if updated.Status != model.StatusInProgress {
		t.Errorf("expected status %q, got %q", model.StatusInProgress, updated.Status)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.UpdateTask(context.Background(), 999, service.UpdateTaskInput{Title: "Ghost"})
	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestDeleteTask_Success(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	created, err := svc.CreateTask(ctx, service.CreateTaskInput{Title: "Temporary"})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := svc.DeleteTask(ctx, created.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := svc.GetTask(ctx, created.ID); !errors.Is(err, service.ErrTaskNotFound) {
		t.Fatalf("expected task to be deleted, got err=%v", err)
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	svc := newTestService()

	err := svc.DeleteTask(context.Background(), 999)
	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
