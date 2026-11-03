// Module marker for the documentation site.
//
// Keeping a go.mod here makes docs/ a nested module, so its large media
// assets are not packed into the Wails module archive downloaded by
// `go install github.com/xel7dev/wails/v3/cmd/wails@...`.
module github.com/xel7dev/wails/docs

go 1.26.0
