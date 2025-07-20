package repositories

import (
	"errors"
	"fmt"
	"sso-service/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	CreatePermission(permission models.Permission) (models.Permission, error)
	GetAllPermissions() ([]models.Permission, error)
	GetPermissionByID(id uuid.UUID) (models.Permission, error)
	GetPermissionByNameAndSite(name string, siteID uuid.UUID) (models.Permission, error)
	UpdatePermission(permission models.Permission) (models.Permission, error)
	DeletePermission(id uuid.UUID) error
	AssignPermissionToRole(roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(roleID, permissionID uuid.UUID) error
	GetPermissionsForUser(userID uuid.UUID) ([]string, error) // Tetap menggunakan Raw SQL
}

// PermissionRepositoryImpl is the implementation of PermissionRepository
type PermissionRepositoryImpl struct {
	db *gorm.DB // Diubah ke *gorm.DB
}

// NewPermissionRepository creates a new instance of PermissionRepositoryImpl
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &PermissionRepositoryImpl{db: db}
}

// CreatePermission inserts a new permission into the database using GORM
func (r *PermissionRepositoryImpl) CreatePermission(permission models.Permission) (models.Permission, error) {
	// GORM akan otomatis mengisi ID, CreatedAt, dan UpdatedAt
	result := r.db.Create(&permission)
	if result.Error != nil {
		return models.Permission{}, fmt.Errorf("failed to create permission: %w", result.Error)
	}
	return permission, nil
}

// GetAllPermissions retrieves all permissions from the database using GORM
func (r *PermissionRepositoryImpl) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	result := r.db.Find(&permissions)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting all permissions: %w", result.Error)
	}
	return permissions, nil
}

// GetPermissionByID retrieves a permission by its ID from the database using GORM
func (r *PermissionRepositoryImpl) GetPermissionByID(id uuid.UUID) (models.Permission, error) {
	var perm models.Permission
	result := r.db.First(&perm, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Permission{}, fmt.Errorf("permission not found: %w", result.Error)
		}
		return models.Permission{}, fmt.Errorf("error getting permission by ID: %w", result.Error)
	}
	return perm, nil
}

// GetPermissionByNameAndSite retrieves a permission by its name and site ID using GORM
func (r *PermissionRepositoryImpl) GetPermissionByNameAndSite(name string, siteID uuid.UUID) (models.Permission, error) {
	var perm models.Permission
	result := r.db.Where("name = ? AND site_id = ?", name, siteID).First(&perm)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Permission{}, fmt.Errorf("permission not found: %w", result.Error)
		}
		return models.Permission{}, fmt.Errorf("error getting permission by name and site: %w", result.Error)
	}
	return perm, nil
}

// UpdatePermission updates an existing permission in the database using GORM
func (r *PermissionRepositoryImpl) UpdatePermission(permission models.Permission) (models.Permission, error) {
	// Pastikan ID ada untuk pembaruan
	if permission.ID == uuid.Nil {
		return models.Permission{}, errors.New("cannot update permission without ID")
	}
	// GORM's Save akan memperbarui semua kolom, atau membuat record baru jika ID tidak ada.
	// Ini juga akan secara otomatis memperbarui field UpdatedAt jika model Anda memiliki tag gorm.
	result := r.db.Save(&permission)
	if result.Error != nil {
		return models.Permission{}, fmt.Errorf("failed to update permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.Permission{}, gorm.ErrRecordNotFound
	}
	return permission, nil
}

// DeletePermission deletes a permission from the database using GORM
func (r *PermissionRepositoryImpl) DeletePermission(id uuid.UUID) error {
	// Menghapus record berdasarkan primary key
	result := r.db.Delete(&models.Permission{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found for deletion")
	}
	return nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *PermissionRepositoryImpl) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	// Untuk operasi join table, menggunakan GORM's Exec lebih sederhana
	// daripada setup GORM's many-to-many associations jika belum ada.
	association := map[string]interface{}{"role_id": roleID, "permission_id": permissionID}
	result := r.db.Table("role_permissions").Create(association)

	if result.Error != nil {
		// Cek duplikasi secara spesifik jika diperlukan
		return fmt.Errorf("failed to assign permission to role: %w", result.Error)
	}
	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *PermissionRepositoryImpl) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	result := r.db.Exec("DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permissionID)
	if result.Error != nil {
		return fmt.Errorf("failed to remove permission from role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("role permission assignment not found")
	}
	return nil
}

// GetPermissionsForUser retrieves all unique permissions for a given user, including hierarchical children.
// FUNGSI INI SENGAJA TIDAK DIUBAH UNTUK MENJAGA PERFORMA.
func (r *PermissionRepositoryImpl) GetPermissionsForUser(userID uuid.UUID) ([]string, error) {
	query := `
		WITH RECURSIVE permission_hierarchy AS (
			SELECT p.id, p.name, p.parent_id FROM permissions p
			JOIN role_permissions rp ON p.id = rp.permission_id
			JOIN user_roles ur ON rp.role_id = ur.role_id
			WHERE ur.user_id = $1
			UNION ALL
			SELECT p_child.id, p_child.name, p_child.parent_id FROM permissions p_child
			JOIN permission_hierarchy ph ON p_child.parent_id = ph.id
		)
		SELECT DISTINCT name FROM permission_hierarchy;`

	var permissions []string
	// Menggunakan Raw SQL dengan GORM
	err := r.db.Raw(query, userID).Scan(&permissions).Error
	if err != nil {
		return nil, fmt.Errorf("error getting hierarchical permissions for user: %w", err)
	}

	if len(permissions) == 0 {
		return nil, errors.New("no permissions found for user")
	}
	return permissions, nil
}
