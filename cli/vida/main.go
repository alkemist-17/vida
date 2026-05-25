package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alkemist-17/vida"
	"github.com/alkemist-17/vida/extensions"
)

const (
	RUN     = "run"
	DEGUG   = "debug"
	TIME    = "time"
	TOKENS  = "tokens"
	AST     = "ast"
	HELP    = "help"
	VERSION = "version"
	ABOUT   = "about"
	CODE    = "code"
	TEST    = "test"
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
			time(args)
		case TOKENS:
			printTokens(args)
		case AST:
			printAST(args)
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
		i, err := vida.NewDebugger(p, extensions.GetLoader())
		handleError(err)
		r, err := i.Debug()
		handleError(err)
		fmt.Println(r)
	} else {
		handleError(errorNoArgsGivenTo(DEGUG))
	}
}

func run(args []string) {
	if len(args) > 2 {
		p, err := filepath.Abs(args[2])
		handleError(err)
		i, err := vida.NewInterpreter(p, extensions.GetLoader())
		handleError(err)
		_, err = i.Run()
		if err != nil {
			printError(err)
			i.PrintCallStack()
		}
	} else {
		handleError(errorNoArgsGivenTo(RUN))
	}
}

func time(args []string) {
	clear()
	printVersion()
	if len(args) > 2 {
		p, err := filepath.Abs(args[2])
		handleError(err)
		i, err := vida.NewInterpreter(p, extensions.GetLoader())
		handleError(err)
		r, err := i.MeasureRunTime()
		if err != nil {
			printError(err)
			i.PrintCallStack()
			fmt.Printf("\tResult : %v ❌\n\n\n\n", r)
			return
		}
		fmt.Printf("\tResult : %v ✅\n\n\n\n", r)
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
			handleError(vida.PrintTokens(p))
		}
	} else {
		handleError(errorNoArgsGivenTo(TOKENS))
	}
}

func printAST(args []string) {
	clear()
	printVersion()
	largs := len(args)
	if largs > 2 {
		for i := 2; i < largs; i++ {
			p, err := filepath.Abs(args[i])
			handleError(err)
			handleError(vida.PrintAST(p))
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
			handleError(vida.PrintMachineCode(p))
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
				fmt.Printf("\t🧪 Running '%v'\n\n\n", v)
				executeScript(filepath.Join(filepath.Dir(args[0]), v))
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
	i, err := vida.NewInterpreter(path, extensions.GetLoader())
	handleTestError(err)
	r, err := i.MeasureRunTime()
	handleTestFailure(r, err)
	fmt.Printf("\tResult : %v ✅\n\n\n\n", r)
}

func handleTestFailure(r vida.Result, err error) {
	if err != nil {
		fmt.Printf("\tResult : %v ❌\n\n", r)
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
