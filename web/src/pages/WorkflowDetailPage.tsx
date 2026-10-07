import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import {
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Connection,
  type Edge as FlowEdge,
  type Node as FlowNodeT,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {
  ArrowLeft,
  CircleCheck,
  CircleX,
  LayoutGrid,
  Loader2,
  MousePointerClick,
  Play,
  Trash2,
} from 'lucide-react'
import { api } from '@/lib/api'
import type { Workflow } from '@/lib/types'
import { cn } from '@/lib/utils'
import { autoLayout, loadPositions, savePositions } from '@/lib/positions'
import { defaultConfigFor } from '@/lib/node-meta'
import { Button } from '@/components/ui/button'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { FlowNode, type FlowNodeData } from '@/components/flow/FlowNode'
import { PalettePanel } from '@/components/flow/PalettePanel'
import { NodeInspector } from '@/components/flow/NodeInspector'
import { FlowActionsContext, type FlowActions } from '@/components/flow/flow-actions'

const nodeTypes = { flow: FlowNode }

function parseJson(text: string): Record<string, unknown> | null {
  try {
    const parsed: unknown = JSON.parse(text)
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) return parsed as Record<string, unknown>
    return null
  } catch {
    return null
  }
}

/** Unique id for a new node of the given type, e.g. "http-3". */
function uniqueId(base: string, taken: Set<string>): string {
  let n = 1
  while (taken.has(`${base}-${n}`) || taken.has(base)) n += 1
  return `${base}-${n}`
}

function Editor({ workflowId }: { workflowId: string }) {
  const reactFlow = useReactFlow()

  const [workflow, setWorkflow] = useState<Workflow | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [nodes, setNodes, onNodesChange] = useNodesState<FlowNodeT<FlowNodeData>>([])
  const [edges, setEdges] = useEdgesState<FlowEdge>([])
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)

  const [runOpen, setRunOpen] = useState(false)
  const [mutating, setMutating] = useState(false)
  const [deletingSelection, setDeletingSelection] = useState(false)
  const [runParams, setRunParams] = useState('{\n  \n}')
  const [running, setRunning] = useState(false)
  const [runResult, setRunResult] = useState<{ ok: boolean; message: string } | null>(null)
  const [dropActive, setDropActive] = useState(false)

  const canvasRef = useRef<HTMLDivElement>(null)
  const dragDepth = useRef(0)
  const positionsRef = useRef<Record<string, { x: number; y: number }>>({})
  const loadTokenRef = useRef(0)
  const mutationQueueRef = useRef<Promise<void>>(Promise.resolve())

  const queueMutation = useCallback((operation: () => Promise<void>) => {
    const nextMutation = mutationQueueRef.current.then(operation)
    mutationQueueRef.current = nextMutation.catch(() => undefined)
    return nextMutation
  }, [])

  /** Rebuild flow state from server data, preserving/migrating positions. */
  const applyWorkflow = useCallback(
    (wf: Workflow) => {
      const saved = loadPositions(workflowId)
      const missing = wf.nodes.filter((n) => !(n.id in saved))
      const computed = missing.length > 0 ? autoLayout(wf.nodes, wf.edges) : {}
      const positions = { ...computed, ...saved }

      // drop positions for nodes that no longer exist
      const liveIds = new Set(wf.nodes.map((n) => n.id))
      const cleaned = Object.fromEntries(Object.entries(positions).filter(([id]) => liveIds.has(id)))
      positionsRef.current = cleaned
      savePositions(workflowId, cleaned)

      setNodes(
        wf.nodes.map((n) => ({
          id: n.id,
          type: 'flow',
          position: positions[n.id] ?? { x: 0, y: 0 },
          data: { nodeId: n.id, type: n.type, config: n.config },
        })),
      )
      setEdges(
        wf.edges.map((e) => ({
          id: `e-${e.from}-${e.to}`,
          source: e.from,
          target: e.to,
          style: { strokeWidth: 1.5 },
        })),
      )
    },
    [workflowId, setNodes, setEdges],
  )

  const load = useCallback(async () => {
    const loadToken = loadTokenRef.current + 1
    loadTokenRef.current = loadToken
    try {
      const wf = await api.getWorkflow(workflowId)
      if (loadToken !== loadTokenRef.current) return
      setWorkflow(wf)
      setError(null)
      applyWorkflow(wf)
    } catch (err) {
      if (loadToken !== loadTokenRef.current) return
      setError((err as Error).message)
    } finally {
      if (loadToken === loadTokenRef.current) setLoading(false)
    }
  }, [workflowId, applyWorkflow])

  useEffect(() => {
    void load()
  }, [load])

  /* ---------- mutations ---------- */

  const addNodeAt = useCallback(
    async (type: string, position: { x: number; y: number }, connectFrom?: string) => {
      return queueMutation(async () => {
        const id = uniqueId(type.split('_')[0], new Set(nodes.map((n) => n.id)))
        try {
          await api.addNode(workflowId, { id, type, config: parseJson(defaultConfigFor(type)) ?? {} })
          if (connectFrom) await api.addEdge(workflowId, connectFrom, id)
          positionsRef.current[id] = position
          savePositions(workflowId, positionsRef.current)
          toast.success(connectFrom ? `Added "${id}" and connected` : `Added "${id}"`)
          await load()
          setSelectedNodeId(id)
        } catch (err) {
          delete positionsRef.current[id]
          savePositions(workflowId, positionsRef.current)
          toast.error((err as Error).message)
        }
      })
    },
    [nodes, workflowId, load, queueMutation],
  )

  const deleteNodeById = useCallback(
    async (nodeId: string) => {
      return queueMutation(async () => {
        try {
          await api.deleteNode(workflowId, nodeId)
          delete positionsRef.current[nodeId]
          savePositions(workflowId, positionsRef.current)
          toast.success(`Node "${nodeId}" deleted`)
          if (selectedNodeId === nodeId) setSelectedNodeId(null)
          await load()
        } catch (err) {
          toast.error((err as Error).message)
        }
      })
    },
    [selectedNodeId, workflowId, load, queueMutation],
  )

  const connectNodes = useCallback(
    async (connection: Connection) => {
      if (!connection.source || !connection.target || connection.source === connection.target) return
      const from = connection.source
      const to = connection.target
      await queueMutation(async () => {
        try {
          if (edges.some((edge) => edge.source === from && edge.target === to)) {
            toast.error('These nodes are already connected')
            return
          }
          await api.addEdge(workflowId, from, to)
          toast.success(`Connected ${from} → ${to}`)
          await load()
        } catch (err) {
          toast.error((err as Error).message)
        }
      })
    },
    [edges, workflowId, load, queueMutation],
  )

  /* keyboard deletion */
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (!['Backspace', 'Delete'].includes(e.key)) return
      const el = document.activeElement
      if (el instanceof HTMLInputElement || el instanceof HTMLTextAreaElement || (el instanceof HTMLElement && el.isContentEditable)) return

      const selected = nodes.filter((n) => n.selected)
      const selectedEdges = edges.filter((ed) => ed.selected)
      if (selected.length === 0 && selectedEdges.length === 0) return

      e.preventDefault()
      if (deletingSelection) return
      setDeletingSelection(true)
      void (async () => {
        try {
          for (const edge of selectedEdges) {
            await api.deleteEdge(workflowId, edge.source, edge.target)
          }
          for (const node of selected) {
            await api.deleteNode(workflowId, node.id)
            if (node.id === selectedNodeId) setSelectedNodeId(null)
          }
          toast.success('Deleted')
          await load()
        } catch (err) {
          toast.error((err as Error).message)
        } finally {
          setDeletingSelection(false)
        }
      })()
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [deletingSelection, nodes, edges, workflowId, selectedNodeId, load])

  /* drag & drop from palette */
  const onDragEnter = useCallback((e: React.DragEvent) => {
    if (!e.dataTransfer.types.includes('application/go-flow-node')) return
    e.preventDefault()
    dragDepth.current += 1
    setDropActive(true)
  }, [])
  const onDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }, [])
  const onDragLeave = useCallback(() => {
    dragDepth.current = Math.max(0, dragDepth.current - 1)
    if (dragDepth.current === 0) setDropActive(false)
  }, [])
  const onDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      dragDepth.current = 0
      setDropActive(false)
      const type = e.dataTransfer.getData('application/go-flow-node')
      if (!type) return
      // center the node under the cursor, like n8n
      const point = reactFlow.screenToFlowPosition({ x: e.clientX, y: e.clientY })
      void addNodeAt(type, { x: point.x - 112, y: point.y - 26 })
    },
    [reactFlow, addNodeAt],
  )

  /** Add a node at the center of the visible canvas (palette click). */
  const addNodeAtCenter = useCallback(
    (type: string) => {
      const rect = canvasRef.current?.getBoundingClientRect()
      const point = rect
        ? reactFlow.screenToFlowPosition({ x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 })
        : { x: 0, y: 0 }
      void addNodeAt(type, { x: point.x - 112, y: point.y - 26 })
    },
    [reactFlow, addNodeAt],
  )

  /* editor actions exposed to nodes via context */
  const flowActions = useMemo<FlowActions>(
    () => ({
      addConnected: (sourceId, type) => {
        const src = nodes.find((n) => n.id === sourceId)
        return addNodeAt(type, { x: (src?.position.x ?? 0) + 300, y: src?.position.y ?? 0 }, sourceId)
      },
      openInspector: setSelectedNodeId,
      deleteNode: (nodeId) => void deleteNodeById(nodeId),
    }),
    [nodes, addNodeAt, deleteNodeById],
  )

  /* auto layout */
  const runAutoLayout = useCallback(() => {
    if (!workflow) return
    const computed = autoLayout(workflow.nodes, workflow.edges)
    setNodes((nds) => nds.map((n) => ({ ...n, position: computed[n.id] ?? n.position })))
    positionsRef.current = computed
    savePositions(workflowId, computed)
    void reactFlow.fitView({ padding: 0.15 })
  }, [workflow, workflowId, setNodes, reactFlow])

  /* run */
  const runWorkflow = async () => {
    if (running) return
    let params: Record<string, unknown> = {}
    const trimmed = runParams.trim()
    if (trimmed && trimmed !== '{}') {
      const parsed = parseJson(trimmed)
      if (parsed === null) return toast.error('Run params must be a valid JSON object')
      params = parsed
    }
    setRunning(true)
    setRunResult(null)
    try {
      const res = await api.runWorkflow(workflowId, params)
      setRunResult({ ok: true, message: res.status })
      toast.success('Workflow completed successfully')
    } catch (err) {
      setRunResult({ ok: false, message: (err as Error).message })
      toast.error((err as Error).message)
    } finally {
      setRunning(false)
    }
  }

  const selectedNode = useMemo(
    () => nodes.find((n) => n.id === selectedNodeId) ?? null,
    [nodes, selectedNodeId],
  )

  /* ---------- render states ---------- */

  if (loading) {
    return (
      <div className="flex h-full">
        <Skeleton className="h-full w-60 shrink-0 rounded-none" />
        <div className="flex-1 p-6">
          <Skeleton className="h-8 w-64 rounded-md" />
        </div>
      </div>
    )
  }

  if (error || !workflow) {
    return (
      <div className="flex h-full flex-col items-center justify-center text-center">
        <p className="text-sm font-medium">{error ?? 'Workflow not found'}</p>
        <Button variant="ghost" size="sm" render={<Link to="/" />} className="mt-3 text-muted-foreground">
          <ArrowLeft data-icon="inline-start" /> Back to workflows
        </Button>
      </div>
    )
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      {/* Header */}
      <header className="flex h-12 shrink-0 items-center justify-between gap-4 border-b bg-card px-4">
        <div className="flex min-w-0 items-center gap-3">
          <Button
            variant="ghost"
            size="icon-sm"
            render={<Link to="/" />}
            title="Back to workflows"
            className="text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft />
          </Button>
          <div className="min-w-0 border-l pl-3">
            <h1 className="truncate text-[14px] font-semibold leading-tight tracking-tight">
              {workflow.name || workflow.id}
            </h1>
            <p className="truncate font-mono text-[11px] text-muted-foreground">{workflow.id}</p>
          </div>
        </div>

        <div className="flex items-center gap-1.5">
          <AlertDialog>
            <AlertDialogTrigger
              render={
                <Button variant="outline" size="sm" className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive">
                  <Trash2 /> Delete
                </Button>
              }
            />
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle className="text-[15px]">Delete "{workflow.name || workflow.id}"?</AlertDialogTitle>
                <AlertDialogDescription className="text-[13px]">
                  This permanently removes the workflow and all of its nodes and edges. This action
                  cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter className="gap-2">
                <AlertDialogCancel size="sm">Cancel</AlertDialogCancel>
            <AlertDialogAction
                  size="sm"
                  className="bg-destructive text-white hover:bg-destructive/90"
                  render={
                    <Link to="/" onClick={() => { void api.deleteWorkflow(workflow.id).catch(() => undefined) }}>
                      Delete
                    </Link>
                  }
                />
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>

          <Dialog open={runOpen} onOpenChange={setRunOpen}>
            <DialogContent>
              <DialogHeader>
                <DialogTitle className="text-[15px]">Run "{workflow.name || workflow.id}"</DialogTitle>
                <DialogDescription className="text-[13px]">
                  Nodes execute in topological order — each output feeds the next input.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-1.5 pt-1">
                <Label htmlFor="run-params" className="text-[13px]">Params (JSON)</Label>
                <Textarea
                  id="run-params"
                  className="min-h-36 font-mono text-xs leading-relaxed"
                  value={runParams}
                  onChange={(e) => setRunParams(e.target.value)}
                  spellCheck={false}
                  placeholder='{ "apiKey": "…" }'
                />
              </div>
              {runResult && (
                <div className="rounded-md border px-3 py-2 text-[13px]">
                  <span className="flex items-center gap-1.5 font-medium">
                    {runResult.ok ? (
                      <>
                        <CircleCheck className="size-3.5 text-emerald-600 dark:text-emerald-500" /> Completed
                      </>
                    ) : (
                      <>
                        <CircleX className="size-3.5 text-destructive" /> Failed
                      </>
                    )}
                  </span>
                  <span className="mt-1 block break-all font-mono text-xs text-muted-foreground">
                    {runResult.message}
                  </span>
                </div>
              )}
              <DialogFooter className="gap-2">
                <Button variant="ghost" size="sm" onClick={() => setRunOpen(false)}>
                  Close
                </Button>
                <Button size="sm" className="font-medium" onClick={runWorkflow} disabled={running}>
                  {running ? <Loader2 className="animate-spin" /> : <Play />}
                  {running ? 'Running…' : 'Run'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <Button size="sm" className="font-medium" onClick={() => setRunOpen(true)}>
            <Play /> Run workflow
          </Button>
        </div>
      </header>

      {/* Body */}
      <div className="flex min-h-0 flex-1">
        <PalettePanel onAdd={addNodeAtCenter} disabled={mutating || loading} />

        <div
          ref={canvasRef}
          className={cn(
            'relative min-w-0 flex-1 transition-shadow duration-150',
            dropActive && 'z-10 ring-1 ring-inset ring-ring/50',
          )}
          onDragEnter={onDragEnter}
          onDragOver={onDragOver}
          onDragLeave={onDragLeave}
          onDrop={onDrop}
        >
          <FlowActionsContext.Provider value={flowActions}>
            <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodesChange={(changes) => {
              onNodesChange(changes)
              const moved = changes.filter(
                (change) => change.type === 'position' && change.dragging === false && change.position,
              )
              if (moved.length === 0) return
              const next = { ...positionsRef.current }
              for (const change of moved) {
                if (change.type === 'position' && change.position) next[change.id] = change.position
              }
              positionsRef.current = next
              savePositions(workflowId, next)
            }}
            onEdgeClick={(_, edge) => {
              void queueMutation(async () => {
                setMutating(true)
                try {
                  await api.deleteEdge(workflowId, edge.source, edge.target)
                  toast.success(`Deleted ${edge.source} → ${edge.target}`)
                  await load()
                } catch (err) {
                  toast.error((err as Error).message)
                } finally {
                  setMutating(false)
                }
              })
            }}
            onConnect={connectNodes}
            onNodeClick={(_, node) => setSelectedNodeId(node.id)}
            onPaneClick={() => setSelectedNodeId(null)}
            deleteKeyCode={null}
            fitView
            fitViewOptions={{ padding: 0.2 }}
            minZoom={0.25}
            maxZoom={2}
            proOptions={{ hideAttribution: true }}
            defaultEdgeOptions={{ style: { strokeWidth: 1.5 } }}
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1.5} className="bg-muted/40" />
            <Controls showInteractive={false} />

            <div className="absolute top-3 right-3 z-10 flex gap-1.5">
              <Button variant="outline" size="sm" className="bg-card font-medium shadow-sm" onClick={runAutoLayout}>
                <LayoutGrid /> Auto layout
              </Button>
            </div>

            <MiniMap
              pannable
              zoomable
              maskColor="oklch(0 0 0 / 8%)"
              nodeColor={() => 'oklch(0.55 0 0)'}
            />

            {nodes.length === 0 && (
              <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center">
                <div className="pointer-events-auto flex flex-col items-center gap-2 rounded-lg border bg-card px-8 py-7 text-center shadow-sm">
                  <MousePointerClick className="size-5 text-muted-foreground/70" strokeWidth={1.5} />
                  <p className="text-sm font-medium">Start building</p>
                  <p className="max-w-56 text-[13px] text-muted-foreground">
                    Drag a block from the left panel onto the canvas — or click it — to add your
                    first node.
                  </p>
                </div>
              </div>
            )}
          </ReactFlow>
          </FlowActionsContext.Provider>
        </div>

        {selectedNode && (
          <NodeInspector
            key={selectedNode.id}
            nodeId={selectedNode.data.nodeId}
            type={selectedNode.data.type}
            config={selectedNode.data.config}
            onClose={() => setSelectedNodeId(null)}
            onSave={(newType, newConfig) =>
              queueMutation(async () => {
                setMutating(true)
                try {
                  await api.deleteNode(workflowId, selectedNode.id)
                  await api.addNode(workflowId, { id: selectedNode.id, type: newType, config: newConfig })
                  toast.success(`Saved "${selectedNode.id}"`)
                  await load()
                } catch (err) {
                  toast.error((err as Error).message)
                  await load()
                } finally {
                  setMutating(false)
                }
              })
            }
            onDelete={async () => deleteNodeById(selectedNode.id)}
          />
        )}
      </div>
    </div>
  )
}

export default function WorkflowDetailPage() {
  const { id = '' } = useParams()
  return (
    <ReactFlowProvider>
      <Editor workflowId={id} />
    </ReactFlowProvider>
  )
}
