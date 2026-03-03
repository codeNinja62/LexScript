package cmd

import (
	"fmt"
	"os"
	"strings"

	"lexscript/pkg/ast"
	"lexscript/pkg/codegen"
	"lexscript/pkg/semantic"

	"github.com/spf13/cobra"
)

var outputPath string
var outputFormat string
var jurisdiction string

var compileCmd = &cobra.Command{
	Use:   "compile <input.lxs>",
	Short: "Compile a .lxs contract into a Markdown or PDF document",
	Long: `Compile a .lxs contract source file through the full pipeline:
  pre-scan → parse → semantic validate → code generation

Output formats (--format / -f):
  md   Markdown document (default)
  pdf  PDF document via go-pdf/fpdf backend

Jurisdiction (--jurisdiction / -j):
  common      Generic common law boilerplate (default)
  delaware    State of Delaware clause library
  california  State of California clause library (JAMS arbitration)
  uk          England and Wales clause library (LCIA arbitration)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]

		// --- Validate --format flag ---
		switch outputFormat {
		case "md", "pdf":
			// valid
		default:
			return fmt.Errorf("unsupported output format %q; valid options: md, pdf", outputFormat)
		}

		// --- Validate --jurisdiction flag (Phase 3) ---
		if !codegen.IsValidJurisdiction(jurisdiction) {
			return fmt.Errorf("unsupported jurisdiction %q; valid options: common, delaware, california, uk", jurisdiction)
		}

		// --- Read source ---
		src, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", inputPath, err)
		}

		// --- REQ-1.3: Pre-scan for forbidden loop/jump keywords ---
		if errs := forbiddenKeywordErrors(string(src)); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "error:", e)
			}
			return fmt.Errorf("compilation failed: %d forbidden-keyword error(s)", len(errs))
		}

		// --- Frontend: Parse DSL → AST ---
		contract, err := ast.Parser.ParseBytes(inputPath, src)
		if err != nil {
			return fmt.Errorf("parse error:\n  %w", err)
		}

		// --- Middle-end: Semantic validation ---
		if errs := semantic.Validate(contract); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "error:", e)
			}
			return fmt.Errorf("compilation failed: %d semantic error(s)", len(errs))
		}

		// --- Determine output path ---
		out := outputPath
		if out == "" {
			base := strings.TrimSuffix(inputPath, ".lxs")
			out = base + "." + outputFormat
		}

		// --- Backend: format-specific code generation ---
		switch outputFormat {
		case "pdf":
			pdfEmitter := codegen.NewPDFEmitter()
			if err := pdfEmitter.EmitPDF(contract, out, jurisdiction); err != nil {
				return fmt.Errorf("PDF generation failed: %w", err)
			}
		default: // "md"
			emitter := codegen.NewEmitter()
			if err := emitter.Emit(contract, out, jurisdiction); err != nil {
				return fmt.Errorf("code generation failed: %w", err)
			}
		}

		fmt.Printf("✓  compiled: %s → %s\n", inputPath, out)
		return nil
	},
}

// forbiddenKeywords lists constructs that would make the DSL Turing-complete (REQ-1.3).
var forbiddenKeywords = []string{
	"while", "for", "loop", "goto", "repeat", "until",
	"do", "foreach", "recurse",
}

// forbiddenKeywordErrors scans raw source for whole-word occurrences of forbidden keywords.
// Word boundaries are checked manually because Go's RE2 does not support \b.
func forbiddenKeywordErrors(src string) []string {
	var errs []string
	for _, kw := range forbiddenKeywords {
		idx := 0
		for {
			pos := strings.Index(src[idx:], kw)
			if pos == -1 {
				break
			}
			abs := idx + pos
			before := abs == 0 || !isIdentRune(rune(src[abs-1]))
			after := abs+len(kw) >= len(src) || !isIdentRune(rune(src[abs+len(kw)]))
			if before && after {
				// Count line number for helpful error message
				line := 1 + strings.Count(src[:abs], "\n")
				errs = append(errs, fmt.Sprintf(
					"line %d: forbidden keyword %q — this DSL is Turing-incomplete and does not support looping or jump constructs (REQ-1.3)",
					line, kw,
				))
			}
			idx = abs + 1
		}
	}
	return errs
}

func isIdentRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}

func init() {
	compileCmd.Flags().StringVarP(&outputPath, "output", "o", "",
		"output path (default: replaces .lxs with .md or .pdf based on --format)")
	compileCmd.Flags().StringVarP(&outputFormat, "format", "f", "md",
		"output format: md (Markdown, default) or pdf")
	compileCmd.Flags().StringVarP(&jurisdiction, "jurisdiction", "j", "common",
		"jurisdiction-specific clause library: common (default), delaware, california, uk")
	rootCmd.AddCommand(compileCmd)
}
