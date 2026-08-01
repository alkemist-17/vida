package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alkemist-17/vida"
	"github.com/alkemist-17/vida/extensions"
)

const (
	RUN          = "run"
	DEGUG        = "debug"
	TIME         = "time"
	TOKENS       = "tokens"
	AST          = "ast"
	SEMANTIC_AST = "astc"
	HELP         = "help"
	VERSION      = "version"
	ABOUT        = "about"
	CODE         = "code"
	TEST         = "test"
	INIT         = "init"
	INSTALL      = "install"
	PATH         = "path"
	COMPILE      = "compile"
	BUNDLE       = "bundle"
	FLAG_I       = "-I"
)

type TestResult struct {
	File     string
	Passed   bool
	Duration time.Duration
	Error    error
}

func main() {
	// f, err := os.Create("vida.prof")
	// handleError(err)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	args := os.Args
	if len(args) > 1 {
		args, includePaths := extractIncludePaths(os.Args)
		applyIncludePaths(includePaths)
		switch parseCMD(args[1]) {
		case RUN:
			run(args)
		case DEGUG:
			runDebug(args)
		case TIME:
			measureRunTime(args)
		case TOKENS:
			printTokens(args)
		case AST:
			printAST(args, false)
		case SEMANTIC_AST:
			printAST(args, true)
		case HELP:
			printHelp()
		case VERSION:
			clear()
			printVersion()
		case ABOUT:
			printAbout()
		case CODE:
			printMachineCode(args)
		case TEST:
			test(args)
		case INIT:
			scaffold(args)
		case INSTALL:
			install(args)
		case PATH:
			path()
		case COMPILE:
			compile(args)
		case BUNDLE:
			bundle(args)
		default:
			handleError(fmt.Errorf("unknown command '%v'.\n\tType 'vida help' for assistance.", parseCMD(args[1])))
		}
	} else {
		printHelp()
	}
}

func runDebug(args []string) {
	clear()
	printVersion()

	var p string
	var err error

	if len(args) > 2 {
		p, err = filepath.Abs(args[2])
		handleError(err)
	} else {
		entry, err := resolveEntryScript()
		handleError(err)
		p = entry
	}

	src, err := vida.LoadScriptFromFile(p)
	handleError(err)
	ctx := vida.NewContext(src, p, extensions.GetLoader())
	err = ctx.RunDebugSession()
	handleError(err)
}

func run(args []string) {
	var p string
	var err error

	if len(args) > 2 {
		p, err = filepath.Abs(args[2])
		handleError(err)
	} else {
		entry, err := resolveEntryScript()
		handleError(err)
		p = entry
	}

	src, err := vida.LoadScriptFromFile(p)
	handleError(err)
	ctx := vida.NewContext(src, p, extensions.GetLoader())
	err = ctx.CompileAndRun()
	if err != nil {
		printError(err)
		ctx.PrintCallStack()
	}
}

func measureRunTime(args []string) {
	clear()
	printVersion()

	var p string
	var err error

	if len(args) > 2 {
		p, err = filepath.Abs(args[2])
		handleError(err)
	} else {
		entry, err := resolveEntryScript()
		handleError(err)
		p = entry
	}

	src, err := vida.LoadScriptFromFile(p)
	handleError(err)
	ctx := vida.NewContext(src, p, extensions.GetLoader())
	ctx.Compile()
	duration, err := ctx.MeasureRunTime()
	if err != nil {
		printError(err)
		ctx.PrintCallStack()
		fmt.Printf("\tFailure ❌\n\n\n\n")
		return
	}
	printDuration(duration)
}

func compile(args []string) {
	var p string
	var err error
	if len(args) > 2 {
		p, err = filepath.Abs(args[2])
		handleError(err)
	} else {
		entry, err := resolveEntryScript()
		handleError(err)
		p = entry
	}
	b, err := vida.BundleSource(p)
	handleError(err)
	bundleAndCompile(b)
}

func printTokens(args []string) {
	clear()
	printVersion()
	largs := len(args)
	if largs > 2 {
		for i := 2; i < largs; i++ {
			p, err := filepath.Abs(args[i])
			handleError(err)
			src, err := vida.LoadScriptFromFile(p)
			handleError(err)
			ctx := vida.NewContext(src, p, extensions.GetLoader())
			handleError(ctx.PrintTokens())
		}
	} else {
		handleError(errorNoArgsGivenTo(TOKENS))
	}
}

func printAST(args []string, withColors bool) {
	clear()
	printVersion()
	largs := len(args)
	if largs > 2 {
		for i := 2; i < largs; i++ {
			p, err := filepath.Abs(args[i])
			handleError(err)
			src, err := vida.LoadScriptFromFile(p)
			handleError(err)
			ctx := vida.NewContext(src, p, extensions.GetLoader())
			handleError(ctx.PrintAST(withColors))
		}
	} else {
		handleError(errorNoArgsGivenTo(AST))
	}
}

func printMachineCode(args []string) {
	clear()
	largs := len(args)
	if largs > 2 {
		for i := 2; i < largs; i++ {
			p, err := filepath.Abs(args[i])
			handleError(err)
			src, err := vida.LoadScriptFromFile(p)
			handleError(err)
			ctx := vida.NewContext(src, p, extensions.GetLoader())
			handleError(ctx.Compile())
			handleError(ctx.PrintMachineCode())
		}
	} else {
		handleError(errorNoArgsGivenTo(CODE))
	}
}

func test(args []string) {
	clear()
	testCount := 0
	if len(args) > 2 {
		dir := args[2]
		info, err := os.Stat(dir)
		handleTestError(err)
		if info.IsDir() {
			scripts := collectScripts(dir)
			if len(scripts) == 0 {
				fmt.Printf("\t❌\tNo vida files were found!\n\t\tTotal files run: %v\n\n\n\n\n\n\n", testCount)
				os.Exit(1)
			}
			printTestResults(runScripts(scripts, &testCount), testCount)
			return
		}
		for _, v := range args[2:] {
			if strings.HasSuffix(v, vida.VidaFileExtension) {
				testCount++
				p, err := filepath.Abs(v)
				handleError(err)
				fmt.Printf("\n\n\n\n\n🧪 Testing '%v'\n\n\n\n\n", p)
				r := executeScript(p)
				if r.Passed {
					fmt.Printf("\tSuccess ✅\n\n\n\n")
				} else {
					fmt.Printf("\tFailure ❌\n\n\n\n")
					fmt.Printf("\t%v\n", r.Error)
				}
				fmt.Printf("\n\n\n\n")
			}
		}
	} else {
		dir, err := os.Getwd()
		handleError(err)
		dir = resolveTestDir(dir)
		scripts := collectScripts(dir)
		if len(scripts) == 0 {
			fmt.Printf("\t❌\tNo vida files were found!\n\t\tTotal files run: %v\n\n\n\n\n\n\n", testCount)
			os.Exit(1)
		}
		printTestResults(runScripts(scripts, &testCount), testCount)
	}
}

func runScripts(scripts []string, textCount *int) []TestResult {
	results := make([]TestResult, len(scripts))
	for _, v := range scripts {
		softclear()
		(*textCount)++
		fmt.Printf("\n\n\n\n\n🧪 Testing '%v'\n\n\n\n\n", v)
		r := executeScript(v)
		results = append(results, r)
		if r.Passed {
			fmt.Printf("\tSuccess ✅\n\n\n\n")
		} else {
			fmt.Printf("\tFailure ❌\n\n\n\n")
			fmt.Printf("\t%v\n", r.Error)
		}
		fmt.Printf("\n\n\n\n")
	}
	return results
}

func executeScript(path string) TestResult {
	src, err := vida.LoadScriptFromFile(path)
	handleError(err)
	ctx := vida.NewContext(src, path, extensions.GetLoader())
	handleError(ctx.Compile())
	dur, err := ctx.MeasureRunTime()
	printDuration(dur)
	return TestResult{
		File:     path,
		Passed:   err == nil,
		Duration: dur,
		Error:    err,
	}
}

func collectScripts(root string) []string {
	var scripts []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(d.Name(), vida.VidaFileExtension) {
			scripts = append(scripts, path)
		}
		return nil
	})
	return scripts
}

func printTestResults(results []TestResult, testCount int) {
	passed, failed := 0, 0
	var totalDuration time.Duration

	softclear()
	fmt.Printf("\n\n\n\n\nSummary\n\n\n\n\n")

	for _, r := range results {
		totalDuration += r.Duration
		if len(r.File) > 0 {
			if r.Passed {
				passed++
				fmt.Printf("PASSED   %23v %11v\n", strings.TrimSuffix(filepath.Base(r.File), vida.VidaFileExtension), r.Duration.Round(time.Millisecond))
			} else {
				failed++
				fmt.Printf("FAILED * %23v %11v\n", strings.TrimSuffix(filepath.Base(r.File), vida.VidaFileExtension), r.Duration.Round(time.Millisecond))
			}
		}
	}

	fmt.Printf("\n\n\n───────────────────────────────────────────────────────\n")
	fmt.Printf("  Total: %d  |  Passed: %d  |  Failed: %d  |  %v\n",
		testCount, passed, failed, totalDuration.Round(time.Millisecond))
	fmt.Printf("───────────────────────────────────────────────────────\n\n\n\n")

}

func handleTestError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleError(err error) {
	if err != nil {
		printVersion()
		fmt.Printf("\t❌ %v\n\n\n\n", err.Error())
		os.Exit(1)
	}
}

func printError(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func parseCMD(cmd string) string {
	cmd = strings.ToLower(cmd)
	switch cmd {
	case RUN, DEGUG, TOKENS, AST, HELP, VERSION, ABOUT, CODE, TIME, INIT, INSTALL, COMPILE, BUNDLE:
		return cmd
	default:
		return cmd
	}
}

func errorNoArgsGivenTo(cmd string) error {
	return fmt.Errorf("no arguments given to the command '%v'", cmd)
}

func printVersion() {
	fmt.Printf("\n\n\n   %v\n   %v\n\n\n", vida.Name(), vida.Version())
}

func printHelp() {
	clear()
	printVersion()
	fmt.Printf("\tVida Tool\n")
	fmt.Println("\tUsage:  vida  [command]  [script]")
	fmt.Printf("\n\n")
	fmt.Printf("\t%-11v compile a script to bytecode and run it\n", RUN)
	fmt.Printf("\t%-11v scaffold a new project (--template=app|lib)\n", INIT)
	fmt.Printf("\t%-11v run focused or all scripts in path|project\n", TEST)
	fmt.Printf("\t%-11v download any dependencies or those listed in vida.toml\n", INSTALL)
	fmt.Printf("\t%-11v compile script to machine code\n", COMPILE)
	fmt.Printf("\t%-11v flatten a script and its local imports into one file (-o out.vida | --compile)\n", BUNDLE)
	fmt.Printf("\t%-11v run a script step by step\n", DEGUG)
	fmt.Printf("\t%-11v compile a script to bytecode and measure its runtime\n", TIME)
	fmt.Printf("\t%-11v show a list of tokens\n", TOKENS)
	fmt.Printf("\t%-11v show the abstact syntax tree\n", AST)
	fmt.Printf("\t%-11v show a colorized abstract syntax tree\n", SEMANTIC_AST)
	fmt.Printf("\t%-11v show this message\n", HELP)
	fmt.Printf("\t%-11v show Vida's version\n", VERSION)
	fmt.Printf("\t%-11v compile a script to bytecode and show it\n", CODE)
	fmt.Printf("\t%-11v gets or sets vidapath in the env vars of the host system\n", PATH)
	fmt.Printf("\t%-11v show Vida's oneline description\n", ABOUT)
	fmt.Printf("\t%-11v add directories to the module search path\n", FLAG_I)
	fmt.Printf("\t%-11v %v", vida.EmptyString, flagIExample)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

const flagIExample = "vida run -I ~/vida-scripts -I ~/shared-libs [script].vida"

func printAbout() {
	clear()
	fmt.Println(vida.About())
}

func clear() {
	fmt.Printf("\u001B[H")
	fmt.Printf("\u001B[2J")
	fmt.Printf("\u001B[3J")
}

func softclear() {
	fmt.Printf("\u001B[H")
	fmt.Printf("\u001B[2J")
}

func printDuration(duration time.Duration) {
	fmt.Printf("\n\n\n\n")
	fmt.Printf("\tDuration in Seconds : %vs\n", duration.Seconds())
	fmt.Printf("\tDuration            : %v", duration)
	fmt.Printf("\n\n\n\n")
}

func extractIncludePaths(args []string) (filtered []string, includes []string) {
	filtered = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "-I" && i+1 < len(args):
			includes = append(includes, args[i+1])
			i++
		case strings.HasPrefix(args[i], "-I="):
			includes = append(includes, strings.TrimPrefix(args[i], "-I="))
		default:
			filtered = append(filtered, args[i])
		}
	}
	return filtered, includes
}

func applyIncludePaths(paths []string) {
	if len(paths) == 0 {
		return
	}
	joined := strings.Join(paths, string(os.PathListSeparator))
	if existing := os.Getenv(vida.VIDAPATH); existing != vida.EmptyString {
		joined = joined + string(os.PathListSeparator) + existing
	}
	os.Setenv(vida.VIDAPATH, joined)
}

// resolveTestDir checks for a vida.toml in cwd; if present and it names a
// test directory, that subdirectory is used instead of cwd itself. This
// keeps 'vida test' from also re-running main.vida or src/ files that
// happen to live in the same project. Falls back to cwd unchanged for
// non-project directories or manifests without a 'test' field.
func resolveTestDir(cwd string) string {
	manifestPath := filepath.Join(cwd, manifestFileName)
	if !fileExistsOnDiskCLI(manifestPath) {
		return cwd
	}
	m, err := readManifest(manifestPath)
	if err != nil || m.Test == vida.EmptyString {
		return cwd
	}
	testDir := filepath.Join(cwd, m.Test)
	if info, err := os.Stat(testDir); err != nil || !info.IsDir() {
		return cwd
	}
	return testDir
}

// path checks whether VIDAPATH has set and notify about that.
// otherwise it creates VIDAPATH at {os.HOME}/vida-cells
func path() {
	clear()
	printVersion()
	if path, exists := os.LookupEnv(vida.VIDAPATH); exists {
		fmt.Printf("\tVIDAPATH has already been set at\n")
		fmt.Printf("\t%v\n\n\n", path)
	} else {
		home, err := os.UserHomeDir()
		handleError(err)
		globalVidaCellsPath := filepath.Join(home, vida.VidaPathDirName)
		handleError(os.Setenv(vida.VIDAPATH, globalVidaCellsPath))
		if path, exists := os.LookupEnv(vida.VIDAPATH); exists {
			fmt.Printf("\tVIDAPATH has been successfully set at\n")
			fmt.Printf("\t%v\n\n\n", path)
		} else {
			fmt.Printf("\tVIDAPATH could not be set in your system\n")
			fmt.Printf("\tPlease look for help online\n\n\n")
		}
	}
}
