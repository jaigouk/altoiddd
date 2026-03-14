# Research: Static Site Generator for alty.ai

**Date:** 2026-03-14
**Spike Ticket:** alty-cli-dkc
**Status:** Final

## Summary

**Recommendation: Astro + Starlight** for the alty.ai website. It is the only candidate that excels at both marketing/landing pages AND documentation from the same project, with verified support for referencing `../docs/*.md` from a `website/` subdirectory via the `glob()` loader. Hugo is a strong runner-up for Go ecosystem alignment but falls short on design quality ceiling for marketing pages.

## Research Question

What is the best static site generator for alty's public-facing website (alty.ai), given:
1. Source content is existing Markdown in `docs/` (PRD.md, DDD.md, ARCHITECTURE.md)
2. The Go repo root must stay clean -- no `package.json`/`node_modules` at root
3. Must support both marketing pages (hero, features, CTAs) AND documentation (nav, search, code blocks)
4. SSG tooling must live entirely in a `website/` subdirectory
5. Must be able to reference `../docs/*.md` as content sources
6. Need to integrate metadata from `.alty/maintenance/doc-registry.toml`

## Options Considered

| Option | Version | License | Language | Marketing Pages | Docs Quality | External Content | TOML Support | Maintenance |
|--------|---------|---------|----------|-----------------|--------------|------------------|--------------|-------------|
| **Astro + Starlight** | Astro 6.0.4 / Starlight 0.38.1 | MIT | JS/TS | Excellent | Excellent | Yes (glob loader) | Built-in (5.12+) | Active, monthly releases |
| **Hugo** | v0.157.0 | Apache 2.0 | Go | Fair | Good | Yes (module mounts) | Built-in (data dir) | Active, monthly releases |
| **MkDocs Material** | MkDocs 2.0 (rewrite) | MIT | Python | Poor | Excellent | Yes (docs_dir) | Via plugin | Unstable (2.0 breaking) |
| **Docusaurus** | 3.9.2 | MIT | JS/React | Fair | Excellent | Yes (path: "../docs") | No built-in | Slowing cadence |
| **mdBook** | 0.4.x | MPL 2.0 | Rust | None | Basic | Limited | No | Minimal updates |

## Detailed Analysis

### 1. Astro + Starlight (Recommended)

**Version:** Astro 6.0.4 (stable, March 2026), Starlight 0.38.1 (March 2025)
**License:** MIT (both Astro and Starlight)
**Source:** [github.com/withastro/astro](https://github.com/withastro/astro), [github.com/withastro/starlight](https://github.com/withastro/starlight)

#### External Content Support (Verified)

Astro's `glob()` loader can reference files from **anywhere on the filesystem**, including outside the project directory. This is documented in the [Content Loader Reference](https://docs.astro.build/en/reference/content-loader-reference/):

> The glob() loader creates collection entries from directories of files located **anywhere on the filesystem**.

Configuration in `website/src/content.config.ts`:
```typescript
import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';

const docs = defineCollection({
  loader: glob({ pattern: "**/*.md", base: "../docs" }),
});
```

This means the `website/` directory can be fully self-contained with its own `package.json`, `node_modules/`, and build config, while pulling content from `../docs/`.

**Source:** [Astro Content Collections docs](https://docs.astro.build/en/guides/content-collections/), [withastro/roadmap Discussion #434](https://github.com/withastro/roadmap/discussions/434)

#### Marketing + Docs in One Project

Starlight provides a `template: splash` layout for landing pages with hero sections, and full access to Astro's page generation for custom marketing pages. Key features:

- **Landing pages:** `template: splash` removes sidebars, supports hero component with images ([Starlight Pages docs](https://starlight.astro.build/guides/pages/))
- **Custom pages:** `<StarlightPage>` component wraps custom Astro pages in Starlight layout ([Starlight Customization](https://starlight.astro.build/guides/customization/))
- **Full Astro pages:** Any `.astro` file in `src/pages/` renders as a standalone page with full design freedom
- **Component ecosystem:** Can use React, Vue, Svelte, or plain Astro components for interactive elements

This means we can have:
- `/` -- Custom landing page with hero, features, CTAs (pure Astro components)
- `/docs/` -- Starlight-powered documentation with nav, search, code blocks

**Source:** [Starlight Pages guide](https://starlight.astro.build/guides/pages/), [Starlight Frontmatter Reference](https://starlight.astro.build/reference/frontmatter/)

#### TOML Integration

Astro 5.12+ has built-in TOML support in content collections via the `glob()` and `file()` loaders. The `file()` loader treats each top-level TOML table as an independent entry.

```typescript
const docRegistry = defineCollection({
  loader: file("../.alty/maintenance/doc-registry.toml"),
  schema: z.object({
    owner: z.string(),
    last_reviewed: z.string(),
    review_cadence_days: z.number(),
  }),
});
```

This means we can pull doc freshness metadata directly into page templates.

**Source:** [Astro Content Loader Reference](https://docs.astro.build/en/reference/content-loader-reference/), [TOML support commit](https://github.com/withastro/astro/commit/6bd5f75806cb4df39d9e4e9b1f2225dcfdd724b0)

#### Design Quality

Starlight provides excellent defaults:
- Built-in search (Pagefind, runs via Node.js API)
- Dark/light mode
- Sidebar auto-generation from file structure
- Code block syntax highlighting
- i18n support
- 95+ Lighthouse scores out of the box

For marketing pages, the full Astro ecosystem of themes and components is available. The [Astro theme directory](https://astro.build/themes/) has dedicated landing page templates.

#### Pros
- Best-in-class for mixed marketing + docs sites
- `glob()` loader verified to work with external directories
- Built-in TOML support for doc-registry integration
- Starlight purpose-built for documentation
- Islands architecture = minimal JS shipped to client
- Active development (Astro 6 just released)
- MIT license

#### Cons
- Requires Node.js in the build pipeline (not Go-native)
- JS/TS ecosystem in `website/` (isolated but still JS)
- Starlight is opinionated -- heavy customization requires understanding Astro internals
- Node.js dependency for CI/CD

### 2. Hugo

**Version:** v0.157.0 (February 25, 2026)
**License:** Apache 2.0
**Source:** [github.com/gohugoio/hugo](https://github.com/gohugoio/hugo)

#### External Content Support (Verified)

Hugo supports mounting external directories via `module.mounts` in `hugo.toml`:

```toml
[[module.mounts]]
source = 'content'
target = 'content'

[[module.mounts]]
source = '/absolute/path/to/alty-cli/docs'
target = 'content/docs'
```

Both absolute and project-relative paths are supported for the main project. Hugo does NOT follow symlinks -- you must use module mounts instead.

**Source:** [Hugo Module Configuration](https://gohugo.io/configuration/module/), [Hugo Directory Structure](https://gohugo.io/getting-started/directory-structure/)

**Important caveat:** When using module mounts, you must mount ALL directories (including the default `content` dir), or Hugo ignores the defaults. This is a known footgun.

#### Marketing + Docs

Hugo's theme ecosystem has several documentation themes:

- **Docsy** (by Google): Full-featured docs theme with landing page support. Has "landing page" template type without left nav. Used by Kubernetes, gRPC, Knative. However, it's "heavyweight" and designed for 20+ page doc sets. ([docsy.dev](https://www.docsy.dev/))
- **Doks:** Beautiful but criticized for mixing content and theme, requiring you to fork the theme to customize. ([Hugo Discourse](https://discourse.gohugo.io/t/looking-for-usable-docs-and-landing-theme-doks-vs-docsy/38512))
- **Book:** Clean, minimal docs theme. No marketing page support.

The fundamental challenge with Hugo for marketing pages is that Hugo's templating language (Go templates) is powerful but verbose for complex layouts. Building a custom hero section, feature grid, or pricing table requires significantly more effort than with Astro components.

#### TOML Integration

Hugo natively reads TOML data files from the `data/` directory, accessible via `.Site.Data`. Module mounts can point the data directory to an external path.

```toml
[[module.mounts]]
source = '/absolute/path/to/.alty/maintenance'
target = 'data/maintenance'
```

Then in templates: `{{ .Site.Data.maintenance.doc_registry }}`.

**Source:** [Hugo Data Sources](https://gohugo.io/content-management/data-sources/)

#### Pros
- Go-native -- single binary, no runtime dependencies
- Fastest build times (1000 pages in ~2 seconds)
- No Node.js, no package.json anywhere
- Apache 2.0 license
- Module mounts support external content
- Mature, stable ecosystem (12+ years)
- Built-in TOML support

#### Cons
- Go template syntax is verbose for complex marketing layouts
- Theme ecosystem quality varies significantly
- Docsy is the best option but still looks "enterprise docs", not "product landing page"
- Custom design requires deep Hugo template knowledge
- Design ceiling lower than Astro for marketing pages
- No component model -- harder to build reusable UI pieces

### 3. MkDocs Material

**Version:** MkDocs 2.0 (ground-up rewrite, announced March 2026)
**License:** MIT
**Source:** [squidfunk.github.io/mkdocs-material](https://squidfunk.github.io/mkdocs-material/)

#### External Content Support

MkDocs supports `docs_dir: ../docs` in `mkdocs.yml` to reference content outside the project directory.

#### Marketing Pages

This is MkDocs Material's biggest weakness. Landing pages require custom template overrides:
- Must create `overrides/home.html` template
- Hero sections require disabling `navigation.instant`
- No built-in support for feature grids, CTAs, or marketing layouts
- Fundamentally a documentation tool, not a website builder

**Source:** [MkDocs Material Issue #2057](https://github.com/squidfunk/mkdocs-material/issues/2057)

#### Stability Concern

MkDocs 2.0 was announced as a "ground-up rewrite" with "potentially significant breaking changes." Material for MkDocs depends on MkDocs as its central dependency but has no control over MkDocs itself. This creates upgrade risk.

**Source:** [Material for MkDocs Blog](https://squidfunk.github.io/mkdocs-material/blog/archive/2026/)

#### Pros
- Beautiful documentation output out of the box
- Excellent search (built-in)
- Simple configuration
- Python-based (lighter than Node.js)

#### Cons
- Marketing pages require significant workarounds
- MkDocs 2.0 rewrite creates stability risk
- Python dependency (not Go-native, not JS either -- a third runtime)
- Design ceiling too low for a product website
- Not suitable for alty.ai's dual marketing + docs requirement

### 4. Docusaurus

**Version:** 3.9.2 (last published ~4 months ago)
**License:** MIT
**Source:** [github.com/facebook/docusaurus](https://github.com/facebook/docusaurus)

#### External Content Support (Verified)

Docusaurus supports `path: "../docs"` in the docs plugin configuration. This is a supported pattern used by React Native and Jest websites (both Meta projects that migrated from Docusaurus v1).

**Source:** [Docusaurus Issue #6256](https://github.com/facebook/docusaurus/issues/6256) (closed as can't-reproduce, confirmed working by maintainers)

#### Marketing + Docs

Docusaurus has a decent landing page system but is React-heavy:
- Default landing page with hero and features
- React components for custom pages
- Blog plugin included

However, the entire site ships as a React SPA, meaning heavier client-side JS than Astro.

#### Pros
- Good docs features (versioning, search via Algolia DocSearch v4 with AI)
- Landing page template included
- React component ecosystem
- MIT license

#### Cons
- Ships full React runtime to client (heavier than Astro)
- Release cadence slowing (3.9.2 is 4+ months old)
- No built-in TOML support
- React-specific (Astro is framework-agnostic)
- Heavier than Astro for the same result
- Less active development momentum compared to Astro

### 5. mdBook

**Version:** 0.4.x
**License:** MPL 2.0
**Source:** [github.com/rust-lang/mdBook](https://github.com/rust-lang/mdBook)

**Eliminated early.** mdBook is a book/documentation tool with no marketing page support, no theme ecosystem for professional websites, and no TOML data integration. It produces clean rendered Markdown and nothing more. Suitable for Rust ecosystem internal docs, not for a product website.

## Recommendation

**Use Astro + Starlight**, deployed from a `website/` subdirectory.

### Rationale

| Decision Driver | Astro + Starlight | Hugo | Others |
|----------------|-------------------|------|--------|
| Marketing + docs in one project | Best (full page freedom + Starlight) | Fair (Go templates limit design) | Poor-Fair |
| External content from `../docs/` | Verified (glob loader) | Verified (module mounts) | Varies |
| Design quality ceiling | Highest (component-driven) | Medium (template-driven) | Low-Medium |
| TOML integration | Built-in (file loader) | Built-in (data dir) | Limited |
| Build simplicity | `npm run build` -> static HTML | `hugo` -> static HTML | Varies |
| Go ecosystem alignment | No (JS/TS) | Yes (Go binary) | No |
| Active development | Very active (Astro 6 just released) | Active | Mixed |

The main trade-off is **Go ecosystem alignment vs. design quality ceiling**. Hugo is Go-native and requires zero JS, but its templating system makes professional marketing pages significantly harder to build. For a product website that represents alty to potential users, design quality wins over ecosystem purity.

The `website/` isolation pattern keeps the JS tooling fully separated:

```
alty-cli/
  cmd/alty/           # Go CLI
  internal/           # Go domain code
  docs/               # Markdown source content
  website/            # Astro + Starlight (self-contained)
    package.json
    node_modules/     # gitignored
    astro.config.mjs
    src/
      content.config.ts  # glob loader pointing to ../docs/
      pages/
        index.astro      # Landing page
      content/
        docs/            # Can be empty if all content comes from ../docs/
```

No JS files pollute the Go project root. CI builds can run `cd website && npm ci && npm run build` as a separate step.

### When Hugo Would Be Better

If the requirement were "documentation only" (no marketing pages), or if zero JS dependency were a hard constraint, Hugo with Docsy would be the right choice. It is faster, simpler, and Go-native.

## References

### Astro / Starlight
- [Astro Content Collections](https://docs.astro.build/en/guides/content-collections/)
- [Astro Content Loader Reference (glob)](https://docs.astro.build/en/reference/content-loader-reference/)
- [Starlight Pages Guide](https://starlight.astro.build/guides/pages/)
- [Starlight Customization](https://starlight.astro.build/guides/customization/)
- [Astro GitHub Releases](https://github.com/withastro/astro/releases) -- v6.0.4 latest
- [Starlight GitHub Releases](https://github.com/withastro/starlight/releases) -- v0.38.1 latest
- [Astro TOML Support Commit](https://github.com/withastro/astro/commit/6bd5f75806cb4df39d9e4e9b1f2225dcfdd724b0)
- [External Content Collections Discussion](https://github.com/withastro/roadmap/discussions/434)
- [Astro License (MIT)](https://github.com/withastro/astro/blob/main/LICENSE)

### Hugo
- [Hugo Module Configuration](https://gohugo.io/configuration/module/)
- [Hugo Directory Structure (Union File System)](https://gohugo.io/getting-started/directory-structure/)
- [Hugo Data Sources](https://gohugo.io/content-management/data-sources/)
- [Hugo GitHub Releases](https://github.com/gohugoio/hugo/releases) -- v0.157.0 latest
- [Hugo License (Apache 2.0)](https://gohugo.io/about/license/)
- [Docsy Theme](https://www.docsy.dev/)

### MkDocs Material
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [Landing Page Issue #2057](https://github.com/squidfunk/mkdocs-material/issues/2057)
- [MkDocs Material Blog (2026)](https://squidfunk.github.io/mkdocs-material/blog/archive/2026/)

### Docusaurus
- [Docusaurus Docs Plugin](https://docusaurus.io/docs/api/plugins/@docusaurus/plugin-content-docs)
- [External Path Issue #6256](https://github.com/facebook/docusaurus/issues/6256)
- [Docusaurus GitHub](https://github.com/facebook/docusaurus) -- v3.9.2 latest

### mdBook
- [mdBook Documentation](https://rust-lang.github.io/mdBook/)

### Comparisons
- [Hygraph SSG Comparison 2026](https://hygraph.com/blog/top-12-ssgs)
- [Starlight vs Docusaurus (LogRocket)](https://blog.logrocket.com/starlight-vs-docusaurus-building-documentation/)

## Follow-up Tasks

1. **Scaffold Astro + Starlight in `website/`** -- Initialize the project, configure glob loader to reference `../docs/`, verify build works. Create landing page skeleton and docs section.

2. **Configure TOML integration** -- Set up content collection for `.alty/maintenance/doc-registry.toml`, display doc freshness metadata (owner, last_reviewed) on documentation pages.

3. **Design landing page** -- Create the alty.ai landing page with hero section, feature grid, and CTA. This is design work, not SSG configuration.

4. **Set up CI build** -- Add `website/` build step to CI pipeline. Output goes to `website/dist/` as static HTML.

5. **Configure deployment** -- Set up static hosting (Cloudflare Pages, Netlify, or similar) with automatic deploys from `website/dist/`.
