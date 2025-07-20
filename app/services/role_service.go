package services

import (
	"errors"
	"fmt"
	"sso-service/app/models"
	"sso-service/app/repositories"

	"github.com/google/uuid"
)

// RoleService defines the interface for role-related services
type RoleService interface {
	CreateRole(claims *JWTClaims, req models.RoleCreateRequest) (models.Role, error)
	GetAllRoles(claims *JWTClaims) ([]models.Role, error)
	GetRoleByID(claims *JWTClaims, roleID uuid.UUID) (models.Role, error)
	UpdateRole(claims *JWTClaims, roleID uuid.UUID, req models.RoleUpdateRequest) (models.Role, error)
	DeleteRole(claims *JWTClaims, roleID uuid.UUID) error
	AssignRoleToUser(claims *JWTClaims, req models.AssignRoleToUserRequest) error
	RemoveRoleFromUser(claims *JWTClaims, req models.AssignRoleToUserRequest) error
}

// RoleServiceImpl is the implementation of RoleService
type RoleServiceImpl struct {
	roleRepository repositories.RoleRepository
	userRepository repositories.UserRepository // To check if user exists
	userService    UserService                 // To use authorization helpers
}

// NewRoleService creates a new instance of RoleServiceImpl
func NewRoleService(roleRepository repositories.RoleRepository, userRepository repositories.UserRepository, userService UserService) RoleService {
	return &RoleServiceImpl{
		roleRepository: roleRepository,
		userRepository: userRepository,
		userService:    userService,
	}
}

// CreateRole creates a new role. Requires 'role:create' permission.
func (s *RoleServiceImpl) CreateRole(claims *JWTClaims, req models.RoleCreateRequest) (models.Role, error) {
	if !s.userService.HasPermission(claims, "role:create") {
		return models.Role{}, errors.New("forbidden: insufficient permissions")
	}

	// Admin specific restriction: cannot create superadmin role
	if s.userService.HasRole(claims, models.RoleAdmin) && req.Name == models.RoleSuperAdmin {
		return models.Role{}, errors.New("forbidden: admin cannot create superadmin role")
	}

	// Check if role name already exists
	_, err := s.roleRepository.GetRoleByName(req.Name)
	if err == nil {
		return models.Role{}, errors.New("role name already exists")
	}

	newRole := models.Role{
		Name: req.Name,
	}
	if req.Description != nil {
		newRole.Description = *req.Description
	}

	createdRole, err := s.roleRepository.CreateRole(newRole)
	if err != nil {
		return models.Role{}, fmt.Errorf("failed to create role: %w", err)
	}
	return createdRole, nil
}

// GetAllRoles retrieves all roles. Requires 'role:read_all' permission.
func (s *RoleServiceImpl) GetAllRoles(claims *JWTClaims) ([]models.Role, error) {
	if !s.userService.HasPermission(claims, "role:read_all") {
		return nil, errors.New("forbidden: insufficient permissions")
	}

	roles, err := s.roleRepository.GetAllRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	return roles, nil
}

// GetRoleByID retrieves a role by ID. Requires 'role:read_all' permission.
func (s *RoleServiceImpl) GetRoleByID(claims *JWTClaims, roleID uuid.UUID) (models.Role, error) {
	if !s.userService.HasPermission(claims, "role:read_all") {
		return models.Role{}, errors.New("forbidden: insufficient permissions")
	}

	role, err := s.roleRepository.GetRoleByID(roleID)
	if err != nil {
		return models.Role{}, fmt.Errorf("failed to get role by ID: %w", err)
	}
	return role, nil
}

// UpdateRole updates an existing role. Requires 'role:update' permission.
// Superadmin can update anything. Admin cannot update superadmin role.
func (s *RoleServiceImpl) UpdateRole(claims *JWTClaims, roleID uuid.UUID, req models.RoleUpdateRequest) (models.Role, error) {
	isSuperAdmin := s.userService.HasRole(claims, models.RoleSuperAdmin)

	if !isSuperAdmin && !s.userService.HasPermission(claims, "role:update") {
		return models.Role{}, errors.New("forbidden: insufficient permissions")
	}

	existingRole, err := s.roleRepository.GetRoleByID(roleID)
	if err != nil {
		return models.Role{}, errors.New("role not found")
	}

	// Admin specific restriction: cannot update superadmin role
	if s.userService.HasRole(claims, models.RoleAdmin) && existingRole.Name == models.RoleSuperAdmin {
		return models.Role{}, errors.New("forbidden: admin cannot update superadmin role")
	}

	if req.Name != nil {
		// Check for name uniqueness if changed
		if *req.Name != existingRole.Name {
			_, err := s.roleRepository.GetRoleByName(*req.Name)
			if err == nil {
				return models.Role{}, errors.New("role name already exists")
			}
		}
		existingRole.Name = *req.Name
	}
	if req.Description != nil {
		existingRole.Description = *req.Description
	}

	updatedRole, err := s.roleRepository.UpdateRole(existingRole)
	if err != nil {
		return models.Role{}, fmt.Errorf("failed to update role: %w", err)
	}
	return updatedRole, nil
}

// DeleteRole deletes a role. Requires 'role:delete' permission.
// Superadmin can delete anything. Admin cannot delete superadmin role.
func (s *RoleServiceImpl) DeleteRole(claims *JWTClaims, roleID uuid.UUID) error {
	isSuperAdmin := s.userService.HasRole(claims, models.RoleSuperAdmin)

	if !isSuperAdmin && !s.userService.HasPermission(claims, "role:delete") {
		return errors.New("forbidden: insufficient permissions")
	}

	existingRole, err := s.roleRepository.GetRoleByID(roleID)
	if err != nil {
		return errors.New("role not found")
	}

	// Admin specific restriction: cannot delete superadmin role
	if s.userService.HasRole(claims, models.RoleAdmin) && existingRole.Name == models.RoleSuperAdmin {
		return errors.New("forbidden: admin cannot delete superadmin role")
	}

	err = s.roleRepository.DeleteRole(roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// AssignRoleToUser assigns a role to a user. Requires 'role:assign' permission.
// Admin cannot assign superadmin role.
func (s *RoleServiceImpl) AssignRoleToUser(claims *JWTClaims, req models.AssignRoleToUserRequest) error {
	isSuperAdmin := s.userService.HasRole(claims, models.RoleSuperAdmin)

	if !isSuperAdmin && !s.userService.HasPermission(claims, "role:assign") {
		return errors.New("forbidden: insufficient permissions")
	}

	// Check if user exists
	_, err := s.userRepository.GetUserByID(req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if role exists
	roleToAssign, err := s.roleRepository.GetRoleByID(req.RoleID)
	if err != nil {
		return errors.New("role not found")
	}

	// Admin specific restriction: cannot assign superadmin role
	if s.userService.HasRole(claims, models.RoleAdmin) && roleToAssign.Name == models.RoleSuperAdmin {
		return errors.New("forbidden: admin cannot assign superadmin role")
	}

	// Superadmin cannot assign superadmin role to themselves (prevent locking out) - optional
	if isSuperAdmin && claims.UserID == req.UserID.String() && roleToAssign.Name == models.RoleSuperAdmin {
		// This is a complex edge case, might need more thought.
		// For now, allow superadmin to assign superadmin role to others.
	}

	err = s.roleRepository.AssignRoleToUser(req.UserID, req.RoleID)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

// RemoveRoleFromUser removes a role from a user. Requires 'role:assign' permission.
// Admin cannot remove superadmin role from a user.
func (s *RoleServiceImpl) RemoveRoleFromUser(claims *JWTClaims, req models.AssignRoleToUserRequest) error {
	isSuperAdmin := s.userService.HasRole(claims, models.RoleSuperAdmin)

	if !isSuperAdmin && !s.userService.HasPermission(claims, "role:assign") {
		return errors.New("forbidden: insufficient permissions")
	}

	// Check if user exists
	_, err := s.userRepository.GetUserByID(req.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if role exists
	roleToRemove, err := s.roleRepository.GetRoleByID(req.RoleID)
	if err != nil {
		return errors.New("role not found")
	}

	// Admin specific restriction: cannot remove superadmin role
	if s.userService.HasRole(claims, models.RoleAdmin) && roleToRemove.Name == models.RoleSuperAdmin {
		return errors.New("forbidden: admin cannot remove superadmin role")
	}

	err = s.roleRepository.RemoveRoleFromUser(req.UserID, req.RoleID)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}
	return nil
}
