package router

import (
	"net/http"

	"github.com/Rithvik-C18/go-flow/internal/handler"
	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handler.AuthHandler, workflowHandler *handler.WorkflowHandler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	registerPublicRoutes(r, authHandler)
	registerProtectedRoutes(r, workflowHandler, authMiddleware)

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
