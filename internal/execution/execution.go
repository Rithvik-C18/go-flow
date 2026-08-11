package execution

import (
	"fmt"

	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
)

type Node interface {
	Execute(*ExecutionContext) error
}

type Executor struct {
}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Run(w *wf.Workflow) error {
	order, err := topologicalSort(w)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	ctx := NewExecutionContext(w.Id, "")

	for _, nodeID := range order {
		nodeDef, ok := w.Nodes[nodeID]
		if !ok {
			return fmt.Errorf("node %s not found in workflow", nodeID)
		}

		node, err := buildNode(nodeDef)
		if err != nil {
			return fmt.Errorf("failed to build node %s: %w", nodeID, err)
		}

		ctx.NodeID = nodeID

		ctx.Input = ctx.Output
		ctx.Output = make(map[string]any)

		if err := node.Execute(ctx); err != nil {
			return fmt.Errorf("execution failed for node %s: %w", nodeID, err)
		}
	}

	return nil
}
