import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { BaseStyles, ThemeProvider } from '@primer/react'

export type ColorMode = 'auto' | 'light' | 'dark'

const STORAGE_KEY = 'kfleet-color-mode'

interface ColorModeContextValue {
  colorMode: ColorMode
  setColorMode: (mode: ColorMode) => void
  toggleColorMode: () => void
  resolvedMode: 'light' | 'dark'
}

const ColorModeContext = createContext<ColorModeContextValue | null>(null)

function readStoredMode(): ColorMode {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark' || stored === 'auto') return stored
  } catch {
    // Private browsing and embedded previews can block storage access.
  }
  return 'auto'
}

function prefersDark(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-color-scheme: dark)').matches
}

/**
 * Wires Primer's ThemeProvider to a persisted color mode preference. `auto`
 * follows the operating system, and an explicit choice is remembered.
 */
export function ColorModeProvider({ children }: { children: ReactNode }) {
  const [colorMode, setColorModeState] = useState<ColorMode>(readStoredMode)
  const [systemDark, setSystemDark] = useState(prefersDark)

  useEffect(() => {
    if (typeof window.matchMedia !== 'function') return

    const query = window.matchMedia('(prefers-color-scheme: dark)')
    const sync = (event: MediaQueryListEvent) => setSystemDark(event.matches)
    query.addEventListener('change', sync)
    return () => query.removeEventListener('change', sync)
  }, [])

  const setColorMode = useCallback((mode: ColorMode) => {
    setColorModeState(mode)
    try {
      window.localStorage.setItem(STORAGE_KEY, mode)
    } catch {
      // Preference is best-effort only.
    }
  }, [])

  const resolvedMode: 'light' | 'dark' = colorMode === 'auto' ? (systemDark ? 'dark' : 'light') : colorMode

  useEffect(() => {
    document.documentElement.style.colorScheme = resolvedMode
  }, [resolvedMode])

  const value = useMemo<ColorModeContextValue>(
    () => ({
      colorMode,
      setColorMode,
      resolvedMode,
      toggleColorMode: () => setColorMode(resolvedMode === 'dark' ? 'light' : 'dark'),
    }),
    [colorMode, resolvedMode, setColorMode],
  )

  return (
    <ColorModeContext.Provider value={value}>
      <ThemeProvider colorMode={colorMode === 'auto' ? 'auto' : resolvedMode} dayScheme="light" nightScheme="dark">
        <BaseStyles>{children}</BaseStyles>
      </ThemeProvider>
    </ColorModeContext.Provider>
  )
}

export function useColorMode(): ColorModeContextValue {
  const context = useContext(ColorModeContext)
  if (context) return context

  // Tests render individual pages without the provider; fall back to a no-op.
  return {
    colorMode: 'auto',
    resolvedMode: 'light',
    setColorMode: () => {},
    toggleColorMode: () => {},
  }
}
