// Package lsp implements a Language Server Protocol server for LexScript.
//
// It provides real-time diagnostics, hover information, autocompletion,
// and go-to-definition for .lxs files. The server communicates over
// stdin/stdout using JSON-RPC 2.0 with Content-Length framing (standard LSP).
//
// This package uses sourcegraph/jsonrpc2 directly (no heavyweight LSP
// framework) so its only external dependency is the same JSON-RPC library
// used throughout the Go LSP ecosystem.
package lsp

import (
	"encoding/json"
)

// ---------------------------------------------------------------------------
// LSP protocol types (subset needed by LexScript)
// These are defined inline to avoid pulling in a framework that brings
// transitive Windows-incompatible dependencies.
// ---------------------------------------------------------------------------

// Position in a text document (0-based line and character).
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

// Range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location represents a location inside a resource.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Diagnostic represents a compiler error or warning.
type Diagnostic struct {
	Range    Range               `json:"range"`
	Severity *DiagnosticSeverity `json:"severity,omitempty"`
	Source   *string             `json:"source,omitempty"`
	Message  string              `json:"message"`
}

type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

// TextDocumentSyncKind defines how the client sends document changes.
type TextDocumentSyncKind int

const (
	SyncNone        TextDocumentSyncKind = 0
	SyncFull        TextDocumentSyncKind = 1
	SyncIncremental TextDocumentSyncKind = 2
)

// ---------------------------------------------------------------------------
// Initialize
// ---------------------------------------------------------------------------

type InitializeParams struct {
	ProcessID    *int            `json:"processId"`
	RootURI      *string         `json:"rootUri"`
	Capabilities json.RawMessage `json:"capabilities"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   *ServerInfo        `json:"serverInfo,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ServerCapabilities struct {
	TextDocumentSync   *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	CompletionProvider *CompletionOptions       `json:"completionProvider,omitempty"`
	HoverProvider      bool                     `json:"hoverProvider,omitempty"`
	DefinitionProvider bool                     `json:"definitionProvider,omitempty"`
}

type TextDocumentSyncOptions struct {
	OpenClose bool                 `json:"openClose"`
	Change    TextDocumentSyncKind `json:"change"`
	Save      *SaveOptions         `json:"save,omitempty"`
}

type SaveOptions struct {
	IncludeText bool `json:"includeText,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// ---------------------------------------------------------------------------
// Text Document Notifications
// ---------------------------------------------------------------------------

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"` // full document text when SyncFull
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         *string                `json:"text,omitempty"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// ---------------------------------------------------------------------------
// Publish Diagnostics
// ---------------------------------------------------------------------------

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// ---------------------------------------------------------------------------
// Hover
// ---------------------------------------------------------------------------

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

// ---------------------------------------------------------------------------
// Completion
// ---------------------------------------------------------------------------

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type CompletionItem struct {
	Label  string              `json:"label"`
	Kind   *CompletionItemKind `json:"kind,omitempty"`
	Detail *string             `json:"detail,omitempty"`
}

type CompletionItemKind int

const (
	CompletionKindKeyword  CompletionItemKind = 14
	CompletionKindVariable CompletionItemKind = 6
	CompletionKindClass    CompletionItemKind = 7
	CompletionKindEnum     CompletionItemKind = 13
	CompletionKindConstant CompletionItemKind = 21
)

// ---------------------------------------------------------------------------
// Definition
// ---------------------------------------------------------------------------

type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}
