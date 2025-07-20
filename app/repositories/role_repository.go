package repositories

import (
	"errors"
	"fmt"
	"sso-service/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	CreateRole(role models.Role) (models.Role, error)
	GetAllRoles() ([]models.Role, error)
	GetRoleByID(id uuid.UUID) (models.Role, error)
	GetRoleByName(name string) (models.Role, error)
	UpdateRole(role models.Role) (models.Role, error)
	DeleteRole(id uuid.UUID) error
	AssignRoleToUser(userID, roleID uuid.UUID) error
	RemoveRoleFromUser(userID, roleID uuid.UUID) error
	GetRolesForUser(userID uuid.UUID) ([]models.Role, error)
}

// RoleRepositoryImpl is the implementation of RoleRepository
type RoleRepositoryImpl struct {
	db *gorm.DB // Diubah ke *gorm.DB
}

// NewRoleRepository creates a new instance of RoleRepositoryImpl
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &RoleRepositoryImpl{db: db}
}

// CreateRole inserts a new role into the database using GORM
func (r *RoleRepositoryImpl) CreateRole(role models.Role) (models.Role, error) {
	result := r.db.Create(&role)
	if result.Error != nil {
		return models.Role{}, fmt.Errorf("failed to create role: %w", result.Error)
	}
	return role, nil
}

// GetAllRoles retrieves all roles from the database using GORM
func (r *RoleRepositoryImpl) GetAllRoles() ([]models.Role, error) {
	var roles []models.Role
	result := r.db.Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting all roles: %w", result.Error)
	}
	return roles, nil
}

// GetRoleByID retrieves a role by its ID from the database using GORM
func (r *RoleRepositoryImpl) GetRoleByID(id uuid.UUID) (models.Role, error) {
	var role models.Role
	result := r.db.First(&role, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Role{}, fmt.Errorf("role not found: %w", result.Error)
		}
		return models.Role{}, fmt.Errorf("error getting role by ID: %w", result.Error)
	}
	return role, nil
}

// GetRoleByName retrieves a role by its name from the database using GORM
func (r *RoleRepositoryImpl) GetRoleByName(name string) (models.Role, error) {
	var role models.Role
	result := r.db.Where("name = ?", name).First(&role)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return models.Role{}, fmt.Errorf("role not found: %w", result.Error)
		}
		return models.Role{}, fmt.Errorf("error getting role by name: %w", result.Error)
	}
	return role, nil
}

// UpdateRole updates an existing role in the database using GORM
func (r *RoleRepositoryImpl) UpdateRole(role models.Role) (models.Role, error) {
	if role.ID == uuid.Nil {
		return models.Role{}, errors.New("cannot update role without ID")
	}
	result := r.db.Save(&role)
	if result.Error != nil {
		return models.Role{}, fmt.Errorf("failed to update role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return models.Role{}, gorm.ErrRecordNotFound
	}
	return role, nil
}

// DeleteRole deletes a role from the database using GORM
func (r *RoleRepositoryImpl) DeleteRole(id uuid.UUID) error {
	result := r.db.Delete(&models.Role{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete role: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("role not found for deletion")
	}
	return nil
}

// AssignRoleToUser assigns a role to a user
func (r *RoleRepositoryImpl) AssignRoleToUser(userID, roleID uuid.UUID) error {
	association := map[string]interface{}{"user_id": userID, "role_id": roleID}
	result := r.db.Table("user_roles").Create(association)
	if result.Error != nil {
		// Cek duplikasi secara spesifik jika diperlukan
		return fmt.Errorf("failed to assign role to user: %w", result.Error)
	}
	return nil
}

// RemoveRoleFromUser removes a role from a user
func (r *RoleRepositoryImpl) RemoveRoleFromUser(userID, roleID uuid.UUID) error {
	result := r.db.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID)
	if result.Error != nil {
		return fmt.Errorf("failed to remove role from user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("user role assignment not found")
	}
	return nil
}

// GetRolesForUser retrieves all roles assigned to a user from the database using GORM
func (r *RoleRepositoryImpl) GetRolesForUser(userID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	// GORM's Joins clause untuk mengambil data dari tabel 'roles'
	// berdasarkan 'user_id' di tabel junction 'user_roles'.
	err := r.db.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error

	if err != nil {
		return nil, fmt.Errorf("error getting roles for user: %w", err)
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found for user %s", userID.String())
	}
	return roles, nil
}
