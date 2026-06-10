// Package jsonl parses Kiro CLI session JSONL transcript files.
// Kiro session format: {"version":"v1","kind":"...","data":{...}}
// Kinds: Prompt, AssistantMessage, ToolResults
package jsonl

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

// --- Raw wire types ---

type rawLine struct {
	Version string          `json:"version"`
	Kind    string          `json:"kind"`
	Data    json.RawMessage `json:"data"`
}

type rawPromptData struct {
	MessageID string `json:"message_id"`
	Content   []struct {
		Kind string `json:"kind"`
		Data string `json:"data"`
	} `json:"content"`
	Meta struct {
		Timestamp int64 `json:"timestamp"`
	} `json:"meta"`
}

type rawAssistantData struct {
	MessageID string `json:"message_id"`
	Content   []struct {
		Kind string          `json:"kind"`
		Data json.RawMessage `json:"data"`
	} `json:"content"`
}

type rawToolUseData struct {
	ToolUseID string          `json:"toolUseId"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
}

type rawToolResultsData struct {
	MessageID string `json:"message_id"`
	Content   []struct {
		Kind string `json:"kind"`
		Data struct {
			ToolUseID string `json:"toolUseId"`
			Status    string `json:"status"`
			Content   []struct {
				Kind string          `json:"kind"`
				Data json.RawMessage `json:"data"`
			} `json:"content"`
		} `json:"data"`
	} `json:"content"`
}

// --- Public types ---

// ToolUse represents a tool invocation in an assistant message.
type ToolUse struct {
	Name string
}

// Message represents a parsed Kiro session message.
type Message struct {
	Kind      string    // "Prompt", "AssistantMessage", "ToolResults"
	MessageID string
	Timestamp time.Time

	// For Prompt
	UserText string

	// For AssistantMessage
	AssistantText string
	Tools         []ToolUse

	// For ToolResults
	ToolStatuses []string // "success" | "error" per tool result
}

// ParseFile parses a Kiro session JSONL file and returns all messages.
// Malformed lines are silently skipped.
func ParseFile(path string) ([]Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var raw rawLine
		if err := json.Unmarshal(line, &raw); err != nil {
			continue
		}
		msg, ok := parseLine(raw)
		if ok {
			msgs = append(msgs, msg)
		}
	}
	return msgs, scanner.Err()
}

func parseLine(raw rawLine) (Message, bool) {
	switch raw.Kind {
	case "Prompt":
		var d rawPromptData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return Message{}, false
		}
		msg := Message{Kind: "Prompt", MessageID: d.MessageID}
		msg.Timestamp = time.Unix(d.Meta.Timestamp, 0)
		for _, c := range d.Content {
			if c.Kind == "text" {
				msg.UserText += c.Data
			}
		}
		return msg, true

	case "AssistantMessage":
		var d rawAssistantData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return Message{}, false
		}
		msg := Message{Kind: "AssistantMessage", MessageID: d.MessageID}
		for _, c := range d.Content {
			switch c.Kind {
			case "text":
				var s string
				if err := json.Unmarshal(c.Data, &s); err == nil {
					msg.AssistantText += s
				}
			case "toolUse":
				var tu rawToolUseData
				if err := json.Unmarshal(c.Data, &tu); err == nil && tu.Name != "" {
					msg.Tools = append(msg.Tools, ToolUse{Name: tu.Name})
				}
			}
		}
		return msg, true

	case "ToolResults":
		var d rawToolResultsData
		if err := json.Unmarshal(raw.Data, &d); err != nil {
			return Message{}, false
		}
		msg := Message{Kind: "ToolResults", MessageID: d.MessageID}
		for _, c := range d.Content {
			if c.Kind == "toolResult" {
				msg.ToolStatuses = append(msg.ToolStatuses, c.Data.Status)
			}
		}
		return msg, true
	}
	return Message{}, false
}

// ExtractAllTools returns all tool names used across all AssistantMessages.
func ExtractAllTools(msgs []Message) []string {
	var tools []string
	for _, m := range msgs {
		if m.Kind != "AssistantMessage" {
			continue
		}
		for _, t := range m.Tools {
			tools = append(tools, t.Name)
		}
	}
	return tools
}

// LastAssistantMessages returns up to n most recent AssistantMessages.
func LastAssistantMessages(msgs []Message, n int) []Message {
	var assistant []Message
	for _, m := range msgs {
		if m.Kind == "AssistantMessage" {
			assistant = append(assistant, m)
		}
	}
	if len(assistant) <= n {
		return assistant
	}
	return assistant[len(assistant)-n:]
}

// HasToolError returns true if any ToolResults entry has a non-success status.
func HasToolError(msgs []Message) bool {
	for _, m := range msgs {
		if m.Kind != "ToolResults" {
			continue
		}
		for _, s := range m.ToolStatuses {
			if s != "success" {
				return true
			}
		}
	}
	return false
}
