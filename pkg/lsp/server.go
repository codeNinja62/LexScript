// Package lsp — server.go wires the JSON-RPC 2.0 LSP server over stdin/stdout.
//
// Uses sourcegraph/jsonrpc2 with Content-Length framing (VSCodeObjectCodec).
package lsp

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/sourcegraph/jsonrpc2"
)

// documentStore keeps the latest content for each open .lxs file.
var (
	docMu    sync.RWMutex
	docStore = make(map[string]string)
)

func getDocument(uri string) (string, bool) {
	docMu.RLock()
	defer docMu.RUnlock()
	s, ok := docStore[uri]
	return s, ok
}

func setDocument(uri string, content string) {
	docMu.Lock()
	defer docMu.Unlock()
	docStore[uri] = content
}

func deleteDocument(uri string) {
	docMu.Lock()
	defer docMu.Unlock()
	delete(docStore, uri)
}

// ---------------------------------------------------------------------------
// RunServer starts the LSP server on stdin/stdout (blocking).
// ---------------------------------------------------------------------------

func RunServer() error {
	ctx := context.Background()
	stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
	conn := jsonrpc2.NewConn(ctx, stream, &handler{})
	<-conn.DisconnectNotify()
	return nil
}

// stdrwc wraps stdin/stdout as an io.ReadWriteCloser.
type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error)  { return os.Stdin.Read(p) }
func (stdrwc) Write(p []byte) (int, error) { return os.Stdout.Write(p) }
func (stdrwc) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	return os.Stdout.Close()
}

// ---------------------------------------------------------------------------
// JSON-RPC Handler
// ---------------------------------------------------------------------------

type handler struct{}

func (h *handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {

	// ---- General Messages ----
	case "initialize":
		h.handleInitialize(ctx, conn, req)
	case "initialized":
		// no-op
	case "shutdown":
		if !req.Notif {
			_ = conn.Reply(ctx, req.ID, nil)
		}
	case "exit":
		os.Exit(0)

	// ---- Text Document Sync ----
	case "textDocument/didOpen":
		h.handleDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		h.handleDidChange(ctx, conn, req)
	case "textDocument/didSave":
		h.handleDidSave(ctx, conn, req)
	case "textDocument/didClose":
		h.handleDidClose(ctx, conn, req)

	// ---- Language Features ----
	case "textDocument/hover":
		h.handleHover(ctx, conn, req)
	case "textDocument/completion":
		h.handleCompletion(ctx, conn, req)
	case "textDocument/definition":
		h.handleDefinition(ctx, conn, req)

	default:
		if !req.Notif {
			_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: "method not supported: " + req.Method,
			})
		}
	}
}

// ---------------------------------------------------------------------------
// initialize
// ---------------------------------------------------------------------------

func (h *handler) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    SyncFull,
				Save:      &SaveOptions{IncludeText: true},
			},
			HoverProvider:      true,
			DefinitionProvider: true,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{" ", "(", ">"},
			},
		},
		ServerInfo: &ServerInfo{
			Name:    "lexscript-lsp",
			Version: "0.4.0",
		},
	}
	_ = conn.Reply(ctx, req.ID, result)
}

// ---------------------------------------------------------------------------
// textDocument/didOpen, didChange, didSave, didClose → publishDiagnostics
// ---------------------------------------------------------------------------

func (h *handler) handleDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	uri := params.TextDocument.URI
	src := params.TextDocument.Text
	setDocument(uri, src)
	publishDiagnostics(ctx, conn, uri, src)
}

func (h *handler) handleDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	uri := params.TextDocument.URI
	if len(params.ContentChanges) > 0 {
		src := params.ContentChanges[0].Text
		setDocument(uri, src)
		publishDiagnostics(ctx, conn, uri, src)
	}
}

func (h *handler) handleDidSave(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	uri := params.TextDocument.URI
	if params.Text != nil {
		setDocument(uri, *params.Text)
		publishDiagnostics(ctx, conn, uri, *params.Text)
	} else if src, ok := getDocument(uri); ok {
		publishDiagnostics(ctx, conn, uri, src)
	}
}

func (h *handler) handleDidClose(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	deleteDocument(params.TextDocument.URI)
}

func publishDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, uri string, src string) {
	diags := Diagnose(uri, src)
	_ = conn.Notify(ctx, "textDocument/publishDiagnostics", &PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diags,
	})
}
