package middleware

import (
	"clank/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireDatabase middleware checks if the database is available before proceeding
func RequireDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip database check for non-database endpoints
		if isNonDatabaseEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		if !db.IsAvailable() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database service is currently unavailable",
				"code":  "DATABASE_UNAVAILABLE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isNonDatabaseEndpoint returns true if the endpoint doesn't require database access
func isNonDatabaseEndpoint(path string) bool {
	// List of endpoints that don't require database access
	nonDBEndpoints := []string{
		"/",
		"/health",
		"/ws",
		"/mcp/sse",
	}

	for _, endpoint := range nonDBEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}
