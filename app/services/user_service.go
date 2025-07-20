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

// UserService defines the interface for user-related services including authentication and authorization
type UserService interface {
	AuthenticateUser(username, password string) (*TokenResponse, error)
	GenerateTokens(user models.User, roles []models.Role, permissions []string) (*TokenResponse, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
	InvalidateToken(tokenString string) error // For logout
	GetUserProfile(userID uuid.UUID) (models.User, error)
	VerifyAccessToken(tokenString string) (*JWTClaims, error)
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

	// Check if refresh token is blacklisted (if you implement refresh token blacklisting)
	// For this example, we are only blacklisting access tokens.
	// isRevoked, err := s.userRepository.IsTokenRevoked(jti)
	// if err != nil || isRevoked {
	// 	return nil, errors.New("refresh token has been revoked")
	// }

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("user ID not found in refresh token claims")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format in refresh token")
	}

	// In a real application, you would verify the refresh token against your database
	// to ensure it hasn't been revoked/invalidated.

	user, err := s.userRepository.GetUserByID(userID) // Use repository
	if err != nil {
		return nil, errors.New("user not found for refresh token")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Fetch user roles and permissions for new JWT claims
	roles, err := s.userRepository.GetRolesForUser(user.ID) // Use repository
	if err != nil {
		fmt.Printf("Warning: User %s has no roles assigned during refresh: %v\n", user.Username, err)
	}
	permissions, err := s.userRepository.GetPermissionsForUser(user.ID) // Use repository
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
		// If token is already invalid/expired, we can still attempt to blacklist its JTI if available
		// or just return success as it's already "invalid"
		fmt.Printf("Attempted to invalidate an already invalid/expired token: %v\n", err)
		// Try to extract JTI even from an invalid token if possible
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if jti, jtiOk := claims["jti"].(string); jtiOk && jti != "" {
				// Blacklist it anyway to be safe, especially if it's invalid due to custom reasons
				if exp, expOk := claims["exp"].(float64); expOk {
					expiresAt := time.Unix(int64(exp), 0)
					return s.userRepository.AddRevokedToken(jti, expiresAt)
				}
			}
		}
		return nil // Consider it successfully "invalidated" if it was already invalid
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

	// Add the token's JTI to the blacklist
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
	// Remove sensitive info before returning
	user.PasswordHash = ""
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
