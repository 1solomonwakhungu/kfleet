import { render, screen, waitFor } from '@testing-library/react'
import { useEffect } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { useDocumentTitle } from './useDocumentTitle'

function Page({ title }: { title: string }) {
  useDocumentTitle(title)
  return <main>{title}</main>
}

describe('useDocumentTitle', () => {
  afterEach(() => {
    document.title = 'kfleet'
  })

  it('sets the document title on mount', () => {
    render(<Page title="Dashboard · kfleet" />)

    expect(document.title).toBe('Dashboard · kfleet')
    expect(screen.getByRole('main')).toBeTruthy()
  })

  it('updates the title when it changes', async () => {
    const { rerender } = render(<Page title="Loading · kfleet" />)

    rerender(<Page title="prod-cluster · kfleet" />)

    await waitFor(() => expect(document.title).toBe('prod-cluster · kfleet'))
  })

  it('restores the previous title on unmount', () => {
    document.title = 'kfleet'
    const { unmount } = render(<Page title="Agents · kfleet" />)

    unmount()

    expect(document.title).toBe('kfleet')
  })
})
