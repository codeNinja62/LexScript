package cmd

import (
	"fmt"

	"lexscript/pkg/playground"

	"github.com/spf13/cobra"
)

var playgroundAddr string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the LexScript web playground",
	Long: `Start an HTTP server hosting the LexScript web playground.

The playground provides a browser-based editor with:
  • Live compilation    — edit .lxs source and see Markdown/PDF output
  • FSM visualization   — interactive state-machine graph (via viz.js)
  • Diagnostics panel   — real-time errors and warnings
  • Jurisdiction picker  — common, delaware, california, uk

Default address: http://localhost:8080`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Starting LexScript playground at http://localhost%s\n", playgroundAddr)
		return playground.Serve(playgroundAddr)
	},
}

func init() {
	serveCmd.Flags().StringVarP(&playgroundAddr, "addr", "a", ":8080",
		"address to listen on (e.g. :8080, :3000)")
	rootCmd.AddCommand(serveCmd)
}
