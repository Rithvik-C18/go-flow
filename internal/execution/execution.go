package execution

import (
	"fmt"
	"sync"

	wf "github.com/Rithvik-C18/go-flow/internal/workflow"
)

type Node interface {
	Execute(*ExecutionContext) error
}

type Executor struct {
	parallel bool
}

func NewExecutor() *Executor {
	return &Executor{
		parallel: true,
	}
}

func (e *Executor) SetParallel(enabled bool) {
	e.parallel = enabled
}

func (e *Executor) Run(w *wf.Workflow, params ...map[string]any) error {
	_, err := e.RunDetailed(w, params...)
	return err
}

func (e *Executor) RunDetailed(w *wf.Workflow, params ...map[string]any) (map[string]map[string]any, error) {
	var runParams map[string]any
	if len(params) > 0 {
		runParams = params[0]
	}

	ctx := NewExecutionContext(w.Id, "", runParams)
	outputs := make(map[string]map[string]any, len(w.Nodes))

	if e.parallel {
		err := e.runParallel(w, ctx, outputs)
		return outputs, err
	}
	err := e.runSequential(w, ctx, outputs)
	return outputs, err
}

func (e *Executor) runSequential(w *wf.Workflow, ctx *ExecutionContext, outputs map[string]map[string]any) error {
	order, err := topologicalSort(w)
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	for _, nodeID := range order {
		if err := e.executeNode(w, ctx, nodeID); err != nil {
			return err
		}
		outputs[nodeID] = ctx.Output
	}

	return nil
}

func (e *Executor) runParallel(w *wf.Workflow, ctx *ExecutionContext, outputs map[string]map[string]any) error {
	levels, err := getExecutionLevels(w)
	if err != nil {
		return fmt.Errorf("failed to get execution levels: %w", err)
	}

	// Shared state for parallel execution
	var (
		mu        sync.Mutex
		errorOnce sync.Once
		execError error
	)

	// Initialize outputs map for all nodes
	for nodeID := range w.Nodes {
		outputs[nodeID] = make(map[string]any)
	}

	for _, nodes := range levels {
		var wg sync.WaitGroup
		for _, nodeID := range nodes {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()

				// Create a copy of context for this goroutine
				nodeCtx := &ExecutionContext{
					WorkflowID: ctx.WorkflowID,
					NodeID:     id,
					Params:     ctx.Params,
					Input:      make(map[string]any),
					Output:     make(map[string]any),
					Logs:       make([]string, 0),
				}

				// Copy input from direct dependencies
				mu.Lock()
				deps := getDirectDependencies(id, w)
				for _, depID := range deps {
					if depOutput, ok := outputs[depID]; ok {
						for k, v := range depOutput {
							nodeCtx.Input[k] = v
						}
					}
				}
				mu.Unlock()

				// Execute the node
				nodeDef, ok := w.Nodes[id]
				if !ok {
					errorOnce.Do(func() {
						execError = fmt.Errorf("node %s not found in workflow", id)
					})
					return
				}

				node, err := buildNode(nodeDef)
				if err != nil {
					errorOnce.Do(func() {
						execError = fmt.Errorf("failed to build node %s: %w", id, err)
					})
					return
				}

				if err := applyInputMapping(nodeDef.Config, nodeCtx); err != nil {
					errorOnce.Do(func() {
						execError = fmt.Errorf("input parsing failed for node %s: %w", id, err)
					})
					return
				}

				if err := node.Execute(nodeCtx); err != nil {
					errorOnce.Do(func() {
						execError = fmt.Errorf("execution failed for node %s: %w", id, err)
					})
					return
				}

				if err := applyOutputMapping(nodeDef.Config, nodeCtx); err != nil {
					errorOnce.Do(func() {
						execError = fmt.Errorf("output parsing failed for node %s: %w", id, err)
					})
					return
				}

				// Store the output
				mu.Lock()
				outputs[id] = nodeCtx.Output
				mu.Unlock()
			}(nodeID)
		}

		wg.Wait()

		// Check if any error occurred during this level
		if execError != nil {
			return execError
		}
	}

	// Merge all outputs into the final context
	for _, output := range outputs {
		for k, v := range output {
			ctx.Output[k] = v
		}
	}

	return nil
}

func (e *Executor) executeNode(w *wf.Workflow, ctx *ExecutionContext, nodeID string) error {
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

	return nil
}

func getDirectDependencies(nodeID string, w *wf.Workflow) []string {
	var deps []string
	for from, neighbors := range w.Edges {
		for _, neighbor := range neighbors {
			if neighbor == nodeID {
				deps = append(deps, from)
			}
		}
	}
	return deps
}
