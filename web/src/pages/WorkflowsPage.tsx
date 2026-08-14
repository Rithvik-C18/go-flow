import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Plus, Play, Trash2, Workflow as WorkflowIcon } from 'lucide-react'
import { api } from '@/lib/api'
import type { WorkflowSummary } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
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
  const [open, setOpen] = useState(false)
  const [id, setId] = useState('')
  const [name, setName] = useState('')
  const [creating, setCreating] = useState(false)

  const load = async () => {
    try {
      setWorkflows(await api.listWorkflows())
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const create = async () => {
    if (!id.trim()) {
      toast.error('Workflow id is required')
      return
    }
    setCreating(true)
    try {
      const wf = await api.createWorkflow(id.trim(), name.trim())
      toast.success(`Workflow "${wf.id}" created`)
      setOpen(false)
      setId('')
      setName('')
      void load()
    } catch (err) {
      toast.error((err as Error).message)
    } finally {
      setCreating(false)
    }
  }

  const remove = async (workflowId: string) => {
    try {
      await api.deleteWorkflow(workflowId)
      toast.success(`Workflow "${workflowId}" deleted`)
      void load()
    } catch (err) {
      toast.error((err as Error).message)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Workflows</h1>
          <p className="text-sm text-muted-foreground">Create and manage your DAG workflows</p>
        </div>

        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger
            render={
              <Button>
                <Plus /> New workflow
              </Button>
            }
          />
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create workflow</DialogTitle>
              <DialogDescription>A workflow holds nodes and the edges between them.</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="wf-id">Id</Label>
                <Input
                  id="wf-id"
                  placeholder="e.g. wf-1"
                  value={id}
                  onChange={(e) => setId(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="wf-name">Name</Label>
                <Input
                  id="wf-name"
                  placeholder="e.g. order pipeline"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>
            </div>
            <DialogFooter>
              <Button onClick={create} disabled={creating}>
                {creating ? 'Creating…' : 'Create'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {workflows === null ? (
        <div className="grid gap-4 sm:grid-cols-2">
          <Skeleton className="h-28" />
          <Skeleton className="h-28" />
        </div>
      ) : workflows.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-16 text-center">
            <WorkflowIcon className="size-10 text-muted-foreground" />
            <div>
              <p className="font-medium">No workflows yet</p>
              <p className="text-sm text-muted-foreground">Create one to start building your DAG</p>
            </div>
            <Button onClick={() => setOpen(true)}>
              <Plus /> Create workflow
            </Button>          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {workflows.map((wf) => (
            <Card key={wf.id} className="transition-colors hover:border-primary/50">
              <CardContent className="flex items-center justify-between gap-4 py-4">
                <Link to={`/workflows/${wf.id}`} className="min-w-0">
                  <p className="truncate font-medium">{wf.name || wf.id}</p>
                  <p className="truncate text-sm text-muted-foreground">{wf.id}</p>
                </Link>
                <div className="flex shrink-0 gap-2">
                  <Button variant="outline" size="icon" onClick={() => navigate(`/workflows/${wf.id}`)} title="Open">
                    <Play />
                  </Button>
                  <AlertDialog>
                    <AlertDialogTrigger
                      render={
                        <Button variant="destructive" size="icon" title="Delete">
                          <Trash2 />
                        </Button>
                      }
                    />
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Delete workflow "{wf.id}"?</AlertDialogTitle>
                        <AlertDialogDescription>
                          This permanently removes the workflow and all of its nodes and edges.
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction onClick={() => void remove(wf.id)}>Delete</AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
