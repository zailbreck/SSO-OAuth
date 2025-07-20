// PermissionMiddleware checks if the authenticated user has the required permission(s)
package middleware

import (
	"net/http"
	"sso-service/app/services"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware(userService services.UserService, requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{ // Forbidden, as authentication passed but claims missing
				"status":  http.StatusForbidden,
				"data":    nil,
				"message": "User claims not found in context for permission check",
			})
			c.Abort()
			return
		}
		jwtClaims := claims.(*services.JWTClaims)

		// Check if user has any of the required permissions
		hasAnyPermission := false
		for _, reqPerm := range requiredPermissions {
			if userService.HasPermission(jwtClaims, reqPerm) {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"data":    nil,
				"message": "Forbidden: Insufficient permissions",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
