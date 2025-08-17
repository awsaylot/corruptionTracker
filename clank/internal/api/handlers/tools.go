package handlers

import "github.com/gin-gonic/gin"

func ToolHandler(c *gin.Context) {
	c.String(200, "Tool execution stub")
}
