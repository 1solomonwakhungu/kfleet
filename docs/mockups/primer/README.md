# kfleet UI mockups — GitHub Primer

Static, dependency-free mockups of the kfleet control plane restyled with GitHub's
[Primer](https://primer.style) design language. They are exploration artifacts, not shipped UI: nothing here is wired to
the hub API and no application code is affected.

## Screens

| File | Screen |
| --- | --- |
| `index.html` | Gallery / entry point |
| `dashboard.html` | Fleet dashboard — summary metrics, filters, cluster cards |
| `cluster-detail.html` | Cluster detail — resource tabs, pod logs, operational timeline |
| `alerts.html` | Alerts — active health transitions and resolved history |
| `agents.html` | Agents — pending approvals, registration token, connected agents |
| `policies.html` | Policy and drift — compliance rollup and failing checks |
| `users.html` | Users — accounts and roles |
| `audit.html` | Audit log — append-only security history |
| `login.html` | Sign in |

## Viewing

Open `index.html` directly in a browser, or serve the folder:

```sh
python3 -m http.server 4173 --directory docs/mockups/primer
```

Then visit <http://localhost:4173/>. The button in the top right toggles between the Primer light and dark themes; the
choice is remembered in `localStorage`.

## How they are built

- **Color, spacing, radii** come from `@primer/primitives` functional tokens loaded from a CDN
  (`--bgColor-*`, `--fgColor-*`, `--borderColor-*`, `--button-*`). No hex values are hardcoded, which is why both themes
  work from the same markup.
- **`assets/mockup.css`** implements the Primer component patterns used here (Box, buttons, labels, state labels,
  counters, underline tabs, flash, timeline, tables) on top of those tokens.
- **`assets/shell.js`** renders the shared sidebar, breadcrumb topbar, and Octicon set, and handles the theme toggle.
  Pages declare their state with `data-page`, `data-breadcrumb`, and `data-shell="none"` for full-bleed screens.

To turn a mockup into real UI, the equivalent production path is `@primer/react` plus `@primer/primitives`, replacing the
current Tailwind + shadcn setup in `web/`.
