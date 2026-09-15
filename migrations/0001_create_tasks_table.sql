-- Creates the tasks table used by the go-task-api service.
-- This file is automatically executed by the official PostgreSQL
-- Docker image on first container startup (see docker-compose.yml,
-- which mounts this folder into /docker-entrypoint-initdb.d).

CREATE TABLE IF NOT EXISTS tasks (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      VARCHAR(20) NOT NULL DEFAULT 'todo'
                CHECK (status IN ('todo', 'in_progress', 'done')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);
