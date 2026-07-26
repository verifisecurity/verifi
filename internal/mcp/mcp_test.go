package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// newTestServer wires two tools: echo returns its raw arguments, boom always
// fails. Enough to exercise the whole dispatch surface.
func newTestServer() *Server {
	s := NewServer("verifi-test", "0.0.0")
	s.AddTool(Tool{
		Name:        "echo",
		Description: "echo the arguments back",
		InputSchema: map[string]any{"type": "object"},
		Handler:     func(args json.RawMessage) (string, error) { return string(args), nil },
	})
	s.AddTool(Tool{
		Name:    "boom",
		Handler: func(args json.RawMessage) (string, error) { return "", errors.New("kaboom") },
	})
	return s
}

// serveLines feeds requests through Serve and returns one decoded response per
// output line. Notifications produce no line, so the count can be lower.
func serveLines(t *testing.T, s *Server, reqs ...string) []map[string]any {
	t.Helper()
	var out bytes.Buffer
	if err := s.Serve(strings.NewReader(strings.Join(reqs, "\n")+"\n"), &out); err != nil {
		t.Fatalf("serve: %v", err)
	}
	var got []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("decode response %q: %v", line, err)
		}
		got = append(got, m)
	}
	return got
}

func TestInitialize(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2099-01-01"}}`)
	if len(got) != 1 {
		t.Fatalf("responses = %d, want 1", len(got))
	}
	res := got[0]["result"].(map[string]any)
	if res["protocolVersion"] != "2099-01-01" {
		t.Errorf("protocolVersion = %v, want the client's echoed back", res["protocolVersion"])
	}
	if info := res["serverInfo"].(map[string]any); info["name"] != "verifi-test" {
		t.Errorf("serverInfo.name = %v, want verifi-test", info["name"])
	}
}

func TestNotificationHasNoResponse(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"ping"}`)
	if len(got) != 1 {
		t.Fatalf("responses = %d, want 1 (the notification is silent)", len(got))
	}
	if _, ok := got[0]["result"]; !ok {
		t.Errorf("ping should return a result, got %v", got[0])
	}
}

func TestToolsList(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	tools := got[0]["result"].(map[string]any)["tools"].([]any)
	names := map[string]bool{}
	for _, tl := range tools {
		names[tl.(map[string]any)["name"].(string)] = true
	}
	if !names["echo"] || !names["boom"] {
		t.Errorf("tools/list = %v, want echo and boom", names)
	}
}

func TestToolCall(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"a":1}}}`)
	res := got[0]["result"].(map[string]any)
	if _, isErr := res["isError"]; isErr {
		t.Fatalf("echo should succeed, got %v", res)
	}
	text := res["content"].([]any)[0].(map[string]any)["text"].(string)
	if text != `{"a":1}` {
		t.Errorf("echo text = %q, want the arguments back", text)
	}
}

func TestToolCallError(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"boom","arguments":{}}}`)
	res := got[0]["result"].(map[string]any)
	if res["isError"] != true {
		t.Errorf("boom should be isError, got %v", res)
	}
}

func TestUnknownTool(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nope","arguments":{}}}`)
	if got[0]["result"].(map[string]any)["isError"] != true {
		t.Errorf("unknown tool should be isError, got %v", got[0])
	}
}

func TestUnknownMethod(t *testing.T) {
	got := serveLines(t, newTestServer(),
		`{"jsonrpc":"2.0","id":1,"method":"does/not/exist"}`)
	rpcErr, ok := got[0]["error"].(map[string]any)
	if !ok {
		t.Fatalf("want a JSON-RPC error, got %v", got[0])
	}
	if rpcErr["code"].(float64) != -32601 {
		t.Errorf("code = %v, want -32601", rpcErr["code"])
	}
}

func TestParseError(t *testing.T) {
	got := serveLines(t, newTestServer(), `{not json`)
	if got[0]["error"].(map[string]any)["code"].(float64) != -32700 {
		t.Errorf("want parse error -32700, got %v", got[0])
	}
}
