import { useEffect, useRef, useState } from 'react'
import { Loader2, Trash2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { nodeTypesMeta } from '@/lib/node-meta'

export type NodeInspectorProps = {
  nodeId: string
  type: string
  config: Record<string, unknown>
  onClose: () => void
  onSave: (type: string, config: Record<string, unknown>) => Promise<void>
  onDelete: () => Promise<void>
}

function tryParse(text: string): Record<string, unknown> | null {
  if (!text.trim()) return {}
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

export function NodeInspector({ nodeId, type, config, onClose, onSave, onDelete }: NodeInspectorProps) {
  const [editType, setEditType] = useState(type)
  const [configText, setConfigText] = useState(() => JSON.stringify(config, null, 2))
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const saveTokenRef = useRef(0)
  const deleteTokenRef = useRef(0)

  useEffect(() => {
    setEditType(type)
    setConfigText(JSON.stringify(config, null, 2))
  }, [nodeId, type, config])

  useEffect(() => () => {
    saveTokenRef.current += 1
    deleteTokenRef.current += 1
  }, [])

  const dirtyType = editType !== type
  const parsed = tryParse(configText)

  const save = async () => {
    if (parsed === null || saving) return
    const saveToken = saveTokenRef.current + 1
    saveTokenRef.current = saveToken
    setSaving(true)
    try {
      await onSave(editType, parsed)
    } finally {
      if (saveToken === saveTokenRef.current) setSaving(false)
    }
  }

  const remove = async () => {
    if (deleting) return
    const deleteToken = deleteTokenRef.current + 1
    deleteTokenRef.current = deleteToken
    setDeleting(true)
    try {
      await onDelete()
    } finally {
      if (deleteToken === deleteTokenRef.current) setDeleting(false)
    }
  }

  return (
    <aside className="flex w-80 shrink-0 flex-col border-l bg-card">
      <div className="flex h-11 shrink-0 items-center justify-between border-b px-3">
        <span className="label-caps">Node settings</span>
        <Button variant="ghost" size="icon-xs" onClick={onClose} title="Close">
          <X />
        </Button>
      </div>

      <div className="flex-1 space-y-4 overflow-y-auto p-3.5">
        <div className="space-y-1.5">
          <Label className="text-[13px]">Id</Label>
          <p className="rounded-md border bg-muted px-2.5 py-1.5 font-mono text-[13px]">{nodeId}</p>
        </div>

        <div className="space-y-1.5">
          <Label className="text-[13px]">Type</Label>
          <Select value={editType} onValueChange={(v) => v && setEditType(v)}>
            <SelectTrigger className="w-full text-[13px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {nodeTypesMeta.map((m) => (
                <SelectItem key={m.type} value={m.type}>
                  {m.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {dirtyType && (
            <p className="text-[11px] text-amber-600 dark:text-amber-500">
              Changing type replaces this node on save.
            </p>
          )}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label className="text-[13px]">Parameters (JSON)</Label>
            {configText !== JSON.stringify(config, null, 2) && (
              <button
                type="button"
                onClick={() => setConfigText(JSON.stringify(config, null, 2))}
                className="text-[11px] font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
              >
                Reset
              </button>
            )}
          </div>
          <Textarea
            value={configText}
            onChange={(e) => setConfigText(e.target.value)}
            spellCheck={false}
            className={`min-h-64 font-mono text-xs leading-relaxed ${parsed === null ? 'border-destructive focus-visible:border-destructive' : ''}`}
          />
          {parsed === null && (
            <p className="text-[11px] text-destructive">Invalid JSON — must be an object.</p>
          )}
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-2 border-t p-3">
        <Button
          variant="outline"
          size="sm"
          className="ml-auto text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
          onClick={() => void remove()}
          disabled={deleting}
        >
          {deleting ? <Loader2 className="animate-spin" /> : <Trash2 />}
          Delete
        </Button>
        <Button
          size="sm"
          className="font-medium"
          onClick={() => void save()}
          disabled={parsed === null || saving || (editType === type && configText === JSON.stringify(config, null, 2))}
        >
          {saving && <Loader2 className="animate-spin" />}
          Save changes
        </Button>
      </div>
    </aside>
  )
}
