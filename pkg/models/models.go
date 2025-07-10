package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Threat struct {
	ID        int       `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll() ([]User, error) {
	query := `
		SELECT id, email, name, created_at 
		FROM users 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int) (*User, error) {
	query := `
		SELECT id, email, name, created_at 
		FROM users 
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, err
	}

	return &user, nil
}

// ThreatRepository handles database operations for threats
type ThreatRepository struct {
	db *sql.DB
}

// NewThreatRepository creates a new threat repository
func NewThreatRepository(db *sql.DB) *ThreatRepository {
	return &ThreatRepository{db: db}
}

// GetAll retrieves all threats from the database
func (r *ThreatRepository) GetAll() ([]Threat, error) {
	query := `
		SELECT id, url, status, created_at 
		FROM threats 
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var threats []Threat
	for rows.Next() {
		var threat Threat
		err := rows.Scan(&threat.ID, &threat.URL, &threat.Status, &threat.CreatedAt)
		if err != nil {
			return nil, err
		}
		threats = append(threats, threat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threats, nil
}

// GetByID retrieves a threat by ID
func (r *ThreatRepository) GetByID(id int) (*Threat, error) {
	query := `
		SELECT id, url, status, created_at 
		FROM threats 
		WHERE id = $1
	`

	var threat Threat
	err := r.db.QueryRow(query, id).Scan(
		&threat.ID, &threat.URL, &threat.Status, &threat.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Threat not found
		}
		return nil, err
	}

	return &threat, nil
}
