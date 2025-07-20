package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"sso-service/app/models"
	"time"

	"github.com/google/uuid"
)

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	CreatePermission(permission models.Permission) (models.Permission, error)
	GetAllPermissions() ([]models.Permission, error)
	GetPermissionByID(id uuid.UUID) (models.Permission, error)
	GetPermissionByNameAndSite(name string, siteID uuid.UUID) (models.Permission, error) // New helper
	UpdatePermission(id uuid.UUID, permission models.Permission) (models.Permission, error)
	DeletePermission(id uuid.UUID) error
	AssignPermissionToRole(roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(roleID, permissionID uuid.UUID) error
	GetPermissionsForUser(userID uuid.UUID) ([]string, error) // Moved from user_repository
}

// PermissionRepositoryImpl is the implementation of PermissionRepository
type PermissionRepositoryImpl struct {
	db *sql.DB
}

// NewPermissionRepository creates a new instance of PermissionRepositoryImpl
func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &PermissionRepositoryImpl{db: db}
}

// CreatePermission inserts a new permission into the database
func (r *PermissionRepositoryImpl) CreatePermission(permission models.Permission) (models.Permission, error) {
	permission.ID = uuid.New()
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()

	query := `INSERT INTO permissions (id, site_id, name, description, parent_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var parentID sql.NullString
	if permission.ParentID != nil {
		parentID.String = permission.ParentID.String()
		parentID.Valid = true
	} else {
		parentID.Valid = false
	}

	err := r.db.QueryRow(query, permission.ID, permission.SiteID, permission.Name, permission.Description, parentID, permission.CreatedAt, permission.UpdatedAt).Scan(&permission.ID)
	if err != nil {
		return models.Permission{}, fmt.Errorf("failed to create permission: %w", err)
	}
	return permission, nil
}

// GetAllPermissions retrieves all permissions from the database
func (r *PermissionRepositoryImpl) GetAllPermissions() ([]models.Permission, error) {
	query := `SELECT id, site_id, name, description, parent_id, created_at, updated_at FROM permissions`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var perm models.Permission
		var parentID sql.NullString // Use sql.NullString for nullable UUID
		if err := rows.Scan(&perm.ID, &perm.SiteID, &perm.Name, &perm.Description, &parentID, &perm.CreatedAt, &perm.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning permission: %w", err)
		}
		if parentID.Valid {
			parsedParentID, err := uuid.Parse(parentID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid parent_id UUID format: %w", err)
			}
			perm.ParentID = &parsedParentID
		} else {
			perm.ParentID = nil
		}
		permissions = append(permissions, perm)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over permissions: %w", err)
	}
	return permissions, nil
}

// GetPermissionByID retrieves a permission by its ID from the database
func (r *PermissionRepositoryImpl) GetPermissionByID(id uuid.UUID) (models.Permission, error) {
	var perm models.Permission
	var parentID sql.NullString
	query := `SELECT id, site_id, name, description, parent_id, created_at, updated_at FROM permissions WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&perm.ID, &perm.SiteID, &perm.Name, &perm.Description, &parentID, &perm.CreatedAt, &perm.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Permission{}, fmt.Errorf("permission not found: %w", err)
		}
		return models.Permission{}, fmt.Errorf("error getting permission by ID: %w", err)
	}
	if parentID.Valid {
		parsedParentID, err := uuid.Parse(parentID.String)
		if err != nil {
			return models.Permission{}, fmt.Errorf("invalid parent_id UUID format: %w", err)
		}
		perm.ParentID = &parsedParentID
	} else {
		perm.ParentID = nil
	}
	return perm, nil
}

// GetPermissionByNameAndSite retrieves a permission by its name and site ID from the database
func (r *PermissionRepositoryImpl) GetPermissionByNameAndSite(name string, siteID uuid.UUID) (models.Permission, error) {
	var perm models.Permission
	var parentID sql.NullString
	query := `SELECT id, site_id, name, description, parent_id, created_at, updated_at FROM permissions WHERE name = $1 AND site_id = $2`
	err := r.db.QueryRow(query, name, siteID).Scan(&perm.ID, &perm.SiteID, &perm.Name, &perm.Description, &parentID, &perm.CreatedAt, &perm.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Permission{}, fmt.Errorf("permission not found: %w", err)
		}
		return models.Permission{}, fmt.Errorf("error getting permission by name and site: %w", err)
	}
	if parentID.Valid {
		parsedParentID, err := uuid.Parse(parentID.String)
		if err != nil {
			return models.Permission{}, fmt.Errorf("invalid parent_id UUID format: %w", err)
		}
		perm.ParentID = &parsedParentID
	} else {
		perm.ParentID = nil
	}
	return perm, nil
}

// UpdatePermission updates an existing permission in the database
func (r *PermissionRepositoryImpl) UpdatePermission(id uuid.UUID, permission models.Permission) (models.Permission, error) {
	permission.UpdatedAt = time.Now()
	query := `UPDATE permissions SET site_id = $1, name = $2, description = $3, parent_id = $4, updated_at = $5 WHERE id = $6 RETURNING id`

	var parentID sql.NullString
	if permission.ParentID != nil {
		parentID.String = permission.ParentID.String()
		parentID.Valid = true
	} else {
		parentID.Valid = false
	}

	err := r.db.QueryRow(query, permission.SiteID, permission.Name, permission.Description, parentID, permission.UpdatedAt, id).Scan(&permission.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Permission{}, fmt.Errorf("permission not found for update: %w", err)
		}
		return models.Permission{}, fmt.Errorf("failed to update permission: %w", err)
	}
	return permission, nil
}

// DeletePermission deletes a permission from the database
func (r *PermissionRepositoryImpl) DeletePermission(id uuid.UUID) error {
	query := `DELETE FROM permissions WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("permission not found for deletion")
	}
	return nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *PermissionRepositoryImpl) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	query := `INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, roleID, permissionID, time.Now())
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "role_permissions_pkey"` { // PostgreSQL specific error
			return errors.New("role already has this permission assigned")
		}
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}
	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *PermissionRepositoryImpl) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	result, err := r.db.Exec(query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after remove permission: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("role permission assignment not found")
	}
	return nil
}

// GetPermissionsForUser retrieves all unique permissions for a given user, including hierarchical children.
// This function is crucial for JWT claims and authorization checks.
// It's placed here because it's a core permission retrieval logic.
func (r *PermissionRepositoryImpl) GetPermissionsForUser(userID uuid.UUID) ([]string, error) {
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
