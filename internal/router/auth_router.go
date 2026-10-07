package router

import (
	"github.com/Rithvik-C18/go-flow/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(r gin.IRouter, authHandler *handler.AuthHandler) {
	r.POST("/login", authHandler.Login)
	r.POST("/refresh", authHandler.RefreshToken)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}
}
