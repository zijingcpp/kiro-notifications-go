// Package dedup prevents duplicate notifications.
package dedup

import (
	"os"
	"path/filepath"
	"time"
)

const lockDuration = 3 * time.Second

type Manager struct{}

func NewManager() *Manager { return &Manager{} }

// Acquire tries to acquire a dedup lock for the given key.
// Returns true if acquired (no recent duplicate).
func (m *Manager) Acquire(key string) bool {
	lockPath := filepath.Join(os.TempDir(), "kiro-notify-"+key+".lock")

	info, err := os.Stat(lockPath)
	if err == nil && time.Since(info.ModTime()) < lockDuration {
		return false // recent lock exists
	}

	// Create/update lock file
	f, err := os.Create(lockPath)
	if err != nil {
		return true // proceed on error
	}
	f.Close()
	return true
}
