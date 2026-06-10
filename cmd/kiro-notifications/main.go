package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zijing/kiro-notifications/internal/hooks"
	"github.com/zijing/kiro-notifications/internal/logging"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "handle-hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: hook event name required (e.g., stop)")
			os.Exit(1)
		}
		runHook(os.Args[2])
	case "version", "--version", "-v":
		fmt.Printf("kiro-notifications v%s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runHook(event string) {
	if os.Getenv("KIRO_NOTIFY_DEBUG") != "" {
		logging.SetVerbose(true)
	}

	pluginRoot := getPluginRoot()
	handler, err := hooks.NewHandler(pluginRoot)
	if err != nil {
		logging.Error("init: %v", err)
		os.Exit(1)
	}

	switch event {
	case "stop", "Stop":
		if err := handler.HandleStop(); err != nil {
			logging.Error("handle-hook stop: %v", err)
			os.Exit(1)
		}
	default:
		logging.Debug("ignoring hook event: %s", event)
	}
}

func getPluginRoot() string {
	// Check env override
	if root := os.Getenv("KIRO_NOTIFICATIONS_ROOT"); root != "" {
		return root
	}
	// Default: executable's parent directory
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(filepath.Dir(exe))
	}
	// Fallback: cwd
	cwd, _ := os.Getwd()
	return cwd
}

func printUsage() {
	fmt.Print(`kiro-notifications - Desktop notifications for Kiro CLI

Usage:
  kiro-notifications handle-hook <event>   Handle Kiro hook event (stop)
  kiro-notifications version               Show version
  kiro-notifications help                  Show this help

Environment:
  KIRO_NOTIFY_DEBUG=1         Enable debug logging
  KIRO_NOTIFICATIONS_ROOT     Override plugin root directory
`)
}
