import { Children, isValidElement, useState, type ReactNode } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { Loader2 } from 'lucide-react'
import { ApiError } from '@/lib/api'
import { useAuth } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

type FieldErrors = Partial<Record<'identifier' | 'email' | 'username' | 'password', string>>

const USERNAME_PATTERN = /^[a-zA-Z0-9_-]+$/
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function getRegistrationErrors(username: string, email: string, password: string): FieldErrors {
  const errors: FieldErrors = {}

  if (username.trim().length < 3 || username.trim().length > 32) {
    errors.username = 'Username must be 3–32 characters'
  } else if (!USERNAME_PATTERN.test(username.trim())) {
    errors.username = 'Use letters, numbers, hyphens and underscores only'
  }

  if (!EMAIL_PATTERN.test(email)) {
    errors.email = 'Enter a valid email address'
  }

  if (password.length < 8) {
    errors.password = 'Password must be at least 8 characters'
  }

  return errors
}

export default function AuthPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login, register } = useAuth()
  const [isLogin, setIsLogin] = useState(true)
  const [identifier, setIdentifier] = useState('')
  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)

  const switchMode = () => {
    setIsLogin((current) => !current)
    setFieldErrors({})
    setFormError(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)

    let errors: FieldErrors
    if (isLogin) {
      errors = {
        identifier: identifier.trim() ? undefined : 'Enter your username or email',
        password: password ? undefined : 'Enter your password',
      }
    } else {
      errors = getRegistrationErrors(username, email, password)
    }

    setFieldErrors(errors)
    if (Object.values(errors).some(Boolean)) {
      return
    }

    setLoading(true)

    try {
      if (isLogin) {
        await login({ identifier: identifier.trim(), password })
        toast.success('Welcome back')
      } else {
        await register({ username: username.trim(), email: email.trim(), password })
        toast.success('Account created successfully')
      }

      const from = location.state?.from
      navigate(typeof from === 'string' && from.startsWith('/') && !from.startsWith('//') ? from : '/')
    } catch (err) {
      const message =
        err instanceof ApiError && err.status === 0
          ? 'Cannot reach the server. Check your connection and try again.'
          : (err as Error).message

      setFormError(message)
      toast.error(message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[320px] pt-16">
      <div className="mb-8">
        <p className="font-mono text-xs uppercase tracking-[0.08em] text-muted-foreground">
          Fluxion / workspace
        </p>
        <h1 className="mt-2 text-lg font-semibold tracking-tight">
          {isLogin ? 'Sign in' : 'Create account'}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {isLogin ? 'Access your workflow automation environment.' : 'Start building automated workflows.'}
        </p>
      </div>

      <form onSubmit={handleSubmit} noValidate className="space-y-4">
        {isLogin ? (
          <Field label="Username or email" error={fieldErrors.identifier}>
            <Input
              id="identifier"
              value={identifier}
              onChange={(event) => setIdentifier(event.target.value)}
              placeholder="username or you@example.com"
              autoComplete="username"
              aria-invalid={Boolean(fieldErrors.identifier)}
              aria-describedby={fieldErrors.identifier ? 'identifier-error' : undefined}
              disabled={loading}
              autoFocus
            />
          </Field>
        ) : (
          <>
            <Field label="Username" error={fieldErrors.username}>
              <Input
                id="username"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                placeholder="ada_lovelace"
                autoComplete="username"
                aria-invalid={Boolean(fieldErrors.username)}
                aria-describedby={fieldErrors.username ? 'username-error' : undefined}
                disabled={loading}
                autoFocus
              />
            </Field>

            <Field label="Email" error={fieldErrors.email}>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="you@example.com"
                autoComplete="email"
                aria-invalid={Boolean(fieldErrors.email)}
                aria-describedby={fieldErrors.email ? 'email-error' : undefined}
                disabled={loading}
              />
            </Field>
          </>
        )}

        <Field label="Password" error={fieldErrors.password}>
          <Input
            id="password"
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            autoComplete={isLogin ? 'current-password' : 'new-password'}
            aria-invalid={Boolean(fieldErrors.password)}
            aria-describedby={fieldErrors.password ? 'password-error' : undefined}
            disabled={loading}
          />
        </Field>

        {formError && (
          <div role="alert" className="border border-destructive/30 bg-destructive/10 px-3 py-2 text-[13px] text-destructive">
            {formError}
          </div>
        )}

        <Button type="submit" size="lg" className="w-full font-medium" disabled={loading}>
          {loading && <Loader2 className="animate-spin" />}
          {loading ? 'Please wait…' : isLogin ? 'Sign in' : 'Create account'}
        </Button>
      </form>

      <p className="mt-6 border-t pt-4 text-center text-[13px] text-muted-foreground">
        {isLogin ? 'No account? ' : 'Already registered? '}
        <button
          type="button"
          onClick={switchMode}
          className="rounded-sm font-medium text-foreground underline-offset-4 transition-colors outline-none hover:underline focus-visible:ring-2 focus-visible:ring-ring/50"
          disabled={loading}
        >
          {isLogin ? 'Sign up' : 'Sign in'}
        </button>
      </p>

      <p className="mt-10 text-center font-mono text-xs leading-relaxed text-muted-foreground">
        HTTP · conditions · Gemini · Google Workspace
      </p>
    </div>
  )
}

function Field({
  label,
  error,
  children,
}: {
  label: string
  error?: string
  children: ReactNode
}) {
  const child = Children.only(children)
  const id = isValidElement<{ id?: string }>(child) ? child.props.id : undefined

  return (
    <div className="space-y-1.5">
      <Label htmlFor={id} className="text-[13px]">{label}</Label>
      {children}
      {error && (
        <p id={`${id}-error`} className="text-xs text-destructive">
          {error}
        </p>
      )}
    </div>
  )
}
