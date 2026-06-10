// Package webhook sends notifications to external services.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zijing/kiro-notifications/internal/config"
	"github.com/zijing/kiro-notifications/internal/logging"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Send sends a webhook notification asynchronously.
func Send(cfg *config.Config, status, body string) {
	if !cfg.Webhook.Enabled || cfg.Webhook.URL == "" {
		return
	}
	go func() {
		if err := send(cfg, status, body); err != nil {
			logging.Error("webhook: %v", err)
		}
	}()
}

func send(cfg *config.Config, status, body string) error {
	payload, err := formatPayload(cfg.Webhook.Preset, status, body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", cfg.Webhook.URL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}
	logging.Debug("webhook sent: %d", resp.StatusCode)
	return nil
}

func formatPayload(preset, status, body string) ([]byte, error) {
	switch preset {
	case "slack":
		return json.Marshal(map[string]string{"text": fmt.Sprintf("%s\n%s", status, body)})
	case "discord":
		return json.Marshal(map[string]interface{}{
			"content":  fmt.Sprintf("%s\n%s", status, body),
			"username": "Kiro Notifications",
		})
	case "lark":
		return json.Marshal(map[string]interface{}{
			"msg_type": "text",
			"content":  map[string]string{"text": fmt.Sprintf("%s\n%s", status, body)},
		})
	default: // custom JSON
		return json.Marshal(map[string]interface{}{
			"status":    status,
			"message":   body,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"source":    "kiro-notifications",
		})
	}
}
