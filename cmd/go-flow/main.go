package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"

	"github.com/Rithvik-C18/go-flow/internal/config"
	"github.com/Rithvik-C18/go-flow/internal/database"
	"github.com/Rithvik-C18/go-flow/internal/execution"
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

	authService := service.NewAuthService(userRepository, cfg)
	workflowService := service.NewWorkflowService(workflowRepository, execution.NewExecutor())

	// Set config for execution factory
	execution.SetConfig(cfg)
	// Set parallel execution mode
	executor := execution.NewExecutor()
	executor.SetParallel(cfg.Execution.Parallel)
	workflowService.SetExecutor(executor)

	authHandler := handler.NewAuthHandler(authService)
	workflowHandler := handler.NewWorkflowHandler(workflowService)

	authMiddleware := middleware.Auth(authService)

	r := router.NewRouter(authHandler, workflowHandler, authMiddleware)
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	r.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server listening on %s", addr)
	server := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
