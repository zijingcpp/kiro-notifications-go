// Package notifier sends cross-platform desktop notifications and sounds.
package notifier

import (
	"os/exec"
	"runtime"

	"github.com/gen2brain/beeep"
	"github.com/zijing/kiro-notifications/internal/config"
	"github.com/zijing/kiro-notifications/internal/logging"
)

// Send sends a desktop notification.
func Send(title, body string, cfg *config.Config) error {
	logging.Debug("notify: %s / %s", title, body)
	// Use a stock icon name; empty string causes notify-send to crash
	// with SIGTRAP in g_variant_new_string() on Ubuntu 24.04.
	return beeep.Notify(title, body, "dialog-information")
}

// PlaySound plays an audio file asynchronously.
func PlaySound(path string) {
	go func() {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("afplay", path)
		case "linux":
			// Try players in order of preference for mp3 support
			players := [][]string{
				{"mpv", "--no-video", "--really-quiet", path},
				{"ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", path},
				{"gst-play-1.0", path},
				{"paplay", path},
			}
			for _, p := range players {
				if _, err := exec.LookPath(p[0]); err == nil {
					cmd = exec.Command(p[0], p[1:]...)
					break
				}
			}
			if cmd == nil {
				logging.Debug("no audio player found")
				return
			}
		default:
			return
		}
		if err := cmd.Run(); err != nil {
			logging.Debug("play sound: %v", err)
		}
	}()
}
