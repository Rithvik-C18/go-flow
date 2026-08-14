import { BrowserRouter, Link, Route, Routes } from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import WorkflowsPage from '@/pages/WorkflowsPage'
import WorkflowDetailPage from '@/pages/WorkflowDetailPage'

export default function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-background">
        <header className="border-b">
          <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-4">
            <Link to="/" className="text-lg font-semibold tracking-tight">
              go<span className="text-primary">-flow</span>
            </Link>
            <span className="text-sm text-muted-foreground">DAG workflow engine</span>
          </div>
        </header>

        <main className="mx-auto max-w-5xl px-4 py-8">
          <Routes>
            <Route path="/" element={<WorkflowsPage />} />
            <Route path="/workflows/:id" element={<WorkflowDetailPage />} />
          </Routes>
        </main>

        <Toaster richColors position="top-center" />
      </div>
    </BrowserRouter>
  )
}
