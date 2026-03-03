// LexScript LSP client — spawns `lexs lsp` and connects over stdio.

import { workspace, ExtensionContext } from "vscode";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
} from "vscode-languageclient/node";
import { resolveLexsBinary } from "./resolveBinary";

let client: LanguageClient | undefined;

export function startClient(context: ExtensionContext): void {
  const command = resolveLexsBinary();

  const serverOptions: ServerOptions = {
    run:   { command, args: ["lsp"] },
    debug: { command, args: ["lsp"] },
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
