package repo

import (
	"time"

	"github.com/gocql/gocql"
)

// NewScyllaDbRepo creates a new ScyllaDbRepo instance
func NewScyllaDbRepo(session *gocql.Session) *ScyllaDbRepo {
	return &ScyllaDbRepo{
		session: session,
	}
}

type ScyllaDbRepo struct {
	session *gocql.Session
}

// CreateExampleTable creates a sample table in ScyllaDB
func (s *ScyllaDbRepo) CreateExampleTable() error {
	query := `
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY,
            name TEXT,
            email TEXT,
            created_at TIMESTAMP
        )
    `
	return s.session.Query(query).Exec()
}

// InsertUser inserts a new user into the users table
// It uses gocql.UUID for the ID, which is a universally unique identifier.
func (s *ScyllaDbRepo) InsertUser(id gocql.UUID, name, email string) error {
	query := `INSERT INTO users (id, name, email, created_at) VALUES (?, ?, ?, ?)`
	return s.session.Query(query, id, name, email, time.Now()).Exec()
}

// GetUser retrieves a user by ID from the users table
// It returns the user's name, email, and creation timestamp.
func (s *ScyllaDbRepo) GetUser(id gocql.UUID) (string, string, time.Time, error) {
	var name, email string
	var createdAt time.Time

	query := `SELECT name, email, created_at FROM users WHERE id = ?`
	err := s.session.Query(query, id).Scan(&name, &email, &createdAt)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return name, email, createdAt, nil
}

// GetAllUsers retrieves all users from the users table
func (s *ScyllaDbRepo) GetAllUsers() ([]User, error) {
	var users []User

	query := `SELECT id, name, email, created_at FROM users`
	iter := s.session.Query(query).Iter()

	var id gocql.UUID
	var name, email string
	var createdAt time.Time

	for iter.Scan(&id, &name, &email, &createdAt) {
		users = append(users, User{
			ID:        id,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates an existing user's name and email in the users table
func (s *ScyllaDbRepo) UpdateUser(id gocql.UUID, name, email string) error {
	query := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	return s.session.Query(query, name, email, id).Exec()
}

// DeleteUser deletes a user by ID from the users table
func (s *ScyllaDbRepo) DeleteUser(id gocql.UUID) error {
	query := `DELETE FROM users WHERE id = ?`
	return s.session.Query(query, id).Exec()
}

// User struct represents a user in the users table
// It includes fields for ID, name, email, and creation timestamp.
type User struct {
	ID        gocql.UUID `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
}

// PreparedInsertUser Example of using prepared statements
func (s *ScyllaDbRepo) PreparedInsertUser() error {
	// Gocql does not support explicit Prepare, but Query uses prepared statements internally.
	id := gocql.TimeUUID()
	return s.session.Query(`INSERT INTO users (id, name, email, created_at) VALUES (?, ?, ?, ?)`,
		id, "John Doe", "john@example.com", time.Now()).Exec()
}

// BatchInsertUsers Example of batch inserting users
// This function takes a slice of User structs and inserts them into the users table in a single batch operation.
func (s *ScyllaDbRepo) BatchInsertUsers(users []User) error {
	batch := s.session.NewBatch(gocql.LoggedBatch)

	for _, user := range users {
		batch.Query(`INSERT INTO users (id, name, email, created_at) VALUES (?, ?, ?, ?)`,
			user.ID, user.Name, user.Email, user.CreatedAt)
	}

	return s.session.ExecuteBatch(batch)
}

// GetUsersWithPagination Example of paginated query
// Note: This is a placeholder function. You can implement your pagination logic here.
func (s *ScyllaDbRepo) GetUsersWithPagination(offset int, limit int) ([]User, error) {
	var users []User

	// ScyllaDB/Cassandra doesn't support OFFSET, use token-based pagination instead
	query := `SELECT id, name, email, created_at FROM users LIMIT ?`
	iter := s.session.Query(query, limit).Iter()

	var id gocql.UUID
	var name, email string
	var createdAt time.Time

	for iter.Scan(&id, &name, &email, &createdAt) {
		users = append(users, User{
			ID:        id,
			Name:      name,
			Email:     email,
			CreatedAt: createdAt,
		})
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return users, nil
}

// CountUsers Example of counting users
// Note: This is a placeholder function. You can implement your counting logic here.
func (s *ScyllaDbRepo) CountUsers() (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM users`
	if err := s.session.Query(query).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
