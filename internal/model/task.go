package model

import "time"

// Task statuses supported by the API.
const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

// ValidStatuses lists every status value accepted by the API.
var ValidStatuses = map[string]bool{
	StatusTodo:       true,
	StatusInProgress: true,
	StatusDone:       true,
}

// Task represents a single task stored in the database.
type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
