// Package semantic implements the middle-end of the LexScript compiler.
//
// Validation passes (executed in order, all errors accumulated):
//
//  1. Duplicate symbol detection (parties, amounts, time_limits, states)
//  2. Type checking — valid currency codes and positive time durations (REQ-2.3)
//  3. Reference resolution — every name used must be declared
//  4. State body completeness — no dead-end states (REQ-2.2)
//  5. Reachability — BFS from the first state flags unreachable nodes (REQ-2.1/2.2)
package semantic

import (
	"fmt"
	"strings"

	"lexscript/pkg/ast"

	"github.com/alecthomas/participle/v2/lexer"
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
	// Pass 3 — Reachability (BFS from first declared state)
	// -----------------------------------------------------------------------
	var initialState string
	for _, decl := range c.Declarations {
		if decl.State != nil {
			initialState = decl.State.Name
			break
		}
	}

	if initialState != "" && len(states) > 1 {
		reachable := make(map[string]bool)
		reachable[initialState] = true
		queue := []string{initialState}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			for _, decl := range c.Declarations {
				if decl.State == nil || decl.State.Name != curr {
					continue
				}
				for _, body := range decl.State.Body {
					if body.Transition != nil {
						tgt := body.Transition.Target
						if !reachable[tgt] {
							reachable[tgt] = true
							queue = append(queue, tgt)
						}
					}
				}
			}
		}

		for name, pos := range states {
			if !reachable[name] {
				errs = append(errs, Error{pos, fmt.Sprintf(
					"state %q is unreachable from the initial state %q (REQ-2.1/2.2)",
					name, initialState,
				)})
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
