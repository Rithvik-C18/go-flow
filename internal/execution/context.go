package execution

type ExecutionContext struct {
	WorkflowID string
	NodeID     string
	Input      map[string]any
	Output     map[string]any
	Logs       []string
}

func NewExecutionContext(workflowID, nodeID string) *ExecutionContext{
	return &ExecutionContext{	
		WorkflowID: workflowID,
		NodeID :nodeID,
		Input : make(map[string]any),
		Output : make(map[string]any),
		Logs : make([]string, 0),
	}
}