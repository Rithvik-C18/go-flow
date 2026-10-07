import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import {
  ArrowRight,
  CircleAlert,
  Loader2,
  Plus,
  Trash2,
  Workflow as WorkflowIcon,
} from 'lucide-react'
import { api } from '@/lib/api'
import type { WorkflowSummary } from '@/lib/types'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Skeleton } from '@/components/ui/skeleton'

export default function WorkflowsPage() {
  const navigate = useNavigate()
  const [workflows, setWorkflows] = useState<WorkflowSummary[] | null>(null)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [open, setOpen] = useState(false)
  const [id, setId] = useState('')
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [deletingIds, setDeletingIds] = useState<ReadonlySet<string>>(new Set())

  const requestRef = useRef(0)
  const workflowsRef = useRef<WorkflowSummary[] | null>(null)
  const deleteSnapshotRef = useRef<Record<string, WorkflowSummary[]>>({})
  const mountedRef = useRef(true)

  useEffect(() => {
    workflowsRef.current = workflows
  }, [workflows])

  useEffect(() => {
    mountedRef.current = true

    return () => {
      mountedRef.current = false
      requestRef.current += 1
    }
  }, [])

  const load = useCallback(async (options?: { showSkeleton?: boolean }) => {
    const requestId = ++requestRef.current
    if (options?.showSkeleton) {
      setWorkflows(null)
      setLoadError(null)
    }

    try {
      const nextWorkflows = await api.listWorkflows()
      if (requestRef.current !== requestId) return
      setWorkflows(nextWorkflows)
      setLoadError(null)
    } catch (err) {
      if (requestRef.current !== requestId) return
      if (workflowsRef.current === null) {
        setLoadError((err as Error).message)
      }
      toast.error((err as Error).message)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  const normalizedId = id.trim().toLowerCase().replace(/[^a-z0-9_-]+/g, '-')
    .replace(/-{2,}/g, '-')
    .replace(/^-+|-+$/g, '')

  const handleCreateDialogChange = (nextOpen: boolean) => {
    if (creating && !nextOpen) return
    setOpen(nextOpen)
    if (!nextOpen) {
      setId('')
      setName('')
      setCreateError(null)
      setCreating(false)
    }
  }

  const create = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (creating) return

    if (!normalizedId) {
      setCreateError('Enter a workflow ID using letters, numbers, hyphens or underscores.')
      return
    }
    if (!name.trim()) {
      setCreateError('Workflow name is required.')
      return
    }

    setCreating(true)
    setCreateError(null)
    let createdWorkflow: WorkflowSummary | null = null
    try {
      createdWorkflow = await api.createWorkflow(normalizedId, name.trim())
      if (!mountedRef.current) return
      toast.success(`Workflow "${createdWorkflow.name || createdWorkflow.id}" created`)
      navigate(`/workflows/${createdWorkflow.id}`)

      if (open) {
        setOpen(false)
      }
      setId('')
      setName('')
      setCreateError(null)
    } catch (err) {
      if (!mountedRef.current) return
      setCreateError((err as Error).message)
    } finally {
      setCreating(false)
    }
  }

  const remove = async (workflow: WorkflowSummary) => {
    if (deletingIds.has(workflow.id)) return
    const previousWorkflows = workflowsRef.current
    if (previousWorkflows === null) return
    deleteSnapshotRef.current[workflow.id] = previousWorkflows
    setDeletingIds((current) => new Set(current).add(workflow.id))
    setWorkflows(currentWorkflows => currentWorkflows?.filter((item) => item.id !== workflow.id) ?? currentWorkflows)

    try {
      await api.deleteWorkflow(workflow.id)
      if (!mountedRef.current) return
      toast.success(`Workflow "${workflow.name || workflow.id}" deleted`)
      await load()
    } catch (err) {
      setWorkflows(deleteSnapshotRef.current[workflow.id])
      if (workflowsRef.current === null) setLoadError((err as Error).message)
      toast.error((err as Error).message)
    } finally {
      delete deleteSnapshotRef.current[workflow.id]
      if (mountedRef.current) {
        setDeletingIds((current) => {
          const nextDeletingIds = new Set(current)
          nextDeletingIds.delete(workflow.id)
          return nextDeletingIds
        })
      }
    }
  }

  return (
    <div>
      <div className="flex h-9 items-center justify-between">
        <h1 className="text-[15px] font-semibold tracking-tight">
          Workflows
          {workflows !== null && workflows.length > 0 && (
            <span className="ml-2 font-mono text-xs font-normal text-muted-foreground tabular-nums">
              {workflows.length}
            </span>
          )}
        </h1>

        <Dialog open={open} onOpenChange={handleCreateDialogChange}>
          <DialogTrigger
            render={
              <Button size="sm" className="font-medium">
                <Plus /> New workflow
              </Button>
            }
          />
          <DialogContent>
            <DialogHeader>
              <DialogTitle className="text-[15px]">New workflow</DialogTitle>
              <DialogDescription className="text-[13px]">
                The id is permanent and used in the API. The name is for display.
              </DialogDescription>
            </DialogHeader>
            <form onSubmit={create} noValidate>
              <div className="space-y-4 pt-1">
                <div className="space-y-1.5">
                  <Label htmlFor="wf-id" className="text-[13px]">Id</Label>
                  <Input
                    id="wf-id"
                    placeholder="order-pipeline"
                    value={id}
                    onChange={(event) => setId(event.target.value)}
                    autoFocus
                    aria-invalid={!normalizedId && Boolean(createError)}
                    disabled={creating}
                    className="rounded-md font-mono text-[13px]"
                  />
                  {id.trim() !== normalizedId && (
                    <p className="text-xs text-muted-foreground">
                      Will be created as “{normalizedId}”.
                    </p>
                  )}
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="wf-name" className="text-[13px]">Name</Label>
                  <Input
                    id="wf-name"
                    placeholder="Order pipeline"
                    value={name}
                    onChange={(event) => setName(event.target.value)}
                    required
                    aria-invalid={!name.trim() && Boolean(createError)}
                    disabled={creating}
                    className="rounded-md"
                  />
                </div>
                {createError && (
                  <p role="alert" className="flex items-start gap-1.5 text-[13px] text-destructive">
                    <CircleAlert className="mt-0.5 size-3.5 shrink-0" />
                    {createError}
                  </p>
                )}
              </div>
              <DialogFooter className="gap-2">
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => handleCreateDialogChange(false)}
                  disabled={creating}
                >
                  Cancel
                </Button>
                <Button type="submit" size="sm" disabled={creating} className="font-medium">
                  {creating ? <Loader2 className="animate-spin" /> : null}
                  {creating ? 'Creating…' : 'Create workflow'}
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="mt-5">
        {workflows === null && !loadError ? (
          <div className="panel divide-y">
            <Skeleton className="h-[52px] rounded-none border-0" />
            <Skeleton className="h-[52px] rounded-none border-0" />
            <Skeleton className="h-[52px] rounded-none border-0" />
          </div>
        ) : loadError ? (
          <div className="panel flex flex-col items-center gap-3 py-16 text-center">
            <CircleAlert className="size-5 text-destructive" strokeWidth={1.5} />
            <div>
              <p className="text-sm font-medium">Could not load workflows</p>
              <p className="mt-1 text-[13px] text-muted-foreground">{loadError}</p>
            </div>
            <Button variant="outline" size="sm" className="mt-1 font-medium" onClick={() => void load()}>
              Try again
            </Button>
          </div>
        ) : workflows?.length === 0 ? (
          <div className="panel flex flex-col items-center gap-3 py-16 text-center">
            <WorkflowIcon className="size-5 text-muted-foreground/70" strokeWidth={1.5} />
            <div>
              <p className="text-sm font-medium">No workflows</p>
              <p className="mt-1 text-[13px] text-muted-foreground">
                Get started by creating your first one.
              </p>
            </div>
            <Button size="sm" variant="outline" className="mt-1 font-medium" onClick={() => setOpen(true)}>
              <Plus /> New workflow
            </Button>
          </div>
        ) : (
          <div className="panel overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b text-left">
                  <th className="label-caps px-4 py-2 font-medium">Name</th>
                  <th className="label-caps hidden px-4 py-2 font-medium sm:table-cell">Id</th>
                  <th className="relative px-4 py-2">
                    <span className="sr-only">Actions</span>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {workflows?.map((wf) => (
                  <tr key={wf.id} className="group row-hover">
                    <td className="max-w-0 px-4 py-2.5">
                      <Link
                        to={`/workflows/${wf.id}`}
                        className="block truncate rounded-sm font-medium outline-none focus-visible:bg-muted focus-visible:ring-2 focus-visible:ring-ring/50"
                      >
                        {wf.name || wf.id}
                      </Link>
                      <span className="truncate font-mono text-xs text-muted-foreground sm:hidden">
                        {wf.id}
                      </span>
                    </td>
                    <td className="hidden max-w-0 px-4 py-2.5 sm:table-cell">
                      <span className="block truncate font-mono text-xs text-muted-foreground">
                        {wf.id}
                      </span>
                    </td>
                    <td className="whitespace-nowrap px-4 py-2.5 text-right">
                      <div className="inline-flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          render={<Link to={`/workflows/${wf.id}`} />}
                          title="Open"
                          className="text-muted-foreground hover:text-foreground"
                        >
                          <ArrowRight />
                        </Button>
                        <AlertDialog>
                          <AlertDialogTrigger
                            render={
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                title={`Delete ${wf.name || wf.id} (${wf.id})`}
                                className="text-muted-foreground hover:text-destructive"
                                disabled={deletingIds.has(wf.id)}
                            >
                                {deletingIds.has(wf.id) ? <Loader2 className="animate-spin" /> : <Trash2 />}
                              </Button>
                            }
                          />
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle className="text-[15px]">
                                Delete "{wf.name || wf.id}"?
                              </AlertDialogTitle>
                              <AlertDialogDescription className="text-[13px]">
                                This permanently removes the workflow and all of its nodes and
                                edges. This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter className="gap-2">
                              <AlertDialogCancel size="sm">Cancel</AlertDialogCancel>
                              <AlertDialogAction
                                size="sm"
                                className="bg-destructive text-white hover:bg-destructive/90"
                                disabled={deletingIds.has(wf.id)}
                                onClick={() => void remove(wf)}
                              >
                                {deletingIds.has(wf.id) ? 'Deleting…' : 'Delete'}
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
