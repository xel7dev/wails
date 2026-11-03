package render

import (
	"github.com/xel7dev/wails/internal/flags"
	"github.com/xel7dev/wails/internal/generator/collect"
	"golang.org/x/tools/go/types/typeutil"
)

// module gathers data that is used when rendering a single JS/TS module.
type module struct {
	*Renderer
	*flags.GenerateBindingsOptions

	Imports *collect.ImportMap

	postponedCreates typeutil.Map
}

// Runtime returns the import path for the Wails JS runtime module.
func (m *module) Runtime() string {
	if m.UseBundledRuntime {
		return "/wails/runtime.js"
	} else {
		return "@wailsio/runtime"
	}
}
