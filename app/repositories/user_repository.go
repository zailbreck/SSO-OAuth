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
	GetRolesForUser(userID uuid.UUID) ([]models.Role, error)
	GetPermissionsForUser(userID uuid.UUID) ([]string, error)
	AddRevokedToken(jti string, expiresAt time.Time) error // New method
	IsTokenRevoked(jti string) (bool, error)               // New method
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
	return exists, nil
}
