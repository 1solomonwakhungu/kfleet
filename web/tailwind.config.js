/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        background: 'var(--color-paper)',
        foreground: 'var(--color-ink)',
        surface: 'var(--color-paper-2)',
        elevated: 'var(--color-paper-3)',
        border: 'var(--color-rule)',
        muted: 'var(--color-muted)',
        accent: 'var(--color-accent)',
        'accent-foreground': 'var(--color-accent-ink)',
        focus: 'var(--color-focus)',
        danger: 'var(--color-danger)',
        'danger-foreground': 'var(--color-danger-ink)',
        'danger-soft': 'var(--color-danger-soft)',
        healthy: 'var(--color-healthy)',
        'healthy-soft': 'var(--color-healthy-soft)',
        degraded: 'var(--color-degraded)',
        'degraded-soft': 'var(--color-degraded-soft)',
        unreachable: 'var(--color-unreachable)',
        'unreachable-soft': 'var(--color-unreachable-soft)',
        unknown: 'var(--color-unknown)',
        'unknown-soft': 'var(--color-unknown-soft)',
      },
      fontFamily: {
        sans: ['IBM Plex Sans', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        display: ['Space Grotesk', 'IBM Plex Sans', 'ui-sans-serif', 'sans-serif'],
        mono: ['JetBrains Mono', 'ui-monospace', 'monospace'],
      },
      transitionTimingFunction: {
        out: 'cubic-bezier(0.16, 1, 0.3, 1)',
      },
    },
  },
  plugins: [],
}
