//go:build windows && !server && production && !devtools

package application

import "github.com/xel7dev/wails/internal/webview2/pkg/edge"

func (w *windowsWebviewWindow) openDevTools() {}

func (w *windowsWebviewWindow) enableDevTools(settings *edge.ICoreWebViewSettings) error {
	return settings.PutAreDevToolsEnabled(false)
}
