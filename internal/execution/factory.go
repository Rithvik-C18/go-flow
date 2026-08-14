package execution

import (
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/workflow"
)

func IsValidNodeType(nodeType string) bool {
	switch nodeType {
	case "http", "if":
		return true
	default:
		return false
	}
}

func buildNode(def *workflow.NodeDefinition) (Node, error) {
	switch def.Type {
	case "http":
		return NewHttpNode(def.Config)
	case "if":
		return NewIfNode(def.Config)
	default:
		return nil, fmt.Errorf("Unknown node type")
	}
}
