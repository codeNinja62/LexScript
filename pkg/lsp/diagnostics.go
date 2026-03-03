// Package lsp — diagnostics.go converts compiler errors into LSP Diagnostics.
//
// This is the bridge between the existing compiler pipeline (ast + semantic)
// and the LSP server — no compiler logic is duplicated here.
package lsp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"lexscript/pkg/ast"
	"lexscript/pkg/semantic"
)

// sourceName is the human-readable source tag shown in the Problems panel.
const sourceName = "lexscript"

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Diagnose runs the full compiler front-end and middle-end on raw source
// text and returns a slice of LSP Diagnostics ready to be published.
func Diagnose(uri string, src string) []Diagnostic {
	var diags []Diagnostic

	// --- Pass 0: Forbidden-keyword pre-scan (REQ-1.3) ---
	if fkErrs := forbiddenKeywordDiagnostics(src); len(fkErrs) > 0 {
		diags = append(diags, fkErrs...)
	}

	// --- Frontend: Parse DSL → AST ---
	contract, err := ast.Parser.ParseString(uri, src)
	if err != nil {
		diags = append(diags, parseDiagnostic(err))
		return diags // parse errors are fatal
	}

	// --- Middle-end: Semantic validation ---
	if semErrs := semantic.Validate(contract); len(semErrs) > 0 {
		diags = append(diags, semanticDiagnostics(semErrs)...)
	}

	return diags
}

// ---------------------------------------------------------------------------
// Forbidden-keyword diagnostics (REQ-1.3)
// ---------------------------------------------------------------------------

var forbiddenKwRegex = regexp.MustCompile(
	`\b(while|for|loop|goto|repeat|until|do|foreach|recurse)\b`,
)

func forbiddenKeywordDiagnostics(src string) []Diagnostic {
	var diags []Diagnostic
	lines := strings.Split(src, "\n")
	for lineIdx, line := range lines {
		matches := forbiddenKwRegex.FindAllStringIndex(line, -1)
		for _, loc := range matches {
			word := line[loc[0]:loc[1]]
			sev := SeverityError
			source := sourceName
			diags = append(diags, Diagnostic{
				Range:    makeRange(lineIdx, loc[0], lineIdx, loc[1]),
				Severity: &sev,
				Source:   &source,
				Message: fmt.Sprintf(
					"forbidden keyword %q — this DSL is Turing-incomplete and does not support looping or jump constructs (REQ-1.3)",
					word,
				),
			})
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// Parse-error diagnostic
// ---------------------------------------------------------------------------

func parseDiagnostic(err error) Diagnostic {
	sev := SeverityError
	source := sourceName
	msg := err.Error()
	line, col := extractPosition(msg)
	return Diagnostic{
		Range:    makeRange(line, col, line, col),
		Severity: &sev,
		Source:   &source,
		Message:  msg,
	}
}

func extractPosition(msg string) (int, int) {
	parts := strings.SplitN(msg, ":", 4)
	if len(parts) >= 3 {
		if l, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			if c, err2 := strconv.Atoi(strings.TrimSpace(parts[2])); err2 == nil {
				return max(l-1, 0), max(c-1, 0)
			}
		}
	}
	return 0, 0
}

// ---------------------------------------------------------------------------
// Semantic-error diagnostics
// ---------------------------------------------------------------------------

func semanticDiagnostics(errs []semantic.Error) []Diagnostic {
	diags := make([]Diagnostic, 0, len(errs))
	for _, e := range errs {
		sev := SeverityError
		source := sourceName
		line := max(e.Pos.Line-1, 0)
		col := max(e.Pos.Column-1, 0)
		diags = append(diags, Diagnostic{
			Range:    makeRange(line, col, line, col),
			Severity: &sev,
			Source:   &source,
			Message:  e.Message,
		})
	}
	return diags
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeRange(startLine, startChar, endLine, endChar int) Range {
	return Range{
		Start: Position{Line: uint32(startLine), Character: uint32(startChar)},
		End:   Position{Line: uint32(endLine), Character: uint32(endChar)},
	}
}
