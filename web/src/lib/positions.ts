export type NodePosition = { x: number; y: number }
type PositionMap = Record<string, NodePosition>

const key = (workflowId: string) => `go-flow:positions:${workflowId}`

export function loadPositions(workflowId: string): PositionMap {
  try {
    const raw = localStorage.getItem(key(workflowId))
    if (!raw) return {}
    const parsed: unknown = JSON.parse(raw)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}
    return Object.fromEntries(
      Object.entries(parsed as Record<string, unknown>).flatMap(([id, value]) => {
        if (!value || typeof value !== 'object' || Array.isArray(value)) return []
        const { x, y } = value as Record<string, unknown>
        return typeof x === 'number' && Number.isFinite(x) && typeof y === 'number' && Number.isFinite(y)
          ? [[id, { x, y }] as [string, NodePosition]]
          : []
      }),
    )
  } catch {
    return {}
  }
}

export function savePositions(workflowId: string, positions: PositionMap): void {
  try {
    localStorage.setItem(key(workflowId), JSON.stringify(positions))
  } catch {
    // storage full or unavailable — positions are best-effort
  }
}

/** Simple layered layout used when a node has no saved position yet. */
export function autoLayout(
  nodes: { id: string }[],
  edges: { from: string; to: string }[],
): PositionMap {
  const depth = new Map<string, number>()
  const incoming = new Map<string, string[]>()
  for (const e of edges) {
    incoming.set(e.to, [...(incoming.get(e.to) ?? []), e.from])
  }

  const resolve = (id: string, seen = new Set<string>()): number => {
    if (depth.has(id)) return depth.get(id)!
    if (seen.has(id)) return 0
    seen.add(id)
    const parents = incoming.get(id) ?? []
    const d = parents.length === 0 ? 0 : Math.max(...parents.map((p) => resolve(p, seen))) + 1
    depth.set(id, d)
    return d
  }
  nodes.forEach((n) => resolve(n.id))

  const columns = new Map<number, string[]>()
  for (const n of nodes) {
    const d = depth.get(n.id) ?? 0
    columns.set(d, [...(columns.get(d) ?? []), n.id])
  }

  const positions: PositionMap = {}
  for (const [d, ids] of columns) {
    ids.forEach((id, row) => {
      positions[id] = { x: d * 300 + 60, y: row * 140 + 60 }
    })
  }
  return positions
}
