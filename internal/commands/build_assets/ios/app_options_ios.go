//go:build ios

package main

import "github.com/xel7dev/wails/pkg/application"

// modifyOptionsForIOS adjusts the application options for iOS
func modifyOptionsForIOS(opts *application.Options) {
	// Disable signal handlers on iOS to prevent crashes
	opts.DisableDefaultSignalHandler = true
}