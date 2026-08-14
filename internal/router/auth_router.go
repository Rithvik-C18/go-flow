package router

import (
	"github.com/Rithvik-C18/go-flow/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(r *gin.Engine, authHandler *handler.AuthHandler) {
	r.POST("/login", authHandler.Login)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}
}
