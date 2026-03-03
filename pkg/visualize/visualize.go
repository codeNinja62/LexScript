// Package visualize emits a Graphviz DOT representation of a LexScript
// contract's finite state machine.
//
// The generated .dot file can be rendered with the graphviz toolchain:
//
//	dot -Tpng contract.dot -o contract.png
//	dot -Tsvg contract.dot -o contract.svg
//
// Node shapes:
//   - Non-terminal states      → rounded rectangle (shape=box, style=rounded)
//   - fulfilled terminal states → double circle, green fill
//   - breached terminal states  → double circle, red fill
//   - expired terminal states   → double circle, gold fill
//
// Edge labels show the trigger type: event names, time_limit(ref),
// or breach(party).
package visualize

import (
	"fmt"
	"strings"

	"lexscript/pkg/ast"
)

// DOT generates a Graphviz DOT source string for the contract's FSM.
func DOT(c *ast.Contract) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("digraph %s {\n", sanitize(c.Name)))
	b.WriteString("    rankdir=LR;\n")
	b.WriteString("    splines=curved;\n")
	b.WriteString("    concentrate=true;\n")
	b.WriteString("    nodesep=0.7;\n")
	b.WriteString("    ranksep=1.4;\n")
	b.WriteString("    graph [fontname=\"Helvetica\" fontsize=13 label=\"")
	b.WriteString(escDOT(c.Name))
	b.WriteString(" — State Machine\" labelloc=t labeljust=c pad=0.4];\n")
	b.WriteString("    node  [fontname=\"Helvetica\" fontsize=11 margin=\"0.2,0.12\"];\n")
	b.WriteString("    edge  [fontname=\"Helvetica\" fontsize=10];\n\n")

	// Find the initial state (first declared state).
	var initialState string
	for _, decl := range c.Declarations {
		if decl.State != nil {
			initialState = decl.State.Name
			break
		}
	}

	// Invisible entry node that points to the initial state.
	if initialState != "" {
		b.WriteString("    // Entry point\n")
		b.WriteString("    __start [label=\"\" shape=point width=0.2];\n")
		b.WriteString(fmt.Sprintf("    __start -> %s;\n\n",
			sanitize(initialState)))
	}

	// Emit node definitions.
	b.WriteString("    // States\n")
	for _, decl := range c.Declarations {
		if decl.State == nil {
			continue
		}
		s := decl.State
		termKind := terminalKind(s)
		b.WriteString(nodeDefinition(s.Name, termKind))
	}

	// Emit edges (transitions).
	b.WriteString("\n    // Transitions\n")
	for _, decl := range c.Declarations {
		if decl.State == nil {
			continue
		}
		s := decl.State
		for _, body := range s.Body {
			if body.Transition == nil {
				continue
			}
			tr := body.Transition
			label := triggerLabel(tr.Trigger)
			b.WriteString(fmt.Sprintf("    %s -> %s [label=%q];\n",
				sanitize(s.Name), sanitize(tr.Target), label))
		}
	}

	b.WriteString("}\n")
	return b.String()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// terminalKind returns the terminate kind ("fulfilled", "breached", "expired")
// for a terminal state, or "" for non-terminal states.
func terminalKind(s *ast.StateDecl) string {
	for _, body := range s.Body {
		if body.Terminate != nil {
			return body.Terminate.Kind
		}
	}
	return ""
}

// nodeDefinition returns the DOT node attribute line for a state.
func nodeDefinition(name, termKind string) string {
	label := splitCamel(name)
	switch termKind {
	case "fulfilled":
		return fmt.Sprintf(
			"    %s [label=%q shape=doublecircle style=filled fillcolor=\"#b7f5b7\" color=\"#2e7d32\" penwidth=2 width=1.1 height=1.1 fixedsize=true];\n",
			sanitize(name), label)
	case "breached":
		return fmt.Sprintf(
			"    %s [label=%q shape=doublecircle style=filled fillcolor=\"#ffcdd2\" color=\"#c62828\" penwidth=2 width=1.1 height=1.1 fixedsize=true];\n",
			sanitize(name), label)
	case "expired":
		return fmt.Sprintf(
			"    %s [label=%q shape=doublecircle style=filled fillcolor=\"#fff9c4\" color=\"#f57f17\" penwidth=2 width=1.1 height=1.1 fixedsize=true];\n",
			sanitize(name), label)
	default:
		return fmt.Sprintf(
			"    %s [label=%q shape=box style=\"rounded,filled\" fillcolor=\"#f5f5f5\" color=\"#555555\"];\n",
			sanitize(name), label)
	}
}

// triggerLabel produces a human-readable edge label from a Trigger node.
func triggerLabel(t *ast.Trigger) string {
	switch {
	case t.TimeLimitRef != nil:
		return fmt.Sprintf("⏱ time_limit(%s)", t.TimeLimitRef.Ref)
	case t.BreachRef != nil:
		return fmt.Sprintf("⚠ breach(%s)", t.BreachRef.Party)
	case t.EventName != nil:
		return splitCamel(*t.EventName)
	default:
		return "?"
	}
}

// sanitize converts a state name to a DOT-safe identifier.
// DOT identifiers may only contain [a-zA-Z0-9_].
func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// escDOT escapes a string for use inside a DOT double-quoted string.
func escDOT(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// splitCamel inserts spaces before uppercase letters that follow lowercase
// letters, turning CamelCase identifiers into human-readable labels.
//
//	"AwaitingDeposit" → "Awaiting Deposit"
//	"TenantBreached"  → "Tenant Breached"
func splitCamel(s string) string {
	// Replace underscores with spaces first
	s = strings.ReplaceAll(s, "_", " ")
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
	return b.String()
}
