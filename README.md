# LexScript

> A compiler that turns structured contract logic into legally formatted agreements — with mathematical guarantees of completeness.

---

## Table of Contents

1. [What Is This? (Plain English)](#1-what-is-this-plain-english)
2. [The Core Problem It Solves](#2-the-core-problem-it-solves)
3. [How It Works — For Lawyers](#3-how-it-works--for-lawyers)
4. [How It Works — For Engineers](#4-how-it-works--for-engineers)
5. [Quick Start](#5-quick-start)
6. [The DSL — Writing a Contract](#6-the-dsl--writing-a-contract)
7. [DSL Language Reference](#7-dsl-language-reference)
8. [CLI Reference](#8-cli-reference)
9. [What the Compiler Checks](#9-what-the-compiler-checks)
10. [Project Structure](#10-project-structure)
11. [Architecture Deep Dive](#11-architecture-deep-dive)
12. [Roadmap](#12-roadmap)
13. [Design Decisions & Constraints](#13-design-decisions--constraints)

---

## 1. What Is This? (Plain English)

**LexScript** is a tool that lets you write the *logic* of a legal agreement in a simple, structured format — and then automatically produces a professionally formatted, legally worded contract document from it.

Think of it like a **spreadsheet formula for contracts**: instead of a lawyer typing out lengthy paragraphs describing what happens when Party A doesn't pay, you write the rule once in a compact, unambiguous form, and the tool generates the full legal language for you — identically, every time.

**The output is a standard Markdown (`.md`) document** or a formatted **PDF (`.pdf`)** — both produced from the same fixed templates. Any word processor, PDF viewer, or legal drafting tool can open them.

**No artificial intelligence is involved.** Every word in the output was written by a human legal drafter and is stored as a fixed template. The compiler simply matches your contract logic to the right template and fills in the blanks. This means:

- The same input always produces the exact same output.
- No hallucinated clauses. No invented obligations.
- The legal text is fully auditable and human-reviewable.

---

## 2. The Core Problem It Solves

### The traditional contract problem

A standard written contract is essentially a description of a *process*: what happens first, what triggers the next step, what happens if someone fails to perform, and how the agreement eventually ends. But natural language is fundamentally ambiguous. Two readers can interpret the same clause differently.

This creates three expensive failure modes:

1. **Ambiguity** — a clause can be argued to mean two opposite things in court.
2. **Incompleteness** — a scenario occurs that the contract simply does not address, leaving parties without recourse.
3. **Inconsistency** — two clauses in the same document contradict each other.

### What this compiler does differently

This compiler treats a contract as a **finite state machine** — a concept from computer science and mathematics. Every contract has:

- A set of **states** (e.g., "Awaiting Payment", "Active", "Fulfilled", "Breached")
- **Transitions** between states, triggered by specific events (e.g., payment received, deadline passed, breach occurred)
- **Terminal states** — points where the contract definitively ends

By forcing the contract author to express the agreement in this form, the compiler can *mathematically verify* properties that are impossible to check in plain text:

| Property | What It Means | How It's Checked |
|---|---|---|
| **No dead ends** | Every situation has a defined outcome | Every state either ends the contract or leads to another state |
| **No unreachable clauses** | No clause exists that can never apply | Graph traversal from the starting state visits all states |
| **Type safety** | "$50,000" is money; "30 days" is time — they cannot be confused | The compiler validates monetary and temporal primitives |
| **Turing-incompleteness** | The contract process always terminates; it cannot loop forever | Looping constructs are forbidden at the language level |

---

## 3. How It Works — For Lawyers

You don't need to know how to program to understand and use this tool. Here is the full process in plain terms.

### Step 1: Write the contract logic

Instead of writing paragraphs, you describe the contract as a series of **phases** (called *states*), **obligations**, and **what causes the phase to change** (called *transitions*).

Here is a rental agreement expressed in the LexScript language:

```
contract RentalAgreement {

    party Landlord;
    party Tenant;

    amount rent             = 1500.00 USD;
    amount security_deposit = 3000.00 USD;

    time_limit lease_duration = 12 months;
    time_limit payment_grace  = 5 business_days;

    state AwaitingDeposit {
        require Tenant pays security_deposit;
        require Tenant pays rent;
        transition PaymentReceived           -> Active;
        transition time_limit(payment_grace) -> Expired;
    }

    state Active {
        require Landlord provides property_access;
        require Tenant pays rent;
        transition breach(Tenant)             -> TenantBreached;
        transition breach(Landlord)           -> LandlordBreached;
        transition time_limit(lease_duration) -> Fulfilled;
    }

    state Fulfilled      { terminate fulfilled; }
    state TenantBreached { terminate breached;   }
    state LandlordBreached { terminate breached; }
    state Expired        { terminate expired;    }
}
```

Reading this requires no technical knowledge. Each line has a direct plain-language meaning:

| Line | Plain English |
|---|---|
| `party Landlord;` | There is a party named Landlord. |
| `amount rent = 1500.00 USD;` | "Rent" means $1,500.00 USD. |
| `time_limit lease_duration = 12 months;` | "Lease Duration" means 12 months. |
| `state AwaitingDeposit { ... }` | There is a phase called "Awaiting Deposit". |
| `require Tenant pays security_deposit;` | The Tenant is obligated to pay the Security Deposit during this phase. |
| `transition PaymentReceived -> Active;` | When payment is received, the contract moves to the "Active" phase. |
| `transition time_limit(payment_grace) -> Expired;` | If the 5-day grace period expires first, the contract moves to "Expired". |
| `transition breach(Tenant) -> TenantBreached;` | If the Tenant materially breaches the agreement, the contract moves to "Tenant Breached". |
| `terminate fulfilled;` | This is the end of the contract — successfully completed. |

### Step 2: Run the compiler

One command (Markdown output, the default):

```
lexs compile RentalAgreement.lxs
```

Or generate a PDF directly:

```
lexs compile RentalAgreement.lxs -f pdf
```

### Step 3: Receive a complete legal document

The compiler produces `RentalAgreement.md` (or `.pdf`) — a formatted agreement containing:

- Title, date, and governing law statement
- Party identification block
- Definitions section (all named amounts and time limits)
- Obligations and Performance section (one clause per phase, per obligation)
- Term and Termination section (one clause per terminal state)
- General Provisions (entire agreement, severability, force majeure, notices, etc.)
- Representations and Warranties
- Signature blocks for each party

> **Important:** The compiler does not replace a lawyer. The output templates were designed to reflect standard common-law structures. A qualified attorney should review any document before execution, particularly for jurisdiction-specific requirements.

---

## 4. How It Works — For Engineers

### Language and toolchain

| Component | Technology |
|---|---|
| Host language | Go 1.24 |
| Parser generator | [`participle/v2`](https://github.com/alecthomas/participle) — grammar expressed as Go struct tags, no codegen step |
| CLI framework | [`cobra`](https://github.com/spf13/cobra) |
| Template engine | Go standard library `text/template` |
| Template embedding | Go `//go:embed` directive — binary is fully self-contained |
| Graph analysis | [`gonum/graph`](https://pkg.go.dev/gonum.org/v1/gonum/graph) — Tarjan SCC + BFS for cycle detection and reachability |
| PDF backend | [`go-pdf/fpdf`](https://github.com/go-pdf/fpdf) — A4 PDF output with inline bold rendering |
| LSP transport | [`sourcegraph/jsonrpc2`](https://github.com/sourcegraph/jsonrpc2) — JSON-RPC 2.0 over stdin/stdout |
| VS Code extension | TypeScript + [`vscode-languageclient`](https://www.npmjs.com/package/vscode-languageclient) |
| Web playground | Go `net/http` + embedded static files + [viz.js](https://github.com/nicknisi/viz.js) for FSM rendering |

### Compiler architecture (three-pass)

```
    Source (.lxs)
       │
       ▼
┌─────────────────────────────────┐
│  FRONTEND                       │
│  1. Pre-scan (forbidden kw.)    │  → hard error on while/for/loop/goto/...
│  2. Lexer (participle/v2)       │  → token stream
│  3. Parser (participle/v2)      │  → AST (*ast.Contract)
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  MIDDLE-END (pkg/semantic)      │
│  Pass 1: Symbol table + dupes   │  duplicate party/state/var detection
│  Pass 2: Type checking          │  currency codes, durations (REQ-2.3)
│  Pass 3: Reference resolution   │  all names must be declared
│  Pass 4: State body completeness│  no dead-end states (REQ-2.2)
│  Pass 5: Tarjan SCC + BFS       │  cycle detection + reachability (REQ-2.1)
└────────────────┬────────────────┘
                 │
                 ▼
┌─────────────────────────────────┐
│  BACKEND (pkg/codegen)          │
│  AST → ContractData model       │
│  text/template render           │  deterministic, no AI (REQ-3.1)
│  Write .md / .pdf output        │
└────────────────┬────────────────┘
                 │
                 ▼
  Output (.md or .pdf)
```

### AST node hierarchy

```
Contract
└── []Declaration
    ├── PartyDecl          { Name string }
    ├── AmountDecl         { Name, Value float64, Currency string }
    ├── TimeLimitDecl      { Name, Value int, Unit string }
    └── StateDecl          { Name, []StateBody }
        └── StateBody (union)
            ├── RequireStmt     { Party, Action, Object string }
            ├── TransitionStmt  { Trigger, Target string }
            │   └── Trigger (union)
            │       ├── TimeLimitTrigger  { Ref string }
            │       ├── BreachTrigger     { Party string }
            │       └── EventName         string
            └── TerminateStmt   { Kind string }  // fulfilled|breached|expired
```

### Why `participle` instead of ANTLR

- No JDK/Java dependency required — the grammar is pure Go struct tags
- The parser struct **is** the AST — no separate tree-building visitor needed
- 2-token lookahead (`UseLookahead(2)`) resolves the `Trigger` alternative without full backtracking
- Sufficient for this LL-style DSL; ANTLR would add complexity without benefit here

### Turing-incompleteness enforcement (REQ-1.3)

The DSL is intentionally **not Turing-complete**. This is a hard design requirement:

1. The lexer does not tokenise looping/jump keywords — they remain as raw identifiers
2. A pre-scan pass runs *before* the parser and scans the raw source bytes for whole-word occurrences of: `while`, `for`, `loop`, `goto`, `repeat`, `until`, `do`, `foreach`, `recurse`
3. Any match produces a compilation error with line number
4. Because no cycles can be expressed in the grammar itself (no loop constructs, no recursive references), every FSM is a Directed Acyclic Graph — guaranteed to halt

---

## 5. Quick Start

### Prerequisites

- Go 1.24 or later ([golang.org](https://golang.org/dl/))

### Install

```bash
# Clone the repository
git clone <repo-url>
cd project

# Download dependencies
go mod tidy

# Build the compiler binary
go build -o bin/lexs .
```

On Windows (PowerShell):
```powershell
go build -o bin\lexs.exe .
```

### Compile your first contract

```bash
# Compile the included rental example (Markdown)
./bin/lexs compile examples/rental.lxs
# Output: examples/rental.md

# Compile to PDF
./bin/lexs compile examples/rental.lxs -f pdf
# Output: examples/rental.pdf
```

### Format a contract source file

```bash
# Print canonical formatting to stdout
./bin/lexs fmt examples/rental.lxs

# Overwrite in-place
./bin/lexs fmt --write examples/rental.lxs
```

### Export the state machine diagram

```bash
./bin/lexs visualize examples/rental.lxs -o examples/rental.dot
# Render with Graphviz: dot -Tpng examples/rental.dot -o examples/rental.png
```

### Verify a contract without generating output

```bash
./bin/lexs validate examples/rental.lxs
# ✓  examples/rental.lxs: no errors
```

### Inspect the parsed AST (developer tool)

```bash
./bin/lexs parse examples/rental.lxs
# Outputs the full Abstract Syntax Tree as JSON
```

---

## 6. The DSL — Writing a Contract

A `.lxs` file has one top-level block: `contract <Name> { ... }`.

Inside, you declare parties, define named values, then describe states.

### Minimal complete example

```
contract SimplePayment {
    party Payer;
    party Payee;

    amount invoice = 5000.00 USD;
    time_limit due_date = 30 days;

    state AwaitingPayment {
        require Payer pays invoice;
        transition PaymentMade           -> Done;
        transition time_limit(due_date)  -> Overdue;
        transition breach(Payer)         -> Defaulted;
    }

    state Done      { terminate fulfilled; }
    state Overdue   { terminate expired;   }
    state Defaulted { terminate breached;  }
}
```

### Three kinds of transitions

| Syntax | When it fires | Generated legal text |
|---|---|---|
| `transition EventName -> TargetState;` | On a named event signal | "Upon the occurrence of **Event Name**, this Agreement shall transition to..." |
| `transition time_limit(varName) -> TargetState;` | When a declared duration elapses | "Upon the expiration of the **Var Name** period without fulfillment..." |
| `transition breach(PartyName) -> TargetState;` | On material breach by a named party | "In the event of a material breach by **Party**, this Agreement shall immediately transition to..." |

### CPI-adjusted amounts (Phase 3)

Append `cpi_adjusted` to any `amount` declaration to generate a Consumer Price Index
adjustment clause in §2 Definitions:

```
amount base_salary = 95000.00 USD cpi_adjusted;
```

Generated: *"Base Salary" means the amount of 95000.00 USD, subject to annual adjustment in accordance with the Consumer Price Index (CPI) as published by the relevant national statistical authority.*

### Named date declarations (Phase 3)

Declare named calendar dates at the top level of a contract:

```
date commencement_date = 2026-04-01;
date contract_expiry   = 2028-04-01;
```

Dates appear in §2 Definitions with both human-readable and ISO 8601 representation:
*"Commencement Date" means April 1, 2026 (2026-04-01).*

Dates are validated as real calendar dates (the semantic pass rejects `2026-02-30`).

### Three terminal states

| Keyword | Meaning | Legal effect |
|---|---|---|
| `terminate fulfilled;` | Successful completion | All obligations discharged; no further liability |
| `terminate breached;` | Material breach | Non-breaching party entitled to all remedies at law |
| `terminate expired;` | Time ran out without performance | Agreement lapses; written extension required to revive |

### Comments

```
// This is a comment — ignored by the compiler
```

---

## 7. DSL Language Reference

### Keywords

| Keyword | Purpose |
|---|---|
| `contract` | Opens the contract block |
| `party` | Declares a named actor |
| `amount` | Declares a named monetary value |
| `time_limit` | Declares a named duration |
| `date` | Declares a named calendar date (ISO 8601) — **Phase 3** |
| `state` | Declares a FSM state / contract phase |
| `require` | Declares an obligation within a state |
| `transition` | Declares a directed edge to another state |
| `terminate` | Marks a terminal state |
| `breach` | Breach trigger in a transition |
| `time_limit(...)` | Time-limit trigger in a transition |
| `cpi_adjusted` | Optional modifier on `amount` — generates CPI-indexed clause — **Phase 3** |

### Valid action verbs (for `require` statements)

`pays` · `provides` · `delivers` · `signs` · `returns` · `transfers` · `notifies`

### Valid currency codes

`USD` · `EUR` · `GBP` · `JPY` · `CAD` · `AUD` · `CHF`

### Valid time units

`days` · `business_days` · `weeks` · `months` · `years` · `hours`

### Forbidden keywords (Turing-incompleteness — REQ-1.3)

The following words cause an immediate compilation error if they appear anywhere in a `.lxs` file:

`while` · `for` · `loop` · `goto` · `repeat` · `until` · `do` · `foreach` · `recurse`

---

## 8. CLI Reference

```
lexs <command> [flags] <input.lxs>
```

### `lexs compile`

Runs the full pipeline: parse → validate → generate.

```
lexs compile <input.lxs> [-f md|pdf] [-j common|delaware|california|uk] [-o <output>]
```

| Flag | Default | Description |
|---|---|---|
| `-f`, `--format` | `md` | Output format: `md` (Markdown) or `pdf` |
| `-j`, `--jurisdiction` | `common` | Jurisdiction clause library: `common`, `delaware`, `california`, `uk` |
| `-o`, `--output` | `<input>.md` or `<input>.pdf` | Explicit output path |

**Jurisdiction variants (Phase 3):**

| Value | Governing Law | Dispute Resolution |
|---|---|---|
| `common` | Common Law | General arbitration / competent jurisdiction |
| `delaware` | State of Delaware | Delaware Court of Chancery / Superior Court |
| `california` | State of California | JAMS arbitration (mediation first) |
| `uk` | England and Wales | LCIA arbitration, London seat |
| `pakistan` | Pakistan (Contract Act 1872) | Arbitration Act 1940 / Karachi seat |

**Exit codes:** `0` = success · `1` = compilation error (full error list printed to stderr)

### `lexs fmt`

Pretty-prints a `.lxs` source file in canonical style. By default writes to stdout; use `--write` to overwrite in-place.

```
lexs fmt [--write] <input.lxs>
```

| Flag | Default | Description |
|---|---|---|
| `-w`, `--write` | false | Overwrite the source file in-place |

### `lexs visualize`

Exports the contract's finite state machine as a Graphviz DOT file.

```
lexs visualize <input.lxs> [-o <output.dot>] [--stdin]
```

| Flag | Default | Description |
|---|---|---|
| `-o`, `--output` | `<input>.dot` | Output path for the `.dot` file |
| `--stdin` | false | Read `.lxs` source from stdin (used by IDE integrations) |

Render with: `dot -Tpng output.dot -o output.png`

### `lexs lsp`

Starts the Language Server Protocol server over stdin/stdout. This is intended to be spawned by an editor extension (e.g. the VS Code LexScript extension), not run manually.

```
lexs lsp
```

**Capabilities:**
- `textDocument/publishDiagnostics` — errors + warnings on every edit
- `textDocument/hover` — keyword and symbol descriptions
- `textDocument/completion` — keywords, declared names, currencies, time units
- `textDocument/definition` — go-to-declaration for parties, amounts, states, dates

### `lexs serve`

Starts the web playground HTTP server.

```
lexs serve [-a <addr>]
```

| Flag | Default | Description |
|---|---|---|
| `-a`, `--addr` | `:8080` | Address to listen on (e.g. `:3000`) |

Open `http://localhost:8080` in a browser to use the playground.

### `lexs validate`

Runs the frontend and middle-end only. Produces no output file. Useful for pre-commit hooks or CI checks.

```
lexs validate <input.lxs>
```

### `lexs parse`

Parses the input and dumps the full Abstract Syntax Tree as formatted JSON to stdout. Useful for debugging grammar issues or inspecting parsed values.

```
lexs parse <input.lxs>
```

### `lexs --help`

Prints help for any command:

```
lexs --help
lexs compile --help
```

---

## 9. What the Compiler Checks

All errors are collected and reported together — the compiler does not stop at the first error.

### Pass 0 — Forbidden keyword pre-scan

Scans raw source bytes *before* the parser runs. Reports line number of any forbidden keyword.

```
error: line 7: forbidden keyword "for" — this DSL is Turing-incomplete and does
not support looping or jump constructs (REQ-1.3)
```

### Pass 1 — Symbol table + duplicate detection

Reports any party, amount, time_limit, or state declared more than once.

```
error: 12:5: duplicate party "Tenant" (previously declared at 9:5)
```

### Pass 2 — Type checking (REQ-2.3)

- Currency codes must be one of the seven valid ISO codes
- Duration values must be positive integers
- Duration units must be one of the six valid unit names
- Monetary values must be non-negative

```
error: 14:5: unknown currency "DOGE" in amount "fee"; valid codes: USD EUR GBP JPY CAD AUD CHF
error: 15:5: time_limit "grace" has non-positive value 0; durations must be positive
```

### Pass 3 — Reference resolution

Every name used must be declared:

```
error: 28:9: transition to undefined state "Payed" (state "Active")
   → did you mean "Paid"?
error: 29:9: time_limit trigger references undeclared variable "deadline" (state "Active")
error: 30:9: breach trigger references undeclared party "Renter" (state "Active")
```

### Pass 4 — State body completeness (REQ-2.2)

Every state must either contain `terminate` or at least one `transition`. A state with neither is a dead end — execution would be trapped there forever.

```
error: 22:5: state "Limbo" is a dead end: has no terminate statement and no transitions;
all execution paths must eventually reach a terminate node (REQ-2.2)
```

### Pass 5 — Cycle detection (Tarjan SCC) + Reachability (REQ-2.1)

Pass 5 builds a directed graph of all states using `gonum/graph`. It then runs two analyses:

**Tarjan's Strongly Connected Components** to detect deadlock cycles — groups of states that can never reach a `terminate` node:

```
error: 4:5: states [A B] form a deadlock cycle; no execution path from this
group can reach a terminate node (REQ-2.1 — Tarjan SCC cycle detection)
```

**Breadth-First Search** from the first declared state to flag unreachable states:

```
error: 35:5: state "GhostState" is unreachable from the initial state "AwaitingPayment" (REQ-2.1)
```

### Pass 6 — Date validation (Phase 3)

Every `date` declaration must be ISO 8601 format (`YYYY-MM-DD`) representing a real calendar date.

```
error: 8:5: invalid date value "2026-02-30" in date "expiry"; expected ISO 8601 format
         YYYY-MM-DD with a real calendar date
error: 9:5: duplicate date "effective_date" (previously declared at 7:5)
```

---

## 10. Project Structure

```
project/
│
├── main.go                          Entry point — delegates to cmd.Execute()
├── go.mod                           Go module definition (lexscript, Go 1.24)
├── go.sum                           Dependency checksums
├── Makefile                         Build, test, lint, and example targets
│
├── grammar/
│   └── grammar.ebnf                 Formal EBNF grammar specification
│
├── cmd/                             CLI layer (Cobra) — no business logic here
│   ├── root.go                      Root command and Execute() entry point
│   ├── compile.go                   `lexs compile` — full pipeline + forbidden-kw scan
│   ├── parse.go                     `lexs parse`      — AST JSON dump (debug)
│   ├── validate.go                  `lexs validate`   — semantic checks only (debug/CI)
│   ├── fmt.go                       `lexs fmt`        — canonical source formatter
│   ├── visualize.go                 `lexs visualize`  — Graphviz DOT export
│   ├── lsp.go                       `lexs lsp`        — start LSP server (Phase 4)
│   └── serve.go                     `lexs serve`      — start web playground (Phase 4)
│
├── pkg/
│   ├── ast/
│   │   └── ast.go                   Lexer definition + all AST node structs + Parser
│   │
│   ├── semantic/
│   │   └── validate.go              5-pass semantic validator (all errors accumulated)
│   │
│   ├── format/
│   │   └── format.go                AST pretty-printer (canonical .lxs formatting)
│   │
│   ├── visualize/
│   │   └── visualize.go             Graphviz DOT emitter (state machine diagram)
│   │
│   ├── codegen/
│   │   ├── emitter.go               AST → ContractData model + text/template renderer
│   │   ├── pdf_emitter.go           AST → PDF document via go-pdf/fpdf
│   │   ├── jurisdiction.go          Jurisdiction-specific boilerplate clauses
│   │   └── templates/
│   │       └── contract.md.tmpl     Master Markdown template (embedded in binary)
│   │
│   ├── lsp/                         Language Server Protocol implementation (Phase 4)
│   │   ├── protocol.go              LSP type definitions (self-contained, no framework)
│   │   ├── server.go                JSON-RPC 2.0 server over stdin/stdout
│   │   ├── diagnostics.go           Compiler pipeline → LSP Diagnostics bridge
│   │   ├── hover.go                 textDocument/hover — keyword & symbol info
│   │   ├── completion.go            textDocument/completion — keywords + declared names
│   │   └── definition.go            textDocument/definition — go-to-declaration
│   │
│   └── playground/                  Web playground HTTP server (Phase 4)
│       ├── server.go                HTTP API: /api/compile, /api/visualize
│       └── static/
│           ├── index.html           Single-page playground UI
│           └── app.js               Client-side compilation & FSM rendering
│
├── vscode-extension/                VS Code extension (Phase 4)
│   ├── package.json                 Extension manifest + language contribution
│   ├── tsconfig.json                TypeScript configuration
│   ├── language-configuration.json  Comment/bracket/folding rules
│   ├── syntaxes/
│   │   └── lexscript.tmLanguage.json  TextMate grammar for syntax highlighting
│   └── src/
│       ├── extension.ts             Extension entry point — activates LSP + commands
│       ├── client.ts                LanguageClient — spawns `lexs lsp` over stdio
│       └── fsmPreview.ts            Webview panel — live FSM graph rendering
│
├── examples/
│   ├── rental.lxs                   Residential rental agreement example
│   ├── rental.md                    Compiled Markdown output
│   ├── rental.pdf                   Compiled PDF output
│   ├── rental.dot                   Graphviz DOT state machine
│   ├── software_dev.lxs             Software development agreement example
│   ├── software_dev.md              Compiled Markdown output
│   ├── software_dev.pdf             Compiled PDF output
│   └── software_dev.dot             Graphviz DOT state machine
│
└── bin/
    └── lexs  (or lexs.exe on Windows) Compiled binary (git-ignored)
```

---

## 11. Architecture Deep Dive

### Frontend: `pkg/ast`

The lexer is defined using `participle/v2`'s `lexer.MustSimple` with seven ordered rules:

| Rule | Pattern | Notes |
|---|---|---|
| `Comment` | `//[^\n]*` | Elided before parser sees tokens |
| `Whitespace` | `\s+` | Elided before parser sees tokens |
| `Arrow` | `\->` | Must precede any `-` rule to avoid greedy mismatch |
| `Float` | `[0-9]+\.[0-9]+` | Must precede `Int` to avoid splitting `1.5` into `1` and `.5` |
| `Int` | `[0-9]+` | |
| `Ident` | `[a-zA-Z_][a-zA-Z0-9_]*` | Covers all keywords; parser disambiguates via literal string matching |
| `Punct` | `[{}();=,]` | Single-character punctuation |

Keywords are not reserved in the lexer — they are matched as string literals in the grammar. This is deliberate: it avoids a common pitfall in `participle` where keyword tokens conflict with identifier tokens.

The parser is built with `UseLookahead(2)`, which enables unambiguous selection of `Trigger` alternatives (`time_limit(...)` vs `breach(...)` vs bare `EventName`) without requiring full backtracking.

### Middle-end: `pkg/semantic`

Validation errors carry source positions from `participle`'s `lexer.Position` (filename + line + column). All five passes run regardless of earlier failures — the programmer sees the complete error list in one compilation attempt.

Pass 5 builds a `gonum/graph/simple.DirectedGraph` from the state transition graph, then runs two analyses:
- **`topo.TarjanSCC`** — detects deadlock cycles (SCCs with more than one node, or single-node SCCs with a self-loop). A cycle means a group of states can never reach a `terminate` node.
- **`traverse.BreadthFirst`** — BFS from the first declared state flags any state not reachable from the contract's entry point.

### Backend: `pkg/codegen`

The template (`contract.md.tmpl`) is embedded into the binary at compile time via `//go:embed`. This means the binary is fully self-contained — no template files need to ship alongside it.

AST-to-legal-text mappings:

| AST construct | Legal clause produced |
|---|---|
| `AmountDecl` | Defined term in §2 Definitions |
| `TimeLimitDecl` | Defined duration in §2 Definitions |
| `RequireStmt` | Obligation paragraph in §3 |
| `TransitionStmt` (event) | Condition precedent clause in §3 |
| `TransitionStmt` (time_limit) | Term and Termination clause in §3 |
| `TransitionStmt` (breach) | Breach and Remedy clause in §3 |
| `TerminateStmt` (fulfilled) | Fulfillment termination clause in §4 |
| `TerminateStmt` (breached) | Breach termination clause in §4 |
| `TerminateStmt` (expired) | Expiry termination clause in §4 |

The template also injects boilerplate §5–§7 sections (General Provisions, Representations & Warranties, Signatures) that are jurisdiction-agnostic and constant across all contracts.

The PDF backend (`pdf_emitter.go`) reuses the same `ContractData` model and renders it using `go-pdf/fpdf` with A4 layout, inline bold text via mixed-font `Write()` calls, and Windows-1252 bullet characters. Both backends produce identical clause content — only the presentation format differs.

---

## 12. Roadmap

### ✓ Phase 2 — Completed

- **Graph-based cycle detection:** `gonum/graph` + Tarjan's SCC replaces BFS for full deadlock cycle detection — REQ-2.1 full implementation
- **PDF output:** `go-pdf/fpdf` backend; `-f pdf` flag on `compile` command
- **`lexs fmt`:** Canonical auto-formatter for `.lxs` source files
- **`lexs visualize`:** Graphviz `.dot` state machine export with colour-coded terminal nodes

### ✓ Phase 3 — Completed

- **Jurisdiction variants:** `--jurisdiction common|delaware|california|uk|pakistan` selects a jurisdiction-specific boilerplate clause library per REQ-3.2. Each jurisdiction provides a distinct governing law statement, severability clause, dispute resolution section (§7), and any additional jurisdiction-specific provisions in §5.
- **`date` primitive:** New `date <name> = YYYY-MM-DD;` top-level declaration. Dates are validated as real ISO 8601 calendar dates by the semantic pass and appear in §2 Definitions with human-readable display (e.g., *April 1, 2026*).
- **CPI-adjusted amounts:** Optional `cpi_adjusted` modifier on any `amount` declaration generates an annual Consumer Price Index adjustment clause in §2 Definitions.

### ✓ Phase 4 — Tooling — Completed

- **LSP server (`lexs lsp`):** Full Language Server Protocol implementation over stdin/stdout using `sourcegraph/jsonrpc2`. Provides real-time diagnostics (error squiggles on every keystroke), hover information (keyword/symbol descriptions), autocompletion (keywords, currencies, time units, declared names), and go-to-definition (jump to party/amount/state/date declarations). Works with any LSP-compatible editor.
- **VS Code extension (`vscode-extension/`):** First-class IDE experience for `.lxs` files — TextMate grammar for syntax highlighting (keywords, currencies, time units, forbidden keywords, dates, numbers, comments), language configuration (bracket matching, folding, indentation), and an FSM Preview command that renders the contract's state machine as an interactive graph in a side panel. The extension spawns `lexs lsp` automatically for diagnostics, hover, completion, and go-to-definition.
- **Web playground (`lexs serve`):** Browser-based editor at `http://localhost:8080` with a split-pane UI — code editor on the left, Markdown output or FSM graph on the right. Supports jurisdiction selection (common/delaware/california/uk), real-time error diagnostics, and FSM visualization via viz.js. All static assets are embedded in the binary via `//go:embed`.

---

## 13. Design Decisions & Constraints

| Decision | Rationale |
|---|---|
| **Turing-incomplete by design** | A contract that loops forever is not a contract — it is a program. Forbidding loops guarantees the FSM halts, which is a legal requirement (parties need certainty). |
| **No generative AI in the backend** | Legal text must be deterministic and auditable. AI-generated clauses introduce non-determinism and cannot be legally reviewed as a fixed artefact (REQ-3.1). |
| **`participle/v2` over ANTLR** | Avoids a JDK/Java runtime dependency. Grammar is co-located with AST types in a single `.go` file. Easier for academic review. |
| **`text/template` over third-party engines** | Zero dependencies. Output is fully predictable. Custom `FuncMap` provides all necessary formatting helpers. |
| **Errors are accumulated, not short-circuited** | A compiler that stops at the first error is frustrating. All five semantic passes run; the programmer sees every problem in one go. |
| **Template embedded with `//go:embed`** | The binary is self-contained. No runtime file path resolution required. Deployment is a single executable. |
| **Go as host language** | Strong typing for AST structs, excellent compilation speed, single-binary deployment, first-class support in `participle/v2`. |
| **`gonum/graph` for cycle detection** | Tarjan's SCC via a well-tested graph library is safer than a hand-rolled DFS. The directed-graph model also makes the BFS reachability pass a natural extension. |
| **`go-pdf/fpdf` for PDF output** | Pure-Go, zero CGO, no external binary dependencies. Core font WinAnsi encoding is handled by using correct byte values for bullet characters rather than UTF-8. |

---

*LexScript v0.4 — Compiler Construction Project, Semester 6*
