package graph

import (
	"clank/internal/db"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleDBError handles database errors and returns appropriate HTTP responses
func handleDBError(c *gin.Context, err error) {
	if !db.IsAvailable() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service is currently unavailable",
			"code":  "DATABASE_UNAVAILABLE",
		})
		return
	}

	// Handle specific Neo4j errors
	switch err.Error() {
	case "Neo.ClientError.Statement.TypeError":
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid data type in request",
			"code":  "INVALID_DATA_TYPE",
		})
	case "Neo.ClientError.Statement.ParameterMissing":
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters",
			"code":  "MISSING_PARAMETERS",
		})
	case "Neo.ClientError.Statement.SyntaxError":
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database query syntax error",
			"code":  "QUERY_SYNTAX_ERROR",
		})
	case "Neo.TransientError.Transaction.LockClientStopped":
		c.JSON(http.StatusConflict, gin.H{
			"error": "Transaction lock acquisition failed",
			"code":  "TRANSACTION_LOCK_ERROR",
		})
	default:
		// Log unexpected errors for debugging
		fmt.Printf("Unexpected database error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "An unexpected database error occurred",
			"code":  "DB_ERROR",
		})
	}
}
