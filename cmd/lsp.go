package cmd

import (
	"lexscript/pkg/lsp"

	"github.com/spf13/cobra"
)

// lspStdio is accepted but ignored — the server always communicates over
// stdin/stdout. The flag exists so vscode-languageclient (TransportKind.stdio)
// can pass --stdio without causing an "unknown flag" error.
var lspStdio bool

var lspCmd = &cobra.Command{
	Use:   "lsp",
	Short: "Start the LexScript Language Server Protocol server (stdin/stdout)",
	Long: `Start a JSON-RPC 2.0 Language Server Protocol (LSP) server over
stdin/stdout. This is intended to be spawned by an editor extension
(e.g. the VS Code LexScript extension).

Capabilities:
  - textDocument/publishDiagnostics  (errors + warnings on every edit)
  - textDocument/hover               (keyword and symbol descriptions)
  - textDocument/completion          (keywords, declared names, currencies)
  - textDocument/definition          (go-to-declaration for state names)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return lsp.RunServer()
	},
}

func init() {
	rootCmd.AddCommand(lspCmd)
	lspCmd.Flags().BoolVar(&lspStdio, "stdio", false, "use stdin/stdout transport (default and only mode; flag accepted for editor compatibility)")
}
