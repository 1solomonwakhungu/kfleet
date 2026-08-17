import { useMemo } from 'react'
import { Label, Text } from '@primer/react'
import { GlobeIcon } from '@primer/octicons-react'

import type { ServiceInfo } from '../../types/resources'
import { ResourceState, ResourceTablePanel, ResourceTableSkeleton } from './ResourceTabState'
import styles from './resource.module.css'

interface ServicesTabProps {
  services: ServiceInfo[]
  loading: boolean
  error: string | null
  search: string
}

export function ServicesTab({ services, loading, error, search }: ServicesTabProps) {
  const query = search.trim().toLowerCase()
  const filtered = useMemo(
    () =>
      services.filter((service) =>
        [
          service.name,
          service.namespace,
          service.type,
          service.clusterIP,
          ...service.ports.flatMap((port) => [port.name, port.protocol, String(port.port), String(port.targetPort)]),
        ].some((value) => value.toLowerCase().includes(query)),
      ),
    [services, query],
  )

  if (error) {
    return <ResourceState kind="error" title="Unable to load services" description={error} />
  }
  if (loading && services.length === 0) {
    return <ResourceTableSkeleton label="Loading services" columns={6} />
  }
  if (filtered.length === 0) {
    return (
      <ResourceState
        kind="empty"
        title={query ? 'No matching services' : 'No services found'}
        description={
          query
            ? `No service matches “${search.trim()}”. Try a name, namespace, type, IP, or port.`
            : 'No services were returned for this namespace.'
        }
      />
    )
  }

  return (
    <ResourceTablePanel label="Services" count={filtered.length} noun="service">
      <table className={styles.table}>
        <caption className={styles.srOnly}>Services, their types, cluster endpoints, port mappings, and age</caption>
        <thead>
          <tr>
            <th scope="col">Name</th>
            <th scope="col">Namespace</th>
            <th scope="col">Type</th>
            <th scope="col">Cluster endpoint</th>
            <th scope="col">Ports</th>
            <th scope="col" className={styles.numeric}>
              Age
            </th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((service) => {
            const headless = !service.clusterIP || service.clusterIP.toLowerCase() === 'none'

            return (
              <tr key={`${service.namespace}/${service.name}`}>
                <td>
                  <span className={styles.readiness}>
                    <GlobeIcon size={16} className={styles.panelIcon} />
                    <Text weight="semibold" className={styles.nameCell}>
                      {service.name}
                    </Text>
                  </span>
                </td>
                <td className={styles.muted}>{service.namespace}</td>
                <td>
                  <Label variant="secondary">{service.type || 'Unknown'}</Label>
                </td>
                <td className={`${styles.mono} ${styles.muted}`}>
                  {headless ? (
                    'Headless'
                  ) : service.ports.length > 0 ? (
                    service.ports.map((port, index) => (
                      <div key={`${port.name || 'port'}-${port.port}-${index}`}>
                        {service.clusterIP}:{port.port}
                      </div>
                    ))
                  ) : (
                    service.clusterIP
                  )}
                </td>
                <td>
                  {service.ports.length > 0 ? (
                    <div className={styles.labels}>
                      {service.ports.map((port, index) => (
                        <Label
                          key={`${port.name || 'port'}-${port.port}-${index}`}
                          variant="secondary"
                          title={`${port.port} routes to target port ${port.targetPort} over ${port.protocol}`}
                        >
                          {port.name ? `${port.name} · ` : ''}
                          {port.port}→{port.targetPort}/{port.protocol}
                        </Label>
                      ))}
                    </div>
                  ) : (
                    <span className={styles.muted}>No ports</span>
                  )}
                </td>
                <td className={`${styles.numeric} ${styles.muted}`}>{service.age || '—'}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </ResourceTablePanel>
  )
}
