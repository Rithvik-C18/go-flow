package router

import (
	"github.com/Rithvik-C18/go-flow/internal/handler"
	"github.com/gin-gonic/gin"
)

func registerProtectedRoutes(r gin.IRouter, workflowHandler *handler.WorkflowHandler, authHandler *handler.AuthHandler, authMiddleware gin.HandlerFunc) {
	r.POST("/logout", authMiddleware, authHandler.Logout)

	workflows := r.Group("/workflows", authMiddleware)
	{
		workflows.GET("", workflowHandler.ListWorkflows)
		workflows.POST("", workflowHandler.CreateWorkflow)

		workflows.GET("/:id", workflowHandler.GetWorkflow)
		workflows.POST("/:id/run", workflowHandler.RunWorkflow)
		workflows.DELETE("/:id", workflowHandler.DeleteWorkflow)

		workflows.GET("/:id/nodes", workflowHandler.ListNodes)
		workflows.POST("/:id/nodes", workflowHandler.AddNode)
		workflows.PUT("/:id/nodes/:nodeId", workflowHandler.UpdateNode)
		workflows.POST("/:id/nodes/:nodeId/run", workflowHandler.RunNode)
		workflows.DELETE("/:id/nodes/:nodeId", workflowHandler.DeleteNode)

		workflows.POST("/:id/edges", workflowHandler.AddEdge)
		workflows.DELETE("/:id/edges", workflowHandler.DeleteEdge)
	}
}
