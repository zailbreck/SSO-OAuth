package services

import (
	"errors"
	"fmt"
	"sso-service/app/models"
	"sso-service/app/repositories" // Import the new repository package
	"time"

	"github.com/golang-jwt/jwt/v5" // For JWT handling
	"github.com/google/uuid"       // For UUID handling
)

// JWTClaims defines the claims structure for JWT
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// TokenResponse represents the structure for returning tokens
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
}

// UserService defines the interface for user-related services including authentication, authorization, and CRUD
type UserService interface {
	AuthenticateUser(username, password string) (*TokenResponse, error)
	GenerateTokens(user models.User, roles []models.Role, permissions []string) (*TokenResponse, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
	InvalidateToken(tokenString string) error
	GetUserProfile(userID uuid.UUID) (models.User, error)
	VerifyAccessToken(tokenString string) (*JWTClaims, error)

	// User Management CRUD Operations
	CreateUser(claims *JWTClaims, req models.UserCreateRequest) (models.User, error)                   // Pass claims for authorization
	GetAllUsers(claims *JWTClaims) ([]models.User, error)                                              // Pass claims for authorization
	GetUserByID(claims *JWTClaims, userID uuid.UUID) (models.User, error)                              // Pass claims for authorization
	UpdateUser(claims *JWTClaims, userID uuid.UUID, req models.UserUpdateRequest) (models.User, error) // Pass claims for authorization
	DeleteUser(claims *JWTClaims, userID uuid.UUID) error                                              // Pass claims for authorization

	// Authorization Helpers
	HasPermission(claims *JWTClaims, requiredPermission string) bool
	HasRole(claims *JWTClaims, roleName string) bool
	HasRoleForUser(userID uuid.UUID, roleName string) bool
}

// UserServiceImpl is the implementation of UserService
type UserServiceImpl struct {
	jwtSecret      string
	userRepository repositories.UserRepository // Dependency on the new repository interface
}

// NewUserService creates a new instance of UserServiceImpl
func NewUserService(jwtSecret string, userRepository repositories.UserRepository) UserService {
	return &UserServiceImpl{
		jwtSecret:      jwtSecret,
		userRepository: userRepository,
	}
}

// AuthenticateUser authenticates a user and generates JWT tokens
func (s *UserServiceImpl) AuthenticateUser(username, password string) (*TokenResponse, error) {
	user, err := s.userRepository.GetUserByUsername(username) // Use repository
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	if !models.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// Fetch user roles and permissions for JWT claims
	roles, err := s.userRepository.GetRolesForUser(user.ID) // Use repository
	if err != nil {
		// Log this, but don't necessarily fail authentication if user has no roles
		fmt.Printf("Warning: User %s has no roles assigned: %v\n", user.Username, err)
	}
	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	permissions, err := s.userRepository.GetPermissionsForUser(user.ID) // Use repository
	if err != nil {
		fmt.Printf("Warning: User %s has no permissions assigned: %v\n", user.Username, err)
	}

	return s.GenerateTokens(user, roles, permissions)
}

// GenerateTokens generates new access and refresh JWT tokens
func (s *UserServiceImpl) GenerateTokens(user models.User, roles []models.Role, permissions []string) (*TokenResponse, error) {
	// Generate a unique JWT ID (JTI) for the access token
	jtiAccessToken := uuid.New().String()
	// Generate a unique JWT ID (JTI) for the refresh token
	jtiRefreshToken := uuid.New().String()

	// Access Token (short-lived)
	accessTokenExp := time.Now().Add(15 * time.Minute) // 15 minutes expiry
	var roleNames []string
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	accessClaims := &JWTClaims{
		UserID:      user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		Roles:       roleNames,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "sso-service",
			Subject:   user.ID.String(),
			Audience:  []string{"sso-client"}, // Audience for the token
			ID:        jtiAccessToken,         // Set JTI for access token
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token (long-lived)
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour) // 7 days expiry
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "sso-service",
		Subject:   user.ID.String(),
		Audience:  []string{"sso-client"},
		ID:        jtiRefreshToken, // Set JTI for refresh token
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// In a real application, you would store the refresh token in a database
	// associated with the user, and potentially invalidate old refresh tokens.
	// For this example, we are not persisting refresh tokens, but their JTI is generated.

	return &TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenExp.Unix() - time.Now().Unix(),
	}, nil
}

// RefreshToken uses a refresh token to issue new access and refresh tokens
func (s *UserServiceImpl) RefreshToken(refreshTokenString string) (*TokenResponse, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid refresh token claims")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return nil, errors.New("refresh token JTI not found")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("user ID not found in refresh token claims")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format in refresh token")
	}

	user, err := s.userRepository.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found for refresh token")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	roles, err := s.userRepository.GetRolesForUser(user.ID)
	if err != nil {
		fmt.Printf("Warning: User %s has no roles assigned during refresh: %v\n", user.Username, err)
	}
	permissions, err := s.userRepository.GetPermissionsForUser(user.ID)
	if err != nil {
		fmt.Printf("Warning: User %s has no permissions assigned: %v\n", user.Username, err)
	}

	return s.GenerateTokens(user, roles, permissions)
}

// InvalidateToken invalidates a given token (e.g., on logout) by blacklisting its JTI
func (s *UserServiceImpl) InvalidateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		fmt.Printf("Attempted to invalidate an already invalid/expired token: %v\n", err)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if jti, jtiOk := claims["jti"].(string); jtiOk && jti != "" {
				if exp, expOk := claims["exp"].(float64); expOk {
					expiresAt := time.Unix(int64(exp), 0)
					return s.userRepository.AddRevokedToken(jti, expiresAt)
				}
			}
		}
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims for invalidation")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return errors.New("token JTI not found for invalidation")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("token expiration time not found for invalidation")
	}
	expiresAt := time.Unix(int64(exp), 0)

	err = s.userRepository.AddRevokedToken(jti, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	fmt.Printf("Token JTI '%s' blacklisted successfully.\n", jti)
	return nil
}

// GetUserProfile retrieves user profile by ID
func (s *UserServiceImpl) GetUserProfile(userID uuid.UUID) (models.User, error) {
	user, err := s.userRepository.GetUserByID(userID) // Corrected: Calls repository
	if err != nil {
		return models.User{}, err
	}
	user.PasswordHash = "" // Remove sensitive info
	return user, nil
}

// VerifyAccessToken verifies the validity of an access token and checks the blacklist
func (s *UserServiceImpl) VerifyAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims or token not valid")
	}

	// Check if token is blacklisted
	isRevoked, err := s.userRepository.IsTokenRevoked(claims.ID) // claims.ID is the JTI
	if err != nil {
		return nil, fmt.Errorf("error checking token blacklist: %w", err)
	}
	if isRevoked {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// --- User Management CRUD Operations ---

// CreateUser creates a new user. Requires 'user:create' permission.
// Superadmin can create any user. Admin cannot create superadmin.
func (s *UserServiceImpl) CreateUser(claims *JWTClaims, req models.UserCreateRequest) (models.User, error) {
	// Authorization check
	if !s.HasPermission(claims, "user:create") {
		return models.User{}, errors.New("forbidden: insufficient permissions")
	}

	// Admin specific restriction: cannot create superadmin
	if s.HasRole(claims, "admin") {
		// This check would require knowing the role of the user being created.
		// For simplicity, we'll assume admin cannot create a user with 'superadmin' role.
		// This logic is better placed in a separate AssignRole service method.
		// For now, if an admin tries to create a user, they cannot assign 'superadmin' role.
		// A more robust check would involve checking the role_id being assigned.
		// If you want to prevent an admin from creating a user that *later* gets assigned superadmin,
		// you'd need to check the assigned role during role assignment.
	}

	// Check if username or email already exists
	_, err := s.userRepository.GetUserByUsername(req.Username)
	if err == nil {
		return models.User{}, errors.New("username already exists")
	}
	// Check for email uniqueness
	// This would require a GetUserByEmail method in the repository
	// For now, relying on DB unique constraint if it exists.

	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		IsActive:     true, // Default to active on creation
	}
	if req.IsActive != nil { // Allow overriding is_active if provided
		newUser.IsActive = *req.IsActive
	}

	createdUser, err := s.userRepository.CreateUser(newUser)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	createdUser.PasswordHash = "" // Clear hash before returning
	return createdUser, nil
}

// GetAllUsers retrieves all users. Requires 'user:read_all' permission.
func (s *UserServiceImpl) GetAllUsers(claims *JWTClaims) ([]models.User, error) {
	// Authorization check
	if !s.HasPermission(claims, "user:read_all") {
		return nil, errors.New("forbidden: insufficient permissions")
	}

	users, err := s.userRepository.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	// Remove password hashes for all users before returning
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

// GetUserByID retrieves a user by ID. Requires 'user:read_all' or 'user:read_own' if it's their own profile.
func (s *UserServiceImpl) GetUserByID(claims *JWTClaims, userID uuid.UUID) (models.User, error) {
	// Check if the user is requesting their own profile
	isOwnProfile := claims.UserID == userID.String()

	// Authorization check
	if !s.HasPermission(claims, "user:read_all") {
		if !isOwnProfile || !s.HasPermission(claims, "user:read_own") {
			return models.User{}, errors.New("forbidden: insufficient permissions to read this user's profile")
		}
	}

	user, err := s.userRepository.GetUserByID(userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user by ID: %w", err)
	}
	user.PasswordHash = "" // Remove sensitive info
	return user, nil
}

// UpdateUser updates an existing user. Requires 'user:update_all' or 'user:update_own'.
// Superadmin can update anything. Admin cannot update superadmin. Consumer can only update own profile.
func (s *UserServiceImpl) UpdateUser(claims *JWTClaims, userID uuid.UUID, req models.UserUpdateRequest) (models.User, error) {
	isOwnProfile := claims.UserID == userID.String()
	isSuperAdmin := s.HasRole(claims, "superadmin")

	// Fetch existing user to check roles and current data
	existingUser, err := s.userRepository.GetUserByID(userID)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	// Authorization check
	if !isSuperAdmin { // Superadmin bypasses all checks
		if !s.HasPermission(claims, "user:update_all") {
			if !isOwnProfile || !s.HasPermission(claims, "user:update_own") {
				return models.User{}, errors.New("forbidden: insufficient permissions to update this user's profile")
			}
		}

		// Admin specific restriction: cannot update superadmin
		if s.HasRole(claims, "admin") && s.HasRoleForUser(existingUser.ID, "superadmin") {
			return models.User{}, errors.New("forbidden: admin cannot update superadmin user")
		}

		// Consumer specific restriction: can only update own profile and specific fields
		if s.HasRole(claims, "consumer") && !isOwnProfile {
			return models.User{}, errors.New("forbidden: consumer can only update their own profile")
		}
		// If consumer is updating own profile, restrict fields they can change
		if s.HasRole(claims, "consumer") && isOwnProfile {
			// Ensure consumer cannot change isActive or roles
			if req.IsActive != nil || req.Username != nil || req.Email != nil {
				// Only allow password change for consumer on their own profile
				if req.Password == nil || (req.Username != nil || req.Email != nil || req.IsActive != nil) {
					return models.User{}, errors.New("forbidden: consumers can only update their own password")
				}
			}
		}
	}

	// Apply updates
	if req.Username != nil {
		existingUser.Username = *req.Username
	}
	if req.Email != nil {
		existingUser.Email = *req.Email
	}
	if req.Password != nil {
		hashedPassword, err := models.HashPassword(*req.Password)
		if err != nil {
			return models.User{}, fmt.Errorf("failed to hash new password: %w", err)
		}
		existingUser.PasswordHash = hashedPassword
	}
	if req.IsActive != nil {
		existingUser.IsActive = *req.IsActive
	}

	updatedUser, err := s.userRepository.UpdateUser(userID, existingUser)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	updatedUser.PasswordHash = "" // Clear hash
	return updatedUser, nil
}

// DeleteUser deletes a user. Requires 'user:delete' permission.
// Superadmin can delete any user. Admin cannot delete superadmin.
func (s *UserServiceImpl) DeleteUser(claims *JWTClaims, userID uuid.UUID) error {
	isSuperAdmin := s.HasRole(claims, "superadmin")

	// Authorization check
	if !isSuperAdmin { // Superadmin bypasses all checks
		if !s.HasPermission(claims, "user:delete") {
			return errors.New("forbidden: insufficient permissions to delete users")
		}

		// Admin specific restriction: cannot delete superadmin
		if s.HasRole(claims, "admin") && s.HasRoleForUser(userID, "superadmin") {
			return errors.New("forbidden: admin cannot delete superadmin user")
		}
	}

	// Prevent user from deleting themselves (optional, but good practice)
	if claims.UserID == userID.String() {
		return errors.New("forbidden: cannot delete your own account via this endpoint")
	}

	err := s.userRepository.DeleteUser(userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// HasPermission checks if the claims contain the required permission
func (s *UserServiceImpl) HasPermission(claims *JWTClaims, requiredPermission string) bool {
	if claims == nil {
		return false
	}
	// Superadmin always has all permissions
	if s.HasRole(claims, "superadmin") {
		return true
	}

	for _, p := range claims.Permissions {
		if p == requiredPermission {
			return true
		}
	}
	return false
}

// HasRole checks if the claims contain the specified role
func (s *UserServiceImpl) HasRole(claims *JWTClaims, roleName string) bool {
	if claims == nil {
		return false
	}
	for _, r := range claims.Roles {
		if r == roleName {
			return true
		}
	}
	return false
}

// HasRoleForUser checks if a specific user (by ID) has a given role.
// This requires a DB lookup, so it's a separate helper.
func (s *UserServiceImpl) HasRoleForUser(userID uuid.UUID, roleName string) bool {
	roles, err := s.userRepository.GetRolesForUser(userID)
	if err != nil {
		return false // User has no roles or error fetching
	}
	for _, role := range roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}
