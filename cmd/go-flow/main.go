package main

import (
	"fmt"
	"log"

	"github.com/Rithvik-C18/go-flow/internal/config"
	"github.com/Rithvik-C18/go-flow/internal/database"
	ex "github.com/Rithvik-C18/go-flow/internal/execution"
	"github.com/Rithvik-C18/go-flow/internal/handler"
	"github.com/Rithvik-C18/go-flow/internal/middleware"
	"github.com/Rithvik-C18/go-flow/internal/repository"
	"github.com/Rithvik-C18/go-flow/internal/router"
	"github.com/Rithvik-C18/go-flow/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.InitDatabase(cfg.DB.ConnString)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	userRepository := repository.NewUserRepository(db)
	workflowRepository := repository.NewWorkflowRepository(db)

	authService := service.NewAuthService(userRepository, cfg.JWT.Secret)
	workflowService := service.NewWorkflowService(workflowRepository, ex.NewExecutor())

	authHandler := handler.NewAuthHandler(authService)
	workflowHandler := handler.NewWorkflowHandler(workflowService)

	authMiddleware := middleware.Auth(authService)

	r := router.NewRouter(authHandler, workflowHandler, authMiddleware)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
