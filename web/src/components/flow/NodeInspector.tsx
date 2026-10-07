import { useEffect, useState } from 'react'
import { Loader2, Play, Trash2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { api } from '@/lib/api'
import { defaultConfigFor, nodeTypeMeta, nodeTypesMeta } from '@/lib/node-meta'
import { Choice, Field, NodeConfigFields, compileForm, initialForm, type FormState } from './NodeConfigFields'

type Config = Record<string, unknown>
export type NodeInspectorProps = {
  workflowId: string
  nodeId: string
  type: string
  config: Config
  onClose: () => void
  onSave: (type: string, config: Config) => Promise<void>
  onDelete: () => Promise<void>
}
function parseObject(raw: string): Config | null {
  try { const v: unknown = JSON.parse(raw); return v && typeof v === 'object' && !Array.isArray(v) ? v as Config : null }
  catch { return null }
}
export function NodeInspector({ workflowId, nodeId, type, config, onClose, onSave, onDelete }: NodeInspectorProps) {
  const [editType, setEditType] = useState(type)
  const [form, setForm] = useState<FormState>(() => initialForm(config))
  const [advanced, setAdvanced] = useState(false)
  const [jsonText, setJsonText] = useState(() => JSON.stringify(config, null, 2))
  const [paramsText, setParamsText] = useState('{}')
  const [result, setResult] = useState<Config | null>(null)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [deleting, setDeleting] = useState(false)
  useEffect(() => { setEditType(type); setForm(initialForm(config)); setJsonText(JSON.stringify(config, null, 2)) }, [nodeId, type, config])
  const parsed = advanced ? parseObject(jsonText) : null
  const validation = advanced ? (parsed ? { config: parsed } : { error: 'Advanced JSON must be an object.' }) : compileForm(editType, form)
  const dirty = editType !== type || (validation.config !== undefined && JSON.stringify(validation.config) !== JSON.stringify(config))
  const changeType = (nextType: string) => {
    const next = parseObject(defaultConfigFor(nextType)) ?? {}
    setEditType(nextType); setForm(initialForm(next)); setJsonText(JSON.stringify(next, null, 2)); setResult(null)
  }
  const changeMode = () => {
    if (advanced) {
      if (!parsed) { setError('Fix the JSON before returning to the form.'); return }
      setForm(initialForm(parsed))
    } else setJsonText(JSON.stringify(validation.config ?? form.config, null, 2))
    setAdvanced(!advanced); setError('')
  }
  const save = async () => {
    if (!validation.config) return
    setSaving(true); setError(''); setResult(null)
    try { await onSave(editType, validation.config) }
    catch (err) { setError((err as Error).message) }
    finally { setSaving(false) }
  }
  const test = async () => {
    if (!validation.config) return
    const params = parseObject(paramsText)
    if (!params) { setError('Test input must be a JSON object.'); return }
    setTesting(true); setError(''); setResult(null)
    try {
      if (dirty) await onSave(editType, validation.config)
      const response = await api.runNode(workflowId, nodeId, params)
      setResult(response.output ?? {})
    } catch (err) { setError((err as Error).message) }
    finally { setTesting(false) }
  }
  return <aside className="flex w-[min(28rem,45vw)] shrink-0 flex-col border-l bg-card" aria-label="Node settings">
    <div className="flex h-12 shrink-0 items-center justify-between border-b px-4"><div className="min-w-0"><p className="truncate text-[13px] font-semibold">{nodeTypeMeta(editType).label}</p><p className="truncate font-mono text-[11px] text-muted-foreground">{nodeId}</p></div><Button variant="ghost" size="icon-xs" onClick={onClose} title="Close settings"><X /></Button></div>
    <div className="flex-1 space-y-5 overflow-y-auto p-4">
      <div className="rounded-md border bg-muted/40 px-3 py-2 text-[12px] leading-relaxed text-muted-foreground">{nodeTypeMeta(editType).description}. Configure the fields, save, then use <strong className="text-foreground">Test step</strong> to see the output.</div>
      <Field label="Node type" hint={editType !== type ? 'Changing type keeps this node and its connections.' : undefined}><Choice value={editType} options={nodeTypesMeta.map(meta => meta.type)} onChange={changeType} /></Field>
      <div className="flex items-center justify-between border-b pb-2"><span className="label-caps">Parameters</span><button type="button" className="text-[12px] text-muted-foreground underline-offset-2 hover:text-foreground hover:underline" onClick={changeMode}>{advanced ? 'Use form' : 'Advanced JSON'}</button></div>
      {advanced ? <Field label="Configuration JSON"><Textarea aria-label="Configuration JSON" value={jsonText} onChange={e => setJsonText(e.target.value)} spellCheck={false} className="min-h-72 font-mono text-xs" /></Field> : <NodeConfigFields type={editType} state={form} onChange={setForm} />}
      <div className="border-t pt-4"><Field label="Test input (JSON)" hint="Optional values available as {{ params.name }}. Test step also runs connected steps before this one."><Textarea aria-label="Test input (JSON)" value={paramsText} onChange={e => setParamsText(e.target.value)} spellCheck={false} className="min-h-16 font-mono text-xs" /></Field></div>
      {(validation.error || error) && <p role="alert" className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-[12px] text-destructive">{validation.error || error}</p>}
      {result && <div className="space-y-2 border-t pt-4"><p className="label-caps">Step output</p><pre aria-label="Step output" className="max-h-64 overflow-auto rounded-md border bg-muted p-3 font-mono text-[11px] leading-relaxed">{JSON.stringify(result, null, 2)}</pre></div>}
    </div>
    <div className="flex shrink-0 flex-wrap items-center gap-2 border-t p-3"><Button variant="ghost" size="sm" className="mr-auto text-muted-foreground hover:text-destructive" disabled={deleting} onClick={() => { setDeleting(true); void onDelete().finally(() => setDeleting(false)) }}>{deleting ? <Loader2 className="animate-spin" /> : <Trash2 />} Delete</Button><Button variant="outline" size="sm" disabled={testing || saving || !!validation.error} onClick={() => void test()}>{testing ? <Loader2 className="animate-spin" /> : <Play />} Test step</Button><Button size="sm" disabled={saving || testing || !!validation.error || !dirty} onClick={() => void save()}>{saving && <Loader2 className="animate-spin" />} Save</Button></div>
  </aside>
}
