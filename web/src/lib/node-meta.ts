import {
  Globe,
  GitBranch,
  Sparkles,
  FileText,
  Table2,
  type LucideIcon,
} from 'lucide-react'
import { nodeTemplates } from '@/lib/api'

export type NodeTypeMeta = {
  type: string
  label: string
  description: string
  icon: LucideIcon
  /** tailwind classes for the icon chip */
  chip: string
}

export const nodeTypesMeta: NodeTypeMeta[] = [
  {
    type: 'http',
    label: 'HTTP Request',
    description: 'Call any REST API',
    icon: Globe,
    chip: 'bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-400',
  },
  {
    type: 'if',
    label: 'Condition (If)',
    description: 'Compare values and output true or false',
    icon: GitBranch,
    chip: 'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-400',
  },
  {
    type: 'gemini',
    label: 'Gemini AI',
    description: 'Generate text with Gemini',
    icon: Sparkles,
    chip: 'bg-violet-100 text-violet-700 dark:bg-violet-500/15 dark:text-violet-400',
  },
  {
    type: 'google_docs',
    label: 'Google Docs',
    description: 'Create or append documents',
    icon: FileText,
    chip: 'bg-blue-100 text-blue-700 dark:bg-blue-500/15 dark:text-blue-400',
  },
  {
    type: 'google_sheets',
    label: 'Google Sheets',
    description: 'Create or update spreadsheets',
    icon: Table2,
    chip: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400',
  },
]

const metaByType = new Map(nodeTypesMeta.map((m) => [m.type, m]))

export function nodeTypeMeta(type: string): NodeTypeMeta {
  return (
    metaByType.get(type) ?? {
      type,
      label: type,
      description: '',
      icon: Globe,
      chip: 'bg-muted text-muted-foreground',
    }
  )
}

export function defaultConfigFor(type: string): string {
  return nodeTemplates[type] ?? '{}'
}
