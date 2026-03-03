// LexScript VS Code Extension — main entry point.
//
// Activates when a .lxs file is opened:
//   1. Starts the lexs LSP server as a child process.
//   2. Registers the FSM Preview command.

import * as vscode from "vscode";
import { startClient, stopClient } from "./client";
import { registerFsmPreview } from "./fsmPreview";

export function activate(context: vscode.ExtensionContext): void {
  startClient(context);
  registerFsmPreview(context);
}

export function deactivate(): Thenable<void> | undefined {
  return stopClient();
}
