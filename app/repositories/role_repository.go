package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"sso-service/app/models"
	"time"

	"github.com/google/uuid"
)

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	CreateRole(role models.Role) (models.Role, error)
	GetAllRoles() ([]models.Role, error)
	GetRoleByID(id uuid.UUID) (models.Role, error)
	GetRoleByName(name string) (models.Role, error) // New helper
	UpdateRole(id uuid.UUID, role models.Role) (models.Role, error)
	DeleteRole(id uuid.UUID) error
	AssignRoleToUser(userID, roleID uuid.UUID) error
	RemoveRoleFromUser(userID, roleID uuid.UUID) error
	GetRolesForUser(userID uuid.UUID) ([]models.Role, error) // Moved from user_repository
}

// RoleRepositoryImpl is the implementation of RoleRepository
type RoleRepositoryImpl struct {
	db *sql.DB
}

// NewRoleRepository creates a new instance of RoleRepositoryImpl
func NewRoleRepository(db *sql.DB) RoleRepository {
	return &RoleRepositoryImpl{db: db}
}

// CreateRole inserts a new role into the database
func (r *RoleRepositoryImpl) CreateRole(role models.Role) (models.Role, error) {
	role.ID = uuid.New()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	query := `INSERT INTO roles (id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, role.ID, role.Name, role.Description, role.CreatedAt, role.UpdatedAt).Scan(&role.ID)
	if err != nil {
		return models.Role{}, fmt.Errorf("failed to create role: %w", err)
	}
	return role, nil
}

// GetAllRoles retrieves all roles from the database
func (r *RoleRepositoryImpl) GetAllRoles() ([]models.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all roles: %w", err)
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
	return roles, nil
}

// GetRoleByID retrieves a role by its ID from the database
func (r *RoleRepositoryImpl) GetRoleByID(id uuid.UUID) (models.Role, error) {
	var role models.Role
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Role{}, fmt.Errorf("role not found: %w", err)
		}
		return models.Role{}, fmt.Errorf("error getting role by ID: %w", err)
	}
	return role, nil
}

// GetRoleByName retrieves a role by its name from the database
func (r *RoleRepositoryImpl) GetRoleByName(name string) (models.Role, error) {
	var role models.Role
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE name = $1`
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Role{}, fmt.Errorf("role not found: %w", err)
		}
		return models.Role{}, fmt.Errorf("error getting role by name: %w", err)
	}
	return role, nil
}

// UpdateRole updates an existing role in the database
func (r *RoleRepositoryImpl) UpdateRole(id uuid.UUID, role models.Role) (models.Role, error) {
	role.UpdatedAt = time.Now()
	query := `UPDATE roles SET name = $1, description = $2, updated_at = $3 WHERE id = $4 RETURNING id`
	err := r.db.QueryRow(query, role.Name, role.Description, role.UpdatedAt, id).Scan(&role.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Role{}, fmt.Errorf("role not found for update: %w", err)
		}
		return models.Role{}, fmt.Errorf("failed to update role: %w", err)
	}
	return role, nil
}

// DeleteRole deletes a role from the database
func (r *RoleRepositoryImpl) DeleteRole(id uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("role not found for deletion")
	}
	return nil
}

// AssignRoleToUser assigns a role to a user
func (r *RoleRepositoryImpl) AssignRoleToUser(userID, roleID uuid.UUID) error {
	query := `INSERT INTO user_roles (user_id, role_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, userID, roleID, time.Now())
	if err != nil {
		// Check for unique constraint violation (user already has this role)
		if err.Error() == `pq: duplicate key value violates unique constraint "user_roles_pkey"` { // PostgreSQL specific error
			return errors.New("user already has this role assigned")
		}
		return fmt.Errorf("failed to assign role to user: %w", err)
	}
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RoleRepositoryImpl) RemoveRoleFromUser(userID, roleID uuid.UUID) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`
	result, err := r.db.Exec(query, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after remove role: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("user role assignment not found")
	}
	return nil
}

// GetRolesForUser retrieves all roles assigned to a user from the database
// This function is crucial for JWT claims and authorization checks
func (r *RoleRepositoryImpl) GetRolesForUser(userID uuid.UUID) ([]models.Role, error) {
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
