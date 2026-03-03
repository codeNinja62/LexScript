// Package lsp — completion.go provides textDocument/completion for LexScript.
package lsp

import (
	"context"
	"encoding/json"

	"lexscript/pkg/ast"

	"github.com/sourcegraph/jsonrpc2"
)

func (h *handler) handleCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	if req.Params == nil {
		_ = conn.Reply(ctx, req.ID, []CompletionItem{})
		return
	}
	var params CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		_ = conn.Reply(ctx, req.ID, []CompletionItem{})
		return
	}

	items := staticCompletions()

	// Add declared symbols from the current document if parseable.
	uri := params.TextDocument.URI
	if src, ok := getDocument(uri); ok {
		contract, err := ast.Parser.ParseString(uri, src)
		if err == nil {
			items = append(items, symbolCompletions(contract)...)
		}
	}

	_ = conn.Reply(ctx, req.ID, items)
}

// ---------------------------------------------------------------------------
// Static completions
// ---------------------------------------------------------------------------

func staticCompletions() []CompletionItem {
	kwKind := CompletionKindKeyword
	enumKind := CompletionKindEnum
	constKind := CompletionKindConstant

	var items []CompletionItem

	keywords := []struct {
		label, detail string
	}{
		{"contract", "Top-level contract block"},
		{"party", "Declare a contract actor"},
		{"amount", "Declare a monetary value"},
		{"time_limit", "Declare a duration"},
		{"date", "Declare a calendar date"},
		{"state", "Declare an FSM state"},
		{"require", "Obligation statement"},
		{"transition", "State transition edge"},
		{"terminate", "Terminal state marker"},
		{"breach", "Breach trigger"},
		{"cpi_adjusted", "CPI adjustment modifier"},
		{"fulfilled", "Terminal: successful completion"},
		{"breached", "Terminal: material breach"},
		{"expired", "Terminal: time expiration"},
	}
	for _, kw := range keywords {
		detail := kw.detail
		items = append(items, CompletionItem{Label: kw.label, Kind: &kwKind, Detail: &detail})
	}

	currencies := []string{"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF"}
	for _, c := range currencies {
		detail := "Currency code"
		items = append(items, CompletionItem{Label: c, Kind: &enumKind, Detail: &detail})
	}

	units := []string{"days", "business_days", "weeks", "months", "years", "hours"}
	for _, u := range units {
		detail := "Time unit"
		items = append(items, CompletionItem{Label: u, Kind: &constKind, Detail: &detail})
	}

	verbs := []string{"pays", "provides", "delivers", "signs", "returns", "transfers", "notifies"}
	for _, v := range verbs {
		detail := "Action verb"
		items = append(items, CompletionItem{Label: v, Kind: &kwKind, Detail: &detail})
	}

	return items
}

// ---------------------------------------------------------------------------
// Symbol completions from AST
// ---------------------------------------------------------------------------

func symbolCompletions(c *ast.Contract) []CompletionItem {
	varKind := CompletionKindVariable
	classKind := CompletionKindClass

	var items []CompletionItem
	for _, decl := range c.Declarations {
		switch {
		case decl.Party != nil:
			detail := "Party"
			items = append(items, CompletionItem{Label: decl.Party.Name, Kind: &varKind, Detail: &detail})
		case decl.Amount != nil:
			detail := "Amount"
			items = append(items, CompletionItem{Label: decl.Amount.Name, Kind: &varKind, Detail: &detail})
		case decl.TimeLimit != nil:
			detail := "Time Limit"
			items = append(items, CompletionItem{Label: decl.TimeLimit.Name, Kind: &varKind, Detail: &detail})
		case decl.Date != nil:
			detail := "Date"
			items = append(items, CompletionItem{Label: decl.Date.Name, Kind: &varKind, Detail: &detail})
		case decl.State != nil:
			detail := "State"
			items = append(items, CompletionItem{Label: decl.State.Name, Kind: &classKind, Detail: &detail})
		}
	}
	return items
}
