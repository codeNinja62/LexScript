// Package semantic implements the middle-end of the LexScript compiler.
//
// Validation passes (executed in order, all errors accumulated):
//
//  1. Duplicate symbol detection (parties, amounts, time_limits, states)
//  2. Type checking — valid currency codes and positive time durations (REQ-2.3)
//  3. Reference resolution — every name used must be declared
//  4. State body completeness — no dead-end states (REQ-2.2)
//  5. Cycle detection via Tarjan's SCC + Reachability from initial state (REQ-2.1)
package semantic

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"lexscript/pkg/ast"

	"github.com/alecthomas/participle/v2/lexer"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"gonum.org/v1/gonum/graph/traverse"
)

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

// Error is a single semantic validation failure with source location.
type Error struct {
	Pos     lexer.Position
	Message string
}

func (e Error) Error() string {
	if e.Pos.Filename != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.Pos.Filename, e.Pos.Line, e.Pos.Column, e.Message)
	}
	return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Column, e.Message)
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Validate runs all semantic analysis passes over a parsed Contract.
// All errors are collected and returned — analysis is NOT short-circuited
// on the first failure, giving the programmer a complete error list.
func Validate(c *ast.Contract) []Error {
	var errs []Error

	// Symbol tables populated during Pass 1
	parties := make(map[string]lexer.Position)
	amounts := make(map[string]lexer.Position)
	timeLimits := make(map[string]lexer.Position)
	dates := make(map[string]lexer.Position)
	states := make(map[string]lexer.Position)

	// -----------------------------------------------------------------------
	// Pass 1 — Build symbol tables + duplicate detection + type checking
	// -----------------------------------------------------------------------
	for _, decl := range c.Declarations {
		switch {

		case decl.Party != nil:
			p := decl.Party
			if prev, ok := parties[p.Name]; ok {
				errs = append(errs, Error{p.Pos, fmt.Sprintf(
					"duplicate party %q (previously declared at %d:%d)",
					p.Name, prev.Line, prev.Column,
				)})
			} else {
				parties[p.Name] = p.Pos
			}

		case decl.Amount != nil:
			a := decl.Amount
			if prev, ok := amounts[a.Name]; ok {
				errs = append(errs, Error{a.Pos, fmt.Sprintf(
					"duplicate amount %q (previously declared at %d:%d)",
					a.Name, prev.Line, prev.Column,
				)})
			} else {
				amounts[a.Name] = a.Pos
			}
			// REQ-2.3 — currency type check
			if !validCurrency(a.Currency) {
				errs = append(errs, Error{a.Pos, fmt.Sprintf(
					"unknown currency %q in amount %q; valid codes: USD EUR GBP JPY CAD AUD CHF",
					a.Currency, a.Name,
				)})
			}
			if a.Value < 0 {
				errs = append(errs, Error{a.Pos, fmt.Sprintf(
					"amount %q is negative (%.2f); currency values must be non-negative",
					a.Name, a.Value,
				)})
			}

		case decl.TimeLimit != nil:
			tl := decl.TimeLimit
			if prev, ok := timeLimits[tl.Name]; ok {
				errs = append(errs, Error{tl.Pos, fmt.Sprintf(
					"duplicate time_limit %q (previously declared at %d:%d)",
					tl.Name, prev.Line, prev.Column,
				)})
			} else {
				timeLimits[tl.Name] = tl.Pos
			}
			// REQ-2.3 — time unit type check
			if !validTimeUnit(tl.Unit) {
				errs = append(errs, Error{tl.Pos, fmt.Sprintf(
					"unknown time unit %q in time_limit %q; valid units: days months years business_days hours weeks",
					tl.Unit, tl.Name,
				)})
			}
			if tl.Value <= 0 {
				errs = append(errs, Error{tl.Pos, fmt.Sprintf(
					"time_limit %q has non-positive value %d; durations must be positive",
					tl.Name, tl.Value,
				)})
			}

		case decl.State != nil:
			s := decl.State
			if prev, ok := states[s.Name]; ok {
				errs = append(errs, Error{s.Pos, fmt.Sprintf(
					"duplicate state %q (previously declared at %d:%d)",
					s.Name, prev.Line, prev.Column,
				)})
			} else {
				states[s.Name] = s.Pos
			}

		case decl.Date != nil:
			// Phase 3 — date declaration: duplicate check + calendar validation
			d := decl.Date
			if prev, ok := dates[d.Name]; ok {
				errs = append(errs, Error{d.Pos, fmt.Sprintf(
					"duplicate date %q (previously declared at %d:%d)",
					d.Name, prev.Line, prev.Column,
				)})
			} else {
				dates[d.Name] = d.Pos
			}
			if !validDate(d.Value) {
				errs = append(errs, Error{d.Pos, fmt.Sprintf(
					"invalid date value %q in date %q; expected ISO 8601 format YYYY-MM-DD with a real calendar date",
					d.Value, d.Name,
				)})
			}
		}
	}

	// -----------------------------------------------------------------------
	// Pass 2 — Reference resolution + state body completeness
	// -----------------------------------------------------------------------
	validTermKinds := map[string]bool{"fulfilled": true, "breached": true, "expired": true}

	for _, decl := range c.Declarations {
		if decl.State == nil {
			continue
		}
		s := decl.State
		hasTerminate := false
		hasTransition := false

		for _, body := range s.Body {
			switch {

			case body.Require != nil:
				req := body.Require
				// Party reference check
				if _, ok := parties[req.Party]; !ok {
					errs = append(errs, Error{req.Pos, fmt.Sprintf(
						"undefined party %q in require statement (state %q)",
						req.Party, s.Name,
					)})
				}

			case body.Transition != nil:
				hasTransition = true
				tr := body.Transition
				// Target state reference check
				if _, ok := states[tr.Target]; !ok {
					errs = append(errs, Error{tr.Pos, fmt.Sprintf(
						"transition to undefined state %q (state %q)",
						tr.Target, s.Name,
					)})
				}
				// Trigger reference checks
				trig := tr.Trigger
				if trig.TimeLimitRef != nil {
					ref := trig.TimeLimitRef.Ref
					if _, ok := timeLimits[ref]; !ok {
						errs = append(errs, Error{tr.Pos, fmt.Sprintf(
							"time_limit trigger references undeclared variable %q (state %q)",
							ref, s.Name,
						)})
					}
				}
				if trig.BreachRef != nil {
					party := trig.BreachRef.Party
					if _, ok := parties[party]; !ok {
						errs = append(errs, Error{tr.Pos, fmt.Sprintf(
							"breach trigger references undeclared party %q (state %q)",
							party, s.Name,
						)})
					}
				}

			case body.Terminate != nil:
				hasTerminate = true
				if !validTermKinds[body.Terminate.Kind] {
					errs = append(errs, Error{body.Terminate.Pos, fmt.Sprintf(
						"invalid terminate kind %q in state %q; expected: fulfilled breached expired",
						body.Terminate.Kind, s.Name,
					)})
				}
			}
		}

		// REQ-2.2 — Every state must either terminate or have at least one outgoing transition
		if !hasTerminate && !hasTransition {
			errs = append(errs, Error{s.Pos, fmt.Sprintf(
				"state %q is a dead end: has no terminate statement and no transitions; "+
					"all execution paths must eventually reach a terminate node (REQ-2.2)",
				s.Name,
			)})
		}
	}

	// -----------------------------------------------------------------------
	// Pass 3 — Cycle detection (Tarjan SCC) + Reachability (REQ-2.1)
	//
	// gonum/graph replaces the previous manual BFS.
	// - TarjanSCC detects deadlock cycles: SCCs with >1 node, or a
	//   single-node SCC with a self-loop.
	// - traverse.BreadthFirst checks reachability from the initial state.
	// -----------------------------------------------------------------------
	var initialState string
	for _, decl := range c.Declarations {
		if decl.State != nil {
			initialState = decl.State.Name
			break
		}
	}

	if initialState != "" {
		// Build directed graph: nodes = states, edges = transitions.
		g := simple.NewDirectedGraph()
		stateToID := make(map[string]int64)
		idToState := make(map[int64]string)
		var nextID int64
		for name := range states {
			stateToID[name] = nextID
			idToState[nextID] = name
			g.AddNode(simple.Node(nextID))
			nextID++
		}
		for _, decl := range c.Declarations {
			if decl.State == nil {
				continue
			}
			fromID := stateToID[decl.State.Name]
			for _, body := range decl.State.Body {
				if body.Transition == nil {
					continue
				}
				toID, ok := stateToID[body.Transition.Target]
				if !ok {
					continue // undefined target already caught in Pass 2
				}
				if !g.HasEdgeFromTo(fromID, toID) {
					g.SetEdge(g.NewEdge(simple.Node(fromID), simple.Node(toID)))
				}
			}
		}

		// --- Cycle detection: Tarjan's Strongly Connected Components ---
		sccs := topo.TarjanSCC(g)
		for _, scc := range sccs {
			switch len(scc) {
			case 1:
				// Single-node SCC — only a cycle when there is a self-loop.
				n := scc[0]
				if g.HasEdgeFromTo(n.ID(), n.ID()) {
					name := idToState[n.ID()]
					errs = append(errs, Error{states[name], fmt.Sprintf(
						"state %q has a self-transition (cycle); all execution paths "+
							"must eventually reach a terminate node "+
							"(REQ-2.1 — Tarjan SCC cycle detection)",
						name,
					)})
				}
			default:
				// Multi-node SCC = deadlock cycle among two or more states.
				names := make([]string, len(scc))
				for i, n := range scc {
					names[i] = idToState[n.ID()]
				}
				sort.Strings(names)
				// Use the position of the lexically first state in the cycle.
				var cyclePos lexer.Position
				for i, name := range names {
					if i == 0 || states[name].Line < cyclePos.Line {
						cyclePos = states[name]
					}
				}
				errs = append(errs, Error{cyclePos, fmt.Sprintf(
					"states %v form a deadlock cycle; no execution path from this "+
						"group can reach a terminate node "+
						"(REQ-2.1 — Tarjan SCC cycle detection)",
					names,
				)})
			}
		}

		// --- Reachability: BFS from first declared state (REQ-2.1) ---
		if len(states) > 1 {
			startID := stateToID[initialState]
			reachable := make(map[int64]bool)
			reachable[startID] = true
			bfsWalker := &traverse.BreadthFirst{}
			bfsWalker.Walk(g, simple.Node(startID), func(n graph.Node, _ int) bool {
				reachable[n.ID()] = true
				return false
			})
			for name, pos := range states {
				if !reachable[stateToID[name]] {
					errs = append(errs, Error{pos, fmt.Sprintf(
						"state %q is unreachable from the initial state %q (REQ-2.1)",
						name, initialState,
					)})
				}
			}
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func validCurrency(code string) bool {
	switch strings.ToUpper(code) {
	case "USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF":
		return true
	}
	return false
}

func validTimeUnit(unit string) bool {
	switch strings.ToLower(unit) {
	case "days", "months", "years", "business_days", "hours", "weeks":
		return true
	}
	return false
}

// validDate checks that s is an ISO 8601 date (YYYY-MM-DD) representing a
// real calendar date. Uses time.Parse for strict validation.
// Phase 3 — date arithmetic primitive.
func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}
