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
	GetAllUsers() ([]models.User, error)                            // New
	CreateUser(user models.User) (models.User, error)               // New
	UpdateUser(id uuid.UUID, user models.User) (models.User, error) // New
	DeleteUser(id uuid.UUID) error                                  // New
	GetRolesForUser(userID uuid.UUID) ([]models.Role, error)
	GetPermissionsForUser(userID uuid.UUID) ([]string, error)
	AddRevokedToken(jti string, expiresAt time.Time) error
	IsTokenRevoked(jti string) (bool, error)
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
	// Note: In a real app, you might update fields selectively and not require all fields.
	// This example assumes all fields are provided for update.
	// For partial updates, you'd build the SQL query dynamically.
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

// GetRolesForUser retrieves all roles assigned to a user from the database
func (r *UserRepositoryImpl) GetRolesForUser(userID uuid.UUID) ([]models.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting roles for user: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over roles: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found for user %s", userID.String())
	}
	return roles, nil
}

// GetPermissionsForUser retrieves all unique permissions for a given user, including hierarchical children.
func (r *UserRepositoryImpl) GetPermissionsForUser(userID uuid.UUID) ([]string, error) {
	query := `
		WITH RECURSIVE permission_hierarchy AS (
			-- Base case: Directly assigned permissions to the user's roles
			SELECT
				p.id,
				p.name,
				p.parent_id
			FROM
				permissions p
			JOIN
				role_permissions rp ON p.id = rp.permission_id
			JOIN
				user_roles ur ON rp.role_id = ur.role_id
			WHERE
				ur.user_id = $1

			UNION ALL

			-- Recursive step: Find children of the permissions found so far
			SELECT
				p_child.id,
				p_child.name,
				p_child.parent_id
			FROM
				permissions p_child
			JOIN
				permission_hierarchy ph ON p_child.parent_id = ph.id
		)
		SELECT DISTINCT name FROM permission_hierarchy;`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting hierarchical permissions for user: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permName string
		if err := rows.Scan(&permName); err != nil {
			return nil, fmt.Errorf("error scanning permission name: %w", err)
		}
		permissions = append(permissions, permName)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over permissions: %w", err)
	}

	if len(permissions) == 0 {
		return nil, fmt.Errorf("no permissions found for user %s", userID.String())
	}
	return permissions, nil
}

// AddRevokedToken adds a token's JTI to the revoked_tokens table
func (r *UserRepositoryImpl) AddRevokedToken(jti string, expiresAt time.Time) error {
	query := `INSERT INTO revoked_tokens (jti, expires_at, revoked_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, jti, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	return nil
}

// IsTokenRevoked checks if a token's JTI exists in the revoked_tokens table
func (r *UserRepositoryImpl) IsTokenRevoked(jti string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti = $1)`
	err := r.db.QueryRow(query, jti).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if token is revoked: %w", err)
	}
	return false, nil
}
