package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"lexscript/pkg/ast"

	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse <input.lxs>",
	Short: "Parse a .lxs file and dump the AST as JSON (debug)",
	Long: `Runs only the Frontend phase (lexer + parser) and prints the
resulting Abstract Syntax Tree as formatted JSON to stdout.
No semantic validation or code generation is performed.`,
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
			return fmt.Errorf("parse failed: forbidden keywords detected")
		}

		contract, err := ast.Parser.ParseBytes(inputPath, src)
		if err != nil {
			return fmt.Errorf("parse error:\n  %w", err)
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(contract)
	},
}

func init() {
	rootCmd.AddCommand(parseCmd)
}
