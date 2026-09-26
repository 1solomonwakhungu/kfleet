import { fireEvent, render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { usePodLogs } from '../../hooks/usePodLogs'
import type { PodInfo } from '../../types/resources'
import { LogsTab } from './LogsTab'

vi.mock('../../hooks/usePodLogs', () => ({ usePodLogs: vi.fn() }))

const pods: PodInfo[] = [
  {
    name: 'api-7d9f',
    namespace: 'default',
    phase: 'Running',
    nodeName: 'node-1',
    restartCount: 0,
    ready: true,
    startTime: new Date().toISOString(),
  },
  {
    name: 'ingress-controller',
    namespace: 'kube-system',
    phase: 'Running',
    nodeName: 'node-2',
    restartCount: 1,
    ready: true,
    startTime: new Date().toISOString(),
  },
  {
    name: 'metrics-agent',
    namespace: 'kube-system',
    phase: 'Running',
    nodeName: 'node-1',
    restartCount: 0,
    ready: false,
    startTime: new Date().toISOString(),
  },
]

function openCombobox(label: string) {
  fireEvent.focus(screen.getByRole('combobox', { name: label }))
  return screen.getByRole('listbox', { name: label })
}

describe('LogsTab namespace and pod selectors', () => {
  const onSelectPod = vi.fn()
  const clear = vi.fn()

  beforeEach(() => {
    vi.mocked(usePodLogs).mockReset()
    vi.mocked(usePodLogs).mockReturnValue({
      lines: [],
      status: 'idle',
      error: null,
      clear,
      retry: vi.fn(),
    })
    onSelectPod.mockReset()
    clear.mockReset()
  })

  it('shows a searchable namespace dropdown with all namespaces', () => {
    render(<LogsTab clusterId="c1" pods={pods} selectedPod={undefined} onSelectPod={onSelectPod} />)
    const list = openCombobox('Namespace for log stream')
    expect(within(list).getByText('All namespaces')).toBeTruthy()
    expect(within(list).getByText('default')).toBeTruthy()
    expect(within(list).getByText('kube-system')).toBeTruthy()
  })

  it('filters pod options when a namespace is chosen', () => {
    render(<LogsTab clusterId="c1" pods={pods} selectedPod={undefined} onSelectPod={onSelectPod} />)
    fireEvent.focus(screen.getByRole('combobox', { name: 'Namespace for log stream' }))
    fireEvent.click(within(screen.getByRole('listbox', { name: 'Namespace for log stream' })).getByText('kube-system'))

    const list = openCombobox('Pod for log stream')
    expect(within(list).getByText('kube-system/ingress-controller')).toBeTruthy()
    expect(within(list).queryByText('default/api-7d9f')).toBeNull()
  })

  it('supports type-to-search in the pod dropdown', () => {
    render(<LogsTab clusterId="c1" pods={pods} selectedPod={undefined} onSelectPod={onSelectPod} />)
    const input = screen.getByRole('combobox', { name: 'Pod for log stream' })
    fireEvent.focus(input)
    fireEvent.change(input, { target: { value: 'ingress' } })
    const list = screen.getByRole('listbox', { name: 'Pod for log stream' })
    expect(within(list).getByText('kube-system/ingress-controller')).toBeTruthy()
    expect(within(list).queryByText('default/api-7d9f')).toBeNull()
  })

  it('supports type-to-search in the namespace dropdown', () => {
    render(<LogsTab clusterId="c1" pods={pods} selectedPod={undefined} onSelectPod={onSelectPod} />)
    const input = screen.getByRole('combobox', { name: 'Namespace for log stream' })
    fireEvent.focus(input)
    fireEvent.change(input, { target: { value: 'kube' } })
    const list = screen.getByRole('listbox', { name: 'Namespace for log stream' })
    expect(within(list).getByText('kube-system')).toBeTruthy()
    expect(within(list).queryByText('default')).toBeNull()
  })

  it('selecting a pod reports it to the parent and adopts its namespace', () => {
    render(<LogsTab clusterId="c1" pods={pods} selectedPod={undefined} onSelectPod={onSelectPod} />)
    const podInput = screen.getByRole('combobox', { name: 'Pod for log stream' })
    fireEvent.focus(podInput)
    fireEvent.click(within(screen.getByRole('listbox', { name: 'Pod for log stream' })).getByText('default/api-7d9f'))
    expect(onSelectPod).toHaveBeenCalledWith(pods[0])
  })
})
