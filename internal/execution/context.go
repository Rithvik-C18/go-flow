package execution

type ExecutionContext struct {
	WorkflowID string
	NodeID     string
	Params     map[string]any
	Input      map[string]any
	Output     map[string]any
	Logs       []string
}

func NewExecutionContext(workflowID, nodeID string, params map[string]any) *ExecutionContext {
	if params == nil {
		params = make(map[string]any)
	}

	return &ExecutionContext{
		WorkflowID: workflowID,
		NodeID:     nodeID,
		Params:     params,
		Input:      make(map[string]any),
		Output:     make(map[string]any),
		Logs:       make([]string, 0),
	}
}

func CloneExecutionContext(ctx *ExecutionContext) *ExecutionContext {
	// Deep copy params
	paramsCopy := make(map[string]any)
	for k, v := range ctx.Params {
		paramsCopy[k] = v
	}

	// Deep copy input
	inputCopy := make(map[string]any)
	for k, v := range ctx.Input {
		inputCopy[k] = v
	}

	return &ExecutionContext{
		WorkflowID: ctx.WorkflowID,
		NodeID:     ctx.NodeID,
		Params:     paramsCopy,
		Input:      inputCopy,
		Output:     make(map[string]any),
		Logs:       make([]string, 0),
	}
}
