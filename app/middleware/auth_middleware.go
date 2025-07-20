package middleware

import (
	"net/http"
	"sso-service/app/services"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies JWT token from Authorization header
func AuthMiddleware(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"data":    nil,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // No "Bearer " prefix found
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"data":    nil,
				"message": "Bearer token required",
			})
			c.Abort()
			return
		}

		claims, err := userService.VerifyAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"data":    nil,
				"message": "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Set user ID and claims in context for subsequent handlers
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)
		c.Set("claims", claims) // Set full claims object for service layer authorization

		c.Next() // Proceed to the next handler (the actual route handler)
	}
}
