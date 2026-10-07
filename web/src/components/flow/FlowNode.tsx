import { memo, useEffect, useRef, useState, type MouseEvent } from 'react'
import { Handle, Position, type Node, type NodeProps } from '@xyflow/react'
import { Plus, Settings2, Trash2 } from 'lucide-react'
import { nodeTypeMeta, nodeTypesMeta } from '@/lib/node-meta'
import { cn } from '@/lib/utils'
import { useFlowActions } from './flow-actions'

export type FlowNodeData = {
  nodeId: string
  type: string
  config: Record<string, unknown>
}

type FlowNodeType = Node<FlowNodeData>

function FlowNodeInner({ data, selected }: NodeProps<FlowNodeType>) {
  const meta = nodeTypeMeta(data.type)
  const Icon = meta.icon
  const actions = useFlowActions()
  const [hovered, setHovered] = useState(false)
  const [pickerOpen, setPickerOpen] = useState(false)
  const addButtonRef = useRef<HTMLButtonElement>(null)
  const stopPropagation = (event: MouseEvent<HTMLButtonElement>) => event.stopPropagation()

  useEffect(() => {
    if (!pickerOpen) return
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.stopPropagation()
        setPickerOpen(false)
        addButtonRef.current?.focus()
      }
    }
    window.addEventListener('keydown', onKeyDown, true)
    return () => window.removeEventListener('keydown', onKeyDown, true)
  }, [pickerOpen])

  const showControls = selected || hovered

  return (
    <div
      className="relative"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => {
        setHovered(false)
        setPickerOpen(false)
      }}
    >
      {/* n8n-style hover/selection toolbar */}
      {showControls && (
        <div className="nodrag absolute -top-3.5 right-2 z-20 flex items-center gap-0.5 rounded-md border bg-card p-0.5 shadow-sm">
          <button
            type="button"
            ref={addButtonRef}
            title="Add connected node"
            aria-label="Add connected node"
            aria-expanded={pickerOpen}
            onClick={(e) => { stopPropagation(e); setPickerOpen((v) => !v) }}
            className={cn(
              'grid size-5 place-items-center rounded-[4px] transition-colors duration-150',
              pickerOpen ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground',
            )}
          >
            <Plus className="size-3" strokeWidth={2} />
          </button>
          <button
            type="button"
            title="Open settings"
            aria-label="Open settings"
            onClick={(e) => { stopPropagation(e); actions.openInspector(data.nodeId) }}
            className="grid size-5 place-items-center rounded-[4px] text-muted-foreground transition-colors duration-150 hover:bg-muted hover:text-foreground"
          >
            <Settings2 className="size-3" strokeWidth={2} />
          </button>
          <button
            type="button"
            title="Delete node"
            aria-label="Delete node"
            onClick={(e) => { stopPropagation(e); actions.deleteNode(data.nodeId) }}
            className="grid size-5 place-items-center rounded-[4px] text-muted-foreground transition-colors duration-150 hover:bg-destructive/10 hover:text-destructive"
          >
            <Trash2 className="size-3" strokeWidth={2} />
          </button>
        </div>
      )}

      {/* quick-add node type picker */}
      {pickerOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onPointerDown={() => { setPickerOpen(false); addButtonRef.current?.focus() }}
          />
          <div className="absolute left-1/2 top-full z-30 mt-2 w-60 -translate-x-1/2 rounded-lg border bg-popover p-1 shadow-md">
            <p className="label-caps px-2 pb-1 pt-1.5">Add connected node</p>
            {nodeTypesMeta.map((m) => (
              <button
                key={m.type}
                type="button"
                onClick={(e) => {
                  e.stopPropagation()
                  setPickerOpen(false)
                  void actions.addConnected(data.nodeId, m.type)
                }}
                className="flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-left transition-colors duration-150 hover:bg-muted"
              >
                <span className={`flex size-6 shrink-0 items-center justify-center rounded-[5px] ${m.chip}`}>
                  <m.icon className="size-3.5" strokeWidth={1.75} />
                </span>
                <span className="min-w-0">
                  <span className="block truncate text-[13px] font-medium leading-tight">{m.label}</span>
                  <span className="block truncate text-[11px] text-muted-foreground">{m.description}</span>
                </span>
              </button>
            ))}
          </div>
        </>
      )}

      <Handle type="target" position={Position.Left} isConnectable />
      <div
        className={cn(
          'flex w-56 items-center gap-3 rounded-lg border bg-card p-3 text-left transition-colors duration-150',
          selected ? 'border-ring ring-2 ring-ring/25' : 'hover:border-input',
        )}
      >
        <span className={cn('flex size-9 shrink-0 items-center justify-center rounded-lg', meta.chip)}>
          <Icon className="size-4.5" strokeWidth={1.75} />
        </span>

        <span className="min-w-0 flex-1">
          <span className="block truncate text-[13px] font-medium leading-tight">{data.nodeId}</span>
          <span className="block truncate text-[11px] text-muted-foreground">{meta.label}</span>
        </span>
      </div>
      <Handle type="source" position={Position.Right} isConnectable />
    </div>
  )
}

export const FlowNode = memo(FlowNodeInner)
