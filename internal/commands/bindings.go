package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xel7dev/wails/internal/flags"
	"github.com/xel7dev/wails/internal/generator"
	"github.com/xel7dev/wails/internal/generator/config"
	"github.com/xel7dev/wails/internal/term"
)

func GenerateBindings(options *flags.GenerateBindingsOptions, patterns []string) error {
	return generateBindingsWithEnvironment(options, patterns, nil)
}

func generateBindingsWithEnvironment(options *flags.GenerateBindingsOptions, patterns, environment []string) error {
	DisableFooter = true

	if options.Silent {
		term.DisableOutput()
		defer term.EnableOutput()
	} else if options.Verbose {
		term.EnableDebug()
		defer term.DisableDebug()
	}

	if !underWake() {
		term.Header("Generate Bindings")
	}

	if len(patterns) == 0 {
		// No input pattern, load package from current directory.
		patterns = []string{"."}
	}

	// Compute absolute path of output directory.
	absPath, err := filepath.Abs(options.OutputDirectory)
	if err != nil {
		return err
	}

	// When clean mode is active and we're writing real files, generate bindings
	// into a dot-prefixed sibling temp directory first, then sync the result
	// into the output directory file by file. This prevents chokidar (used by
	// Vite) from entering a rename-event loop caused by rapid directory
	// delete+recreate, which would otherwise cause the node process to leak
	// memory at ~2-6 MB/s.
	// Dot-prefixed directories are ignored by chokidar's default glob pattern,
	// so no spurious HMR events fire during file generation.
	generationDir := absPath
	swapped := false
	if options.Clean && !options.DryRun {
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			return fmt.Errorf("failed to create bindings parent directory: %w", err)
		}
		tmpDir, err := os.MkdirTemp(filepath.Dir(absPath), ".bindings-tmp-")
		if err != nil {
			return fmt.Errorf("failed to create temp directory for bindings: %w", err)
		}
		generationDir = tmpDir
		defer func() {
			if !swapped {
				_ = os.RemoveAll(tmpDir)
			}
		}()
	} else if options.Clean {
		if err := os.RemoveAll(absPath); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// Initialise file creator.
	var creator config.FileCreator
	if !options.DryRun {
		creator = config.DirCreator(generationDir)
	}

	// Under a wake build, forward progress to the build UI as wire events rather
	// than drawing a competing spinner; otherwise use the interactive spinner.
	var logger config.Logger
	var spinner term.Spinner
	if underWake() {
		logger = newWakeLogger()
	} else {
		spinner = term.StartSpinner("Initialising...")
		logger = spinner.Logger()
	}

	// Initialise and run generator.
	stats, err := generator.NewGenerator(
		options,
		creator,
		logger,
	).WithEnvironment(environment).Generate(patterns...)

	summary := fmt.Sprintf(
		"Processed: %s, %s, %s, %s, %s, %s in %s.",
		pluralise(stats.NumPackages, "Package"),
		pluralise(stats.NumServices, "Service"),
		pluralise(stats.NumMethods, "Method"),
		pluralise(stats.NumEnums, "Enum"),
		pluralise(stats.NumModels, "Model"),
		pluralise(stats.NumEvents, "Event"),
		stats.Elapsed().String(),
	)

	if underWake() {
		logger.Infof("%s", summary)
		logger.Infof("Output directory: %s", absPath)
	} else {
		term.StopSpinner(spinner)
		term.Infof("%s", summary)
		term.Infof("Output directory: %s", absPath)
	}

	// Process generator error.
	if err != nil {
		var report *generator.ErrorReport
		switch {
		case errors.Is(err, generator.ErrNoPackages):
			// Convert to warning message.
			term.Warning(err)
		case errors.As(err, &report):
			if report.HasErrors() {
				// Report error count.
				return err
			} else if report.HasWarnings() {
				// Report warning count.
				term.Warning(report)
			}
		default:
			// Report error.
			return err
		}
	}

	// Sync the generated output into place. Files are updated individually
	// rather than swapping the whole directory: on Windows, deleting or
	// renaming over the output directory fails with "Access is denied" while
	// a file watcher (e.g. Vite's dev server) holds it open (#5515).
	if !swapped && generationDir != absPath {
		if err := syncDirs(generationDir, absPath); err != nil {
			return fmt.Errorf("failed to install new bindings at %q: %w; the bindings directory may be partially updated; rerun the command to retry", absPath, err)
		}
		swapped = true
	}

	if !options.DryRun {
		if err := installTypedEventsCompat(absPath); err != nil {
			return fmt.Errorf("failed to install typed events compatibility bindings: %w", err)
		}
	}

	return nil
}

func pluralise(number int, word string) string {
	if number == 1 {
		return fmt.Sprintf("%d %s", number, word)
	}
	return fmt.Sprintf("%d %ss", number, word)
}

const (
	// wailsModulePath is the import path of this fork's Wails module.
	wailsModulePath = "github.com/xel7dev/wails"
	// typedEventsModulePath is the Wails module path hard-coded by the
	// @wailsio/runtime typed-events Vite plugin when it looks for the
	// generated event bindings.
	typedEventsModulePath = "github.com/wailsapp/wails/v3"
)

// installTypedEventsCompat writes a forwarding bindings module at the module
// path expected by the @wailsio/runtime Vite plugin. The plugin resolves
// <bindings>/github.com/wailsapp/wails/v3/internal/eventcreate at build time,
// which only exists when the project imports upstream Wails; this fork
// generates its bindings under its own module path instead.
func installTypedEventsCompat(bindingsDir string) error {
	if wailsModulePath == typedEventsModulePath {
		return nil
	}
	srcDir := filepath.Join(bindingsDir, filepath.FromSlash(wailsModulePath), "internal")
	var src string
	for _, name := range []string{"eventcreate.ts", "eventcreate.js"} {
		candidate := filepath.Join(srcDir, name)
		if _, err := os.Stat(candidate); err == nil {
			src = candidate
			break
		}
	}
	if src == "" {
		return nil
	}
	dst := filepath.Join(bindingsDir, filepath.FromSlash(typedEventsModulePath), "internal", filepath.Base(src))
	rel, err := filepath.Rel(filepath.Dir(dst), src)
	if err != nil {
		return err
	}
	specifier := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel)) + ".js"
	content := fmt.Sprintf("// This file is automatically generated. DO NOT EDIT\nexport * from %q;\n", specifier)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(content), 0o644)
}
