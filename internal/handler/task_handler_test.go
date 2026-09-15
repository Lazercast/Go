package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"go-task-api/internal/handler"
	"go-task-api/internal/model"
	"go-task-api/internal/service"
	"go-task-api/internal/testutil"
)

// newTestServer builds a fully wired *http.ServeMux backed by an
// in-memory fake repository, so tests exercise the real
// handler -> service wiring without needing PostgreSQL.
func newTestServer() *http.ServeMux {
	repo := testutil.NewFakeTaskRepository()
	svc := service.NewTaskService(repo)
	h := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func doRequest(t *testing.T, mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func decodeTask(t *testing.T, rec *httptest.ResponseRecorder) model.Task {
	t.Helper()
	var task model.Task
	if err := json.NewDecoder(rec.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode task response: %v, body=%s", err, rec.Body.String())
	}
	return task
}

func TestCreateTask_Handler(t *testing.T) {
	mux := newTestServer()

	rec := doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{
		"title":       "Write README",
		"description": "Document the API",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusCreated, rec.Code, rec.Body.String())
	}

	task := decodeTask(t, rec)
	if task.Title != "Write README" {
		t.Errorf("expected title %q, got %q", "Write README", task.Title)
	}
	if task.Status != model.StatusTodo {
		t.Errorf("expected default status %q, got %q", model.StatusTodo, task.Status)
	}
}

func TestCreateTask_Handler_InvalidBody(t *testing.T) {
	mux := newTestServer()

	rec := doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{
		"title": "",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetTask_Handler(t *testing.T) {
	mux := newTestServer()

	created := decodeTask(t, doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{
		"title": "Find me",
	}))

	rec := doRequest(t, mux, http.MethodGet, "/tasks/"+itoa(created.ID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	found := decodeTask(t, rec)
	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
}

func TestGetTask_Handler_NotFound(t *testing.T) {
	mux := newTestServer()

	rec := doRequest(t, mux, http.MethodGet, "/tasks/999", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetTask_Handler_InvalidID(t *testing.T) {
	mux := newTestServer()

	rec := doRequest(t, mux, http.MethodGet, "/tasks/not-a-number", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestListTasks_Handler(t *testing.T) {
	mux := newTestServer()

	doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{"title": "First"})
	doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{"title": "Second"})

	rec := doRequest(t, mux, http.MethodGet, "/tasks", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var tasks []model.Task
	if err := json.NewDecoder(rec.Body).Decode(&tasks); err != nil {
		t.Fatalf("failed to decode task list: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestUpdateTask_Handler(t *testing.T) {
	mux := newTestServer()

	created := decodeTask(t, doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{
		"title": "Original",
	}))

	rec := doRequest(t, mux, http.MethodPut, "/tasks/"+itoa(created.ID), map[string]string{
		"title":  "Updated",
		"status": model.StatusDone,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusOK, rec.Code, rec.Body.String())
	}

	updated := decodeTask(t, rec)
	if updated.Title != "Updated" {
		t.Errorf("expected title %q, got %q", "Updated", updated.Title)
	}
	if updated.Status != model.StatusDone {
		t.Errorf("expected status %q, got %q", model.StatusDone, updated.Status)
	}
}

func TestUpdateTask_Handler_NotFound(t *testing.T) {
	mux := newTestServer()

	rec := doRequest(t, mux, http.MethodPut, "/tasks/999", map[string]string{
		"title": "Ghost",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDeleteTask_Handler(t *testing.T) {
	mux := newTestServer()

	created := decodeTask(t, doRequest(t, mux, http.MethodPost, "/tasks", map[string]string{
		"title": "Temporary",
	}))

	rec := doRequest(t, mux, http.MethodDelete, "/tasks/"+itoa(created.ID), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	getRec := doRequest(t, mux, http.MethodGet, "/tasks/"+itoa(created.ID), nil)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("expected deleted task to be gone, got status %d", getRec.Code)
	}
}

func TestDeleteTask_Handler_NotFound(t *testing.T) {
	mux := newTestServer()

	rec := doRequest(t, mux, http.MethodDelete, "/tasks/999", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
