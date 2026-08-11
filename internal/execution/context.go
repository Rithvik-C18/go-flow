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
