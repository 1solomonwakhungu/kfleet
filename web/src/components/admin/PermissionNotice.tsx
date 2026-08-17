import { ShieldAlert } from 'lucide-react'

import { Card, CardContent } from '../ui/card'

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
    <Card className="ring-1 ring-inset ring-border" role="status">
      <CardContent className="grid min-h-64 place-items-center p-6 text-center">
        <div>
          <span className="mx-auto grid size-12 place-items-center rounded-full bg-elevated text-muted ring-1 ring-inset ring-border">
            <ShieldAlert className="size-6" aria-hidden="true" />
          </span>
          <p className="mt-4 font-display text-xl font-bold">{title}</p>
          <p className="mt-2 max-w-lg text-muted">{description}</p>
        </div>
      </CardContent>
    </Card>
  )
}
