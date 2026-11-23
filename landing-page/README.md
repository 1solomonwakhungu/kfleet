# kfleet landing page

The public marketing site for kfleet, built with React, Vite, Tailwind CSS v4, and shadcn/ui components (`base-nova` style on Base UI primitives).

All copy lives in `src/content/siteCopy.ts`. Page sections are in `src/components/sections/`, shadcn/ui primitives in `src/components/ui/`, and theme tokens in `src/index.css`.

```bash
npm install
npm run dev
```

Build the production bundle with `npm run build`. Vercel should use `landing-page` as the project root.
