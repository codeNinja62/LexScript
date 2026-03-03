// LexScript VS Code Extension — main entry point.
//
// Activates when a .lxs file is opened:
//   1. Starts the lexs LSP server as a child process.
//   2. Registers the FSM Preview command.

import * as vscode from "vscode";
import { startClient, stopClient } from "./client";
import { registerFsmPreview } from "./fsmPreview";

export function activate(context: vscode.ExtensionContext): void {
  // Start LSP client independently — a failure here must not prevent the
  // FSM preview command from being registered.
  try {
    startClient(context);
  } catch (err) {
    vscode.window.showWarningMessage(
      `LexScript: LSP server failed to start: ${err}`
    );
  }
  registerFsmPreview(context);
}

export function deactivate(): Thenable<void> | undefined {
  return stopClient();
}
