package analyzer

import (
	"strings"
	"testing"

	"github.com/zijing/kiro-notifications/pkg/jsonl"
)

func TestAnalyzeSession_TaskComplete(t *testing.T) {
	msgs := []jsonl.Message{
		{Kind: "AssistantMessage", AssistantText: "Done", Tools: []jsonl.ToolUse{{Name: "write"}}},
	}
	if s := AnalyzeSession(msgs); s != StatusTaskComplete {
		t.Errorf("expected task_complete, got %s", s)
	}
}

func TestAnalyzeSession_Question(t *testing.T) {
	msgs := []jsonl.Message{
		{Kind: "AssistantMessage", AssistantText: "Which option do you prefer?"},
	}
	if s := AnalyzeSession(msgs); s != StatusQuestion {
		t.Errorf("expected question, got %s", s)
	}
}

func TestAnalyzeSession_ReviewComplete(t *testing.T) {
	longText := strings.Repeat("x", 201)
	msgs := []jsonl.Message{
		{Kind: "AssistantMessage", AssistantText: longText, Tools: []jsonl.ToolUse{{Name: "read"}}},
	}
	if s := AnalyzeSession(msgs); s != StatusReviewComplete {
		t.Errorf("expected review_complete, got %s", s)
	}
}

func TestAnalyzeSession_Empty(t *testing.T) {
	if s := AnalyzeSession(nil); s != StatusUnknown {
		t.Errorf("expected unknown, got %s", s)
	}
}

func TestAnalyzeSession_ActiveOverridesPassive(t *testing.T) {
	msgs := []jsonl.Message{
		{Kind: "AssistantMessage", AssistantText: "reading", Tools: []jsonl.ToolUse{{Name: "read"}}},
		{Kind: "AssistantMessage", AssistantText: "done", Tools: []jsonl.ToolUse{{Name: "shell"}}},
	}
	if s := AnalyzeSession(msgs); s != StatusTaskComplete {
		t.Errorf("expected task_complete, got %s", s)
	}
}
