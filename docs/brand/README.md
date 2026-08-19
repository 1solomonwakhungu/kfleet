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

## Where the mark is used

| Surface | Asset | Size |
| --- | --- | --- |
| Hub app sidebar (`web`) | `brand/kfleet-mark.svg` | 32 px |
| Hub app mobile topbar | `brand/kfleet-mark.svg` | 24 px |
| Hub app sign-in card | `brand/kfleet-logo{,-dark}.svg` | 40 px |
| Landing page header | `brand/kfleet-mark.svg` | 28 px |
| Landing page footer | `brand/kfleet-mark.svg` | 24 px |
| Landing page social card | inlined in `social-card.svg` | 58 px |
| Primer mockups sidebar / sign-in | `favicon/favicon.svg` | 32 / 48 px |
| Helm charts | `png/kfleet-logo-256.png` via `icon:` | 256 px |
| Repository README | `png/kfleet-logo{,-dark}-128.png` | 128 px |

`web/public/brand/` and `landing-page/public/brand/` are copies of this
directory's SVGs plus the 256 px PNG. Re-copy them after regenerating:

```bash
for app in web landing-page; do
  cp docs/brand/favicon/favicon.svg "$app/public/brand/kfleet-mark.svg"
  cp docs/brand/kfleet-logo*.svg docs/brand/kfleet-lockup*.svg "$app/public/brand/"
  cp docs/brand/png/kfleet-logo-256.png "$app/public/brand/"
done
```

## Regenerating

The SVGs are the source of truth. After editing them, re-render the raster
assets (requires Google Chrome and `python3`):

```bash
./hack/render-brand-assets.sh
```
