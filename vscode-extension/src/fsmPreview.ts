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
    body {
      margin: 0;
      padding: 16px;
      background: var(--vscode-editor-background);
      color: var(--vscode-editor-foreground);
      font-family: var(--vscode-font-family);
      overflow: auto;
    }
    #graph {
      text-align: center;
    }
    #graph svg {
      max-width: 100%;
      height: auto;
    }
    #error {
      color: var(--vscode-errorForeground);
      white-space: pre-wrap;
      font-family: monospace;
      display: none;
    }
    .loading {
      text-align: center;
      padding: 2em;
      opacity: 0.6;
    }
  </style>
</head>
<body>
  <div id="graph"><div class="loading">Waiting for FSM data…</div></div>
  <pre id="error"></pre>

  <!-- viz.js standalone (CDN for simplicity; could be bundled) -->
  <script src="https://unpkg.com/@viz-js/viz@3.4.0/lib/viz-standalone.js"></script>
  <script>
    const vscode = acquireVsCodeApi();
    const graphEl = document.getElementById("graph");
    const errorEl = document.getElementById("error");

    window.addEventListener("message", async (event) => {
      const msg = event.data;
      if (msg.type === "dot") {
        errorEl.style.display = "none";
        try {
          const viz = await Viz.instance();
          const svg = viz.renderSVGElement(msg.dot);
          graphEl.innerHTML = "";
          graphEl.appendChild(svg);
        } catch (e) {
          errorEl.textContent = "Render error: " + e.message;
          errorEl.style.display = "block";
        }
      } else if (msg.type === "error") {
        graphEl.innerHTML = "";
        errorEl.textContent = msg.message;
        errorEl.style.display = "block";
      }
    });
  </script>
</body>
</html>`;
}
