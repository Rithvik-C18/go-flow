import { createContext, useContext } from 'react'

export type FlowActions = {
  /** Create a node of `type` and connect it to the given source node. */
  addConnected: (sourceNodeId: string, type: string) => Promise<void>
  openInspector: (nodeId: string) => void
  deleteNode: (nodeId: string) => void
}

export const FlowActionsContext = createContext<FlowActions | null>(null)

export function useFlowActions(): FlowActions {
  const ctx = useContext(FlowActionsContext)
  if (!ctx) throw new Error('useFlowActions must be used inside <FlowActionsContext.Provider>')
  return ctx
}
