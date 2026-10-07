import {
  BrowserRouter,
  Link,
  Navigate,
  Outlet,
  Route,
  Routes,
  useLocation,
} from 'react-router-dom'
import { Toaster } from '@/components/ui/sonner'
import { AuthProvider, useAuth } from '@/lib/auth'
import WorkflowsPage from '@/pages/WorkflowsPage'
import WorkflowDetailPage from '@/pages/WorkflowDetailPage'
import AuthPage from '@/pages/AuthPage'
import { Button } from '@/components/ui/button'
import { LogOut } from 'lucide-react'

function Logo() {
  return (
    <span className="flex size-5 items-center justify-center rounded-[5px] bg-foreground">
      <span className="size-1.5 rounded-full bg-background" />
    </span>
  )
}

function Header() {
  const { user, logout } = useAuth()

  const handleLogout = async () => {
    await logout()
  }

  return (
    <header className="border-b">
      <div className="mx-auto flex h-12 max-w-5xl items-center justify-between px-6">
        <Link to="/" className="flex items-center gap-2.5 rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50">
          <Logo />
          <span className="text-[13.5px] font-medium tracking-tight">Fluxion</span>
        </Link>

        <div className="flex items-center gap-2">
          <span className="hidden font-mono text-xs text-muted-foreground sm:inline">
            {user?.username}
          </span>
          <Button variant="ghost" size="sm" render={<Link to="/" />} className="text-muted-foreground">
            Workflows
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            className="text-muted-foreground hover:text-foreground"
            onClick={() => void handleLogout()}
            title="Sign out"
          >
            <LogOut />
          </Button>
        </div>
      </div>
    </header>
  )
}

function RequireAuth() {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="flex h-24 items-center justify-center font-mono text-xs text-muted-foreground">
        Restoring session…
      </div>
    )
  }

  if (!user) {
    return <Navigate to="/auth" replace state={{ from: `${location.pathname}${location.search}` }} />
  }

  return <Outlet />
}

function PublicOnlyAuth() {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return (
      <div className="flex h-24 items-center justify-center font-mono text-xs text-muted-foreground">
        Restoring session…
      </div>
    )
  }

  if (user) {
    const from = location.state?.from
    const destination = typeof from === 'string' && from.startsWith('/') && !from.startsWith('//') ? from : '/'
    return <Navigate to={destination} replace />
  }

  return <AuthPage />
}

function Shell() {
  const { user } = useAuth()
  const location = useLocation()
  const isAuthenticated = Boolean(user)
  const isEditor = isAuthenticated && /^\/workflows\/[^/]+$/.test(location.pathname)

  if (isEditor) {
    return (
      <main className="flex h-screen min-h-0 w-full flex-col">
        <Outlet />
      </main>
    )
  }

  return (
    <div className="flex min-h-screen flex-col">
      {isAuthenticated && <Header />}
      <main className="mx-auto w-full max-w-5xl flex-1 px-6 py-10">
        <Outlet />
      </main>
      <footer className="border-t">
        <div className="mx-auto flex h-10 max-w-5xl items-center justify-between px-6 text-xs text-muted-foreground">
          <span>Fluxion</span>
          <span className="font-mono">v1.0</span>
        </div>
      </footer>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<Shell />}>
            <Route path="/auth" element={<PublicOnlyAuth />} />
            <Route element={<RequireAuth />}>
              <Route path="/" element={<WorkflowsPage />} />
              <Route path="/workflows/:id" element={<WorkflowDetailPage />} />
            </Route>
          </Route>
        </Routes>

        <Toaster richColors position="top-center" />
      </AuthProvider>
    </BrowserRouter>
  )
}
