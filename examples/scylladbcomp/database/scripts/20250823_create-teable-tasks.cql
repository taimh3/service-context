-- Create tasks table for ScyllaDB
-- This table corresponds to the Task struct in the Go application

-- Create keyspace if not exists (you may need to adjust replication strategy)
CREATE KEYSPACE IF NOT EXISTS auth WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'replication_factor': 3
};

USE auth;

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    title TEXT,
    description TEXT,
    status TEXT
);

-- Create index on status for efficient querying by task status
CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks (status);

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS tasks_created_at_idx ON tasks (created_at);

-- Optional: Create index on title for search functionality
CREATE INDEX IF NOT EXISTS tasks_title_idx ON tasks (title);
