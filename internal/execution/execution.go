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

func (e *Executor) Run(w *wf.Workflow, params ...map[string]any) error {
	order, err := topologicalSort(w)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	var runParams map[string]any
	if len(params) > 0 {
		runParams = params[0]
	}

	ctx := NewExecutionContext(w.Id, "", runParams)

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

		if err := applyInputMapping(nodeDef.Config, ctx); err != nil {
			return fmt.Errorf("input parsing failed for node %s: %w", nodeID, err)
		}

		if err := node.Execute(ctx); err != nil {
			return fmt.Errorf("execution failed for node %s: %w", nodeID, err)
		}

		if err := applyOutputMapping(nodeDef.Config, ctx); err != nil {
			return fmt.Errorf("output parsing failed for node %s: %w", nodeID, err)
		}
	}

	return nil
}
