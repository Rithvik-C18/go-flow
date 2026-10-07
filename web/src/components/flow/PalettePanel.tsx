import { useState } from 'react'
import { nodeTypesMeta } from '@/lib/node-meta'
import { Input } from '@/components/ui/input'

type PalettePanelProps = {
  /** Adds the node at the center of the visible canvas. */
  onAdd: (type: string) => void
  disabled?: boolean
}

export function PalettePanel({ onAdd, disabled = false }: PalettePanelProps) {
  const [query, setQuery] = useState('')
  const q = query.trim().toLowerCase()
  const items = nodeTypesMeta.filter(
    (m) => m.label.toLowerCase().includes(q) || m.type.toLowerCase().includes(q),
  )

  return (
    <aside className="flex w-60 shrink-0 flex-col border-r bg-card">
      <div className="flex h-11 items-center border-b px-3">
        <span className="label-caps">Add node</span>
      </div>
      <div className="p-2.5">
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search…"
          className="h-8 text-[13px]"
        />
      </div>
      <div className="space-y-1 overflow-y-auto px-2.5 pb-3">
        {items.map((meta) => (
          <button
            key={meta.type}
            type="button"
            draggable
            disabled={disabled}
            aria-disabled={disabled}
            onDragStart={(e) => {
              if (disabled) {
                e.preventDefault()
                return
              }
              e.dataTransfer.setData('application/go-flow-node', meta.type)
              e.dataTransfer.effectAllowed = 'move'
            }}
            onClick={() => !disabled && onAdd(meta.type)}
            title={`Drag onto the canvas or click to add — ${meta.description}`}
            className="flex w-full cursor-grab items-center gap-2.5 rounded-lg border bg-background p-2 text-left transition-colors duration-150 enabled:hover:border-input enabled:hover:bg-muted/60 enabled:active:cursor-grabbing focus-visible:border-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          >
            <span className={`flex size-8 shrink-0 items-center justify-center rounded-md ${meta.chip}`}>
              <meta.icon className="size-4" strokeWidth={1.75} />
            </span>
            <span className="min-w-0">
              <span className="block truncate text-[13px] font-medium leading-tight">{meta.label}</span>
              <span className="block truncate text-[11px] text-muted-foreground">{meta.description}</span>
            </span>
          </button>
        ))}
        {items.length === 0 && (
          <p className="px-1 py-6 text-center text-[13px] text-muted-foreground">No matches</p>
        )}
      </div>
      <p className="mt-auto border-t p-3 text-[11px] leading-relaxed text-muted-foreground">
        Drag a block onto the canvas or click to add it. Hover a node and press + to append the
        next step.
      </p>
    </aside>
  )
}
