package main

import (
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/server"
)

func main() {
	router := server.NewRouter(server.NewStore())

	fmt.Println("server listening on :8080")
	router.Run(":8080")
}
