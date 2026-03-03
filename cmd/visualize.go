package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"lexscript/pkg/ast"
	"lexscript/pkg/visualize"

	"github.com/spf13/cobra"
)

var (
	visualizeOutputPath string
	visualizeStdin      bool
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize <input.lxs>",
	Short: "Export a .lxs contract's state machine as a Graphviz DOT file",
	Long: `Visualize parses a .lxs contract and emits a Graphviz DOT file that
describes the finite state machine defined by the contract.

The .dot file can be rendered with the Graphviz toolchain (https://graphviz.org):

  dot -Tpng contract.dot -o contract.png
  dot -Tsvg contract.dot -o contract.svg

Use --stdin to read the .lxs source from standard input (used by IDE integrations).

Node legend:
  ┌──────────────┐  Non-terminal state (box, rounded)
  │  State Name  │
  └──────────────┘
  ╔══════════════╗  fulfilled terminal (double circle, green)
  ║  Fulfilled   ║  breached  terminal (double circle, red)
  ╚══════════════╝  expired   terminal (double circle, gold)

Edge labels show the transition trigger:
  • Named event    — e.g. "Payment Received"
  • time_limit(x)  — triggers when duration x elapses
  • breach(Party)  — triggers on material breach by Party`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var src []byte
		var inputPath string
		var err error

		if visualizeStdin {
			// Read from stdin (used by the VS Code FSM preview).
			src, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			inputPath = "stdin.lxs"
		} else {
			if len(args) == 0 {
				return fmt.Errorf("provide an input file or use --stdin")
			}
			inputPath = args[0]
			src, err = os.ReadFile(inputPath)
			if err != nil {
				return fmt.Errorf("reading %s: %w", inputPath, err)
			}
		}

		// --- Parse ---
		contract, err := ast.Parser.ParseBytes(inputPath, src)
		if err != nil {
			return fmt.Errorf("parse error:\n  %w", err)
		}

		// --- Determine output path ---
		dot := visualize.DOT(contract)

		// If reading from stdin and no -o flag, write DOT to stdout.
		if visualizeStdin && visualizeOutputPath == "" {
			fmt.Print(dot)
			return nil
		}

		out := visualizeOutputPath
		if out == "" {
			out = strings.TrimSuffix(inputPath, ".lxs") + ".dot"
		}

		// --- Emit DOT ---
		if err := os.WriteFile(out, []byte(dot), 0644); err != nil {
			return fmt.Errorf("writing DOT file %s: %w", out, err)
		}

		fmt.Printf("✓  visualized: %s → %s\n", inputPath, out)
		fmt.Printf("   render with: dot -Tpng %s -o %s\n",
			out, strings.TrimSuffix(out, ".dot")+".png")
		return nil
	},
}

func init() {
	visualizeCmd.Flags().StringVarP(&visualizeOutputPath, "output", "o", "",
		"output path (default: replaces .lxs with .dot)")
	visualizeCmd.Flags().BoolVar(&visualizeStdin, "stdin", false,
		"read .lxs source from stdin instead of a file argument")
	rootCmd.AddCommand(visualizeCmd)
}
