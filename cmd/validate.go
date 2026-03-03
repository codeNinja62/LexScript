package cmd

import (
	"fmt"
	"os"

	"lexscript/pkg/ast"
	"lexscript/pkg/semantic"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <input.lxs>",
	Short: "Validate a .lxs file without generating output (debug)",
	Long: `Runs the Frontend + Middle-end phases:
  1. Checks for forbidden keywords (REQ-1.3)
  2. Lexes and parses the source into an AST
  3. Validates:
       - Duplicate symbol names
       - Unresolved party, state, and time_limit references
       - Dead-end states (no terminate and no transitions)
       - Unreachable states
       - Type constraints (valid currency codes, positive durations)

No output document is produced.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]

		src, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", inputPath, err)
		}

		if errs := forbiddenKeywordErrors(string(src)); len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "error:", e)
			}
			return fmt.Errorf("validation failed: forbidden keywords detected")
		}

		contract, err := ast.Parser.ParseBytes(inputPath, src)
		if err != nil {
			return fmt.Errorf("parse error:\n  %w", err)
		}

		errs := semantic.Validate(contract)
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "error:", e)
			}
			return fmt.Errorf("validation failed: %d error(s)", len(errs))
		}

		fmt.Printf("✓  %s: no errors\n", inputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
