import { useState, type FormEvent } from 'react'
import { Button, Flash, FormControl, Heading, Text, TextInput } from '@primer/react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'

import { useAuth } from '../auth/AuthContext'
import { BrandLogo } from '../components/brand/BrandLogo'
import styles from './Login.module.css'

export function Login() {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (user) return <Navigate to="/" replace />

  const requestedDestination =
    typeof location.state === 'object' &&
    location.state !== null &&
    typeof (location.state as { from?: unknown }).from === 'string'
      ? (location.state as { from: string }).from
      : '/'
  const destination =
    requestedDestination.startsWith('/') &&
    !requestedDestination.startsWith('//') &&
    !requestedDestination.includes('\\')
      ? requestedDestination
      : '/'

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (submitting) return
    setSubmitting(true)
    setError(null)
    try {
      await login(username.trim(), password)
      navigate(destination, { replace: true })
    } catch (loginError) {
      setError(loginError instanceof Error ? loginError.message : 'Sign in failed.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className={styles.page}>
      <section className={styles.card}>
        <div className={styles.brand}>
          <BrandLogo size={40} className={styles.brandMark} />
          <div>
            <Text weight="semibold">kfleet</Text>
            <span className={styles.brandSub}>Control plane</span>
          </div>
        </div>

        <Heading as="h1" variant="medium" className={styles.title}>
          Sign in
        </Heading>
        <Text className={styles.subtitle}>Use an account provisioned by a kfleet administrator.</Text>

        <form className={styles.form} onSubmit={(event) => void submit(event)}>
          <FormControl required>
            <FormControl.Label>Username</FormControl.Label>
            <TextInput
              block
              autoComplete="username"
              autoFocus
              maxLength={128}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
          </FormControl>

          <FormControl required>
            <FormControl.Label>Password</FormControl.Label>
            <TextInput
              block
              type="password"
              autoComplete="current-password"
              maxLength={72}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </FormControl>

          {error && (
            <Flash variant="danger" role="alert">
              {error}
            </Flash>
          )}

          <Button type="submit" variant="primary" block disabled={submitting}>
            {submitting ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>
      </section>
    </main>
  )
}
