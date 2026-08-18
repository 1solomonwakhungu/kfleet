# kfleet brand assets

The kfleet mark is the **Rancher Helm**: a rancher's hat resting on a Kubernetes
helm wheel. It ties the project's fleet-herding idea ("cattle, not pets") to the
Kubernetes helm.

## Palette

| Role | Light mode | Dark mode |
| --- | --- | --- |
| Helm / primary | `#5E6AD2` | `#8D97F0` |
| Hat / accent | `#E8A33D` | `#F0B455` |
| Hat band | `#8A5220` | `#A2641F` |
| Wordmark | `#10121C` | `#F5F3EF` |

## Files

| File | Use |
| --- | --- |
| `kfleet-logo.svg` | Master mark for light backgrounds |
| `kfleet-logo-dark.svg` | Master mark for dark backgrounds |
| `kfleet-logo-mono.svg` | Single colour, inherits `currentColor` |
| `kfleet-lockup.svg` / `-dark.svg` | Mark + wordmark, horizontal |
| `png/kfleet-logo-{16..256}.png` | Light mark, transparent background |
| `png/kfleet-logo-dark-{16..256}.png` | Dark mark, transparent background |
| `favicon/favicon.svg` | Tiled favicon (indigo rounded square) |
| `favicon/favicon.ico` | Multi-resolution ICO (16/32/48) |
| `favicon/apple-touch-icon.png` | 180×180 home-screen icon |

PNG sizes rendered: 16, 24, 32, 48, 64, 96, 128, 180, 256.

## Usage

- At **32 px and below**, use the tiled `favicon/` assets — the transparent mark
  loses its spokes at that size.
- Keep clear space of at least 1/8 of the mark's height on every side.
- Don't recolour the hat band, rotate the mark, or add effects.
- The favicons are copied into `landing-page/public/` and `web/public/`; re-copy
  them after regenerating.

## Regenerating

The SVGs are the source of truth. After editing them, re-render the raster
assets (requires Google Chrome and `python3`):

```bash
./hack/render-brand-assets.sh
```
