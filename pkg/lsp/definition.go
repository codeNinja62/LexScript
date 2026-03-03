// Package lsp — definition.go provides textDocument/definition for LexScript.
package lsp

import (
	"context"
	"encoding/json"

	"lexscript/pkg/ast"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *handler) handleDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params == nil {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}
	var params DefinitionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	uri := params.TextDocument.URI
	src, ok := getDocument(uri)
	if !ok {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	word := wordAtPosition(src, int(params.Position.Line), int(params.Position.Character))
	if word == "" {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	contract, err := ast.Parser.ParseString(uri, src)
	if err != nil {
		_ = conn.Reply(ctx, req.ID, nil)
		return
	}

	for _, decl := range contract.Declarations {
		switch {
		case decl.Party != nil && decl.Party.Name == word:
			_ = conn.Reply(ctx, req.ID, posToLocation(uri, decl.Party.Pos.Line, decl.Party.Pos.Column))
			return
		case decl.Amount != nil && decl.Amount.Name == word:
			_ = conn.Reply(ctx, req.ID, posToLocation(uri, decl.Amount.Pos.Line, decl.Amount.Pos.Column))
			return
		case decl.TimeLimit != nil && decl.TimeLimit.Name == word:
			_ = conn.Reply(ctx, req.ID, posToLocation(uri, decl.TimeLimit.Pos.Line, decl.TimeLimit.Pos.Column))
			return
		case decl.Date != nil && decl.Date.Name == word:
			_ = conn.Reply(ctx, req.ID, posToLocation(uri, decl.Date.Pos.Line, decl.Date.Pos.Column))
			return
		case decl.State != nil && decl.State.Name == word:
			_ = conn.Reply(ctx, req.ID, posToLocation(uri, decl.State.Pos.Line, decl.State.Pos.Column))
			return
		}
	}

	_ = conn.Reply(ctx, req.ID, nil)
}

func posToLocation(uri string, line, col int) Location {
	l := uint32(max(line-1, 0))
	c := uint32(max(col-1, 0))
	return Location{
		URI: uri,
		Range: Range{
			Start: Position{Line: l, Character: c},
			End:   Position{Line: l, Character: c},
		},
	}
}
