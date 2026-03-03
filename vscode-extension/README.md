# LexScript

Language support for **LexScript** (`.lxs`) — a DSL for writing legal contracts as finite state machines.

## Features

- **Syntax highlighting** — keywords, states, currencies, dates, forbidden keywords (shown in red)
- **Error diagnostics** — red squiggles on every keystroke via the LSP server
- **Hover info** — hover over any keyword or symbol for documentation
- **Autocompletion** — `Ctrl+Space` for keywords, currencies, time units, declared names
- **Go-to-definition** — `F12` on a party, amount, state, or date to jump to its declaration
- **FSM Preview** — `Ctrl+Shift+P` → `LexScript: Show FSM Preview` — renders the contract's state machine as an interactive graph

## Requirements

The extension requires the `lexs` binary to be built and available.

1. Build the binary from the project root:
   ```
   go build -o bin/lexs.exe .
   ```
2. The extension auto-resolves `<workspaceRoot>/bin/lexs.exe` by default.  
   To use a custom path, set `lexscript.serverPath` in VS Code settings.

## Extension Settings

| Setting | Default | Description |
|---|---|---|
| `lexscript.serverPath` | `lexs` | Path to the `lexs` binary |
| `lexscript.trace.server` | `off` | LSP trace level (`off` / `messages` / `verbose`) |

## Quick Start

Open any `.lxs` file — the extension activates automatically.
