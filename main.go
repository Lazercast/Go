package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"go-task-api/internal/config"
	"go-task-api/internal/handler"
	"go-task-api/internal/repository"
	"go-task-api/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := openDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.NewPostgresTaskRepository(db)
	svc := service.NewTaskService(repo)
	taskHandler := handler.NewTaskHandler(svc)

	mux := http.NewServeMux()
	taskHandler.RegisterRoutes(mux)

	addr := ":" + cfg.ServerPort
	log.Printf("go-task-api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

// openDatabase opens a connection pool and retries a few times, since
// the API container can start before PostgreSQL is ready to accept
// connections in Docker Compose.
func openDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	const maxAttempts = 10
	var pingErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingErr = db.Ping()
		if pingErr == nil {
			return db, nil
		}
		log.Printf("database not ready yet (attempt %d/%d): %v", attempt, maxAttempts, pingErr)
		time.Sleep(2 * time.Second)
	}

	return nil, pingErr
}
