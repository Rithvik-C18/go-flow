import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { toast } from 'sonner'
import { ArrowRight, GitBranch, Loader2, Play, Plus, Trash2 } from 'lucide-react'
import { api, nodeTemplates } from '@/lib/api'
import type { Workflow } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Textarea } from '@/components/ui/textarea'

function parseJson(text: string): Record<string, unknown> | null {
  try {
    const parsed: unknown = JSON.parse(text)
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>
    }
    return null
  } catch {
    return null
  }
}

function ConfigPreview({ config }: { config: Record<string, unknown> }) {
  const json = JSON.stringify(config)
  const preview = json.length > 90 ? json.slice(0, 90) + '…' : json
  return (
    <pre className="overflow-hidden text-ellipsis rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
      {preview}
    </pre>
  )
}

export default function WorkflowDetailPage() {
  const { id = '' } = useParams()
  const [workflow, setWorkflow] = useState<Workflow | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  const [nodeOpen, setNodeOpen] = useState(false)
  const [nodeId, setNodeId] = useState('')
  const [nodeType, setNodeType] = useState('http')
  const [nodeConfig, setNodeConfig] = useState(nodeTemplates.http)
  const [savingNode, setSavingNode] = useState(false)

  const [edgeOpen, setEdgeOpen] = useState(false)
  const [edgeFrom, setEdgeFrom] = useState('')
  const [edgeTo, setEdgeTo] = useState('')
  const [savingEdge, setSavingEdge] = useState(false)

  const [runOpen, setRunOpen] = useState(false)
  const [runParams, setRunParams] = useState('{\n  \n}')
  const [running, setRunning] = useState(false)
  const [runResult, setRunResult] = useState<{ ok: boolean; message: string } | null>(null)

  const load = useCallback(async () => {
    try {
      const wf = await api.getWorkflow(id)
      setWorkflow(wf)
      setError(null)
    } catch (err) {
      setError((err as Error).message)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    void load()
  }, [load])

  const nodeIds = useMemo(() => workflow?.nodes.map((n) => n.id) ?? [], [workflow])
  const existingEdges = useMemo(
    () => new Set(workflow?.edges.map((e) => `${e.from}|${e.to}`) ?? []),
    [workflow],
  )

  const changeNodeType = (type: string) => {
    setNodeType(type)
    setNodeConfig(nodeTemplates[type] ?? '{}')
  }

  const addNode = async () => {
    const config = parseJson(nodeConfig)
    if (!nodeId.trim()) return toast.error('Node id is required')
    if (config === null) return toast.error('Config must be a valid JSON object')
    setSavingNode(true)
    try {
      await api.addNode(id, { id: nodeId.trim(), type: nodeType, config })
      toast.success(`Node "${nodeId}" added`)
      setNodeOpen(false)
      setNodeId('')
      void load()
    } catch (err) {
      toast.error((err as Error).message)
    } finally {
      setSavingNode(false)
    }
  }

  const deleteNode = async (nodeIdToDelete: string) => {
    try {
      await api.deleteNode(id, nodeIdToDelete)
      toast.success(`Node "${nodeIdToDelete}" deleted`)
      void load()
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  const addEdge = async () => {
    if (!edgeFrom || !edgeTo) return toast.error('Select both nodes')
    if (edgeFrom === edgeTo) return toast.error('Self loops are not allowed')
    if (existingEdges.has(`${edgeFrom}|${edgeTo}`)) return toast.error('Edge already exists')
    setSavingEdge(true)
    try {
      await api.addEdge(id, edgeFrom, edgeTo)
      toast.success(`Edge ${edgeFrom} → ${edgeTo} added`)
      setEdgeOpen(false)
      setEdgeFrom('')
      setEdgeTo('')
      void load()
    } catch (err) {
      toast.error((err as Error).message)
    } finally {
      setSavingEdge(false)
    }
  }

  const deleteEdge = async (from: string, to: string) => {
    try {
      await api.deleteEdge(id, from, to)
      toast.success(`Edge ${from} → ${to} deleted`)
      void load()
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  const runWorkflow = async () => {
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
      const res = await api.runWorkflow(id, params)
      setRunResult({ ok: true, message: res.status })
      toast.success('Workflow completed successfully')
    } catch (err) {
      setRunResult({ ok: false, message: (err as Error).message })
      toast.error((err as Error).message)
    } finally {
      setRunning(false)
    }
  }

  if (loading) {
    return <p className="py-16 text-center text-sm text-muted-foreground">Loading workflow…</p>
  }

  if (error || !workflow) {
    return (
      <div className="py-16 text-center">
        <p className="font-medium">{error ?? 'Workflow not found'}</p>
        <Link to="/" className="mt-2 inline-block text-sm text-primary">
          ← back to workflows
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link to="/" className="text-sm text-muted-foreground hover:text-foreground">
            ← workflows
          </Link>
          <h1 className="mt-1 text-2xl font-semibold tracking-tight">
            {workflow.name || workflow.id}
          </h1>
          <p className="text-sm text-muted-foreground">
            {workflow.id} · {workflow.nodes.length} nodes · {workflow.edges.length} edges
          </p>
        </div>
        <div className="flex gap-2">
          <Dialog open={runOpen} onOpenChange={setRunOpen}>
            <DialogTrigger
              render={
                <Button>
                  <Play /> Run workflow
                </Button>
              }
            />
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Run workflow</DialogTitle>
                <DialogDescription>
                  Optionally provide run-time params, available to nodes as {'{{ params.key }}'}.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-2">
                <Label htmlFor="run-params">Params (JSON)</Label>
                <Textarea
                  id="run-params"
                  className="min-h-28 font-mono text-xs"
                  value={runParams}
                  onChange={(e) => setRunParams(e.target.value)}
                  placeholder='{"apiKey": "…"}'
                />
              </div>
              {runResult && (
                <div
                  className={
                    'rounded-md border px-3 py-2 text-sm ' +
                    (runResult.ok ? 'border-emerald-500/40 bg-emerald-500/10' : 'border-destructive/50 bg-destructive/10')
                  }
                >
                  {runResult.ok ? 'Completed: ' : 'Failed: '}
                  <span className="font-mono text-xs">{runResult.message}</span>
                </div>
              )}
              <DialogFooter>
                <Button onClick={runWorkflow} disabled={running}>
                  {running && <Loader2 className="animate-spin" />}
                  {running ? 'Running…' : 'Run'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          <AlertDialog>
            <AlertDialogTrigger
              render={
                <Button variant="destructive">
                  <Trash2 /> Delete
                </Button>
              }
            />
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete workflow "{workflow.id}"?</AlertDialogTitle>
                <AlertDialogDescription>This cannot be undone.</AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  render={
                    <Link to="/" onClick={() => void api.deleteWorkflow(workflow.id)}>
                      Delete
                    </Link>
                  }
                />
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle className="flex items-center gap-2">
              <GitBranch className="size-4 text-muted-foreground" /> Nodes
            </CardTitle>
            <Dialog open={nodeOpen} onOpenChange={setNodeOpen}>
              <DialogTrigger
                render={
                  <Button size="sm">
                    <Plus /> Add node
                  </Button>
                }
              />
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Add node</DialogTitle>
                  <DialogDescription>
                    Config keys like url, httpMethod, conditions, input and output accept{' '}
                    {'{{ }}'} references to the previous node's output.
                  </DialogDescription>
                </DialogHeader>
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="node-id">Id</Label>
                    <Input
                      id="node-id"
                      placeholder="e.g. fetch-user"
                      value={nodeId}
                      onChange={(e) => setNodeId(e.target.value)}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Type</Label>
                    <Select
                      value={nodeType}
                      onValueChange={(v) => {
                        if (v) changeNodeType(v)
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="http">http</SelectItem>
                        <SelectItem value="if">if</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="node-config">Config (JSON)</Label>
                    <Textarea
                      id="node-config"
                      className="min-h-56 font-mono text-xs"
                      value={nodeConfig}
                      onChange={(e) => setNodeConfig(e.target.value)}
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button onClick={addNode} disabled={savingNode}>
                    {savingNode ? 'Adding…' : 'Add node'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </CardHeader>
          <CardContent className="space-y-3">
            {workflow.nodes.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted-foreground">No nodes yet</p>
            ) : (
              workflow.nodes.map((node) => (
                <div
                  key={node.id}
                  className="rounded-lg border p-3 transition-colors hover:border-primary/40"
                >
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex min-w-0 items-center gap-2">
                      <span className="truncate text-sm font-medium">{node.id}</span>
                      <Badge variant="secondary" className="shrink-0">
                        {node.type}
                      </Badge>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      onClick={() => void deleteNode(node.id)}
                      title={`Delete ${node.id}`}
                    >
                      <Trash2 className="size-3.5" />
                    </Button>
                  </div>
                  <div className="mt-2">
                    <ConfigPreview config={node.config} />
                  </div>
                </div>
              ))
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <CardTitle className="flex items-center gap-2">
              <ArrowRight className="size-4 text-muted-foreground" /> Edges
            </CardTitle>
            <Dialog open={edgeOpen} onOpenChange={setEdgeOpen}>
              <DialogTrigger
                render={
                  <Button size="sm" disabled={workflow.nodes.length < 2}>
                    <Plus /> Add edge
                  </Button>
                }
              />
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Add edge</DialogTitle>
                  <DialogDescription>
                    An edge means "from" runs before "to"; its output becomes "to"'s input.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid grid-cols-[1fr_auto_1fr] items-end gap-3">
                  <div className="space-y-2">
                    <Label>From</Label>
                    <Select
                      value={edgeFrom}
                      onValueChange={(v) => {
                        if (v) setEdgeFrom(v)
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select node" />
                      </SelectTrigger>
                      <SelectContent>
                        {nodeIds.map((n) => (
                          <SelectItem key={n} value={n}>
                            {n}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <ArrowRight className="mb-2 size-4 text-muted-foreground" />
                  <div className="space-y-2">
                    <Label>To</Label>
                    <Select
                      value={edgeTo}
                      onValueChange={(v) => {
                        if (v) setEdgeTo(v)
                      }}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Select node" />
                      </SelectTrigger>
                      <SelectContent>
                        {nodeIds.map((n) => (
                          <SelectItem key={n} value={n} disabled={n === edgeFrom}>
                            {n}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <DialogFooter>
                  <Button onClick={addEdge} disabled={savingEdge}>
                    {savingEdge ? 'Adding…' : 'Add edge'}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </CardHeader>
          <CardContent>
            {workflow.edges.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted-foreground">
                No edges yet — connect nodes to define execution order
              </p>
            ) : (
              <div className="space-y-2">
                {workflow.edges.map((edge) => (
                  <div
                    key={`${edge.from}|${edge.to}`}
                    className="flex items-center justify-between rounded-lg border px-3 py-2"
                  >
                    <div className="flex items-center gap-2 text-sm">
                      <span className="font-medium">{edge.from}</span>
                      <ArrowRight className="size-3.5 text-muted-foreground" />
                      <span className="font-medium">{edge.to}</span>
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="size-7"
                      onClick={() => void deleteEdge(edge.from, edge.to)}
                      title={`Delete ${edge.from} → ${edge.to}`}
                    >
                      <Trash2 className="size-3.5" />
                    </Button>
                  </div>
                ))}
              </div>
            )}
            <Separator className="my-4" />
            <p className="text-xs text-muted-foreground">
              Nodes run in topological order: a node executes after all of its parents, and its
              output becomes the next node's input.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
