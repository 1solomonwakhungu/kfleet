import { useColorMode } from '../../theme/ColorModeProvider'

/**
 * Brand guidance (docs/brand/README.md): the transparent mark loses its helm
 * spokes at 32px and below, so small sizes render the tiled asset instead.
 */
const TILE_THRESHOLD = 32

interface BrandLogoProps {
  size?: number
  className?: string
  /** Rendered as alt text. Empty marks the image decorative next to a wordmark. */
  alt?: string
}

export function BrandLogo({ size = 32, className, alt = '' }: BrandLogoProps) {
  const { resolvedMode } = useColorMode()
  const source =
    size <= TILE_THRESHOLD
      ? '/brand/kfleet-mark.svg'
      : resolvedMode === 'dark'
        ? '/brand/kfleet-logo-dark.svg'
        : '/brand/kfleet-logo.svg'

  return (
    <img
      className={className}
      src={source}
      width={size}
      height={size}
      alt={alt}
      aria-hidden={alt === '' ? true : undefined}
      draggable={false}
    />
  )
}
