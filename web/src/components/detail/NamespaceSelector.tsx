import { useId, useMemo } from 'react'
import { Select } from '@primer/react'

interface NamespaceSelectorProps {
  namespaces: string[]
  value?: string
  onChange: (namespace: string | undefined) => void
}

const ALL = '__all__'

export function NamespaceSelector({ namespaces, value, onChange }: NamespaceSelectorProps) {
  const selectId = useId()
  const options = useMemo(
    () => Array.from(new Set(namespaces.filter(Boolean))).sort((a, b) => a.localeCompare(b)),
    [namespaces],
  )

  return (
    <Select
      id={selectId}
      aria-label="Namespace"
      value={value ?? ALL}
      onChange={(event) => onChange(event.target.value === ALL ? undefined : event.target.value)}
    >
      <Select.Option value={ALL}>All namespaces</Select.Option>
      {options.map((namespace) => (
        <Select.Option key={namespace} value={namespace}>
          {namespace}
        </Select.Option>
      ))}
    </Select>
  )
}
