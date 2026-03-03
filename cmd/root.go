package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "lexs",
	Short: "LexScript: DSL compiler for contract state machines (binary: lexs)",
	Long: `LexScript is a Turing-incomplete DSL compiler (binary: lexs).

	It ingests finite state machine contracts written in the .lxs DSL and
emits legally formatted Markdown documents with standard common-law clauses.

Architecture:
  Frontend  — lexer + parser (participle/v2) → AST
  Middle    — semantic validation (graph reachability, type checks)
  Backend   — deterministic template mapping → Markdown / PDF`,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
