import { Blankslate } from '@primer/react/experimental'
import { ShieldLockIcon } from '@primer/octicons-react'

import layout from '../../styles/layout.module.css'

interface PermissionNoticeProps {
  title: string
  description: string
}

/**
 * Rendered in place of an admin view when the signed-in user's role is below
 * the level the hub requires for the underlying endpoints.
 */
export function PermissionNotice({ title, description }: PermissionNoticeProps) {
  return (
    <div className={layout.box} role="status">
      <Blankslate>
        <Blankslate.Visual>
          <ShieldLockIcon size={24} />
        </Blankslate.Visual>
        <Blankslate.Heading as="h2">{title}</Blankslate.Heading>
        <Blankslate.Description>{description}</Blankslate.Description>
      </Blankslate>
    </div>
  )
}
