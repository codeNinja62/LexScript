// LexScript FSM Preview — renders the contract state machine as an
// interactive graph inside a VS Code Webview panel.
//
// Flow:
//   1. On command, read the active .lxs file.
//   2. Spawn `lexs visualize -` (stdin) → stdout DOT output.
//   3. Pass DOT string to viz.js inside the webview to render SVG.
//   4. On document change, re-run and update the preview.

import * as vscode from "vscode";
import * as path from "path";
import { execFile } from "child_process";
import { resolveLexsBinary } from "./resolveBinary";

let panel: vscode.WebviewPanel | undefined;
let extensionContext: vscode.ExtensionContext | undefined;

export function registerFsmPreview(context: vscode.ExtensionContext): void {
  extensionContext = context;
  const disposable = vscode.commands.registerCommand(
    "lexscript.showFsmPreview",
    () => openPreview(context)
  );
  context.subscriptions.push(disposable);

  // Live-update on save.
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((doc) => {
      if (doc.languageId === "lexscript" && panel) {
        updatePreview(doc.getText());
      }
    })
  );
}

function openPreview(context: vscode.ExtensionContext): void {
  const editor = vscode.window.activeTextEditor;
  if (!editor || editor.document.languageId !== "lexscript") {
    vscode.window.showWarningMessage(
      "Open a .lxs file first to preview its FSM."
    );
    return;
  }

  if (panel) {
    panel.reveal(vscode.ViewColumn.Beside);
  } else {
    panel = vscode.window.createWebviewPanel(
      "lexscriptFsmPreview",
      "LexScript FSM Preview",
      vscode.ViewColumn.Beside,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [
          vscode.Uri.file(path.join(context.extensionPath, "node_modules", "@viz-js", "viz", "dist"))
        ],
      }
    );
    panel.onDidDispose(() => {
      panel = undefined;
    });
    panel.webview.html = getWebviewHtml(panel.webview, context);
  }

  updatePreview(editor.document.getText());
}

function updatePreview(source: string): void {
  if (!panel) return;

  const command = resolveLexsBinary();

  // Run `lexs visualize --stdin` feeding the .lxs source via stdin.
  const child = execFile(
    command,
    ["visualize", "--stdin"],
    { timeout: 10000, maxBuffer: 1024 * 1024 },
    (error, stdout, stderr) => {
      if (error) {
        panel?.webview.postMessage({
          type: "error",
          message: stderr || error.message,
        });
        return;
      }
      panel?.webview.postMessage({ type: "dot", dot: stdout });
    }
  );

  if (child.stdin) {
    child.stdin.write(source);
    child.stdin.end();
  }
}

function getWebviewHtml(webview: vscode.Webview, context: vscode.ExtensionContext): string {
  const vizDiskPath = vscode.Uri.file(
    path.join(context.extensionPath, "node_modules", "@viz-js", "viz", "dist", "viz-global.js")
  );
  const vizUri = webview.asWebviewUri(vizDiskPath);
  const nonce = getNonce();
  return /* html */ `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <meta http-equiv="Content-Security-Policy"
        content="default-src 'none';
                 script-src 'nonce-${nonce}';
                 style-src 'unsafe-inline';" />
  <title>LexScript FSM Preview</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: var(--vscode-editor-background);
      color: var(--vscode-editor-foreground);
      font-family: var(--vscode-font-family);
      overflow: hidden;
      width: 100vw;
      height: 100vh;
    }
    #toolbar {
      position: fixed;
      top: 8px;
      right: 12px;
      display: flex;
      gap: 6px;
      z-index: 10;
    }
    #toolbar button {
      background: var(--vscode-button-background);
      color: var(--vscode-button-foreground);
      border: none;
      border-radius: 4px;
      padding: 3px 9px;
      cursor: pointer;
      font-size: 14px;
    }
    #toolbar button:hover {
      background: var(--vscode-button-hoverBackground);
    }
    #viewport {
      width: 100%;
      height: 100%;
      overflow: hidden;
      cursor: grab;
    }
    #viewport.panning { cursor: grabbing; }
    #canvas {
      transform-origin: 0 0;
      display: inline-block;
    }
    #canvas svg {
      display: block;
    }
    #error {
      position: fixed;
      top: 40px;
      left: 12px;
      right: 12px;
      color: var(--vscode-errorForeground);
      background: var(--vscode-inputValidation-errorBackground);
      border: 1px solid var(--vscode-inputValidation-errorBorder);
      padding: 8px 12px;
      border-radius: 4px;
      white-space: pre-wrap;
      font-family: monospace;
      font-size: 12px;
      display: none;
    }
    #loading {
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      opacity: 0.5;
    }
  </style>
</head>
<body>
  <div id="toolbar">
    <button id="btn-zoom-in"  title="Zoom in">+</button>
    <button id="btn-zoom-out" title="Zoom out">−</button>
    <button id="btn-fit"      title="Fit to window">⊡</button>
  </div>
  <div id="viewport">
    <div id="canvas"></div>
  </div>
  <div id="error"></div>
  <div id="loading">Waiting for FSM data…</div>

  <script nonce="${nonce}" src="${vizUri}"></script>
  <script nonce="${nonce}">
    const vscode    = acquireVsCodeApi();
    const viewport  = document.getElementById("viewport");
    const canvas    = document.getElementById("canvas");
    const errorEl   = document.getElementById("error");
    const loadingEl = document.getElementById("loading");

    // ── Transform state ──────────────────────────────────────────────────────
    let scale = 1, tx = 0, ty = 0;

    function applyTransform() {
      canvas.style.transform = \`translate(\${tx}px, \${ty}px) scale(\${scale})\`;
    }

    function fitToWindow() {
      const svg = canvas.querySelector("svg");
      if (!svg) return;
      const vw = viewport.clientWidth  - 32;
      const vh = viewport.clientHeight - 32;
      const sw = svg.getBBox ? svg.getBBox().width  : svg.clientWidth;
      const sh = svg.getBBox ? svg.getBBox().height : svg.clientHeight;
      if (!sw || !sh) return;
      scale = Math.min(vw / sw, vh / sh, 1);
      tx = (vw - sw * scale) / 2 + 16;
      ty = (vh - sh * scale) / 2 + 16;
      applyTransform();
    }

    // ── Zoom ─────────────────────────────────────────────────────────────────
    viewport.addEventListener("wheel", (e) => {
      e.preventDefault();
      const factor = e.deltaY < 0 ? 1.1 : 0.91;
      const rect   = viewport.getBoundingClientRect();
      const mx = e.clientX - rect.left;
      const my = e.clientY - rect.top;
      tx = mx - (mx - tx) * factor;
      ty = my - (my - ty) * factor;
      scale *= factor;
      applyTransform();
    }, { passive: false });

    document.getElementById("btn-zoom-in").addEventListener("click", () => {
      const cx = viewport.clientWidth / 2, cy = viewport.clientHeight / 2;
      tx = cx - (cx - tx) * 1.2; ty = cy - (cy - ty) * 1.2; scale *= 1.2;
      applyTransform();
    });
    document.getElementById("btn-zoom-out").addEventListener("click", () => {
      const cx = viewport.clientWidth / 2, cy = viewport.clientHeight / 2;
      tx = cx - (cx - tx) * 0.83; ty = cy - (cy - ty) * 0.83; scale *= 0.83;
      applyTransform();
    });
    document.getElementById("btn-fit").addEventListener("click", fitToWindow);

    // ── Pan ──────────────────────────────────────────────────────────────────
    let dragging = false, dragX = 0, dragY = 0;
    viewport.addEventListener("mousedown", (e) => {
      dragging = true; dragX = e.clientX - tx; dragY = e.clientY - ty;
      viewport.classList.add("panning");
    });
    window.addEventListener("mousemove", (e) => {
      if (!dragging) return;
      tx = e.clientX - dragX; ty = e.clientY - dragY;
      applyTransform();
    });
    window.addEventListener("mouseup", () => {
      dragging = false; viewport.classList.remove("panning");
    });

    // ── Messages from extension ───────────────────────────────────────────
    window.addEventListener("message", async (event) => {
      const msg = event.data;
      if (msg.type === "dot") {
        errorEl.style.display  = "none";
        loadingEl.style.display = "none";
        try {
          const viz = await Viz.instance();
          const svg = viz.renderSVGElement(msg.dot);
          svg.removeAttribute("width");
          svg.removeAttribute("height");
          canvas.innerHTML = "";
          canvas.appendChild(svg);
          // Small delay to let the browser measure the SVG before fitting.
          requestAnimationFrame(() => setTimeout(fitToWindow, 50));
        } catch (e) {
          errorEl.textContent    = "Render error: " + e.message;
          errorEl.style.display  = "block";
        }
      } else if (msg.type === "error") {
        loadingEl.style.display = "none";
        canvas.innerHTML        = "";
        errorEl.textContent     = msg.message;
        errorEl.style.display   = "block";
      }
    });
  </script>
</body>
</html>`;
}

function getNonce(): string {
  let text = "";
  const possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  for (let i = 0; i < 32; i++) {
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  }
  return text;
}
