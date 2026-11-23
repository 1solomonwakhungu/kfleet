# kfleet landing page

The public marketing site for kfleet, built with React, Vite, Tailwind CSS v4, and the **Tonal** editorial design system (Google Sans, flat tonal surfaces, 24px bento tiles, and Soft Pill controls).

All copy lives in `src/content/siteCopy.ts`. Page sections are in `src/components/sections/`, Tonal primitives in `src/components/tonal/`, and Tonal design tokens in `src/styles/tonal-tokens.css`.

```bash
npm install
npm run dev
```

Build the production bundle with `npm run build`. Vercel should use `landing-page` as the project root.
