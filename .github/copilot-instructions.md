# LexScript — Copilot Instructions

## What This Project Is

**LexScript** is a Turing-incomplete DSL compiler (binary: `lexs`) that compiles `.lxs` contract files — expressed as finite state machines — into deterministic, legally-formatted Markdown or PDF documents. No AI is involved in output generation; every clause comes from a fixed Go template. (SRS §3.3 REQ-3.1)

## Architecture (Three-Phase Pipeline)

```
.lxs source → [Pre-scan] → [Frontend] → [Middle-end] → [Backend] → .md / .pdf output
                forbidden    participle   semantic       template
                keywords     /v2 parser   validate.go    emitter.go / pdf_emitter.go
```

| Phase | Package | Key file |
|-------|---------|----------|
| Pre-scan (forbidden keywords) | `cmd` | `cmd/compile.go` → `forbiddenKeywordErrors()` |
| Frontend — lexer + parser → AST | `pkg/ast` | `pkg/ast/ast.go` |
| Middle-end — semantic validation | `pkg/semantic` | `pkg/semantic/validate.go` |
| Backend — Markdown codegen | `pkg/codegen` | `pkg/codegen/emitter.go` |
| Backend — PDF codegen | `pkg/codegen` | `pkg/codegen/pdf_emitter.go` |
| Formatter | `pkg/format` | `pkg/format/format.go` |
| DOT visualizer | `pkg/visualize` | `pkg/visualize/visualize.go` |

## SRS Requirement Index

| REQ | Description | Implementation location |
|-----|-------------|------------------------|
| REQ-1.1 | Tokenise `party`, `state`, `require`, `transition`, `terminate`, `time_limit`, `amount` | `pkg/ast/ast.go` lexer rules |
| REQ-1.2 | Parser constructs AST of actors (nodes), obligations (edges), terminal states | `pkg/ast/ast.go` struct tags |
| REQ-1.3 | Reject looping keywords (`while`, `for`, `loop`, `goto`, `repeat`, `until`, `do`, `foreach`, `recurse`) before parsing | `cmd/compile.go` `forbiddenKeywordErrors()` |
| REQ-2.1 | Tarjan SCC cycle detection (deadlock cycles) + BFS reachability from first declared state | `pkg/semantic/validate.go` Pass 3 (gonum/graph) |
| REQ-2.2 | Every state must either `terminate` or have ≥1 `transition` (no dead ends) | `pkg/semantic/validate.go` Pass 4 |
| REQ-2.3 | Type-check currency codes and positive time durations | `pkg/semantic/validate.go` Pass 1 |
| REQ-3.1 | No generative AI — identical inputs always produce identical legal text | `pkg/codegen/emitter.go` + template |
| REQ-3.2 | AST nodes map deterministically to common-law clauses (see table below) | `pkg/codegen/emitter.go` `buildTransitionData()` |
| REQ-3.3 | Output is `.md` with no residual DSL syntax | `pkg/codegen/templates/contract.md.tmpl` |

## AST → Legal Clause Mapping (REQ-3.2)

| AST construct | Legal clause produced |
|---------------|----------------------|
| `AmountDecl` | Defined term in §2 Definitions |
| `TimeLimitDecl` | Defined duration in §2 Definitions |
| `RequireStmt` | Obligation paragraph in §3 |
| `TransitionStmt` (event trigger) | Condition Precedent clause in §3 |
| `TransitionStmt` (time_limit trigger) | Term and Termination clause in §3 |
| `TransitionStmt` (breach trigger) | Breach and Remedy clause in §3 |
| `TerminateStmt fulfilled` | Fulfillment termination clause in §4 |
| `TerminateStmt breached` | Breach termination clause in §4 |
| `TerminateStmt expired` | Expiry termination clause in §4 |

## AST / Parser Pattern

Grammar rules live **directly in Go struct field tags** via [participle/v2](https://github.com/alecthomas/participle). There is no generated code — the struct _is_ the grammar rule.

```go
// pkg/ast/ast.go
type AmountDecl struct {
    Name     string  `parser:"'amount' @Ident '='"`
    Value    float64 `parser:"@(Float|Int)"`
    Currency string  `parser:"@Ident ';'"`
}
```

The canonical grammar is in [grammar/grammar.ebnf](grammar/grammar.ebnf). Keep that file in sync when adding new AST nodes.

Keywords are **not** reserved in the lexer — all identifiers use the `Ident` token; grammar struct tags do string-literal matching (`'contract'`, `'state'`, etc.). The lexer rule order matters: `Arrow` before `-`, `Float` before `Int`.

## Semantic Validation Conventions

- **Never short-circuit** — `Validate()` in `pkg/semantic/validate.go` always accumulates _all_ errors before returning. Do not add early returns.
- Three code sections (covering five conceptual passes): duplicate detection + type checking (REQ-2.3) → reference resolution + state body completeness (REQ-2.2) → Tarjan SCC cycle detection + BFS reachability (REQ-2.1).
- Pass 3 builds a `gonum/graph/simple.DirectedGraph`, runs `topo.TarjanSCC` for deadlock cycle detection, then uses `traverse.BreadthFirst` for reachability. Do **not** revert Pass 3 to a manual BFS loop.
- REQ tags are requirement IDs from the SRS — preserve them in comments when editing related code.

## Codegen / Template Pattern

The Markdown template (`pkg/codegen/templates/contract.md.tmpl`) is **embedded at compile time** via `//go:embed`. Changing the template requires no build step beyond `go build`.

`Emitter.Emit()` converts the validated AST into a `ContractData` struct (strongly typed template model), then executes the template. To add a new DSL construct:
1. Add the AST node struct with parser tags to `pkg/ast/ast.go` and update `grammar/grammar.ebnf`.
2. Add validation logic to `pkg/semantic/validate.go` (accumulate, don't short-circuit).
3. Add fields to the relevant `*Data` struct in `pkg/codegen/emitter.go` and update the template.

## Developer Workflows

```bash
make build                  # compile → bin/lexs
make test                   # go test ./... -v
make compile-rental         # end-to-end: examples/rental.lxs → examples/rental.md
make compile-rental-pdf     # Phase 2: examples/rental.lxs → examples/rental.pdf
make visualize-rental       # Phase 2: examples/rental.lxs → examples/rental.dot
make fmt-rental             # Phase 2: format examples/rental.lxs in-place
make parse-rental           # dump AST as JSON (frontend debug)
make validate-rental        # semantic check only (no output file)

# Direct CLI (after build)
./bin/lexs compile examples/rental.lxs
./bin/lexs compile examples/rental.lxs -o out.md          # Markdown (default)
./bin/lexs compile examples/rental.lxs -f pdf -o out.pdf  # PDF (Phase 2)
./bin/lexs fmt    examples/rental.lxs                     # print canonical source to stdout
./bin/lexs fmt -w examples/rental.lxs                     # overwrite in-place
./bin/lexs visualize examples/rental.lxs -o rental.dot    # Graphviz DOT export
./bin/lexs parse    examples/rental.lxs                   # JSON AST dump
./bin/lexs validate examples/rental.lxs                   # errors only, no output
```

## DSL Key Facts

- Source extension: `.lxs`; output extensions: `.md` (default) or `.pdf` (`-f pdf`)
- Forbidden looping keywords (`while`, `for`, `loop`, `goto`, `repeat`, `until`, `do`, `foreach`, `recurse`) are rejected at pre-scan before parsing (REQ-1.3).
- Every state must either `terminate` or have at least one `transition` (no dead-end states, REQ-2.2).
- Terminal states use `terminate fulfilled|breached|expired`.
- Reachability is checked via gonum BFS from the **first declared state** in the contract (REQ-2.1).
- Deadlock cycles (states that can never reach `terminate`) are detected by Tarjan’s SCC via `gonum/graph/topo.TarjanSCC` (REQ-2.1 full impl).
- Valid currencies: `USD EUR GBP JPY CAD AUD CHF`. Valid time units: `days months years business_days hours weeks`.- **Phase 3 — `date` declarations:** `date <name> = YYYY-MM-DD;` adds a named calendar date to §2 Definitions. Validated as a real ISO 8601 date. Lexer `Date` token rule precedes `Int` to prevent greedy mismatch.
- **Phase 3 — `cpi_adjusted`:** Optional modifier on `amount` declarations. Generates a CPI-indexed adjustment clause in §2.
- **Phase 3 — `--jurisdiction`:** Selects boilerplate clause library: `common` (default), `delaware`, `california`, `uk`. Affects §1 preamble, §2 catchall, §5 severability + additional provisions, and §7 Dispute Resolution.
## Planned but Not Yet Implemented (Roadmap)

Do not implement these without explicit instruction — they conflict with planned architecture decisions:
- **VS Code extension + LSP:** Syntax highlighting and error squiggles for `.lxs` files.
- **Date arithmetic expressions:** Date arithmetic within `require` / `transition` contexts (currently only top-level declarations are supported).
- **Web playground:** Browser-based editor and live contract preview.

## Key Files Reference

| Purpose | File |
|---------|------|
| Root entry point | `main.go` → `cmd/root.go` |
| All CLI subcommands | `cmd/compile.go`, `cmd/parse.go`, `cmd/validate.go`, `cmd/fmt.go`, `cmd/visualize.go` |
| AST node definitions + lexer | `pkg/ast/ast.go` |
| Semantic validation (all passes) | `pkg/semantic/validate.go` |
| Markdown code generation + data model | `pkg/codegen/emitter.go` |
| PDF code generation (Phase 2) | `pkg/codegen/pdf_emitter.go` |
| Jurisdiction clause library (Phase 3) | `pkg/codegen/jurisdiction.go` |
| Markdown output template | `pkg/codegen/templates/contract.md.tmpl` |
| Source formatter (Phase 2) | `pkg/format/format.go` |
| DOT visualizer (Phase 2) | `pkg/visualize/visualize.go` |
| Formal grammar (EBNF) | `grammar/grammar.ebnf` |
| Example contracts | `examples/rental.lxs`, `examples/software_dev.lxs`, `examples/employment.lxs` |
