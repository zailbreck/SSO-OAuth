package repositories

import (
	"database/sql"
	"fmt"
	"sso-service/app/models"
	"time"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetUserByUsername(username string) (models.User, error)
	GetUserByID(id uuid.UUID) (models.User, error)
	GetAllUsers() ([]models.User, error)
	CreateUser(user models.User) (models.User, error)
	UpdateUser(id uuid.UUID, user models.User) (models.User, error)
	DeleteUser(id uuid.UUID) error
}

// UserRepositoryImpl is the implementation of UserRepository
type UserRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepositoryImpl
func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// GetUserByUsername retrieves a user by their username from the database
func (r *UserRepositoryImpl) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, is_active, created_at, updated_at FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found: %w", err)
		}
		return models.User{}, fmt.Errorf("error getting user by username: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by their ID from the database
func (r *UserRepositoryImpl) GetUserByID(id uuid.UUID) (models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password_hash, is_active, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found: %w", err)
		}
		return models.User{}, fmt.Errorf("error getting user by ID: %w", err)
	}
	return user, nil
}

// GetAllUsers retrieves all users from the database
func (r *UserRepositoryImpl) GetAllUsers() ([]models.User, error) {
	query := `SELECT id, username, email, is_active, created_at, updated_at FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		// Note: password_hash is not selected here for security
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over users: %w", err)
	}
	return users, nil
}

// CreateUser inserts a new user into the database
func (r *UserRepositoryImpl) CreateUser(user models.User) (models.User, error) {
	user.ID = uuid.New() // Generate a new UUID for the user
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := r.db.QueryRow(query, user.ID, user.Username, user.Email, user.PasswordHash, user.IsActive, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user in the database
func (r *UserRepositoryImpl) UpdateUser(id uuid.UUID, user models.User) (models.User, error) {
	user.UpdatedAt = time.Now()
	query := `UPDATE users SET username = $1, email = $2, password_hash = $3, is_active = $4, updated_at = $5 WHERE id = $6 RETURNING id`
	err := r.db.QueryRow(query, user.Username, user.Email, user.PasswordHash, user.IsActive, user.UpdatedAt, id).Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found for update: %w", err)
		}
		return models.User{}, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// DeleteUser deletes a user from the database
func (r *UserRepositoryImpl) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found for deletion")
	}
	return nil
}
