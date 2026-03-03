// Package playground provides an HTTP server that hosts a browser-based
// LexScript editor with live compilation and FSM preview.
//
// Static files (HTML, CSS, JS) are embedded at compile time via //go:embed.
// The server exposes two API endpoints:
//
//	POST /api/compile   — compile .lxs source → Markdown output + diagnostics
//	POST /api/visualize — compile .lxs source → Graphviz DOT string
//
// The default address is :8080.
package playground

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"lexscript/pkg/ast"
	"lexscript/pkg/codegen"
	"lexscript/pkg/semantic"
	"lexscript/pkg/visualize"
)

//go:embed static
var staticFS embed.FS

// Serve starts the playground HTTP server on the given address (e.g. ":8080").
func Serve(addr string) error {
	mux := http.NewServeMux()

	// Strip the "static" prefix so index.html is served at /.
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("creating sub-filesystem: %w", err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// API endpoints.
	mux.HandleFunc("/api/compile", handleCompile)
	mux.HandleFunc("/api/visualize", handleVisualize)

	log.Printf("LexScript playground listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// ---------------------------------------------------------------------------
// /api/compile — returns Markdown output + diagnostics
// ---------------------------------------------------------------------------

type compileRequest struct {
	Source       string `json:"source"`
	Jurisdiction string `json:"jurisdiction,omitempty"`
}

type compileResponse struct {
	Markdown    string       `json:"markdown,omitempty"`
	Diagnostics []diagEntry  `json:"diagnostics"`
	Success     bool         `json:"success"`
}

type diagEntry struct {
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
}

func handleCompile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var req compileRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp := compileResponse{Diagnostics: []diagEntry{}}

	// Check forbidden keywords.
	fkDiags := checkForbiddenKeywords(req.Source)
	resp.Diagnostics = append(resp.Diagnostics, fkDiags...)

	// Parse.
	contract, parseErr := ast.Parser.ParseString("playground.lxs", req.Source)
	if parseErr != nil {
		resp.Diagnostics = append(resp.Diagnostics, diagEntry{
			Line:    1,
			Column:  1,
			Message: parseErr.Error(),
		})
		writeJSON(w, resp)
		return
	}

	// Semantic validation.
	if semErrs := semantic.Validate(contract); len(semErrs) > 0 {
		for _, e := range semErrs {
			resp.Diagnostics = append(resp.Diagnostics, diagEntry{
				Line:    e.Pos.Line,
				Column:  e.Pos.Column,
				Message: e.Message,
			})
		}
		writeJSON(w, resp)
		return
	}

	// Code generation.
	jurisdiction := req.Jurisdiction
	if jurisdiction == "" {
		jurisdiction = "common"
	}
	emitter := codegen.NewEmitter()
	md, emitErr := emitter.EmitString(contract, jurisdiction)
	if emitErr != nil {
		resp.Diagnostics = append(resp.Diagnostics, diagEntry{
			Line:    1,
			Column:  1,
			Message: emitErr.Error(),
		})
		writeJSON(w, resp)
		return
	}

	resp.Markdown = md
	resp.Success = true
	writeJSON(w, resp)
}

// ---------------------------------------------------------------------------
// /api/visualize — returns DOT source
// ---------------------------------------------------------------------------

type visualizeRequest struct {
	Source string `json:"source"`
}

type visualizeResponse struct {
	DOT         string      `json:"dot,omitempty"`
	Diagnostics []diagEntry `json:"diagnostics"`
	Success     bool        `json:"success"`
}

func handleVisualize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var req visualizeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	resp := visualizeResponse{Diagnostics: []diagEntry{}}

	contract, parseErr := ast.Parser.ParseString("playground.lxs", req.Source)
	if parseErr != nil {
		resp.Diagnostics = append(resp.Diagnostics, diagEntry{
			Line:    1,
			Column:  1,
			Message: parseErr.Error(),
		})
		writeJSON(w, resp)
		return
	}

	resp.DOT = visualize.DOT(contract)
	resp.Success = true
	writeJSON(w, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// checkForbiddenKeywords mirrors cmd/compile.go's pre-scan.
func checkForbiddenKeywords(src string) []diagEntry {
	forbidden := []string{
		"while", "for", "loop", "goto", "repeat",
		"until", "do", "foreach", "recurse",
	}
	var diags []diagEntry
	for lineIdx, line := range strings.Split(src, "\n") {
		for _, kw := range forbidden {
			if idx := strings.Index(line, kw); idx >= 0 {
				diags = append(diags, diagEntry{
					Line:    lineIdx + 1,
					Column:  idx + 1,
					Message: fmt.Sprintf("forbidden keyword %q — this DSL is Turing-incomplete (REQ-1.3)", kw),
				})
			}
		}
	}
	return diags
}
