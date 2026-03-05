# VSCode Extension + Svelte 5 + @xyflow/svelte Setup Patterns

**Date:** 2026-03-05
**Type:** Research Spike
**Status:** Complete

## Summary

This document provides verified, copy-pasteable setup patterns for building a VSCode
extension with a Svelte 5 (runes-based) webview containing an @xyflow/svelte node-graph
canvas, bundled with Vite, tested with Vitest, and linted with ESLint flat config.

## Version Matrix (verified March 2026)

| Package | Version | License | Source |
|---------|---------|---------|--------|
| svelte | 5.53.7 | MIT | [npm](https://www.npmjs.com/package/svelte) |
| @sveltejs/vite-plugin-svelte | 6.2.4 | MIT | [npm](https://www.npmjs.com/package/@sveltejs/vite-plugin-svelte) |
| @xyflow/svelte | 1.5.1 | MIT | [npm](https://www.npmjs.com/package/@xyflow/svelte) |
| vite | 6.x | MIT | [npm](https://www.npmjs.com/package/vite) |
| vitest | 3.x | MIT | [npm](https://www.npmjs.com/package/vitest) |
| eslint-plugin-svelte | 3.x | MIT | [npm](https://www.npmjs.com/package/eslint-plugin-svelte) |
| typescript-eslint | 8.x | MIT | [npm](https://www.npmjs.com/package/typescript-eslint) |

---

## 1. Project Structure

```
my-vscode-extension/
├── .vscode/
│   ├── launch.json              # Extension debug config
│   └── tasks.json               # Build tasks
├── src/
│   └── extension/               # Extension host code (Node.js, CJS)
│       ├── extension.ts         # activate/deactivate entry
│       └── panels/
│           └── GraphPanel.ts    # Webview panel manager
├── webview/                     # Webview UI code (Browser, ESM)
│   ├── App.svelte               # Root Svelte component
│   ├── main.ts                  # Webview entry point
│   ├── lib/
│   │   ├── nodes/               # Custom @xyflow/svelte node components
│   │   │   └── CustomNode.svelte
│   │   └── stores/
│   │       └── graph.svelte.ts  # Svelte 5 rune-based state
│   └── vite-env.d.ts
├── dist/
│   ├── extension/               # Compiled extension host
│   │   └── extension.js
│   └── webview/                 # Compiled webview
│       └── index.html
├── package.json
├── tsconfig.json                # Shared base config
├── tsconfig.extension.json      # Extension host (Node)
├── tsconfig.webview.json        # Webview (Browser)
├── vite.config.extension.ts     # Vite config for extension host
├── vite.config.webview.ts       # Vite config for webview
├── vitest.config.ts             # Test config
├── eslint.config.js             # Flat config
└── index.html                   # Webview HTML template
```

---

## 2. package.json

```jsonc
{
  "name": "my-graph-extension",
  "displayName": "My Graph Extension",
  "description": "VSCode extension with Svelte 5 node-graph webview",
  "version": "0.1.0",
  "publisher": "your-publisher",
  "engines": {
    "vscode": "^1.100.0"
  },
  "categories": ["Other"],
  "activationEvents": [],
  "main": "./dist/extension/extension.js",
  "contributes": {
    "commands": [
      {
        "command": "myExtension.openGraph",
        "title": "Open Graph View"
      }
    ],
    "viewsContainers": {
      "activitybar": [
        {
          "id": "graph-explorer",
          "title": "Graph Explorer",
          "icon": "media/icon.svg"
        }
      ]
    }
  },
  "scripts": {
    "build": "npm run build:extension && npm run build:webview",
    "build:extension": "vite build --config vite.config.extension.ts",
    "build:webview": "vite build --config vite.config.webview.ts",
    "watch": "concurrently \"npm:watch:*\"",
    "watch:extension": "vite build --config vite.config.extension.ts --watch",
    "watch:webview": "vite build --config vite.config.webview.ts --watch",
    "dev:webview": "vite --config vite.config.webview.ts",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint .",
    "lint:fix": "eslint . --fix",
    "typecheck": "tsc --noEmit -p tsconfig.extension.json && tsc --noEmit -p tsconfig.webview.json",
    "package": "npm run build && vsce package"
  },
  "devDependencies": {
    "@eslint/js": "^9.0.0",
    "@sveltejs/vite-plugin-svelte": "^6.2.4",
    "@testing-library/svelte": "^5.0.0",
    "@testing-library/jest-dom": "^6.0.0",
    "@types/vscode": "^1.100.0",
    "@xyflow/svelte": "^1.5.1",
    "concurrently": "^9.0.0",
    "eslint": "^9.0.0",
    "eslint-plugin-svelte": "^3.0.0",
    "globals": "^16.0.0",
    "jsdom": "^26.0.0",
    "svelte": "^5.53.0",
    "svelte-check": "^4.0.0",
    "typescript": "^5.7.0",
    "typescript-eslint": "^8.0.0",
    "vite": "^6.0.0",
    "vitest": "^3.0.0"
  }
}
```

**Key decisions:**
- `engines.vscode` is `^1.100.0` because ESM extension host support landed in v1.100.0
  ([source](https://github.com/microsoft/vscode/issues/130367)).
- `main` points to the CJS-bundled extension output. Even with ESM host support, CJS
  remains the safer default for broad compatibility.
- Two separate Vite configs because the extension host targets Node (CJS) while the
  webview targets the browser (ESM).

---

## 3. Vite Configuration -- Dual Build Targets

### 3a. Extension Host -- `vite.config.extension.ts`

The extension host runs in Node.js. VSCode expects CJS (`require`/`module.exports`).
The `vscode` module is external -- it is injected by the host at runtime.

```typescript
import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    lib: {
      entry: './src/extension/extension.ts',
      formats: ['cjs'],
      fileName: () => 'extension.js',
    },
    outDir: 'dist/extension',
    rollupOptions: {
      external: [
        'vscode',
        // Add any Node built-in modules your extension uses
        'path',
        'fs',
        'os',
        'child_process',
      ],
    },
    sourcemap: true,
    // Do NOT minify extension code -- VSCode debugging needs readable output
    minify: false,
    emptyOutDir: true,
  },
  // No plugins needed -- this is plain TypeScript, no Svelte
});
```

### 3b. Webview -- `vite.config.webview.ts`

The webview runs in a browser-like sandbox. It uses ESM and bundles Svelte + xyflow.

```typescript
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [
    svelte({
      compilerOptions: {
        // Svelte 5 runes mode is the default; no legacy flag needed
      },
    }),
  ],
  build: {
    outDir: 'dist/webview',
    rollupOptions: {
      input: './index.html',
      output: {
        entryFileNames: 'assets/[name].js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]',
      },
    },
    sourcemap: true,
    emptyOutDir: true,
  },
  // Used only during `dev:webview` for standalone browser development
  server: {
    port: 5173,
  },
});
```

**Why two configs instead of one?** The extension host is a Node CJS library with
`vscode` as an external. The webview is a browser app with an HTML entry point. These
are fundamentally different build targets. Keeping them separate is cleaner than
conditional logic in a single config.

**Alternative:** The `@tomjs/vite-plugin-vscode` package
([npm](https://www.npmjs.com/package/@tomjs/vite-plugin-vscode)) unifies both into a
single config and provides HMR injection for webviews during development. It uses
`tsdown` internally for the extension host. Consider it if you want a more opinionated
all-in-one solution.

---

## 4. TypeScript Configuration

### 4a. Base -- `tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022"],
    "module": "Node16",
    "moduleResolution": "Node16",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "exactOptionalPropertyTypes": true
  },
  "exclude": ["node_modules", "dist"]
}
```

### 4b. Extension Host -- `tsconfig.extension.json`

```json
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "outDir": "dist/extension",
    "types": ["node"]
  },
  "include": ["src/extension/**/*.ts"]
}
```

### 4c. Webview -- `tsconfig.webview.json`

```json
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "outDir": "dist/webview",
    "types": []
  },
  "include": [
    "webview/**/*.ts",
    "webview/**/*.svelte",
    "webview/**/*.svelte.ts"
  ]
}
```

**Key differences:**
- Extension host uses `Node16` module resolution (CJS, `require()`).
- Webview uses `Bundler` module resolution (ESM, Vite processes it).
- Webview has `DOM` lib; extension host has `node` types.

---

## 5. VSCode Extension Host -- Extension Entry + Webview Panel

### 5a. `src/extension/extension.ts`

```typescript
import * as vscode from 'vscode';
import { GraphPanel } from './panels/GraphPanel';

export function activate(context: vscode.ExtensionContext): void {
  const openGraphCommand = vscode.commands.registerCommand(
    'myExtension.openGraph',
    () => {
      GraphPanel.createOrShow(context.extensionUri);
    }
  );

  context.subscriptions.push(openGraphCommand);

  // If the panel already exists when the extension reactivates
  if (GraphPanel.currentPanel) {
    GraphPanel.currentPanel.reveal();
  }
}

export function deactivate(): void {
  // cleanup if needed
}
```

### 5b. `src/extension/panels/GraphPanel.ts`

This is the critical glue between the extension host and the Svelte webview.

```typescript
import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

/**
 * Typed message protocol for extension <-> webview communication.
 * Both sides share this contract.
 */
export type ExtensionMessage =
  | { type: 'graph:load'; payload: { nodes: unknown[]; edges: unknown[] } }
  | { type: 'graph:nodeSelected'; payload: { nodeId: string } }
  | { type: 'graph:update'; payload: { nodes: unknown[]; edges: unknown[] } };

export type WebviewMessage =
  | { type: 'webview:ready' }
  | { type: 'webview:nodeClicked'; payload: { nodeId: string } }
  | { type: 'webview:graphChanged'; payload: { nodes: unknown[]; edges: unknown[] } };

export class GraphPanel {
  public static currentPanel: GraphPanel | undefined;
  private static readonly viewType = 'graphView';
  private readonly panel: vscode.WebviewPanel;
  private readonly extensionUri: vscode.Uri;
  private disposables: vscode.Disposable[] = [];

  public static createOrShow(extensionUri: vscode.Uri): GraphPanel {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    if (GraphPanel.currentPanel) {
      GraphPanel.currentPanel.panel.reveal(column);
      return GraphPanel.currentPanel;
    }

    const panel = vscode.window.createWebviewPanel(
      GraphPanel.viewType,
      'Graph View',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [
          vscode.Uri.joinPath(extensionUri, 'dist', 'webview'),
        ],
      }
    );

    GraphPanel.currentPanel = new GraphPanel(panel, extensionUri);
    return GraphPanel.currentPanel;
  }

  private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
    this.panel = panel;
    this.extensionUri = extensionUri;

    this.panel.webview.html = this.getHtmlForWebview(this.panel.webview);

    // Listen for messages from the webview
    this.panel.webview.onDidReceiveMessage(
      (message: WebviewMessage) => {
        switch (message.type) {
          case 'webview:ready':
            // Webview is mounted -- send initial data
            this.postMessage({
              type: 'graph:load',
              payload: { nodes: [], edges: [] },
            });
            break;
          case 'webview:nodeClicked':
            vscode.window.showInformationMessage(
              `Node clicked: ${message.payload.nodeId}`
            );
            break;
          case 'webview:graphChanged':
            // Handle graph state changes from webview
            break;
        }
      },
      null,
      this.disposables
    );

    this.panel.onDidDispose(() => this.dispose(), null, this.disposables);
  }

  public postMessage(message: ExtensionMessage): void {
    this.panel.webview.postMessage(message);
  }

  public reveal(): void {
    this.panel.reveal();
  }

  private dispose(): void {
    GraphPanel.currentPanel = undefined;
    this.panel.dispose();
    while (this.disposables.length) {
      const d = this.disposables.pop();
      if (d) d.dispose();
    }
  }

  private getHtmlForWebview(webview: vscode.Webview): string {
    const distPath = vscode.Uri.joinPath(
      this.extensionUri,
      'dist',
      'webview'
    );

    // Read the built index.html and rewrite asset paths to webview URIs
    const indexHtmlPath = vscode.Uri.joinPath(distPath, 'index.html');
    let html = fs.readFileSync(indexHtmlPath.fsPath, 'utf-8');

    // Replace relative asset paths with webview-safe URIs
    // Vite outputs paths like "./assets/main.js" or "/assets/main.js"
    html = html.replace(
      /(src|href)="(\.?\/?assets\/[^"]+)"/g,
      (_match, attr, assetPath) => {
        const cleanPath = assetPath.replace(/^\.?\//, '');
        const assetUri = webview.asWebviewUri(
          vscode.Uri.joinPath(distPath, cleanPath)
        );
        return `${attr}="${assetUri}"`;
      }
    );

    // Inject CSP nonce for security
    const nonce = getNonce();
    html = html.replace(
      '<head>',
      `<head>
        <meta http-equiv="Content-Security-Policy"
          content="default-src 'none';
                   style-src ${webview.cspSource} 'unsafe-inline';
                   script-src 'nonce-${nonce}';
                   font-src ${webview.cspSource};
                   img-src ${webview.cspSource} data:;">
      `
    );

    // Add nonce to script tags
    html = html.replace(/<script/g, `<script nonce="${nonce}"`);

    return html;
  }
}

function getNonce(): string {
  let text = '';
  const possible =
    'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  for (let i = 0; i < 32; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}
```

---

## 6. Svelte 5 Webview with @xyflow/svelte

### 6a. `index.html` (webview entry)

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Graph View</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/webview/main.ts"></script>
  </body>
</html>
```

### 6b. `webview/main.ts`

```typescript
import { mount } from 'svelte';
import App from './App.svelte';

const app = mount(App, {
  target: document.getElementById('app')!,
});

export default app;
```

### 6c. `webview/App.svelte`

```svelte
<script lang="ts">
  import {
    SvelteFlow,
    Background,
    Controls,
    MiniMap,
    BackgroundVariant,
    type Node,
    type Edge,
    type FitViewOptions,
    type DefaultEdgeOptions,
  } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';

  import CustomNode from './lib/nodes/CustomNode.svelte';
  import { createVscodeMessenger } from './lib/stores/vscode.svelte';

  import type { ExtensionMessage } from '../src/extension/panels/GraphPanel';

  // -- Custom node types registration --
  const nodeTypes = {
    custom: CustomNode,
  };

  // -- Svelte 5 runes for reactive state --
  let nodes = $state.raw<Node[]>([
    {
      id: '1',
      type: 'input',
      position: { x: 0, y: 0 },
      data: { label: 'Start' },
    },
    {
      id: '2',
      type: 'custom',
      position: { x: 200, y: 100 },
      data: { label: 'Process', description: 'A custom node' },
    },
    {
      id: '3',
      type: 'output',
      position: { x: 400, y: 0 },
      data: { label: 'End' },
    },
  ]);

  let edges = $state.raw<Edge[]>([
    { id: 'e1-2', source: '1', target: '2', type: 'smoothstep' },
    { id: 'e2-3', source: '2', target: '3', type: 'smoothstep' },
  ]);

  const fitViewOptions: FitViewOptions = { padding: 0.2 };
  const defaultEdgeOptions: DefaultEdgeOptions = { animated: true };

  // -- VSCode messenger (postMessage bridge) --
  const vscode = createVscodeMessenger();

  // Listen for messages from extension host
  $effect(() => {
    const handler = (event: MessageEvent<ExtensionMessage>) => {
      const message = event.data;
      switch (message.type) {
        case 'graph:load':
          nodes = message.payload.nodes as Node[];
          edges = message.payload.edges as Edge[];
          break;
        case 'graph:nodeSelected':
          // Handle external node selection
          break;
      }
    };

    window.addEventListener('message', handler);

    // Signal to extension that webview is ready
    vscode.postMessage({ type: 'webview:ready' });

    return () => {
      window.removeEventListener('message', handler);
    };
  });

  // -- Event handlers --
  function handleNodeClick(event: CustomEvent) {
    const node = event.detail.node;
    vscode.postMessage({
      type: 'webview:nodeClicked',
      payload: { nodeId: node.id },
    });
  }

  function handleConnect(event: CustomEvent) {
    const { connection } = event.detail;
    edges = [
      ...edges,
      {
        id: `e${connection.source}-${connection.target}`,
        source: connection.source,
        target: connection.target,
        type: 'smoothstep',
      },
    ];
  }
</script>

<div class="flow-container">
  <SvelteFlow
    bind:nodes
    bind:edges
    {nodeTypes}
    {fitViewOptions}
    {defaultEdgeOptions}
    onnodeclick={handleNodeClick}
    onconnect={handleConnect}
    fitView
  >
    <Background variant={BackgroundVariant.Dots} />
    <Controls />
    <MiniMap />
  </SvelteFlow>
</div>

<style>
  .flow-container {
    width: 100vw;
    height: 100vh;
  }
</style>
```

### 6d. `webview/lib/stores/vscode.svelte.ts`

The VSCode API bridge using Svelte 5 runes.

```typescript
/**
 * Type-safe VSCode webview API wrapper.
 *
 * acquireVsCodeApi() can only be called ONCE per webview session.
 * This module caches the instance.
 */

interface VsCodeApi {
  postMessage(message: unknown): void;
  getState(): unknown;
  setState(state: unknown): void;
}

declare function acquireVsCodeApi(): VsCodeApi;

let vsCodeApiInstance: VsCodeApi | undefined;

function getVsCodeApi(): VsCodeApi {
  if (!vsCodeApiInstance) {
    vsCodeApiInstance = acquireVsCodeApi();
  }
  return vsCodeApiInstance;
}

export function createVscodeMessenger() {
  const api = getVsCodeApi();

  return {
    postMessage(message: unknown): void {
      api.postMessage(message);
    },
    getState<T>(): T | undefined {
      return api.getState() as T | undefined;
    },
    setState<T>(state: T): void {
      api.setState(state);
    },
  };
}
```

### 6e. `webview/lib/nodes/CustomNode.svelte`

A custom @xyflow/svelte node with typed props (Svelte 5 `$props()`).

```svelte
<script lang="ts" module>
  import type { Node } from '@xyflow/svelte';

  export type CustomNodeData = {
    label: string;
    description?: string;
  };

  export type CustomNodeType = Node<CustomNodeData, 'custom'>;
</script>

<script lang="ts">
  import { Handle, Position, type NodeProps } from '@xyflow/svelte';

  let { id, data }: NodeProps<CustomNodeType> = $props();
</script>

<div class="custom-node">
  <div class="custom-node__header">{data.label}</div>
  {#if data.description}
    <div class="custom-node__body">{data.description}</div>
  {/if}
  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} />
</div>

<style>
  .custom-node {
    padding: 10px 15px;
    border-radius: 8px;
    background: var(--vscode-editor-background, #1e1e1e);
    color: var(--vscode-editor-foreground, #cccccc);
    border: 1px solid var(--vscode-panel-border, #333);
    font-family: var(--vscode-font-family);
    font-size: var(--vscode-font-size, 13px);
    min-width: 150px;
  }

  .custom-node__header {
    font-weight: 600;
    margin-bottom: 4px;
  }

  .custom-node__body {
    font-size: 0.85em;
    opacity: 0.8;
  }
</style>
```

**Key Svelte 5 patterns used:**
- `$state.raw()` for nodes/edges (raw because xyflow manages its own deep reactivity;
  `.raw` avoids Svelte's proxy overhead on large arrays)
  ([source](https://svelteflow.dev/api-reference/svelte-flow))
- `$props()` for component props (replaces `export let` from Svelte 4)
- `$effect()` for side effects (replaces `onMount` + `$:` reactive statements)
- `$derived()` available for computed values
- `<script module>` block for type exports (replaces `<script context="module">`)

---

## 7. Vitest Configuration

### 7a. `vitest.config.ts`

```typescript
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { svelteTesting } from '@testing-library/svelte/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  plugins: [svelte(), svelteTesting()],
  test: {
    // Use jsdom for component testing
    environment: 'jsdom',
    // Enable .svelte.ts test files to use runes
    include: [
      'webview/**/*.test.ts',
      'webview/**/*.test.svelte.ts',
      'src/**/*.test.ts',
    ],
    setupFiles: ['./vitest.setup.ts'],
    // Resolve Svelte browser conditions in test
    resolve: {
      conditions: ['browser'],
    },
  },
});
```

### 7b. `vitest.setup.ts`

```typescript
import '@testing-library/jest-dom/vitest';
```

### 7c. Example test: `webview/lib/stores/vscode.svelte.test.ts`

Note: Test files that use runes must have `.svelte.ts` extension.

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock acquireVsCodeApi (not available outside VSCode)
const mockPostMessage = vi.fn();
const mockGetState = vi.fn();
const mockSetState = vi.fn();

vi.stubGlobal('acquireVsCodeApi', () => ({
  postMessage: mockPostMessage,
  getState: mockGetState,
  setState: mockSetState,
}));

// Must import AFTER mocking
const { createVscodeMessenger } = await import('./vscode.svelte');

describe('createVscodeMessenger', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('sends messages via postMessage', () => {
    const messenger = createVscodeMessenger();
    messenger.postMessage({ type: 'webview:ready' });
    expect(mockPostMessage).toHaveBeenCalledWith({ type: 'webview:ready' });
  });

  it('delegates getState to VSCode API', () => {
    mockGetState.mockReturnValue({ selectedNode: '1' });
    const messenger = createVscodeMessenger();
    expect(messenger.getState()).toEqual({ selectedNode: '1' });
  });
});
```

### 7d. Example component test: `webview/lib/nodes/CustomNode.test.ts`

```typescript
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import CustomNode from './CustomNode.svelte';

describe('CustomNode', () => {
  it('renders label', () => {
    render(CustomNode, {
      props: {
        id: 'test-1',
        data: { label: 'My Node', description: 'Test description' },
        // Minimal NodeProps stubs -- xyflow injects these at runtime
        type: 'custom',
        dragging: false,
        zIndex: 0,
        isConnectable: true,
        positionAbsoluteX: 0,
        positionAbsoluteY: 0,
        width: 150,
        height: 50,
        selected: false,
        dragHandle: undefined,
        parentId: undefined,
        sourcePosition: undefined,
        targetPosition: undefined,
      },
    });
    expect(screen.getByText('My Node')).toBeInTheDocument();
    expect(screen.getByText('Test description')).toBeInTheDocument();
  });

  it('hides description when not provided', () => {
    render(CustomNode, {
      props: {
        id: 'test-2',
        data: { label: 'Minimal' },
        type: 'custom',
        dragging: false,
        zIndex: 0,
        isConnectable: true,
        positionAbsoluteX: 0,
        positionAbsoluteY: 0,
        width: 150,
        height: 50,
        selected: false,
        dragHandle: undefined,
        parentId: undefined,
        sourcePosition: undefined,
        targetPosition: undefined,
      },
    });
    expect(screen.getByText('Minimal')).toBeInTheDocument();
    expect(screen.queryByText('Test description')).not.toBeInTheDocument();
  });
});
```

**Critical Svelte 5 testing note:**
Files that use runes (`$state`, `$derived`, `$effect`) in test code must be named
`*.svelte.ts` or `*.svelte.js`. Regular `.ts` test files can test Svelte components
but cannot declare runes directly in test code.
([source](https://svelte.dev/docs/svelte/testing))

---

## 8. ESLint Flat Config

### `eslint.config.js`

```javascript
import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import ts from 'typescript-eslint';

export default ts.config(
  // -- Global ignores --
  {
    ignores: ['dist/', 'node_modules/', '*.config.js', '*.config.ts'],
  },

  // -- Base rules --
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs['flat/recommended'],

  // -- TypeScript settings for all TS files --
  {
    files: ['**/*.ts'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },

  // -- Svelte files: use svelte-eslint-parser with TS --
  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
      },
    },
  },

  // -- Extension host files: Node.js globals --
  {
    files: ['src/extension/**/*.ts'],
    languageOptions: {
      globals: {
        ...globals.node,
      },
    },
  },

  // -- Webview files: Browser globals --
  {
    files: ['webview/**/*.ts', 'webview/**/*.svelte'],
    languageOptions: {
      globals: {
        ...globals.browser,
        acquireVsCodeApi: 'readonly',
      },
    },
  },

  // -- Test files --
  {
    files: ['**/*.test.ts', '**/*.test.svelte.ts'],
    languageOptions: {
      globals: {
        ...globals.node,
      },
    },
  },

  // -- Shared rule overrides --
  {
    rules: {
      // Svelte 5 uses $state, $derived, etc. which look like unused vars
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_|^\\$',
        },
      ],
    },
  }
);
```

**Key points:**
- Uses `ts.config()` wrapper from `typescript-eslint` for proper config merging
  ([source](https://sveltejs.github.io/eslint-plugin-svelte/user-guide/))
- Separate `globals` blocks for Node (extension host) vs Browser (webview)
- `acquireVsCodeApi` declared as a readonly global for webview files
- `projectService: true` enables type-aware linting without specifying tsconfig paths
  manually (typescript-eslint v8+)

---

## 9. VSCode Launch/Tasks Configuration

### `.vscode/launch.json`

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Run Extension",
      "type": "extensionHost",
      "request": "launch",
      "args": ["--extensionDevelopmentPath=${workspaceFolder}"],
      "outFiles": ["${workspaceFolder}/dist/extension/**/*.js"],
      "preLaunchTask": "build"
    }
  ]
}
```

### `.vscode/tasks.json`

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build",
      "type": "shell",
      "command": "npm run build",
      "group": { "kind": "build", "isDefault": true },
      "problemMatcher": ["$tsc"]
    },
    {
      "label": "watch",
      "type": "shell",
      "command": "npm run watch",
      "isBackground": true,
      "group": "build",
      "problemMatcher": ["$tsc-watch"]
    }
  ]
}
```

---

## 10. postMessage Communication Pattern -- Detailed

The communication between extension host and webview uses a typed message protocol.

### Architecture

```
Extension Host (Node.js)          Webview (Browser)
     |                                  |
     |  panel.webview.postMessage(msg)  |
     |  ---------------------------->>  |
     |                                  |  window.addEventListener('message', handler)
     |                                  |
     |  panel.webview.onDidReceiveMessage
     |  <<----------------------------  |
     |                                  |  vscode.postMessage(msg)
```

### Shared type definitions (`shared/messages.ts`)

For larger projects, extract message types into a shared module:

```typescript
// shared/messages.ts
// Importable by both extension host and webview code

export type NodeData = {
  id: string;
  label: string;
  type: string;
  position: { x: number; y: number };
  data: Record<string, unknown>;
};

export type EdgeData = {
  id: string;
  source: string;
  target: string;
  type?: string;
};

/** Messages from extension host -> webview */
export type ExtensionToWebview =
  | { type: 'graph:load'; payload: { nodes: NodeData[]; edges: EdgeData[] } }
  | { type: 'graph:selectNode'; payload: { nodeId: string } }
  | { type: 'theme:changed'; payload: { kind: 'light' | 'dark' } };

/** Messages from webview -> extension host */
export type WebviewToExtension =
  | { type: 'webview:ready' }
  | { type: 'webview:nodeClicked'; payload: { nodeId: string } }
  | { type: 'webview:edgeCreated'; payload: EdgeData }
  | { type: 'webview:graphExport'; payload: { nodes: NodeData[]; edges: EdgeData[] } };
```

### Request/Response pattern (for async operations)

```typescript
// In webview -- request-response with Promise
let requestId = 0;
const pendingRequests = new Map<number, (value: unknown) => void>();

export function request<T>(type: string, payload?: unknown): Promise<T> {
  return new Promise((resolve) => {
    const id = ++requestId;
    pendingRequests.set(id, resolve as (value: unknown) => void);
    vscode.postMessage({ type, payload, requestId: id });
  });
}

// Listen for responses
window.addEventListener('message', (event) => {
  const { requestId: id, ...rest } = event.data;
  if (id && pendingRequests.has(id)) {
    pendingRequests.get(id)!(rest);
    pendingRequests.delete(id);
  }
});
```

---

## 11. Gotchas and Pitfalls

### Vite CJS deprecation warning

Vite 6 logs a deprecation warning for CJS builds. This is cosmetic -- CJS output via
`build.lib.formats: ['cjs']` still works. The warning refers to Vite's own Node API
being consumed as CJS, not your output format. Suppress it or ignore it.
([source](https://www.eliostruyf.com/vite-bundling-visual-studio-code-extension/))

### CSP in webviews

VSCode webviews enforce Content Security Policy. Inline scripts are blocked by default.
You must inject a `nonce` into all `<script>` tags (shown in GraphPanel.ts above).
The `style-src 'unsafe-inline'` is needed for Svelte's scoped styles and xyflow's
inline CSS.

### `$state.raw` vs `$state` for xyflow

Use `$state.raw()` for nodes and edges arrays. The regular `$state()` wraps values in
Svelte's deep reactivity proxy, which conflicts with xyflow's internal state management
and causes performance issues with large graphs.
([source](https://svelteflow.dev/api-reference/svelte-flow))

### `acquireVsCodeApi()` is singleton

This function can only be called once. Calling it a second time throws an error. Cache
the result in a module-level variable (as shown in `vscode.svelte.ts` above).
([source](https://code.visualstudio.com/api/extension-guides/webview))

### Svelte 5 event syntax

Svelte 5 uses `onnodeclick` (lowercase, no colon) instead of Svelte 4's `on:nodeclick`.
The xyflow docs show both patterns; use the Svelte 5 pattern.
([source](https://svelteflow.dev/api-reference/svelte-flow))

### Test file naming for runes

If your test file needs to use `$state()`, `$derived()`, or `$effect()` directly,
it MUST be named `*.svelte.ts`. Regular `.ts` files cannot use runes.
([source](https://svelte.dev/docs/svelte/testing))

---

## Sources

- [Svelte 5 docs -- runes](https://svelte.dev/docs/svelte/v5-migration-guide)
- [Svelte 5 testing](https://svelte.dev/docs/svelte/testing)
- [@xyflow/svelte docs](https://svelteflow.dev/learn)
- [@xyflow/svelte TypeScript guide](https://svelteflow.dev/learn/advanced/typescript)
- [@xyflow/svelte custom nodes](https://svelteflow.dev/learn/customization/custom-nodes)
- [VSCode webview API](https://code.visualstudio.com/api/extension-guides/webview)
- [VSCode ESM support (v1.100+)](https://github.com/microsoft/vscode/issues/130367)
- [Vite for VSCode extensions](https://www.eliostruyf.com/vite-bundling-visual-studio-code-extension/)
- [@tomjs/vite-plugin-vscode](https://github.com/tomjs/vite-plugin-vscode)
- [eslint-plugin-svelte user guide](https://sveltejs.github.io/eslint-plugin-svelte/user-guide/)
- [ESLint flat config for Svelte+TS](https://gist.github.com/pboling/e8945f4009e5e521c094616783bd4c13)
- [@sveltejs/vite-plugin-svelte](https://www.npmjs.com/package/@sveltejs/vite-plugin-svelte)
- [VSCode extension 2026 guide](https://abdulkadersafi.com/blog/building-vs-code-extensions-in-2026-the-complete-modern-guide)
- [ESM VSCode extension (2025)](https://jan.miksovsky.com/posts/2025/03-17-vs-code-extension)
