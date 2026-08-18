#!/usr/bin/env bash
# Renders the kfleet brand PNG/ICO assets from the master SVGs.
# Requires Google Chrome (headless) and python3.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BRAND="$ROOT/docs/brand"
CHROME="${CHROME:-/Applications/Google Chrome.app/Contents/MacOS/Google Chrome}"
[ -x "$CHROME" ] || CHROME="$(command -v google-chrome || command -v chromium)"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

render() { # svg out size
  local svg="$1" out="$2" size="$3"
  cat > "$TMP/page.html" <<HTML
<!doctype html><meta charset="utf-8">
<style>html,body{margin:0;padding:0;background:transparent}
img{display:block;width:${size}px;height:${size}px;image-rendering:auto}</style>
<img src="file://$svg">
HTML
  "$CHROME" --headless --disable-gpu --hide-scrollbars \
    --default-background-color=00000000 --force-device-scale-factor=1 \
    --window-size="$size,$size" --screenshot="$out" \
    "file://$TMP/page.html" >/dev/null 2>&1
}

SIZES="16 24 32 48 64 96 128 180 256"

for size in $SIZES; do
  render "$BRAND/kfleet-logo.svg"      "$BRAND/png/kfleet-logo-${size}.png"      "$size"
  render "$BRAND/kfleet-logo-dark.svg" "$BRAND/png/kfleet-logo-dark-${size}.png" "$size"
done

for size in 16 32 48 64 180 256; do
  render "$BRAND/favicon/favicon.svg" "$BRAND/favicon/favicon-${size}.png" "$size"
done
cp "$BRAND/favicon/favicon-180.png" "$BRAND/favicon/apple-touch-icon.png"

python3 - "$BRAND/favicon" <<'PY'
import struct, sys, pathlib
d = pathlib.Path(sys.argv[1])
sizes = [16, 32, 48]
imgs = [(s, (d / f"favicon-{s}.png").read_bytes()) for s in sizes]
out = bytearray(struct.pack("<HHH", 0, 1, len(imgs)))
offset = 6 + 16 * len(imgs)
for s, data in imgs:
    out += struct.pack("<BBBBHHII", s % 256, s % 256, 0, 0, 1, 32, len(data), offset)
    offset += len(data)
for _, data in imgs:
    out += data
(d / "favicon.ico").write_bytes(bytes(out))
print("wrote favicon.ico", len(out), "bytes")
PY

echo "Brand assets rendered into $BRAND"
