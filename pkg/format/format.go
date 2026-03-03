// Package format implements the LexScript source-code pretty-printer.
//
// Format() parses a .lxs source string (or uses a pre-parsed AST) and
// re-serialises it in canonical form:
//
//   - 4-space indentation for top-level declarations
//   - 8-space indentation for state-body statements
//   - Blank line between declaration groups (parties / amounts+time_limits / states)
//   - Single blank line between state blocks
//   - Normalised spacing around operators and punctuation
//
// The idempotency property holds: Format(Format(src)) == Format(src).
package format

import (
	"fmt"
	"strings"

	"lexscript/pkg/ast"
)

// Format pretty-prints a validated *ast.Contract and returns the canonical
// .lxs source as a string.
func Format(c *ast.Contract) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("contract %s {\n", c.Name))

	// Collect declarations by kind so they can be emitted in canonical group order:
	// parties → amounts → time_limits → states.
	// Within each group, the original declaration order is preserved.
	type kindedDecl struct {
		kind int // 0=party 1=amount 2=timelimit 3=state
		decl *ast.Declaration
	}
	groups := [4][]kindedDecl{}
	for _, d := range c.Declarations {
		switch {
		case d.Party != nil:
			groups[0] = append(groups[0], kindedDecl{0, d})
		case d.Amount != nil:
			groups[1] = append(groups[1], kindedDecl{1, d})
		case d.TimeLimit != nil:
			groups[2] = append(groups[2], kindedDecl{2, d})
		case d.State != nil:
			groups[3] = append(groups[3], kindedDecl{3, d})
		}
	}

	first := true
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false
		for _, kd := range group {
			d := kd.decl
			switch kd.kind {
			case 0: // party
				b.WriteString(fmt.Sprintf("    party %s;\n", d.Party.Name))
			case 1: // amount
				b.WriteString(fmt.Sprintf("    amount %s = %s %s;\n",
					d.Amount.Name,
					formatFloat(d.Amount.Value),
					d.Amount.Currency,
				))
			case 2: // time_limit
				b.WriteString(fmt.Sprintf("    time_limit %s = %d %s;\n",
					d.TimeLimit.Name,
					d.TimeLimit.Value,
					d.TimeLimit.Unit,
				))
			case 3: // state
				b.WriteString("\n")
				b.WriteString(formatState(d.State))
			}
		}
	}

	b.WriteString("}\n")
	return b.String()
}

// formatState serialises a StateDecl in canonical form.
func formatState(s *ast.StateDecl) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("    state %s {\n", s.Name))
	for _, body := range s.Body {
		switch {
		case body.Require != nil:
			req := body.Require
			b.WriteString(fmt.Sprintf("        require %s %s %s;\n",
				req.Party, req.Action, req.Object))
		case body.Transition != nil:
			tr := body.Transition
			b.WriteString(fmt.Sprintf("        transition %s -> %s;\n",
				formatTrigger(tr.Trigger), tr.Target))
		case body.Terminate != nil:
			b.WriteString(fmt.Sprintf("        terminate %s;\n",
				body.Terminate.Kind))
		}
	}
	b.WriteString("    }\n")
	return b.String()
}

// formatTrigger serialises a Trigger node back to DSL source.
func formatTrigger(t *ast.Trigger) string {
	switch {
	case t.TimeLimitRef != nil:
		return fmt.Sprintf("time_limit(%s)", t.TimeLimitRef.Ref)
	case t.BreachRef != nil:
		return fmt.Sprintf("breach(%s)", t.BreachRef.Party)
	case t.EventName != nil:
		return *t.EventName
	default:
		return "?"
	}
}

// formatFloat renders a float64 with at least two decimal places, trimming
// unnecessary trailing zeros beyond the second decimal place.
//
//	1500.0   → "1500.00"
//	5000.5   → "5000.50"
//	1234.567 → "1234.567"
func formatFloat(v float64) string {
	s := fmt.Sprintf("%.10f", v)
	// Ensure at least two decimal places while removing trailing zeros beyond them.
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return fmt.Sprintf("%.2f", v)
	}
	dec := strings.TrimRight(parts[1], "0")
	if len(dec) < 2 {
		dec = parts[1][:2]
	}
	return parts[0] + "." + dec
}
