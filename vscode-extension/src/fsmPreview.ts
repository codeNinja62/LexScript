// LexScript FSM Preview — renders the contract state machine as an
// interactive graph inside a VS Code Webview panel.
//
// Flow:
//   1. On command, read the active .lxs file.
//   2. Spawn `lexs visualize -` (stdin) → stdout DOT output.
//   3. Pass DOT string to viz.js inside the webview to render SVG.
//   4. On document change, re-run and update the preview.

import * as vscode from "vscode";
import { execFile } from "child_process";
import { resolveLexsBinary } from "./resolveBinary";

let panel: vscode.WebviewPanel | undefined;

export function registerFsmPreview(context: vscode.ExtensionContext): void {
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
      }
    );
    panel.onDidDispose(() => {
      panel = undefined;
    });
    panel.webview.html = getWebviewHtml();
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

function getWebviewHtml(): string {
  return /* html */ `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>LexScript FSM Preview</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background: var(--vscode-editor-background);
      color: var(--vscode-editor-foreground);
      font-family: var(--vscode-font-family);
      overflow: hidden;
      height: 100vh;
      display: flex;
      flex-direction: column;
    }
    #toolbar {
      display: flex;
      align-items: center;
      gap: 4px;
      padding: 4px 8px;
      background: var(--vscode-editorWidget-background, #252526);
      border-bottom: 1px solid var(--vscode-editorWidget-border, #454545);
      flex-shrink: 0;
      user-select: none;
    }
    #toolbar button {
      background: var(--vscode-button-secondaryBackground, #3a3d41);
      color: var(--vscode-button-secondaryForeground, #ccc);
      border: 1px solid var(--vscode-button-border, #555);
      border-radius: 3px;
      padding: 2px 8px;
      font-size: 14px;
      cursor: pointer;
      line-height: 1.4;
    }
    #toolbar button:hover {
      background: var(--vscode-button-secondaryHoverBackground, #45494e);
    }
    #toolbar span {
      font-size: 11px;
      opacity: 0.7;
      margin-left: 4px;
    }
    #graph {
      flex: 1;
      overflow: hidden;
      position: relative;
    }
    #graph svg {
      width: 100%;
      height: 100%;
      display: block;
    }
    #error {
      color: var(--vscode-errorForeground);
      white-space: pre-wrap;
      font-family: monospace;
      padding: 16px;
      display: none;
    }
    .loading {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100%;
      opacity: 0.6;
    }
  </style>
</head>
<body>
  <div id="toolbar">
    <button id="btn-zoom-in"  title="Zoom in (scroll up)">+</button>
    <button id="btn-zoom-out" title="Zoom out (scroll down)">−</button>
    <button id="btn-reset"    title="Reset zoom &amp; pan">⊙ Reset</button>
    <span id="zoom-level">100%</span>
  </div>
  <div id="graph"><div class="loading">Waiting for FSM data…</div></div>
  <pre id="error"></pre>

  <script src="https://unpkg.com/@viz-js/viz@3.4.0/lib/viz-standalone.js"></script>
  <script src="https://unpkg.com/svg-pan-zoom@3.6.1/dist/svg-pan-zoom.min.js"></script>
  <script>
    const graphEl   = document.getElementById("graph");
    const errorEl   = document.getElementById("error");
    const zoomLabel = document.getElementById("zoom-level");
    let panZoom = null;

    function updateZoomLabel() {
      if (panZoom) {
        zoomLabel.textContent = Math.round(panZoom.getZoom() * 100) + "%";
      }
    }

    document.getElementById("btn-zoom-in").addEventListener("click", () => {
      panZoom?.zoomIn(); updateZoomLabel();
    });
    document.getElementById("btn-zoom-out").addEventListener("click", () => {
      panZoom?.zoomOut(); updateZoomLabel();
    });
    document.getElementById("btn-reset").addEventListener("click", () => {
      panZoom?.resetZoom(); panZoom?.resetPan(); updateZoomLabel();
    });

    window.addEventListener("message", async (event) => {
      const msg = event.data;
      if (msg.type === "dot") {
        errorEl.style.display = "none";
        try {
          const viz = await Viz.instance();
          const svg = viz.renderSVGElement(msg.dot);
          svg.removeAttribute("width");
          svg.removeAttribute("height");
          svg.style.width  = "100%";
          svg.style.height = "100%";

          if (panZoom) { panZoom.destroy(); panZoom = null; }
          graphEl.innerHTML = "";
          graphEl.appendChild(svg);

          panZoom = svgPanZoom(svg, {
            zoomEnabled:    true,
            panEnabled:     true,
            controlIconsEnabled: false,
            fit:            true,
            center:         true,
            minZoom:        0.1,
            maxZoom:        10,
            zoomScaleSensitivity: 0.3,
            onZoom:         updateZoomLabel,
          });
          updateZoomLabel();
        } catch (e) {
          errorEl.textContent = "Render error: " + e.message;
          errorEl.style.display = "block";
        }
      } else if (msg.type === "error") {
        if (panZoom) { panZoom.destroy(); panZoom = null; }
        graphEl.innerHTML = "";
        errorEl.textContent = msg.message;
        errorEl.style.display = "block";
      }
    });
  </script>
</body>
</html>`;
}
