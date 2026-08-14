export interface NodeDefinition {
  id: string
  type: string
  config: Record<string, unknown>
}

export interface Edge {
  from: string
  to: string
}

export interface WorkflowSummary {
  id: string
  name: string
}

export interface Workflow {
  id: string
  name: string
  nodes: NodeDefinition[]
  edges: Edge[]
}
