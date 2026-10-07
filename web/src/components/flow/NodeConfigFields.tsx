import { Plus, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

type Config = Record<string, unknown>
type Pair = { name: string; value: string }
export type FormState = {
  config: Config
  query: Pair[]
  headers: Pair[]
  body: string
  data: string
  values: string
}
export const text = (v: unknown, fallback = '') => typeof v === 'string' ? v : fallback
export function pairsFrom(v: unknown): Pair[] {
  if (!v || typeof v !== 'object' || Array.isArray(v)) return [{ name: '', value: '' }]
  const rows = Object.entries(v).map(([name, value]) => ({ name, value: String(value ?? '') }))
  return rows.length ? rows : [{ name: '', value: '' }]
}
export function initialForm(config: Config): FormState {
  return {
    config,
    query: pairsFrom(config.queryParameters),
    headers: pairsFrom(config.headers),
    body: JSON.stringify(config.body ?? {}, null, 2),
    data: JSON.stringify(config.data ?? [['Name', 'Value']], null, 2),
    values: JSON.stringify(config.values ?? ['value'], null, 2),
  }
}
function arrayJson(raw: string): unknown[] | null {
  try { const v: unknown = JSON.parse(raw); return Array.isArray(v) ? v : null } catch { return null }
}
export function compileForm(type: string, state: FormState): { config?: Config; error?: string } {
  const value = { ...state.config }
  if (type === 'http') {
    try {
      const url = new URL(text(value.url))
      if (!['http:', 'https:'].includes(url.protocol)) throw new Error()
    } catch { return { error: 'Enter a complete http:// or https:// URL.' } }
    if ([...state.query, ...state.headers].some(row => !row.name.trim() && row.value.trim())) return { error: 'Every query parameter and header value needs a name.' }
    value.queryParameters = Object.fromEntries(state.query.filter(row => row.name.trim()).map(row => [row.name.trim(), row.value]))
    value.headers = Object.fromEntries(state.headers.filter(row => row.name.trim()).map(row => [row.name.trim(), row.value]))
    if (['POST', 'PUT', 'PATCH'].includes(text(value.httpMethod, 'GET'))) {
      try { value.body = JSON.parse(state.body) as unknown } catch { return { error: 'Request body must be valid JSON.' } }
    } else delete value.body
  }
  if (type === 'if') {
    const conditions = value.conditions
    if (!Array.isArray(conditions) || !conditions.length || conditions.some(c => typeof c !== 'object' || !c || !('operator' in c) || !('left' in c) || !('right' in c) || c.left === '' || (!['exists','notExists','isEmpty','notEmpty','isTruthy','isFalsy'].includes(c.operator) && c.right === ''))) return { error: 'Add at least one complete condition.' }
  }
  if (type === 'gemini' && !text(value.prompt).trim()) return { error: 'Enter a prompt.' }
  if (type === 'google_docs') {
    if (value.operation === 'create' && !text(value.title).trim()) return { error: 'Enter a document title.' }
    if (value.operation !== 'create' && !text(value.documentId).trim()) return { error: 'Enter a document ID.' }
    if (value.operation === 'append' && !text(value.text).trim()) return { error: 'Enter text to append.' }
  }
  if (type === 'google_sheets') {
    if (value.operation === 'create') {
      if (!text(value.title).trim()) return { error: 'Enter a spreadsheet title.' }
      const data = arrayJson(state.data)
      if (!data || data.some(row => !Array.isArray(row))) return { error: 'Initial rows must be a JSON array of rows.' }
      value.data = data
    } else {
      if (!text(value.spreadsheetId).trim()) return { error: 'Enter a spreadsheet ID.' }
      if (value.operation === 'update' && !text(value.range).trim()) return { error: 'Enter a sheet range.' }
      if (['append', 'update'].includes(text(value.operation))) {
        const values = arrayJson(state.values)
        if (!values?.length) return { error: 'Row values must be a nonempty JSON array.' }
        value.values = values
      }
    }
  }
  return { config: value }
}
function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return <div className="space-y-1.5"><Label className="text-[13px] font-medium">{label}</Label>{children}{hint && <p className="text-[11px] leading-relaxed text-muted-foreground">{hint}</p>}</div>
}
export { Field }
export function Choice({ value, options, onChange }: { value: string; options: string[]; onChange: (v: string) => void }) {
  return <select aria-label="Choose option" value={value} onChange={e => onChange(e.target.value)} className="h-8 w-full rounded-md border border-input bg-background px-2.5 text-[13px] outline-none focus-visible:ring-2 focus-visible:ring-ring/40">{options.map(option => <option key={option} value={option}>{option}</option>)}</select>
}
function Pairs({ label, hint, rows, onChange }: { label: string; hint: string; rows: Pair[]; onChange: (rows: Pair[]) => void }) {
  return <Field label={label} hint={hint}><div className="space-y-1.5">{rows.map((row, index) => <div key={index} className="flex gap-1.5"><Input aria-label={`${label} name ${index + 1}`} placeholder="Name" value={row.name} className="min-w-0 flex-1 font-mono text-xs" onChange={e => onChange(rows.map((r, i) => i === index ? { ...r, name: e.target.value } : r))} /><Input aria-label={`${label} value ${index + 1}`} placeholder="Value" value={row.value} className="min-w-0 flex-1 font-mono text-xs" onChange={e => onChange(rows.map((r, i) => i === index ? { ...r, value: e.target.value } : r))} /><Button size="icon-sm" variant="ghost" aria-label={`Remove ${label} row ${index + 1}`} disabled={rows.length === 1 && !row.name && !row.value} onClick={() => onChange(rows.filter((_, i) => i !== index))}><X /></Button></div>)}<Button variant="outline" size="sm" onClick={() => onChange([...rows, { name: '', value: '' }])}><Plus /> Add row</Button></div></Field>
}
export function NodeConfigFields({ type, state, onChange }: { type: string; state: FormState; onChange: (value: FormState) => void }) {
  const c = state.config
  const set = (key: string, value: unknown) => onChange({ ...state, config: { ...c, [key]: value } })
  const input = (label: string, key: string, placeholder = '', hint?: string) => <Field label={label} hint={hint}><Input aria-label={label} placeholder={placeholder} value={text(c[key])} onChange={e => set(key, e.target.value)} /></Field>
  const area = (label: string, key: string, placeholder = '', hint?: string) => <Field label={label} hint={hint}><Textarea aria-label={label} placeholder={placeholder} value={text(c[key])} onChange={e => set(key, e.target.value)} className="min-h-24 font-mono text-xs" /></Field>
  const operation = text(c.operation, 'create')
  const conditions: Config[] = Array.isArray(c.conditions) ? c.conditions : [{ left: '', operator: '==', right: '' }]
  const condition = (index: number, key: string, value: unknown) => set('conditions', conditions.map((item, i) => i === index ? { ...item, [key]: value } : item))
  return <div className="space-y-5">
    {type === 'http' && <>
      <div className="grid grid-cols-[6rem_1fr] gap-2"><Field label="Method"><Choice value={text(c.httpMethod, 'GET')} options={['GET', 'POST', 'PUT', 'PATCH', 'DELETE']} onChange={v => set('httpMethod', v)} /></Field>{input('Request URL', 'url', 'https://api.example.com/items', 'Paste the complete API endpoint, including https://.')}</div>
      <Pairs label="Query parameters" hint="Optional name/value pairs are added after ? in the URL." rows={state.query} onChange={rows => onChange({ ...state, query: rows })} />
      <Pairs label="Headers" hint="Use headers for API keys, such as Authorization: Bearer …" rows={state.headers} onChange={rows => onChange({ ...state, headers: rows })} />
      {['POST', 'PUT', 'PATCH'].includes(text(c.httpMethod, 'GET')) && <Field label="JSON body" hint={'Example: {"name":"Alice"}.'}><Textarea aria-label="JSON body" value={state.body} onChange={e => onChange({ ...state, body: e.target.value })} spellCheck={false} className="min-h-32 font-mono text-xs" /></Field>}
      <p className="text-[11px] text-muted-foreground">Use <code>{'{{ params.id }}'}</code> for a run input or <code>{'{{ input.id }}'}</code> from the previous node in a field value.</p>
    </>}
    {type === 'if' && <>
      <Field label="Match"><Choice value={text(c.combine, 'and')} options={['and', 'or']} onChange={v => set('combine', v)} /></Field>
      <Field label="Conditions" hint="Compare an incoming value such as {{ statusCode }} with 200."><div className="space-y-2">{conditions.map((item, i) => <div key={i} className="space-y-1 rounded-md border p-2"><div className="flex gap-1"><Input aria-label={`Condition ${i + 1} left`} value={String(item.left ?? '')} placeholder="{{ statusCode }}" onChange={e => condition(i, 'left', e.target.value)} className="min-w-0 flex-1 font-mono text-xs" /><Choice value={text(item.operator, '==')} options={['==', '!=', '>', '>=', '<', '<=', 'contains', 'notContains', 'startsWith', 'endsWith', 'exists', 'notExists', 'isEmpty', 'notEmpty', 'isTruthy', 'isFalsy']} onChange={v => condition(i, 'operator', v)} /><Input aria-label={`Condition ${i + 1} right`} value={String(item.right ?? '')} placeholder="200" onChange={e => condition(i, 'right', e.target.value)} className="min-w-0 flex-1 font-mono text-xs" /></div><Button variant="ghost" size="sm" disabled={conditions.length === 1} onClick={() => set('conditions', conditions.filter((_, j) => j !== i))}><X /> Remove</Button></div>)}<Button variant="outline" size="sm" onClick={() => set('conditions', [...conditions, { left: '', operator: '==', right: '' }])}><Plus /> Add condition</Button></div></Field>
    </>}
    {type === 'gemini' && <>{area('Prompt', 'prompt', 'Summarize {{ input.content }}', 'Use {{ input.field }} from an earlier node or {{ params.field }} from test input.')}{area('System instruction', 'systemPrompt', 'You are a helpful assistant.')}{input('Model', 'model', 'gemini-3.1-flash-lite')}<div className="grid grid-cols-2 gap-3"><Field label="Temperature" hint="0 to 2"><Input aria-label="Temperature" type="number" min="0" max="2" step="0.1" value={Number(c.temperature ?? 0.7)} onChange={e => set('temperature', Number(e.target.value))} /></Field><Field label="Max output tokens"><Input aria-label="Max output tokens" type="number" min="1" value={Number(c.maxTokens ?? 1024)} onChange={e => set('maxTokens', Number(e.target.value))} /></Field></div><p className="text-[11px] text-muted-foreground">The server needs GEMINI_API_KEY. No key is stored in this workflow.</p></>}
    {type === 'google_docs' && <><Field label="Operation"><Choice value={operation} options={['create', 'read', 'append']} onChange={v => set('operation', v)} /></Field>{operation === 'create' ? <>{input('Document title', 'title', 'Weekly report')}{area('Initial content', 'content', 'Hello World')}</> : <>{input('Document ID', 'documentId', 'ID from docs.google.com/document/d/…', 'Paste the ID between /d/ and /edit in the document URL.')}{operation === 'append' && area('Text to append', 'text', 'Add this paragraph')}</>}<p className="text-[11px] text-muted-foreground">The server needs GOOGLE_CREDENTIALS with Docs access.</p></>}
    {type === 'google_sheets' && <><Field label="Operation"><Choice value={operation} options={['create', 'read', 'append', 'update']} onChange={v => set('operation', v)} /></Field>{operation === 'create' ? <>{input('Spreadsheet title', 'title', 'Contacts')}<Field label="Initial rows (JSON)" hint={'Each inner array is a row. Example: [["Name","Email"],["Alice","alice@example.com"]].'}><Textarea aria-label="Initial rows (JSON)" value={state.data} onChange={e => onChange({ ...state, data: e.target.value })} spellCheck={false} className="min-h-28 font-mono text-xs" /></Field></> : <>{input('Spreadsheet ID', 'spreadsheetId', 'ID from docs.google.com/spreadsheets/d/…', 'Paste the ID between /d/ and /edit in the spreadsheet URL.')}{input('Range', 'range', operation === 'read' ? 'Sheet1!A1:Z1000' : 'Sheet1!A1', 'Use A1 notation, such as Sheet1!A1:C10.')}{operation !== 'read' && <Field label="Row values (JSON)" hint={'One row as an array, for example ["Alice","alice@example.com"].'}><Textarea aria-label="Row values (JSON)" value={state.values} onChange={e => onChange({ ...state, values: e.target.value })} spellCheck={false} className="min-h-24 font-mono text-xs" /></Field>}</>}<p className="text-[11px] text-muted-foreground">The server needs GOOGLE_CREDENTIALS with Sheets access.</p></>}
  </div>
}
