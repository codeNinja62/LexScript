// Shared binary resolution for the LexScript extension.
// Handles the case where `lexs` is not on PATH by resolving it to
// <workspaceRoot>/bin/lexs[.exe] on Windows automatically.

import * as path from "path";
import { workspace } from "vscode";

/**
 * Returns the absolute path to the `lexs` binary, respecting the
 * `lexscript.serverPath` setting and the current platform.
 */
export function resolveLexsBinary(): string {
  const config = workspace.getConfiguration("lexscript");
  const serverPath: string = config.get("serverPath", "lexs");
  const winExt = process.platform === "win32" ? ".exe" : "";

  if (serverPath === "lexs") {
    // Default: auto-resolve to <workspaceRoot>/bin/lexs[.exe].
    if (workspace.workspaceFolders?.length) {
      return path.join(
        workspace.workspaceFolders[0].uri.fsPath,
        "bin",
        `lexs${winExt}`
      );
    }
    return `lexs${winExt}`;
  }

  if (!path.isAbsolute(serverPath)) {
    const base = workspace.workspaceFolders?.length
      ? workspace.workspaceFolders[0].uri.fsPath
      : ".";
    const resolved = path.join(base, serverPath);
    if (process.platform === "win32" && !path.extname(resolved)) {
      return resolved + ".exe";
    }
    return resolved;
  }

  // Absolute path — use as-is, append .exe on Windows if missing.
  if (process.platform === "win32" && !path.extname(serverPath)) {
    return serverPath + ".exe";
  }
  return serverPath;
}
