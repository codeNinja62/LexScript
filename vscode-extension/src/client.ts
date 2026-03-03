// LexScript LSP client — spawns `lexs lsp` and connects over stdio.

import * as path from "path";
import { workspace, ExtensionContext } from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from "vscode-languageclient/node";

let client: LanguageClient | undefined;

export function startClient(context: ExtensionContext): void {
  const config = workspace.getConfiguration("lexscript");
  const serverPath: string = config.get("serverPath", "lexs");
  const winExt = process.platform === "win32" ? ".exe" : "";

  let command: string;

  if (serverPath === "lexs") {
    // Default: auto-resolve to <workspaceRoot>/bin/lexs[.exe] so the extension
    // works out-of-the-box without requiring lexs on PATH.
    // Node's child_process.spawn does NOT resolve PATH extensions on Windows,
    // so we must use the explicit binary name.
    if (workspace.workspaceFolders?.length) {
      command = path.join(
        workspace.workspaceFolders[0].uri.fsPath,
        "bin",
        `lexs${winExt}`
      );
    } else {
      // No workspace open — fall back to PATH, append .exe on Windows.
      command = `lexs${winExt}`;
    }
  } else if (!path.isAbsolute(serverPath)) {
    // Relative path: resolve against workspace root.
    const base = workspace.workspaceFolders?.length
      ? workspace.workspaceFolders[0].uri.fsPath
      : ".";
    command = path.join(base, serverPath);
    // Append .exe on Windows if no extension provided.
    if (process.platform === "win32" && !path.extname(command)) {
      command += ".exe";
    }
  } else {
    // Absolute path — use as-is, but append .exe on Windows if missing.
    command = serverPath;
    if (process.platform === "win32" && !path.extname(command)) {
      command += ".exe";
    }
  }

  const serverOptions: ServerOptions = {
    run: { command, args: ["lsp"], transport: TransportKind.stdio },
    debug: { command, args: ["lsp"], transport: TransportKind.stdio },
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "lexscript" }],
    synchronize: {
      fileEvents: workspace.createFileSystemWatcher("**/*.lxs"),
    },
  };

  client = new LanguageClient(
    "lexscript",
    "LexScript Language Server",
    serverOptions,
    clientOptions
  );

  client.start();
}

export function stopClient(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}
