import { Button } from '@primer/react'
import { Blankslate } from '@primer/react/experimental'
import { AlertIcon } from '@primer/octicons-react'
import { Link } from 'react-router-dom'

import { useDocumentTitle } from '../hooks/useDocumentTitle'
import layout from '../styles/layout.module.css'
import styles from './NotFound.module.css'

export default function NotFound() {
  useDocumentTitle('Page not found · kfleet')

  return (
    <main className={layout.page}>
      <div className={layout.box}>
        <Blankslate>
          <Blankslate.Visual>
            <AlertIcon size={24} />
          </Blankslate.Visual>
          <Blankslate.Heading as="h1">Page not found</Blankslate.Heading>
          <Blankslate.Description>
            This address does not match any kfleet page. It may have been moved or mistyped.
          </Blankslate.Description>
          <div className={styles.action}>
            <Button as={Link} to="/" variant="primary">
              Back to dashboard
            </Button>
          </div>
        </Blankslate>
      </div>
    </main>
  )
}
