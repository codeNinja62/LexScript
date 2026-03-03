package cmd

import (
	"fmt"
	"os"

	"lexscript/pkg/ast"
	"lexscript/pkg/format"

	"github.com/spf13/cobra"
)

var fmtWrite bool

var fmtCmd = &cobra.Command{
	Use:   "fmt <input.lxs>",
	Short: "Format a .lxs contract source file in canonical style",
	Long: `Format parses a .lxs source file and re-emits it in canonical style:

  • 4-space indentation for top-level declarations
  • 8-space indentation for state-body statements
  • Declaration groups ordered: parties → amounts → time_limits → states
  • Single blank line between groups; blank line before each state block

By default output is written to stdout so you can preview unchanged.
Use --write / -w to overwrite the file in-place.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputPath := args[0]

		// --- Read source ---
		src, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", inputPath, err)
		}

		// --- Parse ---
		contract, err := ast.Parser.ParseBytes(inputPath, src)
		if err != nil {
			return fmt.Errorf("parse error (fix syntax errors before formatting):\n  %w", err)
		}

		// --- Format ---
		formatted := format.Format(contract)

		if fmtWrite {
			// Overwrite file in-place.
			if err := os.WriteFile(inputPath, []byte(formatted), 0644); err != nil {
				return fmt.Errorf("writing %s: %w", inputPath, err)
			}
			fmt.Printf("✓  formatted: %s\n", inputPath)
		} else {
			// Print to stdout.
			fmt.Print(formatted)
		}
		return nil
	},
}

func init() {
	fmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false,
		"write result to source file instead of stdout")
	rootCmd.AddCommand(fmtCmd)
}
