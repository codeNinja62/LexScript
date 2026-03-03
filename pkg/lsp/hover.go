// Package lsp — hover.go provides textDocument/hover for LexScript.
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"lexscript/pkg/ast"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *handler) handleHover(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params == nil {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}
	var params HoverParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	uri := params.TextDocument.URI
	src, ok := getDocument(uri)
	if !ok {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	word := wordAtPosition(src, int(params.Position.Line), int(params.Position.Character))
	if word == "" {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	// 1. Keywords
	if desc, ok := keywordDescriptions[word]; ok {
		_ = conn.Reply(ctx, req.ID, &Hover{Contents: MarkupContent{Kind: "markdown", Value: desc}})
		return
	}
	// 2. Currency codes
	if desc, ok := currencyDescriptions[strings.ToUpper(word)]; ok {
		_ = conn.Reply(ctx, req.ID, &Hover{Contents: MarkupContent{Kind: "markdown", Value: desc}})
		return
	}
	// 3. Time units
	if desc, ok := timeUnitDescriptions[strings.ToLower(word)]; ok {
		_ = conn.Reply(ctx, req.ID, &Hover{Contents: MarkupContent{Kind: "markdown", Value: desc}})
		return
	}
	// 4. Action verbs
	if desc, ok := actionVerbDescriptions[strings.ToLower(word)]; ok {
		_ = conn.Reply(ctx, req.ID, &Hover{Contents: MarkupContent{Kind: "markdown", Value: desc}})
		return
	}
	// 5. Forbidden keywords
	if desc, ok := forbiddenDescriptions[strings.ToLower(word)]; ok {
		_ = conn.Reply(ctx, req.ID, &Hover{Contents: MarkupContent{Kind: "markdown", Value: desc}})
		return
	}
	// 6. Declared symbols from AST
	contract, err := ast.Parser.ParseString(uri, src)
	if err == nil {
		if desc := symbolDescription(contract, word); desc != "" {
			_ = conn.Reply(ctx, req.ID, &Hover{Contents: MarkupContent{Kind: "markdown", Value: desc}})
			return
		}
	}

	_ = conn.Reply(ctx, req.ID, nil)
}

// ---------------------------------------------------------------------------
// Symbol lookup
// ---------------------------------------------------------------------------

func symbolDescription(c *ast.Contract, name string) string {
	for _, decl := range c.Declarations {
		switch {
		case decl.Party != nil && decl.Party.Name == name:
			return fmt.Sprintf("**party** `%s`\n\nDeclared contract actor.", name)
		case decl.Amount != nil && decl.Amount.Name == name:
			a := decl.Amount
			cpi := ""
			if a.CpiAdjusted {
				cpi = " *(CPI-adjusted)*"
			}
			return fmt.Sprintf("**amount** `%s` = %.2f %s%s", name, a.Value, a.Currency, cpi)
		case decl.TimeLimit != nil && decl.TimeLimit.Name == name:
			tl := decl.TimeLimit
			return fmt.Sprintf("**time_limit** `%s` = %d %s", name, tl.Value, tl.Unit)
		case decl.Date != nil && decl.Date.Name == name:
			return fmt.Sprintf("**date** `%s` = %s", name, decl.Date.Value)
		case decl.State != nil && decl.State.Name == name:
			s := decl.State
			var parts []string
			for _, b := range s.Body {
				switch {
				case b.Require != nil:
					parts = append(parts, fmt.Sprintf("- require %s %s %s", b.Require.Party, b.Require.Action, b.Require.Object))
				case b.Transition != nil:
					parts = append(parts, fmt.Sprintf("- transition → %s", b.Transition.Target))
				case b.Terminate != nil:
					parts = append(parts, fmt.Sprintf("- terminate %s", b.Terminate.Kind))
				}
			}
			body := ""
			if len(parts) > 0 {
				body = "\n\n" + strings.Join(parts, "\n")
			}
			return fmt.Sprintf("**state** `%s`%s", name, body)
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Static description tables
// ---------------------------------------------------------------------------

var keywordDescriptions = map[string]string{
	"contract":     "**contract** — Top-level block declaring a named contract.",
	"party":        "**party** — Declares a named actor (e.g. Landlord, Tenant).",
	"amount":       "**amount** — Declares a named monetary value.\n\nSyntax: `amount <name> = <value> <currency> [cpi_adjusted];`",
	"time_limit":   "**time_limit** — Declares a named duration.\n\nSyntax: `time_limit <name> = <value> <unit>;`",
	"date":         "**date** — Declares a named calendar date (ISO 8601).\n\nSyntax: `date <name> = YYYY-MM-DD;`",
	"state":        "**state** — Declares a Finite State Machine node (contract phase).",
	"require":      "**require** — Obligation: a party must perform an action.\n\nSyntax: `require <Party> <action> <object>;`",
	"transition":   "**transition** — Directed FSM edge to another state.\n\nTriggers: event name, `time_limit(ref)`, or `breach(party)`.",
	"terminate":    "**terminate** — Marks a terminal state.\n\nKinds: `fulfilled` · `breached` · `expired`",
	"breach":       "**breach(party)** — Transition trigger on material breach.",
	"cpi_adjusted": "**cpi_adjusted** — Optional modifier on `amount`; generates CPI clause.",
	"fulfilled":    "**fulfilled** — Terminal: contract successfully completed.",
	"breached":     "**breached** — Terminal: material breach occurred.",
	"expired":      "**expired** — Terminal: time ran out without performance.",
}

var currencyDescriptions = map[string]string{
	"USD": "**USD** — United States Dollar",
	"EUR": "**EUR** — Euro",
	"GBP": "**GBP** — British Pound Sterling",
	"JPY": "**JPY** — Japanese Yen",
	"CAD": "**CAD** — Canadian Dollar",
	"AUD": "**AUD** — Australian Dollar",
	"CHF": "**CHF** — Swiss Franc",
}

var timeUnitDescriptions = map[string]string{
	"days":          "**days** — Calendar days",
	"business_days": "**business_days** — Weekdays excluding public holidays",
	"weeks":         "**weeks** — 7-day periods",
	"months":        "**months** — Calendar months",
	"years":         "**years** — Calendar years",
	"hours":         "**hours** — Clock hours",
}

var actionVerbDescriptions = map[string]string{
	"pays":      "**pays** — Obligation verb: monetary payment",
	"provides":  "**provides** — Obligation verb: service or access provision",
	"delivers":  "**delivers** — Obligation verb: physical or digital delivery",
	"signs":     "**signs** — Obligation verb: document execution",
	"returns":   "**returns** — Obligation verb: return of property or deposit",
	"transfers": "**transfers** — Obligation verb: transfer of rights or assets",
	"notifies":  "**notifies** — Obligation verb: formal notice delivery",
}

var forbiddenDescriptions = map[string]string{
	"while":   "⚠ **while** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"for":     "⚠ **for** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"loop":    "⚠ **loop** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"goto":    "⚠ **goto** — Forbidden keyword (REQ-1.3). Jump constructs are not supported.",
	"repeat":  "⚠ **repeat** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"until":   "⚠ **until** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"do":      "⚠ **do** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"foreach": "⚠ **foreach** — Forbidden keyword (REQ-1.3). Looping constructs are not supported.",
	"recurse": "⚠ **recurse** — Forbidden keyword (REQ-1.3). Recursion is not supported.",
}

// ---------------------------------------------------------------------------
// Text helpers
// ---------------------------------------------------------------------------

func wordAtPosition(src string, line, char int) string {
	lines := strings.Split(src, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	l := lines[line]
	if char < 0 || char >= len(l) {
		return ""
	}
	start := char
	for start > 0 && isIdentChar(l[start-1]) {
		start--
	}
	end := char
	for end < len(l) && isIdentChar(l[end]) {
		end++
	}
	if start == end {
		return ""
	}
	return l[start:end]
}

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_'
}
