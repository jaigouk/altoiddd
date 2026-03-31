# VHS Terminal Recording Research

**Date:** 2026-03-31
**Topic:** Charmbracelet VHS for automated terminal GIF generation
**Purpose:** Evaluate VHS feasibility for automated demo recordings of alto CLI

---

## Summary

VHS (v0.11.0, MIT license) is viable for automated alto CLI demo recording, including on headless Linux servers via Docker. The tool is actively maintained by Charmbracelet, uses a ttyd+Chromium+xterm.js+ffmpeg pipeline, and handles interactive CLI programs well via the `Wait` command. The main risk is the Chromium dependency, which adds complexity to headless environments -- mitigated entirely by the official Docker image.

---

## 1. Runtime Dependencies

VHS has **four** runtime dependencies, not two as commonly stated:

| Dependency | Role | Required? |
|------------|------|-----------|
| **ttyd** | Web-based terminal emulator; runs the shell and exposes it over HTTP | Yes |
| **Chromium** (via go-rod) | Headless browser that connects to ttyd, renders xterm.js terminal | Yes |
| **ffmpeg** | Converts captured PNG frames into GIF/MP4/WebM | Yes |
| **Shell** (bash/zsh/fish) | Executes the actual commands | Yes (system default) |

**Rendering pipeline:**
```
Tape file -> Parser -> Evaluator -> ttyd (terminal server)
                                       |
                                    go-rod (browser automation)
                                       |
                                    Chromium + xterm.js (render)
                                       |
                                    PNG frame capture
                                       |
                                    ffmpeg -> GIF/MP4/WebM
```

Source: [DeepWiki VHS architecture](https://deepwiki.com/charmbracelet/vhs), [GitHub discussion #291](https://github.com/charmbracelet/vhs/discussions/291)

**Why a browser?** VHS uses xterm.js inside Chromium to render the terminal. This gives cross-platform consistency (macOS, Windows, Linux) without maintaining a custom terminal renderer. The maintainers acknowledge "there is room for a version that does not use the browser" but have not implemented one.

Source: [Discussion #291](https://github.com/charmbracelet/vhs/discussions/291)

---

## 2. Headless Linux Server Operation

**Yes, VHS runs headlessly.** It does not need an X11 display -- Chromium runs in headless mode via go-rod.

### Bare metal / VM requirements

On headless Linux servers, the Chromium dependency requires these shared libraries:
- `libgbm1`
- `libnss3`
- `libatk1.0-0`
- `libatk-bridge2.0-0`
- `libcups2`
- `libxcomposite1`
- `libxdamage1`

Source: [Issue #45](https://github.com/charmbracelet/vhs/issues/45)

### Root user sandbox issue

Running VHS as root triggers Chromium's sandbox restriction. Fix: set `VHS_NO_SANDBOX=true` environment variable (implemented in PR #195, undocumented in README as of 2026-03).

Source: [Issue #504](https://github.com/charmbracelet/vhs/issues/504)

### Docker (recommended for CI)

```bash
docker run --rm -v $PWD:/vhs ghcr.io/charmbracelet/vhs <tape-file>.tape
```

The Docker image bundles all dependencies (ttyd, ffmpeg, Chromium). This is the simplest path for CI/automation.

Source: [VHS README](https://github.com/charmbracelet/vhs)

### Server mode (SSH-based)

VHS has a `vhs serve` mode that accepts tape files over SSH:

```bash
# Server side
VHS_PORT=1976 vhs serve

# Client side
ssh vhs.example.com < demo.tape > demo.gif
```

Environment variables: `VHS_PORT`, `VHS_HOST`, `VHS_GID`, `VHS_UID`, `VHS_KEY_PATH`, `VHS_AUTHORIZED_KEYS_PATH`.

Source: [VHS README](https://github.com/charmbracelet/vhs)

---

## 3. Complete .tape File Syntax

### Output commands

```tape
Output demo.gif          # GIF output
Output demo.mp4          # MP4 output
Output demo.webm         # WebM output
Output frames/           # PNG sequence to directory
# Multiple outputs allowed in single tape
```

### Require (dependency check)

```tape
Require git
Require node
# Must appear at top; fails early if program not in $PATH
```

### Set commands (terminal configuration)

**Must be at top of tape file before non-setting/non-output commands.** Exception: `TypingSpeed` can be changed mid-tape.

| Command | Description | Example |
|---------|-------------|---------|
| `Set Shell "<shell>"` | Shell to use | `Set Shell "zsh"` |
| `Set FontSize <n>` | Font size in pixels | `Set FontSize 16` |
| `Set FontFamily "<font>"` | Font name | `Set FontFamily "JetBrains Mono"` |
| `Set Width <n>` | Terminal width in pixels | `Set Width 1200` |
| `Set Height <n>` | Terminal height in pixels | `Set Height 600` |
| `Set LetterSpacing <n>` | Character spacing | `Set LetterSpacing 1` |
| `Set LineHeight <n>` | Line height | `Set LineHeight 1.2` |
| `Set Padding <n>` | Internal padding in pixels | `Set Padding 20` |
| `Set Margin <n>` | External margin in pixels | `Set Margin 20` |
| `Set MarginFill "<color>"` | Margin background color | `Set MarginFill "#1e1e2e"` |
| `Set BorderRadius <n>` | Corner radius | `Set BorderRadius 10` |
| `Set WindowBar <style>` | Window decoration | `Colorful`, `ColorfulRight`, `Rings`, `RingsRight` |
| `Set Theme "<name>"` | Named theme (340+ available) | `Set Theme "Catppuccin Mocha"` |
| `Set Theme {...}` | Custom JSON theme | See below |
| `Set Framerate <n>` | GIF framerate | `Set Framerate 30` |
| `Set PlaybackSpeed <n>` | Playback multiplier | `Set PlaybackSpeed 1.5` |
| `Set TypingSpeed <dur>` | Typing delay per char | `Set TypingSpeed 50ms` |
| `Set LoopOffset <n>` | GIF loop offset | `Set LoopOffset 50%` |
| `Set CursorBlink <bool>` | Cursor blink | `Set CursorBlink false` |

### Interaction commands

| Command | Description | Example |
|---------|-------------|---------|
| `Type "<text>"` | Type text with typing animation | `Type "echo hello"` |
| `Type@<dur> "<text>"` | Type with custom speed | `Type@100ms "fast typing"` |
| `Enter` | Press Enter | `Enter` |
| `Tab` | Press Tab (autocomplete) | `Tab` |
| `Space` | Press Space | `Space` |
| `Backspace` | Delete character | `Backspace 3` (repeat) |
| `Up` / `Down` | Arrow keys | `Up 2` (repeat) |
| `Left` / `Right` | Arrow keys | `Left` |
| `PageUp` / `PageDown` | Page navigation | `PageUp` |
| `ScrollUp` / `ScrollDown` | Viewport scrolling (v0.11.0+) | `ScrollDown 5` |
| `Ctrl+<key>` | Control combinations | `Ctrl+C`, `Ctrl+L` |
| `Ctrl+Alt+<key>` | Multi-modifier | `Ctrl+Alt+Delete` |
| `Shift+<key>` | Shift modifier (v0.7.0+) | `Shift+Tab` |
| `Insert` / `Delete` | Insert/Delete keys (v0.7.0+) | `Delete` |

### Timing and synchronization

```tape
# Fixed delay
Sleep 500ms
Sleep 2s
Sleep 0.5

# Wait for terminal content (v0.9.0+)
Wait                           # Default: waits for shell prompt (/$/)
Wait /World/                   # Wait for "World" anywhere
Wait+Screen /Loading complete/ # Wait for text anywhere on screen
Wait+Line /\$\s*$/             # Wait for text on current line
Wait@10ms /pattern/            # Wait with polling interval
```

The `Wait` command is the key to handling interactive programs -- it blocks until the expected output appears, then proceeds.

### Visibility control

```tape
Hide                           # Stop recording frames
Type "cd ~/project && clear"   # Setup not shown
Enter
Wait+Line /\$/
Show                           # Resume recording
```

### Utility commands

```tape
Screenshot build-result.png    # Capture current frame as PNG
Copy "text to clipboard"       # Clipboard operations
Paste
Source other-tape.tape         # Include commands from another file
Env DEMO "true"                # Set environment variable
```

Source: [VHS README](https://github.com/charmbracelet/vhs/blob/main/README.md), [Context7 docs](https://context7.com/charmbracelet/vhs/llms.txt)

---

## 4. Interactive CLI Program Support

**Yes, VHS handles interactive prompts.** The mechanism is:

1. Use `Type` to enter responses to prompts
2. Use `Wait+Screen /prompt text/` to synchronize before typing
3. Use arrow keys (`Up`, `Down`) and `Tab` for selection-style prompts
4. Use `Enter` to confirm

**Example pattern for a CLI that asks questions:**

```tape
# Program asks: "What is your project name?"
Type "my-app"
Enter

# Program asks: "Select language:"
Wait+Screen /Select language/
Down 2
Enter

# Program shows: "Generating..."
Wait+Screen /Complete/
Sleep 2s
```

The `Wait` command (introduced v0.9.0, January 2025) was specifically designed for this use case -- it replaces fragile `Sleep` timing with regex-based content matching.

Source: [VHS v0.9.0 release notes](https://github.com/charmbracelet/vhs/releases)

---

## 5. Bubbletea / Survey-Style Prompt Limitations

### Known issues

1. **Rendering inconsistency** (Issue #412): Some bubbletea programs render inconsistently -- the same tape file can produce working recordings sometimes and blank viewports other times. This appears to be timing-related, not a configuration issue.

2. **Broken styling** (Issue #362): Lip Gloss styled backgrounds can render full-width instead of text-width. **Resolution:** Upgrading ttyd to the latest version fixes this. The issue was in ttyd, not VHS.

3. **Non-deterministic rendering**: Custom bubbletea apps may fail to render event messages even when official bubbletea examples work fine.

### Mitigations

- **Use latest ttyd** -- most rendering issues trace to outdated ttyd versions
- **Use `Wait+Screen` instead of `Sleep`** -- eliminates timing race conditions
- **Add generous `Sleep` after `Wait`** -- gives the TUI time to fully render before the next frame
- **Set explicit terminal dimensions** -- bubbletea programs may resize differently in VHS vs local terminal

### Assessment for alto CLI

alto uses Cobra (not bubbletea) for its CLI. Cobra-based CLIs with standard stdin prompts are much simpler than full TUI programs. VHS should handle alto prompts reliably with `Type` + `Enter` + `Wait` patterns.

Source: [Issue #412](https://github.com/charmbracelet/vhs/issues/412), [Issue #362](https://github.com/charmbracelet/vhs/issues/362)

---

## 6. Font and Theme Options

### Fonts

Any font installed on the system or available in the Chromium rendering context can be used. Common choices:

```tape
Set FontFamily "JetBrains Mono"
Set FontFamily "Fira Code"
Set FontFamily "SF Mono"
Set FontFamily "Monoflow"
```

The Docker image includes a default monospace font. Custom fonts must be mounted into the container.

### Themes

VHS includes **340+ pre-defined themes** stored in an embedded `themes.json`.

```bash
vhs themes    # List all available theme names
```

Examples: `Catppuccin Mocha`, `Catppuccin Frappe`, `Dracula`, `Nord`, `Solarized Dark`, `Tokyo Night`.

**Custom themes** via inline JSON:

```tape
Set Theme {
  "black": "#171421",
  "red": "#c01c28",
  "green": "#26a269",
  "yellow": "#a2734c",
  "blue": "#12488b",
  "magenta": "#a347ba",
  "cyan": "#2aa1b3",
  "white": "#d0cfcc",
  "brightBlack": "#5e5c64",
  "brightRed": "#f66151",
  "brightGreen": "#33d17a",
  "brightYellow": "#e9ad0c",
  "brightBlue": "#2a7bde",
  "brightMagenta": "#c061cb",
  "brightCyan": "#33c7de",
  "brightWhite": "#ffffff",
  "background": "#171421",
  "foreground": "#d0cfcc",
  "selection": "#b4d5ff",
  "cursor": "#d0cfcc"
}
```

This means we can define an alto-branded theme matching our Midnight Teal palette.

Source: [VHS README](https://github.com/charmbracelet/vhs), [Context7 docs](https://context7.com/charmbracelet/vhs/llms.txt)

---

## 7. GIF Quality and Size Control

### Levers for controlling output size

| Control | Effect on Size | Command |
|---------|---------------|---------|
| **Terminal dimensions** | Smaller = smaller file | `Set Width 800` / `Set Height 400` |
| **Font size** | Smaller = smaller frames | `Set FontSize 14` |
| **Framerate** | Lower = fewer frames | `Set Framerate 15` (default varies) |
| **Duration** | Shorter = fewer frames | Minimize `Sleep` durations |
| **PlaybackSpeed** | Faster = shorter duration | `Set PlaybackSpeed 2` |
| **Output format** | WebM/MP4 much smaller than GIF | `Output demo.webm` |

### Typical GIF sizes

Based on VHS examples and community reports:
- **Simple command demo** (5-10s, 1200x600): 500KB - 2MB
- **Interactive session** (15-30s, 1200x600): 2MB - 8MB
- **Complex TUI recording** (30s+, 1200x600): 5MB - 15MB+

### Size optimization strategies

1. **Use WebM instead of GIF** -- 5-10x smaller for equivalent quality
2. **Reduce dimensions** -- `Set Width 800` + `Set Height 400` + `Set FontSize 12`
3. **Lower framerate** -- `Set Framerate 15` (sufficient for typing animations)
4. **Minimize idle time** -- Use `Wait` instead of long `Sleep` values
5. **Post-process with gifsicle** -- `gifsicle -O3 --lossy=80 output.gif -o optimized.gif`
6. **Use PNG sequence + external encoder** -- `Output frames/` for maximum control

### SVG alternative

A community fork (agentstation/vhs) supports SVG output, which scales perfectly and can be significantly smaller than raster GIFs for text-heavy recordings. However, this is a fork, not mainline VHS.

Source: [VHS README](https://github.com/charmbracelet/vhs), [libraries.io agentstation/vhs](https://libraries.io/go/github.com%2Fagentstation%2Fvhs)

---

## 8. Version and Maintenance

| Attribute | Value |
|-----------|-------|
| **Current version** | v0.11.0 (March 10, 2026) |
| **License** | MIT |
| **Language** | Go |
| **Repository** | [github.com/charmbracelet/vhs](https://github.com/charmbracelet/vhs) |
| **Release cadence** | ~3-4 months between major releases |
| **Recent releases** | v0.11.0 (Mar 2026), v0.10.0 (Jun 2025), v0.9.0 (Jan 2025) |

Actively maintained by Charmbracelet (same team behind bubbletea, lip gloss, charm).

Source: [VHS releases](https://github.com/charmbracelet/vhs/releases)

---

## 9. Feasibility for alto Demo Automation

### CI-like environment (Docker)

```bash
# Build alto, then record demo
docker run --rm \
  -v $PWD:/vhs \
  ghcr.io/charmbracelet/vhs \
  demo.tape
```

The Docker image handles all dependencies. The tape file would:
1. `Hide` the build step
2. `Show` the actual demo
3. Use `Wait+Screen` for prompt synchronization
4. Output to GIF and/or WebM

### Claude Code session (local machine)

Prerequisites: install VHS (`go install github.com/charmbracelet/vhs@latest`), ttyd, ffmpeg. On Linux, also install Chromium's shared library dependencies or use Docker.

A Claude Code agent could:
1. Write a `.tape` file based on the demo scenario
2. Run `vhs demo.tape` via Bash tool
3. The GIF appears at the specified output path

### Example tape for alto

```tape
Output alto-demo.gif
Output alto-demo.webm

Require alto

Set Shell "bash"
Set FontFamily "JetBrains Mono"
Set FontSize 16
Set Width 1200
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set Framerate 24
Set WindowBar Colorful
Set CursorBlink false

Hide
Type "export PATH=$PWD/bin:$PATH"
Enter
Wait+Line /\$/
Type "clear"
Enter
Wait+Line /\$/
Show

Type "alto init"
Enter
Wait+Screen /project name/
Sleep 500ms

Type "my-saas-app"
Enter
Wait+Screen /describe your project/
Sleep 500ms

Type "A SaaS platform for managing restaurant reservations with real-time availability"
Enter
Wait+Screen /Generating/
Sleep 1s

Wait+Screen /complete/
Sleep 3s
```

---

## 10. Risks and Limitations

| Risk | Severity | Mitigation |
|------|----------|------------|
| **Chromium dependency size** | Medium | Use Docker image; do not install Chromium per-machine |
| **Bubbletea rendering flakes** | Low (alto uses Cobra, not bubbletea) | Use `Wait` instead of `Sleep`; latest ttyd |
| **GIF file size for long demos** | Medium | Use WebM for web; GIF only for README |
| **Non-deterministic timing** | Medium | `Wait+Screen` regex matching eliminates timing issues |
| **Font availability in Docker** | Low | Mount custom fonts or use default monospace |
| **VHS_NO_SANDBOX undocumented** | Low | Set env var; documented in issue #504 |

---

## Recommendation

**VHS is well-suited for automated alto CLI demo recording.** Use the Docker image (`ghcr.io/charmbracelet/vhs`) for CI and headless environments. The `Wait+Screen` command (v0.9.0+) handles interactive prompts reliably. Define a custom theme matching alto's Midnight Teal palette. Output WebM for the website, GIF for README/docs.

### Follow-up work needed

1. **Create alto demo tape files** -- one per major workflow (init, discover, doc-health, generate)
2. **Define alto VHS theme** -- custom JSON matching Midnight Teal palette
3. **CI integration** -- add Docker-based VHS recording step
4. **Font bundling** -- verify JetBrains Mono availability in Docker image or mount it
