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

var compileCmd = &cobra.Command{
	Use:   "compile <input.lxs>",
	Short: "Compile a .lxs contract into a Markdown document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]

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
			out = strings.TrimSuffix(inputPath, ".lxs") + ".md"
		}

		// --- Backend: Template-driven code generation ---
		emitter := codegen.NewEmitter()
		if err := emitter.Emit(contract, out); err != nil {
			return fmt.Errorf("code generation failed: %w", err)
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
		"output path (default: replaces .lxs with .md)")
	rootCmd.AddCommand(compileCmd)
}
