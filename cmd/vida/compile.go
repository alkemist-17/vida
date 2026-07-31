package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

const goRuntimeTemplate = `
package main

import(
	"fmt"
	"os"
	"github.com/alkemist-17/vida"
	"github.com/alkemist-17/vida/extensions"
)

func main() {
	script := {{.Script | quote}}
	ctx := vida.NewContext([]byte(script), {{.Id | quote}}, extensions.GetLoader())
	err := ctx.CompileAndRun()
	handleError(err)
	if err != nil {
		printError(err)
		ctx.PrintCallStack()
	}

}

func handleError(err error) {
	if err != nil {
		printVersion()
		fmt.Printf("\t❌ %v\n\n\n\n", err.Error())
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("\n\n\n   %v\n   %v\n\n\n", vida.Name(), vida.Version())
}
	
func printError(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
`

type CompileScript struct {
	Script string
	Id     string
}

func compileVidaScript(path string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("the Go toolchain is required to compile Vida scripts: %w", err)
	}

	vidaCode, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("could not read %s: %w", path, err)
	}

	tempDir, err := os.MkdirTemp("", "vida-build-*")

	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	funcMap := template.FuncMap{
		"quote": strconv.Quote,
	}

	t := template.Must(template.New("vida_runtime").Funcs(funcMap).Parse(goRuntimeTemplate))

	var buf bytes.Buffer

	data := CompileScript{Script: string(vidaCode), Id: path}

	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	formattedCode, err := format.Source(buf.Bytes())

	if err != nil {
		return fmt.Errorf("failed to format generated go code: %w", err)
	}

	tmpGoFile := filepath.Join(tempDir, "main.go")

	if err := os.WriteFile(tmpGoFile, formattedCode, 0644); err != nil {
		return err
	}

	runCmd := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = tempDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Println("Initializing module ...")

	if err := runCmd("go", "mod", "init", "vida_autogen"); err != nil {
		return fmt.Errorf("failed to init go mod: %w", err)
	}

	fmt.Println("Fetching Vida context ...")

	if err := runCmd("go", "get", "github.com/alkemist-17/vida@latest"); err != nil {
		return fmt.Errorf("failed to get Vida context: %w", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	outputBinaryName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	goos := os.Getenv("GOOS")
	if goos == "windows" || (goos == "" && runtime.GOOS == "windows") {
		outputBinaryName += ".exe"
	}
	outputBinaryPath := filepath.Join(currentDir, outputBinaryName)
	fmt.Printf("Compiling %s -> %s ... \n", path, outputBinaryName)

	if err := runCmd("go", "build", "-o", outputBinaryPath, tmpGoFile); err != nil {
		return fmt.Errorf("failed to compile script: %w", err)
	}

	fmt.Println("Done.")
	return nil
}
