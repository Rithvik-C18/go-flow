package main

import (
	"fmt"

	ex "github.com/Rithvik-C18/go-flow/internal/execution"
	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
	// "github.com/Rithvik-C18/go-flow/internal/config"
	// "github.com/gin-gonic/gin"
	// "gorm.io/driver/postgres"
	// "gorm.io/gorm"
)

func main() {
	w := wf.NewWorkflow("wf-1", "test Workflow")

	node1, err := wf.NewNodeDefinition("node-1", "http", map[string]any{
		"url":        "https://jsonplaceholder.typicode.com/todos/1",
		"httpMethod": "GET",
	})
	if err != nil {
		fmt.Println("error creating model:", err)
		return
	}

	node2, err := wf.NewNodeDefinition("node-2", "http", map[string]any{
		"url":        "https://jsonplaceholder.typicode.com/todos/2",
		"httpMethod": "GET",
	})
	if err != nil {
		fmt.Println("error creating node2:", err)
		return
	}

	if err := w.AddNode(node1); err != nil {
		fmt.Println("error adding node1:", err)
		return
	}
	if err := w.AddNode(node2); err != nil {
		fmt.Println("error adding node2:", err)
		return
	}

	if err := w.AddEdge("node-1", "node-2"); err != nil {
		fmt.Println("error adding edge:", err)
		return
	}

	exec := ex.NewExecutor()
	if err := exec.Run(w); err != nil {
		fmt.Println("execution failed:", err)
		return
	}

	fmt.Println("workflow completed successfully")

	// cfg := config.LoadConfig()
	// dsn := cfg.DB.ConnString
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	fmt.Println("Your databse is not connected")
	// }

	// router := gin.Default()

	// router.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })
	// fmt.Print()
	// router.Run("localhost:8888")
}

// Create a workflow using NewWorkflow()
// Create nodes using NewNodeDefinition()
// Add nodes and edges to the workflow
// Create executor using NewExecutor()
// Call Run(workflow)
