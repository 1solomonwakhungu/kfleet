import { useEffect } from 'react'

/**
 * Sets the browser tab title while the calling page is mounted and restores
 * the previous title on unmount, so navigating away never leaves a stale page
 * name behind.
 */
export function useDocumentTitle(title: string) {
  useEffect(() => {
    const previousTitle = document.title
    document.title = title
    return () => {
      document.title = previousTitle
    }
  }, [title])
}