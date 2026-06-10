// Package analyzer determines task status from Kiro session transcript.
package analyzer

import (
	"strings"

	"github.com/zijing/kiro-notifications/pkg/jsonl"
)

// Kiro tool categories
var (
	ActiveTools  = []string{"write", "shell", "use_aws", "subagent"}
	PassiveTools = []string{"read", "grep", "glob", "web_search", "web_fetch", "code", "knowledge", "introspect"}
)

// Status represents the task notification status.
type Status string

const (
	StatusTaskComplete   Status = "task_complete"
	StatusReviewComplete Status = "review_complete"
	StatusQuestion       Status = "question"
	StatusError          Status = "error"
	StatusUnknown        Status = "unknown"
)

// AnalyzeSession determines the status from parsed session messages.
// Only considers the last 15 assistant messages for relevance.
func AnalyzeSession(msgs []jsonl.Message) Status {
	recent := jsonl.LastAssistantMessages(msgs, 15)
	if len(recent) == 0 {
		return StatusUnknown
	}

	last := recent[len(recent)-1]

	// Check if last message ends with a question
	text := strings.TrimSpace(last.AssistantText)
	if strings.HasSuffix(text, "?") || strings.HasSuffix(text, "？") {
		return StatusQuestion
	}

	// Check tools used in the last message
	if len(last.Tools) > 0 {
		lastTool := last.Tools[len(last.Tools)-1].Name
		if isActiveTool(lastTool) {
			return StatusTaskComplete
		}
	}

	// Check all tools across recent messages
	allTools := jsonl.ExtractAllTools(recent)
	hasActive := false
	hasPassive := false
	for _, t := range allTools {
		if isActiveTool(t) {
			hasActive = true
		}
		if isPassiveTool(t) {
			hasPassive = true
		}
	}

	if hasActive {
		return StatusTaskComplete
	}

	// Only passive tools used and response >200 chars → review
	if hasPassive && len(text) > 200 {
		return StatusReviewComplete
	}

	// Fallback: any text response
	if len(text) > 0 {
		return StatusTaskComplete
	}

	return StatusUnknown
}

func isActiveTool(name string) bool {
	for _, t := range ActiveTools {
		if t == name {
			return true
		}
	}
	return false
}

func isPassiveTool(name string) bool {
	for _, t := range PassiveTools {
		if t == name {
			return true
		}
	}
	return false
}
