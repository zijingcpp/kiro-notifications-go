package jsonl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	content := `{"version":"v1","kind":"Prompt","data":{"message_id":"p1","content":[{"kind":"text","data":"hello"}],"meta":{"timestamp":1700000000}}}
{"version":"v1","kind":"AssistantMessage","data":{"message_id":"a1","content":[{"kind":"text","data":"response text"},{"kind":"toolUse","data":{"toolUseId":"t1","name":"shell","input":{}}}]}}
{"version":"v1","kind":"ToolResults","data":{"message_id":"r1","content":[{"kind":"toolResult","data":{"toolUseId":"t1","status":"success","content":[]}}]}}
`
	tmp := filepath.Join(t.TempDir(), "test.jsonl")
	os.WriteFile(tmp, []byte(content), 0644)

	msgs, err := ParseFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	// Prompt
	if msgs[0].Kind != "Prompt" || msgs[0].UserText != "hello" {
		t.Errorf("prompt: got %+v", msgs[0])
	}

	// AssistantMessage
	if msgs[1].Kind != "AssistantMessage" || msgs[1].AssistantText != "response text" {
		t.Errorf("assistant text: got %q", msgs[1].AssistantText)
	}
	if len(msgs[1].Tools) != 1 || msgs[1].Tools[0].Name != "shell" {
		t.Errorf("tools: got %+v", msgs[1].Tools)
	}

	// ToolResults
	if msgs[2].Kind != "ToolResults" || len(msgs[2].ToolStatuses) != 1 || msgs[2].ToolStatuses[0] != "success" {
		t.Errorf("results: got %+v", msgs[2])
	}
}

func TestParseFile_MalformedLines(t *testing.T) {
	content := `not json at all
{"version":"v1","kind":"Prompt","data":{"message_id":"p1","content":[{"kind":"text","data":"ok"}],"meta":{"timestamp":1}}}
{broken
`
	tmp := filepath.Join(t.TempDir(), "test.jsonl")
	os.WriteFile(tmp, []byte(content), 0644)

	msgs, err := ParseFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 valid message, got %d", len(msgs))
	}
}

func TestExtractAllTools(t *testing.T) {
	msgs := []Message{
		{Kind: "AssistantMessage", Tools: []ToolUse{{Name: "read"}, {Name: "write"}}},
		{Kind: "Prompt"},
		{Kind: "AssistantMessage", Tools: []ToolUse{{Name: "shell"}}},
	}
	tools := ExtractAllTools(msgs)
	if len(tools) != 3 || tools[0] != "read" || tools[2] != "shell" {
		t.Errorf("got %v", tools)
	}
}

func TestLastAssistantMessages(t *testing.T) {
	msgs := []Message{
		{Kind: "AssistantMessage", MessageID: "1"},
		{Kind: "Prompt"},
		{Kind: "AssistantMessage", MessageID: "2"},
		{Kind: "AssistantMessage", MessageID: "3"},
	}
	last := LastAssistantMessages(msgs, 2)
	if len(last) != 2 || last[0].MessageID != "2" || last[1].MessageID != "3" {
		t.Errorf("got %+v", last)
	}
}
