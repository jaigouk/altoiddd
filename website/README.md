# AltoiDDD Website

Static site for [jaigouk.com/altoiddd](https://jaigouk.com/altoiddd), built with [Astro](https://astro.build/) + [Starlight](https://starlight.astro.build/) + [Tailwind](https://tailwindcss.com/) + [daisyUI](https://daisyui.com/).

## Development

```bash
cd website
npm install
npm run dev        # http://localhost:4321
```

## Build

```bash
npm run build      # static output in dist/
npm run preview    # preview the build
```

## Content

Guide content lives in `../docs/guides/*.md` and is loaded via Astro's `glob()` content loader. Edit the markdown files there — they'll appear in the site automatically.

## Design System

- **Palette:** Midnight Teal (oklch) — see `src/styles/global.css`
- **Fonts:** Inter Variable, JetBrains Mono, Inter Tight
- **Components:** `src/components/` (Hero, Features, CTA, HowItWorks, NavHeader, Footer)
- **Theme bridge:** ThemeProvider + ThemeSelect override Starlight's theme system for daisyUI compatibility

## Deployment

### GitHub Pages (via GitHub Actions)

The site deploys automatically to `https://jaigouk.com/altoiddd` when changes are pushed to `main`.
See `.github/workflows/website.yaml` for the workflow.

### Manual deployment

```bash
cd website
npm run build
# Upload dist/ to any static hosting
```
