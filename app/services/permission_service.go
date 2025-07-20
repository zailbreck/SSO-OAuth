package services

import (
	"errors"
	"fmt"
	"sso-service/app/models"
	"sso-service/app/repositories"

	"github.com/google/uuid"
)

// PermissionService defines the interface for permission-related services
type PermissionService interface {
	CreatePermission(claims *JWTClaims, req models.PermissionCreateRequest) (models.Permission, error)
	GetAllPermissions(claims *JWTClaims) ([]models.Permission, error)
	GetPermissionByID(claims *JWTClaims, permissionID uuid.UUID) (models.Permission, error)
	UpdatePermission(claims *JWTClaims, permissionID uuid.UUID, req models.PermissionUpdateRequest) (models.Permission, error)
	DeletePermission(claims *JWTClaims, permissionID uuid.UUID) error
	AssignPermissionToRole(claims *JWTClaims, req models.AssignPermissionToRoleRequest) error
	RemovePermissionFromRole(claims *JWTClaims, req models.AssignPermissionToRoleRequest) error
}

// PermissionServiceImpl is the implementation of PermissionService
type PermissionServiceImpl struct {
	permissionRepository repositories.PermissionRepository
	roleRepository       repositories.RoleRepository // To check if role exists
	siteRepository       repositories.SiteRepository // To check if site exists
	userService          UserService                 // To use authorization helpers
}

// NewPermissionService creates a new instance of PermissionServiceImpl
func NewPermissionService(
	permissionRepository repositories.PermissionRepository,
	roleRepository repositories.RoleRepository,
	siteRepository repositories.SiteRepository,
	userService UserService,
) PermissionService {
	return &PermissionServiceImpl{
		permissionRepository: permissionRepository,
		roleRepository:       roleRepository,
		siteRepository:       siteRepository,
		userService:          userService,
	}
}

// CreatePermission creates a new permission. Requires 'permission:create' permission.
func (s *PermissionServiceImpl) CreatePermission(claims *JWTClaims, req models.PermissionCreateRequest) (models.Permission, error) {
	if !s.userService.HasPermission(claims, "permission:create") {
		return models.Permission{}, errors.New("forbidden: insufficient permissions")
	}

	// Check if site exists
	_, err := s.siteRepository.GetSiteByID(req.SiteID)
	if err != nil {
		return models.Permission{}, errors.New("site not found")
	}

	// Check if parent permission exists if provided
	if req.ParentID != nil {
		_, err := s.permissionRepository.GetPermissionByID(*req.ParentID)
		if err != nil {
			return models.Permission{}, errors.New("parent permission not found")
		}
	}

	// Check if permission name already exists (globally unique)
	_, err = s.permissionRepository.GetPermissionByNameAndSite(req.Name, req.SiteID) // Check uniqueness per site
	if err == nil {
		return models.Permission{}, errors.New("permission name already exists for this site")
	}

	newPermission := models.Permission{
		SiteID:   req.SiteID,
		Name:     req.Name,
		ParentID: req.ParentID,
	}
	if req.Description != nil {
		newPermission.Description = *req.Description
	}

	createdPermission, err := s.permissionRepository.CreatePermission(newPermission)
	if err != nil {
		return models.Permission{}, fmt.Errorf("failed to create permission: %w", err)
	}
	return createdPermission, nil
}

// GetAllPermissions retrieves all permissions. Requires 'permission:read_all' permission.
func (s *PermissionServiceImpl) GetAllPermissions(claims *JWTClaims) ([]models.Permission, error) {
	if !s.userService.HasPermission(claims, "permission:read_all") {
		return nil, errors.New("forbidden: insufficient permissions")
	}

	permissions, err := s.permissionRepository.GetAllPermissions()
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions: %w", err)
	}
	return permissions, nil
}

// GetPermissionByID retrieves a permission by ID. Requires 'permission:read_all' permission.
func (s *PermissionServiceImpl) GetPermissionByID(claims *JWTClaims, permissionID uuid.UUID) (models.Permission, error) {
	if !s.userService.HasPermission(claims, "permission:read_all") {
		return models.Permission{}, errors.New("forbidden: insufficient permissions")
	}

	permission, err := s.permissionRepository.GetPermissionByID(permissionID)
	if err != nil {
		return models.Permission{}, fmt.Errorf("failed to get permission by ID: %w", err)
	}
	return permission, nil
}

// UpdatePermission updates an existing permission. Requires 'permission:update' permission.
func (s *PermissionServiceImpl) UpdatePermission(claims *JWTClaims, permissionID uuid.UUID, req models.PermissionUpdateRequest) (models.Permission, error) {
	if !s.userService.HasPermission(claims, "permission:update") {
		return models.Permission{}, errors.New("forbidden: insufficient permissions")
	}

	existingPermission, err := s.permissionRepository.GetPermissionByID(permissionID)
	if err != nil {
		return models.Permission{}, errors.New("permission not found")
	}

	if req.SiteID != nil {
		// Check if new site exists
		_, err := s.siteRepository.GetSiteByID(*req.SiteID)
		if err != nil {
			return models.Permission{}, errors.New("new site not found")
		}
		existingPermission.SiteID = *req.SiteID
	}
	if req.Name != nil {
		// Check for name uniqueness if changed and site ID is provided
		if *req.Name != existingPermission.Name || (req.SiteID != nil && *req.SiteID != existingPermission.SiteID) {
			targetSiteID := existingPermission.SiteID
			if req.SiteID != nil {
				targetSiteID = *req.SiteID
			}
			_, err := s.permissionRepository.GetPermissionByNameAndSite(*req.Name, targetSiteID)
			if err == nil {
				return models.Permission{}, errors.New("permission name already exists for the target site")
			}
		}
		existingPermission.Name = *req.Name
	}
	if req.Description != nil {
		existingPermission.Description = *req.Description
	}
	if req.ParentID != nil {
		// Check if new parent permission exists
		if *req.ParentID != uuid.Nil { // Check if not setting to NULL
			_, err := s.permissionRepository.GetPermissionByID(*req.ParentID)
			if err != nil {
				return models.Permission{}, errors.New("parent permission not found")
			}
		}
		existingPermission.ParentID = req.ParentID
	} else {
		existingPermission.ParentID = nil // Explicitly set to NULL if nil is passed
	}

	updatedPermission, err := s.permissionRepository.UpdatePermission(existingPermission)
	if err != nil {
		return models.Permission{}, fmt.Errorf("failed to update permission: %w", err)
	}
	return updatedPermission, nil
}

// DeletePermission deletes a permission. Requires 'permission:delete' permission.
func (s *PermissionServiceImpl) DeletePermission(claims *JWTClaims, permissionID uuid.UUID) error {
	if !s.userService.HasPermission(claims, "permission:delete") {
		return errors.New("forbidden: insufficient permissions")
	}

	err := s.permissionRepository.DeletePermission(permissionID)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

// AssignPermissionToRole assigns a permission to a role. Requires 'permission:assign' permission.
func (s *PermissionServiceImpl) AssignPermissionToRole(claims *JWTClaims, req models.AssignPermissionToRoleRequest) error {
	if !s.userService.HasPermission(claims, "permission:assign") {
		return errors.New("forbidden: insufficient permissions")
	}

	// Check if role exists
	_, err := s.roleRepository.GetRoleByID(req.RoleID)
	if err != nil {
		return errors.New("role not found")
	}

	// Check if permission exists
	_, err = s.permissionRepository.GetPermissionByID(req.PermissionID)
	if err != nil {
		return errors.New("permission not found")
	}

	err = s.permissionRepository.AssignPermissionToRole(req.RoleID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("failed to assign permission: %w", err)
	}
	return nil
}

// RemovePermissionFromRole removes a permission from a role. Requires 'permission:assign' permission.
func (s *PermissionServiceImpl) RemovePermissionFromRole(claims *JWTClaims, req models.AssignPermissionToRoleRequest) error {
	if !s.userService.HasPermission(claims, "permission:assign") {
		return errors.New("forbidden: insufficient permissions")
	}

	// Check if role exists
	_, err := s.roleRepository.GetRoleByID(req.RoleID)
	if err != nil {
		return errors.New("role not found")
	}

	// Check if permission exists
	_, err = s.permissionRepository.GetPermissionByID(req.PermissionID)
	if err != nil {
		return errors.New("permission not found")
	}

	err = s.permissionRepository.RemovePermissionFromRole(req.RoleID, req.PermissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}
	return nil
}
