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
)

func main() {
	// f, err := os.Create("vida.prof")
	// handleError(err)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	args := os.Args
	if len(args) > 1 {
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
		default:
			handleError(fmt.Errorf("unknown command '%v'.\n\tType 'vida help' for assistance.", parseCMD(args[1])))
		}
	} else {
		printHelp()
	}
}

func runDebug(args []string) {
	clear()
	if len(args) > 2 {
		printVersion()
		p, err := filepath.Abs(args[2])
		handleError(err)
		src, err := vida.LoadScriptFromFile(p)
		handleError(err)
		ctx := vida.NewContext(src, p, extensions.GetLoader())
		err = ctx.RunDebugSession()
		handleError(err)
	} else {
		handleError(errorNoArgsGivenTo(DEGUG))
	}
}

func run(args []string) {
	if len(args) > 2 {
		p, err := filepath.Abs(args[2])
		handleError(err)
		src, err := vida.LoadScriptFromFile(p)
		handleError(err)
		ctx := vida.NewContext(src, p, extensions.GetLoader())
		err = ctx.CompileAndRun()
		if err != nil {
			printError(err)
			ctx.PrintCallStack()
		}
	} else {
		handleError(errorNoArgsGivenTo(RUN))
	}
}

func measureRunTime(args []string) {
	clear()
	printVersion()
	if len(args) > 2 {
		p, err := filepath.Abs(args[2])
		handleError(err)
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
	} else {
		handleError(errorNoArgsGivenTo(TIME))
	}
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
	printVersion()
	testCount := 0
	if len(args) > 2 {
		dir := args[2]
		info, err := os.Stat(dir)
		handleTestError(err)
		if info.IsDir() {
			scripts, err := os.ReadDir(dir)
			handleTestError(err)
			runScripts(dir, scripts, &testCount)
			goto stats
		}
		for _, v := range args[2:] {
			if strings.HasSuffix(v, vida.VidaFileExtension) {
				testCount++
				p, err := filepath.Abs(v)
				handleError(err)
				fmt.Printf("\t🧪 Running '%v'\n\n\n", p)
				executeScript(p)
				fmt.Printf("\n\n\n\n")
			}
		}
	} else {
		dir, err := os.Getwd()
		scripts, err := os.ReadDir(dir)
		handleTestError(err)
		runScripts(dir, scripts, &testCount)
	}
stats:
	if testCount > 0 {
		fmt.Printf("\t🧪\tAll tests were ok!\n\t\tTotal files run: %v\n\n\n\n\n\n\n", testCount)
	} else {
		fmt.Printf("\t❌\tNo vida files were found!\n\t\tTotal files run: %v\n\n\n\n\n\n\n", testCount)
	}
}

func runScripts(dir string, scripts []os.DirEntry, textCount *int) {
	for _, v := range scripts {
		if !v.IsDir() && strings.HasSuffix(v.Name(), vida.VidaFileExtension) {
			(*textCount)++
			fmt.Printf("\t🧪 Running '%v'\n\n\n", v.Name())
			executeScript(filepath.Join(dir, v.Name()))
			fmt.Printf("\n\n\n\n")
		}
	}
}

func executeScript(path string) {
	src, err := vida.LoadScriptFromFile(path)
	handleError(err)
	ctx := vida.NewContext(src, path, extensions.GetLoader())
	handleError(ctx.Compile())
	dur, err := ctx.MeasureRunTime()
	printDuration(dur)
	handleTestError(err)
	handleTestFailure(err)
	fmt.Printf("\tSuccess ✅\n\n\n\n")
}

func handleTestFailure(err error) {
	if err != nil {
		fmt.Printf("\tFailure ❌\n\n")
		fmt.Println(err)
		fmt.Printf("\n\n")
		os.Exit(1)
	}
}

func handleTestError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleError(err error) {
	if err != nil {
		//clear()
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
	case RUN, DEGUG, TOKENS, AST, HELP, VERSION, ABOUT, CODE, TIME:
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
	fmt.Printf("\tCLI Tool\n")
	fmt.Println("\tUsage:  vida  [command]  [script]")
	fmt.Printf("\n\n")
	fmt.Printf("\t%-11v compile and run a script\n", RUN)
	fmt.Printf("\t%-11v run focused or all scripts in path\n", TEST)
	fmt.Printf("\t%-11v compile and run a script step by step\n", DEGUG)
	fmt.Printf("\t%-11v compile and run a script measuring their runtime\n", TIME)
	fmt.Printf("\t%-11v show the token list\n", TOKENS)
	fmt.Printf("\t%-11v show the syntax tree\n", AST)
	fmt.Printf("\t%-11v show a colorized syntax tree\n", SEMANTIC_AST)
	fmt.Printf("\t%-11v show this message\n", HELP)
	fmt.Printf("\t%-11v show the language version\n", VERSION)
	fmt.Printf("\t%-11v compile and show the compiled code\n", CODE)
	fmt.Printf("\t%-11v show the description of Vida\n", ABOUT)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

func printAbout() {
	clear()
	fmt.Println(vida.About())
}

func clear() {
	fmt.Printf("\u001B[H")
	fmt.Printf("\u001B[2J")
	fmt.Printf("\u001B[3J")
}

func printDuration(duration time.Duration) {
	fmt.Printf("\n\n\n\n")
	fmt.Printf("\tDuration in Seconds : %vs\n", duration.Seconds())
	fmt.Printf("\tDuration            : %v", duration)
	fmt.Printf("\n\n\n\n")
}
