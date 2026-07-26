// Package mcp is a minimal Model Context Protocol server over stdio. It speaks
// JSON-RPC 2.0, one message per line, with just the methods an editor needs to
// discover and run tools: initialize, tools/list, tools/call, and ping. It
// knows nothing about verifi. Callers register tools whose handlers do the
// work, so this stays a thin, reusable transport and the real logic lives in
// the command it wraps. Stdlib only (see docs/adr/0001).
package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

// protocolVersion is the MCP revision we implement. We echo the client's
// requested version when it sends one, so a newer client still connects.
const protocolVersion = "2025-06-18"

// Tool is one callable task. The handler receives the raw JSON arguments and
// returns the text result, or an error that becomes an isError tool result.
type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(args json.RawMessage) (string, error)
}

// Server serves a fixed set of tools over stdio.
type Server struct {
	name, version string
	tools         []Tool
	byName        map[string]Tool
}

// NewServer returns a server that identifies itself with name and version.
func NewServer(name, version string) *Server {
	return &Server{name: name, version: version, byName: map[string]Tool{}}
}

// AddTool registers a tool. It panics on a duplicate name, which is a wiring
// bug, not a runtime condition.
func (s *Server) AddTool(t Tool) {
	if _, dup := s.byName[t.Name]; dup {
		panic("mcp: duplicate tool " + t.Name)
	}
	s.tools = append(s.tools, t)
	s.byName[t.Name] = t
}

// ToolNames returns the registered tool names, in registration order.
func (s *Server) ToolNames() []string {
	names := make([]string, len(s.tools))
	for i, t := range s.tools {
		names[i] = t.Name
	}
	return names
}

// Serve reads newline-delimited JSON-RPC from in and writes responses to out,
// one per line, until in reaches EOF. The editor owns the process lifetime, so
// EOF is a clean shutdown, not an error.
func (s *Server) Serve(in io.Reader, out io.Writer) error {
	r := bufio.NewReader(in)
	w := bufio.NewWriter(out)
	for {
		line, err := r.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			if resp, ok := s.handle(line); ok {
				if werr := writeMessage(w, resp); werr != nil {
					return werr
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// handle dispatches one message. The second return is false for notifications,
// which get no response.
func (s *Server) handle(line []byte) (response, bool) {
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		return errorResponse(nil, -32700, "parse error"), true
	}
	isNotification := len(req.ID) == 0

	switch req.Method {
	case "initialize":
		return s.ok(req.ID, s.initializeResult(req.Params)), true
	case "notifications/initialized":
		return response{}, false
	case "ping":
		return s.ok(req.ID, map[string]any{}), true
	case "tools/list":
		return s.ok(req.ID, map[string]any{"tools": s.toolList()}), true
	case "tools/call":
		return s.ok(req.ID, s.callTool(req.Params)), true
	default:
		if isNotification {
			return response{}, false
		}
		return errorResponse(req.ID, -32601, "method not found: "+req.Method), true
	}
}

func (s *Server) initializeResult(params json.RawMessage) map[string]any {
	pv := protocolVersion
	if len(params) > 0 {
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(params, &p) == nil && p.ProtocolVersion != "" {
			pv = p.ProtocolVersion
		}
	}
	return map[string]any{
		"protocolVersion": pv,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": s.name, "version": s.version},
	}
}

func (s *Server) toolList() []map[string]any {
	out := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		schema := t.InputSchema
		if schema == nil {
			schema = map[string]any{"type": "object"}
		}
		out = append(out, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": schema,
		})
	}
	return out
}

func (s *Server) callTool(params json.RawMessage) map[string]any {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		return toolError("invalid tools/call params: " + err.Error())
	}
	t, ok := s.byName[call.Name]
	if !ok {
		return toolError("unknown tool: " + call.Name)
	}
	text, err := t.Handler(call.Arguments)
	if err != nil {
		return toolError(err.Error())
	}
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": text}},
	}
}

// toolError is a tool-execution failure, returned as a result the model sees,
// not a JSON-RPC protocol error.
func toolError(msg string) map[string]any {
	return map[string]any{
		"content": []map[string]any{{"type": "text", "text": msg}},
		"isError": true,
	}
}

func (s *Server) ok(id json.RawMessage, result any) response {
	raw, err := json.Marshal(result)
	if err != nil {
		return errorResponse(id, -32603, "internal error marshalling result")
	}
	return response{JSONRPC: "2.0", ID: id, Result: raw}
}

func errorResponse(id json.RawMessage, code int, msg string) response {
	return response{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

func writeMessage(w *bufio.Writer, resp response) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.WriteByte('\n'); err != nil {
		return err
	}
	return w.Flush()
}
