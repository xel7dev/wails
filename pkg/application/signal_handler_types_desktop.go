//go:build !ios

package application

import "github.com/xel7dev/wails/internal/signal"

// platformSignalHandler holds the signal handler for desktop platforms
type platformSignalHandler struct {
	signalHandler *signal.SignalHandler
}