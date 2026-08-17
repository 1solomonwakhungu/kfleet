import { useCallback, useEffect, useRef, useState } from 'react'

import { adminApi } from '../lib/adminApi'
import type { UserAccount } from '../types/admin'
import { isAbortError, messageFrom } from '../lib/errors'

export interface UsersState {
  users: UserAccount[]
  loading: boolean
  error: string | null
  reload: () => Promise<void>
  replaceUser: (user: UserAccount) => void
  removeUser: (id: string) => void
}

/** Loads the hub user directory. Admin-only: the hub answers 403 otherwise. */
export function useUsers(enabled = true): UsersState {
  const [users, setUsers] = useState<UserAccount[]>([])
  const [loading, setLoading] = useState(enabled)
  const [error, setError] = useState<string | null>(null)
  const controllerRef = useRef<AbortController | null>(null)

  const reload = useCallback(async () => {
    if (!enabled) return
    controllerRef.current?.abort()
    const controller = new AbortController()
    controllerRef.current = controller
    setLoading(true)
    setError(null)

    try {
      const loaded = await adminApi.listUsers(controller.signal)
      if (controller.signal.aborted) return
      setUsers(loaded)
    } catch (caught) {
      if (!isAbortError(caught)) setError(messageFrom(caught, 'Users could not be loaded.'))
    } finally {
      if (!controller.signal.aborted) setLoading(false)
    }
  }, [enabled])

  useEffect(() => {
    void reload()
    return () => controllerRef.current?.abort()
  }, [reload])

  const replaceUser = useCallback((user: UserAccount) => {
    setUsers((current) => {
      const exists = current.some((candidate) => candidate.id === user.id)
      if (!exists) return [...current, user]
      return current.map((candidate) => (candidate.id === user.id ? user : candidate))
    })
  }, [])

  const removeUser = useCallback((id: string) => {
    setUsers((current) => current.filter((candidate) => candidate.id !== id))
  }, [])

  return { users, loading, error, reload, replaceUser, removeUser }
}
