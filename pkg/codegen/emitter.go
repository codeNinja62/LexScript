// Package codegen implements the backend of the LexScript compiler.
//
// The Emitter walks the validated AST, constructs a strongly-typed
// ContractData model, and feeds it to a Go text/template to produce a
// Markdown document.
//
// REQ-3.1: No generative AI is involved. Output is deterministic:
// identical inputs always produce identical legal text.
//
// REQ-3.2: AST node types are mapped to common-law clauses:
//   - amount + time_limit declarations → §2 Definitions
//   - require statements              → §3 Obligations
//   - transition (time_limit trigger) → Term and Termination clause
//   - transition (breach trigger)     → Breach and Remedy clause
//   - transition (event)              → Condition Precedent clause
//   - terminate fulfilled/breached/expired → §4 Term and Termination sub-clauses
//
// REQ-3.3: Output is a .md file containing no residual DSL syntax.
package codegen

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"lexscript/pkg/ast"
)

// ---------------------------------------------------------------------------
// Template (embedded at compile time — binary is self-contained)
// ---------------------------------------------------------------------------

//go:embed templates/contract.md.tmpl
var contractTemplate string

// ---------------------------------------------------------------------------
// Template Data Model
// ---------------------------------------------------------------------------

// ContractData is the complete data model passed to the Markdown template.
type ContractData struct {
	Title        string
	Date         string
	Parties      []string
	Amounts      []AmountData
	TimeLimits   []TimeLimitData
	Dates        []DateData // Phase 3 — date primitive
	States       []StateData
	Jurisdiction JurisdictionData // Phase 3 — jurisdiction variant
}

// AmountData is the template representation of an AmountDecl.
type AmountData struct {
	Name        string
	Display     string // e.g. "50,000.00 USD"
	Value       float64
	Currency    string
	CpiAdjusted bool // Phase 3 — triggers CPI-indexed definitions clause
}

// TimeLimitData is the template representation of a TimeLimitDecl.
type TimeLimitData struct {
	Name    string
	Display string // e.g. "30 days"
	Value   int
	Unit    string
}

// DateData is the template representation of a DateDecl.
// Phase 3 — native Date primitive for date arithmetic provisions.
type DateData struct {
	Name    string
	Value   string // YYYY-MM-DD (raw ISO 8601 value from source)
	Display string // e.g. "March 1, 2026" (human-readable)
}

// StateData is the template representation of a StateDecl.
type StateData struct {
	Name          string
	IsTerminal    bool
	TermKind      string // "fulfilled" | "breached" | "expired"
	SectionNumber int    // 1-based index among non-terminal states (used for §3 subsection numbering)
	Obligations   []ObligationData
	Transitions   []TransitionData
}

// ObligationData represents a require statement.
type ObligationData struct {
	Party       string
	Action      string // raw DSL verb (e.g. "pays")
	LegalAction string // legal phrasing (e.g. "pay")
	Object      string
}

// TransitionData represents a transition statement.
type TransitionData struct {
	TriggerKind string // "event" | "time_limit" | "breach"
	TriggerText string // human-readable trigger description
	Target      string // target state name
	LegalClause string // full legal clause text ready for template insertion
}

// ---------------------------------------------------------------------------
// Emitter
// ---------------------------------------------------------------------------

// Emitter converts a validated AST into a Markdown document.
type Emitter struct {
	tmpl *template.Template
}

// NewEmitter constructs an Emitter with the embedded template and FuncMap loaded.
func NewEmitter() *Emitter {
	funcMap := template.FuncMap{
		// inc increments an integer (used for 1-based list numbering in templates)
		"inc": func(i int) int { return i + 1 },

		// add adds two integers (used for dynamic section numbering in templates)
		"add": func(a, b int) int { return a + b },

		// titleCase converts snake_case / camelCase identifiers to Title Case words
		"titleCase": titleCase,

		// legalAction maps raw DSL action verbs to infinitive legal phrasing
		"legalAction": legalAction,

		// termClause generates the boilerplate paragraph for a terminal state kind
		"termClause": termClause,
	}

	tmpl := template.Must(
		template.New("contract").Funcs(funcMap).Parse(contractTemplate),
	)
	return &Emitter{tmpl: tmpl}
}

// Emit walks the contract AST, builds ContractData, and renders the template
// to an output Markdown file at outPath.
//
// jurisdiction selects the boilerplate clause library (Phase 3).
// Valid values: "common" (default), "delaware", "california", "uk".
func (e *Emitter) Emit(c *ast.Contract, outPath string, jurisdiction string) error {
	data := e.buildData(c, jurisdiction)

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", outPath, err)
	}
	defer f.Close()

	if err := e.tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// AST → Data Model conversion
// ---------------------------------------------------------------------------

func (e *Emitter) buildData(c *ast.Contract, jurisdiction string) ContractData {
	data := ContractData{
		Title:        c.Name,
		Date:         time.Now().Format("January 2, 2006"),
		Jurisdiction: GetJurisdiction(jurisdiction),
	}

	for _, decl := range c.Declarations {
		switch {
		case decl.Party != nil:
			data.Parties = append(data.Parties, decl.Party.Name)

		case decl.Amount != nil:
			a := decl.Amount
			data.Amounts = append(data.Amounts, AmountData{
				Name:        a.Name,
				Value:       a.Value,
				Currency:    a.Currency,
				Display:     fmt.Sprintf("%.2f %s", a.Value, a.Currency),
				CpiAdjusted: a.CpiAdjusted,
			})

		case decl.TimeLimit != nil:
			tl := decl.TimeLimit
			data.TimeLimits = append(data.TimeLimits, TimeLimitData{
				Name:    tl.Name,
				Value:   tl.Value,
				Unit:    tl.Unit,
				Display: fmt.Sprintf("%d %s", tl.Value, tl.Unit),
			})

		case decl.Date != nil:
			// Phase 3 — date primitive: parse and format for human-readable display
			d := decl.Date
			display := d.Value // fallback: raw YYYY-MM-DD
			if t, err := time.Parse("2006-01-02", d.Value); err == nil {
				display = t.Format("January 2, 2006")
			}
			data.Dates = append(data.Dates, DateData{
				Name:    d.Name,
				Value:   d.Value,
				Display: display,
			})

		case decl.State != nil:
			sd := e.buildStateData(decl.State)
			data.States = append(data.States, sd)
		}
	}
	// Assign 1-based section numbers to non-terminal states for §3 subsection numbering
	nonTermIdx := 0
	for i := range data.States {
		if !data.States[i].IsTerminal {
			nonTermIdx++
			data.States[i].SectionNumber = nonTermIdx
		}
	}
	return data
}

func (e *Emitter) buildStateData(s *ast.StateDecl) StateData {
	sd := StateData{Name: s.Name}

	for _, body := range s.Body {
		switch {
		case body.Require != nil:
			req := body.Require
			sd.Obligations = append(sd.Obligations, ObligationData{
				Party:       req.Party,
				Action:      req.Action,
				LegalAction: legalAction(req.Action),
				Object:      req.Object,
			})

		case body.Transition != nil:
			tr := body.Transition
			sd.Transitions = append(sd.Transitions, buildTransitionData(tr))

		case body.Terminate != nil:
			sd.IsTerminal = true
			sd.TermKind = body.Terminate.Kind
		}
	}
	return sd
}

func buildTransitionData(tr *ast.TransitionStmt) TransitionData {
	td := TransitionData{Target: tr.Target}
	trig := tr.Trigger

	switch {
	case trig.TimeLimitRef != nil:
		ref := trig.TimeLimitRef.Ref
		td.TriggerKind = "time_limit"
		td.TriggerText = fmt.Sprintf("expiration of the %s period", titleCase(ref))
		td.LegalClause = fmt.Sprintf(
			"Upon the expiration of the **%s** period without fulfillment of the "+
				"obligations set forth herein, this Agreement shall automatically "+
				"transition to the **%s** phase.",
			titleCase(ref), titleCase(tr.Target),
		)

	case trig.BreachRef != nil:
		party := trig.BreachRef.Party
		td.TriggerKind = "breach"
		td.TriggerText = fmt.Sprintf("material breach by %s", party)
		td.LegalClause = fmt.Sprintf(
			"In the event of a material breach by **%s**, this Agreement shall "+
				"immediately transition to the **%s** phase, and the non-breaching "+
				"party shall be entitled to seek all remedies available at law or in equity.",
			party, titleCase(tr.Target),
		)

	case trig.EventName != nil:
		event := *trig.EventName
		td.TriggerKind = "event"
		td.TriggerText = event
		td.LegalClause = fmt.Sprintf(
			"Upon the occurrence of **%s**, this Agreement shall transition to the **%s** phase.",
			titleCase(event), titleCase(tr.Target),
		)
	}
	return td
}

// ---------------------------------------------------------------------------
// Template helper functions
// ---------------------------------------------------------------------------

// titleCase converts a snake_case or camelCase identifier to Title Case words.
//
//	"lease_duration" → "Lease Duration"
//	"PaymentReceived" → "Payment Received"
func titleCase(s string) string {
	// Replace underscores with spaces
	s = strings.ReplaceAll(s, "_", " ")
	// Insert space before uppercase letters that follow lowercase letters (camelCase)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if prev >= 'a' && prev <= 'z' {
				b.WriteRune(' ')
			}
		}
		b.WriteRune(r)
	}
	// Capitalise first letter of each word
	words := strings.Fields(b.String())
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// legalAction maps a raw DSL action verb to its legal infinitive form.
func legalAction(action string) string {
	switch strings.ToLower(action) {
	case "pays":
		return "pay"
	case "provides":
		return "provide"
	case "delivers":
		return "deliver"
	case "signs":
		return "execute"
	case "returns":
		return "return"
	case "transfers":
		return "transfer"
	case "notifies":
		return "notify"
	default:
		return action
	}
}

// termClause returns the boilerplate termination paragraph for a given kind.
func termClause(kind string) string {
	switch kind {
	case "fulfilled":
		return "This Agreement terminates upon the successful performance of all obligations " +
			"by all parties, constituting complete fulfillment of the Agreement. " +
			"Upon termination by fulfillment, all rights and obligations under this " +
			"Agreement shall cease, and the parties shall have no further liability to one another."
	case "breached":
		return "This Agreement terminates immediately upon a material breach by any party. " +
			"The non-breaching party shall be entitled to all remedies available at law or in equity, " +
			"including but not limited to damages, specific performance, and injunctive relief. " +
			"Termination under this clause does not limit any other rights or remedies available " +
			"to the non-breaching party."
	case "expired":
		return "This Agreement terminates automatically upon the expiration of the applicable " +
			"time limit without fulfillment of the obligations set forth herein. " +
			"Unless the parties execute a written extension agreement prior to expiry, " +
			"no party shall have any further obligation to the others after termination by expiry."
	default:
		return "This Agreement terminates as described herein."
	}
}
