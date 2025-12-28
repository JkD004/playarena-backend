// api/middleware.go
package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/JkD004/playarena-backend/user"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware creates a "guard" for your routes
func AuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the "Authorization" header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// 2. Extract the Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			return
		}

		// 3. Prepare claims
		claims := &user.Claims{}

		// --- SECURITY FIX: Read JWT Secret from ENV ---
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server config error"})
			return
		}
		// ------------------------------------------------

		// 4. Parse the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 5. Check role permissions
		isAllowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You do not have permission"})
			return
		}

		// 6. Success — attach user data to context
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)

		// Continue to next handler
		c.Next()
	}
}

// 👇 ADD THIS NEW FUNCTION 👇
// MaintenanceMiddleware blocks all requests when MAINTENANCE_MODE is "true"
func MaintenanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check Env Variable
		if os.Getenv("MAINTENANCE_MODE") == "true" {
			
			// 2. EXCEPTION: Always allow Health Check (so UptimeRobot doesn't panic)
			if c.Request.URL.Path == "/api/v1/health" {
				c.Next()
				return
			}

			// 3. EXCEPTION: Optional - Allow Admin Login if needed
			// if strings.Contains(c.Request.URL.Path, "/login") || strings.Contains(c.Request.URL.Path, "/admin") {
			// 	c.Next()
			// 	return
			// }

			// 4. Reject everything else with 503
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "System is under maintenance. Please try again later.",
				"code":  503,
			})
			return
		}

		c.Next()
	}
}
