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

  // Resolve the binary path.  If it's a relative path, resolve against
  // the workspace root so users can set "serverPath": "./bin/lexs".
  let command = serverPath;
  if (
    !path.isAbsolute(serverPath) &&
    serverPath !== "lexs" &&
    workspace.workspaceFolders?.length
  ) {
    command = path.join(
      workspace.workspaceFolders[0].uri.fsPath,
      serverPath
    );
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
