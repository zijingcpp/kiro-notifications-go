// Package hooks orchestrates Kiro notification handling.
// Called by Kiro's stop hook — no stdin data, finds session file itself.
package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/zijing/kiro-notifications/internal/analyzer"
	"github.com/zijing/kiro-notifications/internal/config"
	"github.com/zijing/kiro-notifications/internal/dedup"
	"github.com/zijing/kiro-notifications/internal/logging"
	"github.com/zijing/kiro-notifications/internal/notifier"
	"github.com/zijing/kiro-notifications/internal/summary"
	"github.com/zijing/kiro-notifications/internal/webhook"
	"github.com/zijing/kiro-notifications/pkg/jsonl"
)

// Handler handles Kiro hook events.
type Handler struct {
	cfg        *config.Config
	dedupMgr   *dedup.Manager
	notifierFn func(status analyzer.Status, title, body string) error
	webhookFn  func(status analyzer.Status, body string)
	pluginRoot string
}

// NewHandler creates a new hook handler.
func NewHandler(pluginRoot string) (*Handler, error) {
	cfg, err := config.Load(pluginRoot)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &Handler{
		cfg:        cfg,
		dedupMgr:   dedup.NewManager(),
		pluginRoot: pluginRoot,
		notifierFn: func(status analyzer.Status, title, body string) error {
			return notifier.Send(title, body, cfg)
		},
		webhookFn: func(status analyzer.Status, body string) {
			webhook.Send(cfg, string(status), body)
		},
	}, nil
}

// HandleStop is called by Kiro's stop hook.
func (h *Handler) HandleStop() error {
	// Find latest session file
	sessionPath, err := findLatestSession()
	if err != nil {
		return fmt.Errorf("find session: %w", err)
	}
	logging.Debug("session: %s", sessionPath)

	// Dedup check
	sessionID := filepath.Base(sessionPath)
	if !h.dedupMgr.Acquire(sessionID) {
		logging.Debug("duplicate, skipping")
		return nil
	}

	// Parse session
	msgs, err := jsonl.ParseFile(sessionPath)
	if err != nil {
		return fmt.Errorf("parse session: %w", err)
	}

	// Analyze status
	status := analyzer.AnalyzeSession(msgs)
	if status == analyzer.StatusUnknown {
		logging.Debug("status unknown, skipping notification")
		return nil
	}

	// Generate summary
	body := summary.Generate(msgs)

	// Get title from config
	title := h.cfg.GetTitle(status)

	logging.Info("notify: status=%s title=%s body=%s", status, title, body)

	// Send desktop notification
	if h.cfg.Desktop.Enabled {
		if err := h.notifierFn(status, title, body); err != nil {
			logging.Error("desktop notification: %v", err)
		}
	}

	// Send webhook
	if h.cfg.Webhook.Enabled {
		h.webhookFn(status, body)
	}

	// Play sound
	if h.cfg.Desktop.Sound {
		soundPath := h.cfg.GetSoundPath(status, h.pluginRoot)
		if soundPath != "" {
			notifier.PlaySound(soundPath)
		}
	}

	return nil
}

// findLatestSession finds the most recently modified session JSONL in Kiro's session dir.
func findLatestSession() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sessionDir := filepath.Join(homeDir, ".kiro", "sessions", "cli")

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return "", fmt.Errorf("read session dir: %w", err)
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var jsonlFiles []fileInfo
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".jsonl" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		jsonlFiles = append(jsonlFiles, fileInfo{
			path:    filepath.Join(sessionDir, e.Name()),
			modTime: info.ModTime(),
		})
	}

	if len(jsonlFiles) == 0 {
		return "", fmt.Errorf("no session files found in %s", sessionDir)
	}

	sort.Slice(jsonlFiles, func(i, j int) bool {
		return jsonlFiles[i].modTime.After(jsonlFiles[j].modTime)
	})

	return jsonlFiles[0].path, nil
}
