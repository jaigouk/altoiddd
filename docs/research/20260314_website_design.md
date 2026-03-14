# Research: Website Design System — Tailwind + daisyUI + Starlight

**Date:** 2026-03-14
**Spike Ticket:** alty-cli-qzv.3
**Status:** Final

## Summary

Starlight + daisyUI v5 integration is feasible but requires careful CSS layering and theme-switcher overrides. daisyUI v5 uses a pure CSS `@plugin` syntax (no `tailwind.config.js`), which simplifies Astro integration but introduces a `data-theme` conflict with Starlight's built-in dark/light switcher. The recommended approach is to override Starlight's `ThemeProvider` and `ThemeSelect` components to bridge both systems, using CSS cascade layers (`@layer`) to ensure daisyUI styles override Starlight defaults without !important hacks.

**Recommendation:** Use daisyUI v5 with a custom theme (`alty-dark` / `alty-light`), override Starlight's theme components, and adopt a restrained "developer-tool" aesthetic inspired by Linear and Warp — not the neon-glow cliché.

---

## 1. Starlight + daisyUI v5 Compatibility

### 1.1 The Core Conflict

Starlight uses its own dark/light theme system based on:
- A `<ThemeProvider>` component rendered in `<head>` that sets/reads a `data-theme` attribute (or `starlight-theme` localStorage key)
- CSS custom properties like `--sl-color-accent-*` and `--sl-color-gray-*`

daisyUI v5 uses:
- The `data-theme` attribute on the `<html>` element to select themes
- Its own CSS custom properties like `--color-base-100`, `--color-primary`, etc.

**Conflict:** Both systems use `data-theme` but with incompatible values and variable namespaces. Starlight's theme switcher sets `data-theme="dark"` or `data-theme="light"`, but daisyUI expects its own theme names (e.g., `"night"`, `"nord"`).

Source: [tricked.dev/blog/daisyui-starlight](https://tricked.dev/blog/daisyui-starlight/) — the author confirms: *"The theme switcher of starlight will not play nicely with the new daisyui themes and you will have to overwrite it."* The author also notes that *"with the new daisyui and tailwind versions this no longer works"* (as of July 2025), meaning the v4-era approach is broken.

### 1.2 Resolution Strategy

**Approach: Override Starlight's theme components to bridge both systems.**

Starlight supports [component overrides](https://starlight.astro.build/guides/overriding-components/). The two components to override:

| Component | Purpose | Override Strategy |
|-----------|---------|-------------------|
| `ThemeProvider` | Sets up dark/light theme in `<head>` | Bridge to set daisyUI's `data-theme` when Starlight theme changes |
| `ThemeSelect` | UI dropdown for theme switching | Keep UI, but toggle both Starlight vars AND daisyUI `data-theme` |

Source: [Starlight Overrides Reference](https://starlight.astro.build/reference/overrides/)

**Implementation outline:**

```astro
<!-- src/components/ThemeProvider.astro (override) -->
<script is:inline>
  // Read Starlight's theme preference
  const theme = localStorage.getItem('starlight-theme') ||
    (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  // Map to daisyUI theme names
  const daisyTheme = theme === 'dark' ? 'alty-dark' : 'alty-light';
  document.documentElement.setAttribute('data-theme', daisyTheme);
</script>
```

### 1.3 CSS Cascade Layer Strategy

Starlight v0.34+ uses CSS cascade layers internally. Custom unlayered CSS automatically overrides Starlight defaults.

```css
/* src/styles/global.css */
@layer base, starlight, daisy-overrides, theme, components, utilities;

@import '@astrojs/starlight-tailwind';
@import 'tailwindcss/theme.css' layer(theme);
@import 'tailwindcss/utilities.css' layer(utilities);
```

daisyUI styles in the `daisy-overrides` layer will take precedence over Starlight's `starlight` layer while being overridable by utility classes.

Source: [Starlight CSS & Styling docs](https://starlight.astro.build/guides/css-and-tailwind/)

### 1.4 Variable Mapping

Map daisyUI's semantic colors to Starlight's expected variables:

```css
/* Bridge daisyUI colors → Starlight variables */
@layer daisy-overrides {
  :root {
    --sl-color-accent-high: var(--color-primary);
    --sl-color-accent-low: var(--color-primary-content);
    --sl-color-text: var(--color-base-content);
    --sl-color-bg-nav: var(--color-base-200);
    --sl-color-bg-sidebar: var(--color-base-100);
  }
}
```

### 1.5 Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Theme flash on page load | Medium | Inline script in `<head>` sets `data-theme` before paint |
| Starlight updates break overrides | Medium | Pin `@astrojs/starlight` minor version; test on upgrade |
| daisyUI class names collide with Starlight | Low | Use `prefix: "d-"` in daisyUI config if needed |
| View transitions reset theme | Low | Persist in `localStorage`, re-apply on `astro:after-swap` |

---

## 2. daisyUI v5 Configuration for Astro

### 2.1 Installation

```bash
npm i vite@latest tailwindcss @tailwindcss/vite
npm i -D daisyui@latest
```

Source: [dav.one — Setting up Astro with Tailwind v4 + daisyUI v5](https://dav.one/setting-up-astro-project-with-tailwindcss-and-daisyui/)

### 2.2 Vite Plugin Setup

```javascript
// astro.config.mjs
import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import starlight from "@astrojs/starlight";

export default defineConfig({
  integrations: [
    starlight({
      title: "alty",
      customCss: ["./src/styles/global.css"],
      components: {
        ThemeProvider: "./src/components/ThemeProvider.astro",
        ThemeSelect: "./src/components/ThemeSelect.astro",
      },
    }),
  ],
  vite: {
    plugins: [tailwindcss()],
  },
});
```

### 2.3 CSS Configuration (No tailwind.config.js)

daisyUI v5 is configured entirely in CSS via the `@plugin` directive:

```css
/* src/styles/global.css */
@layer base, starlight, daisy-overrides, theme, components, utilities;

@import "tailwindcss";
@import "@astrojs/starlight-tailwind";

@plugin "daisyui" {
  themes: alty-light --default, alty-dark --prefersdark;
  root: ":root";
  logs: false;
}
```

Source: [daisyUI Config docs](https://daisyui.com/docs/config/)

### 2.4 Custom Theme Definition

daisyUI v5 uses `@plugin "daisyui/theme"` for custom themes with oklch colors:

```css
@plugin "daisyui/theme" {
  name: "alty-dark";
  default: false;
  prefersdark: true;
  color-scheme: dark;

  --color-base-100: oklch(15% 0.01 260);
  --color-base-200: oklch(12% 0.01 260);
  --color-base-300: oklch(10% 0.01 260);
  --color-base-content: oklch(90% 0.01 260);

  --color-primary: oklch(70% 0.15 190);
  --color-primary-content: oklch(15% 0.02 190);

  --color-secondary: oklch(65% 0.12 280);
  --color-secondary-content: oklch(95% 0.01 280);

  --color-accent: oklch(75% 0.10 160);
  --color-accent-content: oklch(15% 0.02 160);

  --color-neutral: oklch(25% 0.02 260);
  --color-neutral-content: oklch(85% 0.01 260);

  --radius-field: 0.375rem;
  --radius-box: 0.5rem;
  --border: 1px;
  --depth: 0;
  --noise: 0;
}
```

Key v5 features:
- **oklch color format** — perceptually uniform, P3 gamut support
- **`--depth`** — soft 3D effect on buttons (set to 0 for flat design)
- **`--noise`** — textured background noise (set to 0 for clean look)
- **`--size-field`** / **`--size-selector`** — global component sizing
- **`color-mix()`** — opacity handled via CSS `color-mix()`, no JS needed

Source: [daisyUI v5 release notes](https://daisyui.com/docs/v5/), [daisyUI themes docs](https://daisyui.com/docs/themes/)

---

## 3. Reference Site Analysis

### 3.1 Comparison Matrix

| Site | Primary Palette | Typography | Hero Pattern | Dark Mode | Key Design Move |
|------|----------------|------------|--------------|-----------|-----------------|
| **Linear** | Neutral grays, semantic accents | System sans + monospace | Animated grid dots, centered text | CSS vars, token-based | Refined motion (step-based keyframes, 3200ms) |
| **Vercel** | B&W + gradient accents | Geist (custom), monospace | Dark bg, gradient mesh | Geist design system tokens | Radical simplicity, anti-decoration |
| **Supabase** | Jungle Green #34B27B + Bunker #11181C | System sans + code font | Dark, code-editor aesthetic | Dark-first (not afterthought) | Green = "success/go" signal, developer-native |
| **Charm** | Minimal neutrals, playful accents | Clean sans-serif | Centered text, "glamorous CLI" | Light base | Playful mascots, CLI product cards |
| **Buf** | Brand blue #0E5DF5, purple→cyan gradients | Ubuntu + JetBrains Mono + Public Sans | Value-first headline, logo carousel | Light only (enterprise B2B) | Interactive card carousel, gradient bars |
| **Warp** | Deep charcoal #121212, neutral text | Inter + Geist Mono + Fragment Mono | Sticky nav, flex grid | CSS custom properties, token-based | Minimalist, performance-focused transitions |
| **Railway** | Purple-based hsl(250,24%,9%), cyan accents | Inter Tight + JetBrains Mono + IBM Plex Serif | Layered gradients (327°, 245°, 147°) | Three modes: dark, light, vaporwave | Multi-theme personality, angled gradients |
| **Resend** | Black bg, white text, glass-morphism | Domaine + Inter Variable + Commit Mono | Layered bg images, hero-1.jpg | Forced dark (no toggle) | Glass-morphism buttons, backdrop-blur |

Sources: Direct WebFetch analysis of each site's CSS/markup.

### 3.2 Pattern Synthesis

**What the best sites have in common:**

1. **Typography stack = 2-3 fonts max:** One sans for UI, one mono for code, optionally one serif for editorial weight
2. **Dark mode is the default**, not an afterthought. Light mode exists but dark is primary
3. **Motion is purposeful:** Linear uses 3200ms step animations; Warp uses cubic-bezier; nobody uses gratuitous particle effects
4. **Color restraint:** One accent color maximum (Supabase = green, Vercel = gradient, Linear = semantic). Not multiple neon colors
5. **Hero = text-first:** Large headline, short subtext, 1-2 CTAs. No hero images or stock photos
6. **Code-native typography:** JetBrains Mono and Geist Mono dominate the developer tool space

**What separates "premium" from "template":**
- Cascade layer management instead of `!important`
- Token-based color systems (CSS custom properties)
- Performance-conscious animation (GPU-friendly `opacity`, `transform`)
- View Transition API for page navigation
- Custom fonts loaded strategically (not 8 Google Fonts)

---

## 4. Design Tokens

### 4.1 Proposed Color Palette: "Midnight Teal"

Inspired by Linear's restraint, Supabase's developer-native feel, and Warp's charcoal depth. Deliberately avoids neon purple/cyan "AI slop" (see Section 6).

**Primary palette (oklch):**

| Token | oklch Value | Hex Approx | Usage |
|-------|-------------|------------|-------|
| `--color-base-100` | `oklch(16% 0.015 250)` | #1a1d2e | Page background |
| `--color-base-200` | `oklch(13% 0.012 250)` | #141726 | Sidebar, nav |
| `--color-base-300` | `oklch(10% 0.010 250)` | #0f111d | Cards, code blocks |
| `--color-base-content` | `oklch(92% 0.008 250)` | #e8e9ed | Primary text |
| `--color-primary` | `oklch(72% 0.12 190)` | #3db8a9 | Accent (teal, not neon) |
| `--color-primary-content` | `oklch(15% 0.03 190)` | #0d2b27 | Text on primary bg |
| `--color-secondary` | `oklch(60% 0.08 270)` | #7b6fbd | Secondary (muted violet) |
| `--color-secondary-content` | `oklch(95% 0.01 270)` | #f0eff5 | Text on secondary bg |
| `--color-accent` | `oklch(80% 0.06 80)` | #d4b87a | Warm accent (gold) |
| `--color-accent-content` | `oklch(15% 0.02 80)` | #2a2418 | Text on accent bg |
| `--color-neutral` | `oklch(22% 0.015 250)` | #282c42 | Borders, dividers |
| `--color-neutral-content` | `oklch(80% 0.01 250)` | #c0c2cc | Muted text |

**Rationale:**
- **Teal primary** — signals "go/success" like Supabase's green, but cooler and more unique. Avoids the purple/blue/neon trap
- **Muted violet secondary** — provides contrast without screaming "AI"
- **Warm gold accent** — breaks the cold-color monotony, adds personality
- **Deep blue-gray base** — not pure black (avoids harsh contrast), not purple-tinted (avoids AI slop)

**Light theme variant (`alty-light`):**

| Token | oklch Value | Usage |
|-------|-------------|-------|
| `--color-base-100` | `oklch(98% 0.005 250)` | Page background |
| `--color-base-200` | `oklch(95% 0.008 250)` | Sidebar, nav |
| `--color-base-300` | `oklch(92% 0.006 250)` | Cards |
| `--color-base-content` | `oklch(18% 0.015 250)` | Primary text |
| `--color-primary` | `oklch(55% 0.14 190)` | Teal (darker for contrast) |

### 4.2 Typography

| Token | Value | Usage |
|-------|-------|-------|
| `--font-sans` | `"Inter Variable", "Inter", system-ui, sans-serif` | UI text, body copy |
| `--font-mono` | `"JetBrains Mono", "Geist Mono", ui-monospace, monospace` | Code blocks, CLI references |
| `--font-display` | `"Inter Tight", "Inter Variable", system-ui, sans-serif` | Hero headlines, section titles |

**Why these fonts:**
- **Inter Variable** — universal readability, variable font (single file), used by Linear and Warp
- **JetBrains Mono** — ligatures for code, used by Buf, Railway, and Warp
- **Inter Tight** — geometric, condensed variant for headlines (used by Railway)

All fonts are free and open-source (SIL Open Font License).

### 4.3 Spacing Scale

Follow Tailwind's default 4px-based scale. No custom overrides needed:

```
4px (1), 8px (2), 12px (3), 16px (4), 20px (5), 24px (6), 32px (8), 40px (10), 48px (12), 64px (16), 80px (20), 96px (24)
```

Key usage:
- Section padding: `py-16` to `py-24` (64px–96px)
- Card padding: `p-6` to `p-8` (24px–32px)
- Gap between elements: `gap-4` to `gap-6` (16px–24px)

### 4.4 Border Radius

| Element | Token | Value |
|---------|-------|-------|
| Buttons, inputs | `--radius-field` | `0.375rem` (6px) |
| Cards, panels | `--radius-box` | `0.5rem` (8px) |
| Pills, badges | Tailwind `rounded-full` | `9999px` |

Avoid over-rounding. Ref sites use subtle radii (6-8px), not the bubbly 16px+ common in AI-generated designs.

### 4.5 Shadows and Effects

```css
/* Subtle elevation, not glow */
--shadow-sm: 0 1px 2px oklch(0% 0 0 / 0.05);
--shadow-md: 0 4px 6px oklch(0% 0 0 / 0.07), 0 2px 4px oklch(0% 0 0 / 0.06);
--shadow-lg: 0 10px 15px oklch(0% 0 0 / 0.1), 0 4px 6px oklch(0% 0 0 / 0.05);

/* Accent glow (use SPARINGLY — one element per viewport max) */
--glow-primary: 0 0 20px oklch(72% 0.12 190 / 0.15);
```

**Rule:** Glow effects on at most ONE focal element per viewport (hero CTA or key feature). Not on every card, button, and heading.

---

## 5. Tailwind Plugins

### 5.1 Recommended

| Plugin | Purpose | Size | Source |
|--------|---------|------|--------|
| `@tailwindcss/typography` | Prose styling for markdown docs | Built-in | [github.com/tailwindlabs/tailwindcss-typography](https://github.com/tailwindlabs/tailwindcss-typography) |
| `tailwindcss-motion` | CSS-first animations (no JS) | ~5KB | [rombo.co/tailwind](https://rombo.co/tailwind/) |

### 5.2 @tailwindcss/typography

Essential for Starlight docs content. Usage:

```html
<article class="prose dark:prose-invert lg:prose-lg">
  <!-- Markdown-rendered content -->
</article>
```

Starlight applies this automatically to docs pages. For landing pages, add explicitly.

Source: [Astro docs — Style rendered Markdown with Tailwind Typography](https://docs.astro.build/en/recipes/tailwind-rendered-markdown/)

### 5.3 tailwindcss-motion

CSS-first approach — zero JavaScript, hardware-accelerated, ~5KB:

```html
<!-- Fade in on scroll -->
<div class="motion-preset-fade motion-duration-500">
  ...
</div>

<!-- Slide up with delay -->
<div class="motion-preset-slide-up motion-delay-200">
  ...
</div>
```

Presets: `fade`, `slide-up`, `slide-down`, `slide-left`, `slide-right`, `bounce`, `expand`, `shrink`, `blur-in`.

Source: [blog.brightcoding.dev — Tailwind Animations Plugin](https://www.blog.brightcoding.dev/2026/03/10/tailwind-animations-the-revolutionary-plugin-for-effortless-ui-motion)

### 5.4 NOT Recommended

| Plugin | Why Not |
|--------|---------|
| `framer-motion` | JS bundle overhead, React-only, overkill for static site |
| `Aceternity UI` | Beautiful but React + Framer dependency, doesn't fit Astro SSG |
| `tailwind-scrollbar` | Starlight handles scrollbar styling |
| Multiple animation libraries | Pick one. CSS-first > JS-based for SSG. |

---

## 6. Anti-Patterns to Avoid

### 6.1 "AI Slop" Patterns (CRITICAL)

These patterns are now instantly recognizable as lazy AI-generated design. Avoid all of them:

| Pattern | Why It's "Slop" | What To Do Instead |
|---------|-----------------|-------------------|
| Purple/cyan neon gradients | Default output of every AI image generator | Restrained teal + warm gold accents |
| Glowing orbs/particles in background | Screams "I typed 'tech background' into Midjourney" | Subtle grid patterns or solid dark backgrounds |
| Inter font + purple gradient + rounded corners | The "AI trifecta" — recognizable from a mile away | Inter is fine, but pair with distinctive color + sharp radii |
| Robot/brain/circuit illustrations | "Digital brain" stock imagery is exhausted | Terminal screenshots, code snippets, architecture diagrams |
| Multiple neon accent colors | Cyberpunk aesthetic peaked in 2023 | One accent color, used sparingly |
| Glass-morphism on everything | Was trendy in 2022, now a cliché when overused | Use on 1-2 focal elements max (hero CTA, pricing card) |
| Dark purple (#2D1B69-ish) backgrounds | The literal "Purple AI Slop" color | Blue-gray (#1a1d2e) or near-black |

Sources:
- [zeroskillai.com — How to Fix Purple AI Slop](https://zeroskillai.com/how-to-fix-purple-ai-slop/)
- [medium.com/design-bootcamp — Aesthetics in the AI Era](https://medium.com/design-bootcamp/aesthetics-in-the-ai-era-visual-web-design-trends-for-2026-5a0f75a10e98)

### 6.2 Other Anti-Patterns

| Anti-Pattern | Better Approach |
|--------------|-----------------|
| Loading 5+ Google Fonts | 2-3 variable fonts max (Inter Variable + JetBrains Mono) |
| Hero image with stock photo | Text-first hero with code example or terminal demo |
| Animations on every element | Motion on 1-2 key elements per viewport |
| `!important` overrides | CSS cascade layers (`@layer`) |
| Full-page gradient backgrounds | Solid dark bg with subtle gradient on ONE section |
| Centered-everything layout | Left-aligned text with asymmetric grid (like Linear) |
| "Built with AI" aesthetic | "Built by developers" aesthetic — show the tool in action |

### 6.3 The "alty Test"

Before shipping any design, ask: *"If I showed this to a senior developer, would they think an AI generated the landing page?"* If yes, simplify.

**What makes developer-tool sites credible:**
1. Show the actual product (terminal output, CLI commands)
2. Real code examples, not decorative code
3. Restraint in color and motion
4. Information density > decoration
5. Fast load times (< 1s LCP)

---

## 7. Implementation Approach

### 7.1 File Structure

```
website/
├── astro.config.mjs              # Starlight + Tailwind config
├── src/
│   ├── styles/
│   │   ├── global.css            # Tailwind + daisyUI + layer config
│   │   └── starlight-bridge.css  # Variable mapping overrides
│   ├── components/
│   │   ├── ThemeProvider.astro   # Override: bridges Starlight ↔ daisyUI themes
│   │   └── ThemeSelect.astro    # Override: syncs both theme systems
│   └── content/
│       └── docs/                 # Starlight docs (from docs/guides/)
└── package.json
```

### 7.2 Integration Order

1. Set up Astro + Starlight (vanilla, no customization)
2. Add Tailwind v4 via `@tailwindcss/vite`
3. Add daisyUI v5 with `@plugin "daisyui"` in CSS
4. Define `alty-dark` and `alty-light` custom themes
5. Override `ThemeProvider` and `ThemeSelect` to bridge systems
6. Add CSS cascade layer definitions
7. Map daisyUI variables → Starlight variables
8. Add `@tailwindcss/typography` for prose styling
9. Add `tailwindcss-motion` for landing page animations
10. Test theme switching, View Transitions, and mobile responsiveness

---

## References

### Primary Sources
- [daisyUI v5 Config docs](https://daisyui.com/docs/config/)
- [daisyUI v5 Themes docs](https://daisyui.com/docs/themes/)
- [daisyUI v5 Release Notes](https://daisyui.com/docs/v5/)
- [daisyUI Install for Astro](https://daisyui.com/docs/install/astro/?lang=en)
- [Starlight CSS & Styling](https://starlight.astro.build/guides/css-and-tailwind/)
- [Starlight Overrides Reference](https://starlight.astro.build/reference/overrides/)
- [Starlight Configuration Reference](https://starlight.astro.build/reference/configuration/)

### Integration Guides
- [tricked.dev — Combining daisyUI and Starlight](https://tricked.dev/blog/daisyui-starlight/) (v4 era — approach is outdated but conflict analysis still valid)
- [dav.one — Astro + Tailwind v4 + daisyUI v5 Setup](https://dav.one/setting-up-astro-project-with-tailwindcss-and-daisyui/)
- [fogbender.com — Tailwind CSS for Starlight](https://fogbender.com/blog/setting-up-tailwind-with-starlight)

### Design Systems (Reference Sites)
- [Vercel Geist Design System](https://vercel.com/geist/colors)
- [Supabase Design System](https://supabase.com/design-system)
- [Supabase brand colors](https://mobbin.com/colors/brand/supabase) — Jungle Green #34B27B, Bunker #11181C
- [Linear](https://linear.app) — token-based neutral system, step animations
- [Warp](https://warp.dev) — Inter + Geist Mono, charcoal #121212, 4px spacing scale
- [Buf](https://buf.build) — Brand blue #0E5DF5, Ubuntu + JetBrains Mono
- [Railway](https://railway.com) — hsl(250,24%,9%), Inter Tight + JetBrains Mono, multi-theme
- [Resend](https://resend.com) — forced dark, Domaine + Inter Variable + Commit Mono, glass-morphism

### Anti-Pattern Sources
- [zeroskillai.com — How to Fix Purple AI Slop](https://zeroskillai.com/how-to-fix-purple-ai-slop/)
- [medium.com/design-bootcamp — Aesthetics in the AI Era: 2026](https://medium.com/design-bootcamp/aesthetics-in-the-ai-era-visual-web-design-trends-for-2026-5a0f75a10e98)
- [paddo.dev — Breaking the AI Slop Aesthetic](https://paddo.dev/blog/claude-code-plugins-frontend-design/)

### Fonts (All SIL Open Font License)
- [Inter Variable](https://rsms.me/inter/) — UI text
- [JetBrains Mono](https://www.jetbrains.com/lp/mono/) — Code
- [Inter Tight](https://rsms.me/inter/) — Headlines

### Tailwind Plugins
- [@tailwindcss/typography](https://github.com/tailwindlabs/tailwindcss-typography) — Prose styling
- [tailwindcss-motion](https://rombo.co/tailwind/) — CSS-first animations (~5KB)
