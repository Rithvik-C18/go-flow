package execution

import (
	"fmt"

	"github.com/Rithvik-C18/go-flow/internal/workflow"
)


func buildNode(def *workflow.NodeDefinition) (Node, error) {
	switch def.Type {
	case "http":
		return NewHttpNode(def.Config)
	default:
		return nil, fmt.Errorf("Unknown node type")
	}
}
