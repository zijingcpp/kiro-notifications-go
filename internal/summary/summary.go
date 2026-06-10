// Package summary generates concise notification messages from session data.
package summary

import (
	"strings"

	"github.com/zijing/kiro-notifications/pkg/jsonl"
)

const maxLen = 200

// Generate produces a summary from the last assistant message.
func Generate(msgs []jsonl.Message) string {
	last := jsonl.LastAssistantMessages(msgs, 1)
	if len(last) == 0 {
		return "Task completed"
	}

	text := strings.TrimSpace(last[0].AssistantText)
	if text == "" {
		if len(last[0].Tools) > 0 {
			return "Used: " + last[0].Tools[len(last[0].Tools)-1].Name
		}
		return "Task completed"
	}

	// Clean up markdown
	text = cleanMarkdown(text)

	if len(text) > maxLen {
		text = text[:maxLen] + "..."
	}
	return text
}

func cleanMarkdown(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		// Remove markdown headers
		l = strings.TrimLeft(l, "# ")
		// Remove bullet points
		l = strings.TrimPrefix(l, "- ")
		l = strings.TrimPrefix(l, "* ")
		result = append(result, l)
	}
	return strings.Join(result, " ")
}
