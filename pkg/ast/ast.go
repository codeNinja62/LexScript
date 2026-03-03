// Package ast defines the Abstract Syntax Tree for the LexScript DSL.
//
// The lexer and parser are built with participle/v2.  Grammar rules are
// expressed directly as Go struct field tags, making each struct both the
// grammar rule and the AST node — no separate generated code required.
//
// Formal grammar is documented in grammar/grammar.ebnf.
package ast

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// ---------------------------------------------------------------------------
// Lexer
// ---------------------------------------------------------------------------

// l2lLexer uses a simple ordered regex ruleset.
// Rules are tested top-to-bottom; the first match wins.
// Keywords are matched as string literals in the grammar (e.g. 'contract'),
// so all keywords fall into the generic Ident token — no keyword reservation
// in the lexer is needed.  This avoids word-boundary issues absent from RE2.
var l2lLexer = lexer.MustSimple([]lexer.SimpleRule{
	// Ignored tokens — elided before the parser sees them
	{Name: "Comment", Pattern: `//[^\n]*`},
	{Name: "Whitespace", Pattern: `\s+`},
	// Punctuation / operators (order matters: -> before standalone -)
	{Name: "Arrow", Pattern: `\->`},
	// Date literal must come before Int to prevent 2026-03-01 tokenising as Int
	// Phase 3: native Date primitive (REQ Phase 3 — date arithmetic)
	{Name: "Date", Pattern: `[0-9]{4}-[0-9]{2}-[0-9]{2}`},
	// Numeric literals (float before int so 3.14 is not split into 3 and .14)
	{Name: "Float", Pattern: `[0-9]+\.[0-9]+`},
	{Name: "Int", Pattern: `[0-9]+`},
	// Identifiers (covers keywords; grammar disambiguates via literal matching)
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	// Punctuation
	{Name: "Punct", Pattern: `[{}();=,]`},
})

// ---------------------------------------------------------------------------
// AST Node Types
// ---------------------------------------------------------------------------

// Contract is the root AST node — one per .l2l source file.
//
//	contract <Name> { <declaration>* }
type Contract struct {
	Pos          lexer.Position
	Name         string         `parser:"'contract' @Ident '{'"`
	Declarations []*Declaration `parser:"@@* '}'"`
}

// Declaration is a top-level item inside a contract block.
// Exactly one field is non-nil after parsing.
type Declaration struct {
	Pos       lexer.Position
	Party     *PartyDecl     `parser:"( @@"`
	Amount    *AmountDecl    `parser:"| @@"`
	TimeLimit *TimeLimitDecl `parser:"| @@"`
	Date      *DateDecl      `parser:"| @@"`
	State     *StateDecl     `parser:"| @@ )"`
}

// PartyDecl declares a named actor in the contract.
//
//	party <Name>;
type PartyDecl struct {
	Pos  lexer.Position
	Name string `parser:"'party' @Ident ';'"`
}

// AmountDecl declares a named monetary value (REQ-2.3 currency primitive).
//
//	amount <Name> = <Value> <Currency>;
//	amount <Name> = <Value> <Currency> cpi_adjusted;
//	e.g.: amount rent = 1500.00 USD cpi_adjusted;
//
// Phase 3: optional cpi_adjusted modifier triggers annual CPI-indexed legal clause.
type AmountDecl struct {
	Pos         lexer.Position
	Name        string  `parser:"'amount' @Ident '='"`
	Value       float64 `parser:"@(Float|Int)"`
	Currency    string  `parser:"@Ident"`
	CpiAdjusted bool    `parser:"( @'cpi_adjusted' )? ';'"`
}

// TimeLimitDecl declares a named duration (REQ-2.3 time primitive).
//
//	time_limit <Name> = <Value> <Unit>;
//	e.g.: time_limit lease_duration = 12 months;
type TimeLimitDecl struct {
	Pos   lexer.Position
	Name  string `parser:"'time_limit' @Ident '='"`
	Value int    `parser:"@Int"`
	Unit  string `parser:"@Ident ';'"`
}

// StateDecl defines a contract state — a node in the Finite State Machine.
//
//	state <Name> { <StateBody>* }
type StateDecl struct {
	Pos  lexer.Position
	Name string       `parser:"'state' @Ident '{'"`
	Body []*StateBody `parser:"@@* '}'"`
}

// StateBody is a single statement within a state block.
// Exactly one field is non-nil after parsing.
type StateBody struct {
	Pos        lexer.Position
	Require    *RequireStmt    `parser:"( @@"`
	Transition *TransitionStmt `parser:"| @@"`
	Terminate  *TerminateStmt  `parser:"| @@ )"`
}

// RequireStmt mandates that a party perform an action — an obligation (edge label).
//
//	require <Party> <action> <object>;
//	e.g.: require Buyer pays invoice;
type RequireStmt struct {
	Pos    lexer.Position
	Party  string `parser:"'require' @Ident"`
	Action string `parser:"@Ident"`
	Object string `parser:"@Ident ';'"`
}

// TransitionStmt specifies a directed FSM edge: Trigger → TargetState.
//
//	transition <Trigger> -> <TargetState>;
type TransitionStmt struct {
	Pos     lexer.Position
	Trigger *Trigger `parser:"'transition' @@"`
	Target  string   `parser:"'->' @Ident ';'"`
}

// Trigger is the guard condition that fires a state transition.
// Alternatives are tried left-to-right (participle.UseLookahead handles backtracking):
//
//	time_limit(<ref>)   — triggers when the named duration elapses
//	breach(<party>)     — triggers on a party's material breach
//	<EventName>         — triggers on an explicit named signal/event
type Trigger struct {
	Pos          lexer.Position
	TimeLimitRef *TimeLimitTrigger `parser:"( @@"`
	BreachRef    *BreachTrigger    `parser:"| @@"`
	EventName    *string           `parser:"| @Ident )"`
}

// TimeLimitTrigger fires when a declared time_limit duration elapses.
//
//	time_limit(<ref>)
type TimeLimitTrigger struct {
	Pos lexer.Position
	Ref string `parser:"'time_limit' '(' @Ident ')'"`
}

// BreachTrigger fires on a named party's material breach.
//
//	breach(<party>)
type BreachTrigger struct {
	Pos   lexer.Position
	Party string `parser:"'breach' '(' @Ident ')'"`
}

// DateDecl declares a named calendar date value (Phase 3 — date arithmetic primitive).
//
//	date <Name> = YYYY-MM-DD;
//	e.g.: date effective_date = 2026-03-01;
type DateDecl struct {
	Pos   lexer.Position
	Name  string `parser:"'date' @Ident '='"`
	Value string `parser:"@Date ';'"`
}

// TerminateStmt marks this state as a terminal (leaf) node of the FSM.
// Guarantees the FSM always halts (REQ-2.2).
//
//	terminate fulfilled|breached|expired;
type TerminateStmt struct {
	Pos  lexer.Position
	Kind string `parser:"'terminate' @Ident ';'"`
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

// Parser is the compiled participle parser for the Contract root rule.
//
// Options:
//   - Lexer:       the l2l simple lexer defined above
//   - Elide:       Whitespace and Comment tokens are stripped before parsing
//   - UseLookahead: 2-token lookahead enables unambiguous Trigger alternative
//     selection without full backtracking overhead
var Parser = participle.MustBuild[Contract](
	participle.Lexer(l2lLexer),
	participle.Elide("Whitespace", "Comment"),
	participle.UseLookahead(2),
)
