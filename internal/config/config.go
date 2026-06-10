// Package config handles kiro-notifications configuration.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/zijing/kiro-notifications/internal/analyzer"
)

type Config struct {
	Desktop  DesktopConfig          `json:"desktop"`
	Webhook  WebhookConfig          `json:"webhook"`
	Statuses map[string]StatusConfig `json:"statuses"`
}

type DesktopConfig struct {
	Enabled bool `json:"enabled"`
	Sound   bool `json:"sound"`
}

type WebhookConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Preset  string `json:"preset"` // "slack", "discord", "lark", "custom"
}

type StatusConfig struct {
	Title string `json:"title"`
	Sound string `json:"sound"`
}

func Load(pluginRoot string) (*Config, error) {
	cfg := defaultConfig()

	configPath := filepath.Join(pluginRoot, "config", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) GetTitle(status analyzer.Status) string {
	if s, ok := c.Statuses[string(status)]; ok && s.Title != "" {
		return s.Title
	}
	switch status {
	case analyzer.StatusTaskComplete:
		return "✅ Task Complete"
	case analyzer.StatusReviewComplete:
		return "🔍 Review Complete"
	case analyzer.StatusQuestion:
		return "❓ Question"
	case analyzer.StatusError:
		return "🔴 Error"
	default:
		return "Kiro"
	}
}

func (c *Config) GetSoundPath(status analyzer.Status, pluginRoot string) string {
	if s, ok := c.Statuses[string(status)]; ok && s.Sound != "" {
		return expandPath(s.Sound, pluginRoot)
	}
	// Default sounds
	switch status {
	case analyzer.StatusTaskComplete:
		return filepath.Join(pluginRoot, "sounds", "task-complete.mp3")
	case analyzer.StatusReviewComplete:
		return filepath.Join(pluginRoot, "sounds", "review-complete.mp3")
	case analyzer.StatusQuestion:
		return filepath.Join(pluginRoot, "sounds", "question.mp3")
	case analyzer.StatusError:
		return filepath.Join(pluginRoot, "sounds", "error.mp3")
	}
	return ""
}

func expandPath(path, pluginRoot string) string {
	return strings.ReplaceAll(path, "${PLUGIN_ROOT}", pluginRoot)
}

func defaultConfig() *Config {
	return &Config{
		Desktop: DesktopConfig{Enabled: true, Sound: true},
		Webhook: WebhookConfig{Enabled: false},
		Statuses: map[string]StatusConfig{
			"task_complete":   {Title: "✅ Task Complete", Sound: "${PLUGIN_ROOT}/sounds/task-complete.mp3"},
			"review_complete": {Title: "🔍 Review Complete", Sound: "${PLUGIN_ROOT}/sounds/review-complete.mp3"},
			"question":        {Title: "❓ Question", Sound: "${PLUGIN_ROOT}/sounds/question.mp3"},
			"error":           {Title: "🔴 Error", Sound: "${PLUGIN_ROOT}/sounds/error.mp3"},
		},
	}
}
