# Contributing to LexScript

Thank you for your interest in contributing! Here is everything you need to get started.

---

## Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/) — only needed to rebuild the VS Code extension
- [Graphviz](https://graphviz.org/download/) — only needed to render `.dot` files to PNG/SVG

### Clone and build

```bash
git clone https://github.com/<your-org>/lexscript.git
cd lexscript
make build          # compiles → bin/lexs  (bin/lexs.exe on Windows)
make setup          # build + install VS Code extension
```

### Run the tests

```bash
make test
```

### Verify the full pipeline

```bash
make validate-rental
make compile-rental
make compile-rental-pdf
make visualize-rental
```

---

## Project Layout

```
cmd/              CLI subcommands (compile, fmt, parse, validate, visualize, lsp, serve)
pkg/
  ast/            Lexer + parser (participle/v2 struct-tag grammar)
  semantic/       Five-pass semantic validator (accumulates all errors)
  codegen/        Markdown + PDF backends; jurisdiction clause library
  format/         Source formatter
  visualize/      Graphviz DOT emitter
  lsp/            LSP server (diagnostics, hover, completion, definition)
  playground/     Web playground HTTP server + embedded static assets
grammar/          Formal EBNF grammar (keep in sync with ast/ast.go)
vscode-extension/ VS Code extension (TypeScript)
examples/         Sample .lxs contracts with generated outputs
```

---

## Making Changes

### Adding a new DSL keyword or construct

1. Add the AST node struct with `participle/v2` parser tags to `pkg/ast/ast.go`.
2. Update `grammar/grammar.ebnf` to reflect the new grammar rule.
3. Add validation logic to `pkg/semantic/validate.go` — **accumulate errors, never short-circuit**.
4. Add fields to the relevant `*Data` struct in `pkg/codegen/emitter.go` and update `pkg/codegen/templates/contract.md.tmpl`.
5. Update the formatter in `pkg/format/format.go`.
6. Update the LSP hover/completion tables in `pkg/lsp/hover.go` and `pkg/lsp/completion.go`.

### Adding a new jurisdiction

Edit `pkg/codegen/jurisdiction.go` and add a new case to the `JurisdictionClauses` function following the existing pattern.

### Changing the Markdown template

Edit `pkg/codegen/templates/contract.md.tmpl`. The template is embedded at compile time via `//go:embed` — no extra build step needed.

---

## Code Style

- **Go:** `gofmt`-formatted. Run `go fmt ./...` before committing.
- **TypeScript:** Use `tsc` to verify compilation. Run `npm run compile` inside `vscode-extension/`.
- **Semantic validator:** Never add early returns — all five passes must always run to completion.
- **Error messages:** Include the `(REQ-x.x)` tag for errors tied to an SRS requirement.

---

## Pull Request Guidelines

1. Keep PRs focused — one logical change per PR.
2. Include a short description of *why* the change is needed, not just what it does.
3. Regenerate example outputs if you change the compiler or templates:
   ```bash
   make compile-rental compile-software
   make visualize-rental visualize-software
   ```
4. Do not implement the items listed under **Planned but Not Yet Implemented** in `.github/copilot-instructions.md` without discussion — they conflict with planned architecture decisions.

---

## Reporting Issues

Open a GitHub issue with:
- The `.lxs` source that reproduces the problem (or a minimal repro).
- The command you ran and the full output.
- Your OS and Go version (`go version`).
