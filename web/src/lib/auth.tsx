import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import {
  api,
  clearAccessToken,
  getAccessToken,
  getAccessTokenClaims,
  setAccessToken,
  subscribeToAccessToken,
  type User,
} from '@/lib/api'

const CACHED_USER_KEY = 'go-flow:user'

type Credentials = {
  identifier: string
  password: string
}

type RegistrationDetails = {
  username: string
  email: string
  password: string
}

type AuthState = {
  user: User | null
  loading: boolean
  login: (credentials: Credentials) => Promise<User>
  register: (details: RegistrationDetails) => Promise<User>
  logout: () => Promise<void>
}

function getCachedUser(): User | null {
  try {
    const value = localStorage.getItem(CACHED_USER_KEY)
    if (!value) return null

    const parsed = JSON.parse(value) as Partial<User>
    const user: User | null =
      typeof parsed.id === 'number' &&
      typeof parsed.username === 'string' &&
      typeof parsed.email === 'string'
        ? { id: parsed.id, username: parsed.username, email: parsed.email }
        : null

    return user
  } catch {
    return null
  }
}

function setCachedUser(user: User): void {
  localStorage.setItem(CACHED_USER_KEY, JSON.stringify(user))
}

function clearCachedUser(): void {
  localStorage.removeItem(CACHED_USER_KEY)
}

function getUserFromToken(token: string | null): User | null {
  if (!token) return null

  const claims = getAccessTokenClaims(token)
  if (!claims) {
    clearAccessToken()
    clearCachedUser()
    return null
  }

  const cachedUser = getCachedUser()
  return cachedUser?.id === claims.uid
    ? cachedUser
    : { id: claims.uid, username: claims.username, email: '' }
}

function validateAuthResponse(value: unknown): asserts value is { access_token: string; user: User } {
  if (
    !value ||
    typeof value !== 'object' ||
    !('access_token' in value) ||
    typeof value.access_token !== 'string' ||
    !('user' in value) ||
    !value.user ||
    typeof value.user !== 'object'
  ) {
    throw new Error('The server returned an invalid authentication response')
  }
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => getUserFromToken(getAccessToken()))
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    let expirationTimer: number | undefined

    const synchronizeUser = (token: string | null) => {
      const nextUser = getUserFromToken(token)
      setUser(nextUser)

      if (expirationTimer) {
        window.clearTimeout(expirationTimer)
      }

      if (!token) return

      const claims = getAccessTokenClaims(token)
      if (!claims) return

      expirationTimer = window.setTimeout(
        () => {
          clearAccessToken()
          clearCachedUser()
          setUser(null)
        },
        Math.max(0, claims.exp * 1000 - Date.now()),
      )
    }

    synchronizeUser(getAccessToken())
    const unsubscribe = subscribeToAccessToken(synchronizeUser)

    return () => {
      unsubscribe()
      if (expirationTimer) {
        window.clearTimeout(expirationTimer)
      }
    }
  }, [])

  const login = useCallback(async ({ identifier, password }: Credentials) => {
    setLoading(true)

    try {
      const response = await api.login(identifier, password)
      validateAuthResponse(response)
      setAccessToken(response.access_token)
      setCachedUser(response.user)
      setUser(response.user)
      return response.user
    } finally {
      setLoading(false)
    }
  }, [])

  const register = useCallback(async ({ username, email, password }: RegistrationDetails) => {
    setLoading(true)

    try {
      const response = await api.register(username, email, password)
      validateAuthResponse(response)
      setAccessToken(response.access_token)
      setCachedUser(response.user)
      setUser(response.user)
      return response.user
    } finally {
      setLoading(false)
    }
  }, [])

  const logout = useCallback(async () => {
    try {
      await api.logout()
    } finally {
      clearAccessToken()
      clearCachedUser()
      setUser(null)
    }
  }, [])

  const value = useMemo(
    () => ({ user, loading, login, register, logout }),
    [user, loading, login, register, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
