package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alkemist-17/vida"
)

func bundle(args []string) {
	clear()
	printVersion()

	var scriptArg, outPath string
	var doCompile bool

	if len(args) == 2 {
		entry, err := resolveEntryScript()
		handleError(err)
		scriptArg = entry
		goto jumpHere
	}

	if len(args) == 3 && args[2] == "--compile" {
		entry, err := resolveEntryScript()
		handleError(err)
		scriptArg = entry
		doCompile = true
		goto jumpHere
	}

	if len(args) < 3 {
		handleError(fmt.Errorf("no script given.\n\tUsage: vida bundle <script.vida> [-o out.vida] [--compile]"))
	}

	scriptArg, outPath, doCompile = parseBundleArgs(args[2:])

jumpHere:
	p, err := filepath.Abs(scriptArg)
	handleError(err)

	b, err := vida.BundleSource(p)
	handleError(err)

	if doCompile {
		bundleAndCompile(b)
	}

	if outPath == vida.EmptyString {
		outPath = defaultBundleOutput(b.Entry)
	}
	handleError(os.WriteFile(outPath, b.Source, 0o644))

	fmt.Printf("\n\n\n\t✅ bundled %v file(s) into %v\n", len(b.Files), outPath)
	fmt.Printf("\n\n\n\n")
}

func parseBundleArgs(args []string) (script, out string, doCompile bool) {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-o" && i+1 < len(args):
			out = args[i+1]
			i++
		case strings.HasPrefix(args[i], "-o="):
			out = strings.TrimPrefix(args[i], "-o=")
		case args[i] == "--compile":
			doCompile = true
		default:
			if script == vida.EmptyString {
				script = args[i]
			}
		}
	}
	return
}

func defaultBundleOutput(entry string) string {
	base := strings.TrimSuffix(filepath.Base(entry), vida.VidaFileExtension)
	return base + ".bundle.vida"
}

// bundleAndCompile writes the flattened bundle to a temp .vida file named
// after the original entry script (so compileVidaScript derives a sensible
// binary name from it), then reuses the exact same native-compile pipeline
// that 'vida compile' already uses. Because the bundle has no remaining
// local import(...) calls, the resulting binary is fully self-contained --
// unlike compiling the original entry script directly, which still
// resolves its imports from disk relative to wherever that script
// originally lived, so moving the binary elsewhere would break it.
func bundleAndCompile(b *vida.Bundle) {
	tempDir, err := os.MkdirTemp(vida.EmptyString, "vida-bundle-*")
	handleError(err)
	defer os.RemoveAll(tempDir)

	bundleFileName := strings.TrimSuffix(filepath.Base(b.Entry), vida.VidaFileExtension) + vida.VidaFileExtension
	bundlePath := filepath.Join(tempDir, bundleFileName)
	handleError(os.WriteFile(bundlePath, b.Source, 0o644))

	fmt.Printf("\tBundled %v file(s)\n", len(b.Files))
	handleError(compileVidaScript(bundlePath))
}
