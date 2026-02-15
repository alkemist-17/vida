package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alkemist-17/vida"
	"github.com/alkemist-17/vida/extension"
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
	CORELIB = "corelib"
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
		case CORELIB:
			printCoreLib()
		case TEST:
			test(args)
		default:
			handleError(fmt.Errorf("unknown command '%v'. Type 'vida help' for assistance", parseCMD(args[1])))
		}
	} else {
		printHelp()
	}
}

func runDebug(args []string) {
	clear()
	if len(args) > 2 {
		extensions := extension.LoadExtensions()
		printVersion()
		i, err := vida.NewDebugger(args[2], extensions)
		handleError(err)
		r, err := i.Debug()
		handleError(err)
		fmt.Println(r)
	} else {
		handleError(errorNoArgsGivenTo(DEGUG))
	}
}

func run(args []string) {
	extensions := extension.LoadExtensions()
	if len(args) > 2 {
		i, err := vida.NewInterpreter(args[2], extensions)
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
	extensions := extension.LoadExtensions()
	if len(args) > 2 {
		i, err := vida.NewInterpreter(args[2], extensions)
		handleError(err)
		r, err := i.MeasureRunTime()
		if err != nil {
			printError(err)
			i.PrintCallStack()
		}
		fmt.Printf("   Interpretation Result : %v\n\n\n\n", r)
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
			err := vida.PrintTokens(args[i])
			handleError(err)
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
			err := vida.PrintAST(args[i])
			handleError(err)
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
			err := vida.PrintMachineCode(args[i])
			handleError(err)
		}
	} else {
		handleError(errorNoArgsGivenTo(CODE))
	}
}

func test(args []string) {
	clear()
	printVersion()
	count := 0
	basePath := "./"
	scripts, err := os.ReadDir(basePath)
	handleTestError(err, basePath)
	if len(args) > 2 {
		for _, v := range args[2:] {
			if strings.HasSuffix(v, vida.VidaFileExtension) {
				count++
				fmt.Printf("🧪 Running tests from '%v'\n", v)
				executeScript(v)
				fmt.Printf("\n\n\n")
			}
		}
	} else {
		for _, v := range scripts {
			if !v.IsDir() && strings.HasSuffix(v.Name(), vida.VidaFileExtension) {
				count++
				fmt.Printf("🧪 Running tests from '%v'\n", v.Name())
				executeScript(v.Name())
				fmt.Printf("\n\n\n")
			}
		}
	}
	fmt.Printf("🧪  All tests were ok!\n    Total files run: %v\n\n\n\n\n\n\n", count)
}

func executeScript(path string) {
	i, err := vida.NewInterpreter(path, extension.LoadExtensions())
	handleTestError(err, path)
	r, err := i.MeasureRunTime()
	handleTestFailure(r, err)
	fmt.Printf("   Interpretation Result : %v ✅\n\n\n\n", r)
}

func handleTestFailure(r vida.Result, err error) {
	if err != nil {
		fmt.Printf("   Interpretation Result : %v ❌\n\n", r)
		fmt.Println(err)
		os.Exit(0)
	}
}

func handleTestError(err error, path string) {
	if err != nil {
		fmt.Println(err)
		fmt.Println(path)
		os.Exit(0)
	}
}

func handleError(err error) {
	if err != nil {
		clear()
		printVersion()
		fmt.Printf("   ❌ %v\n\n\n\n", err.Error())
		os.Exit(0)
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
	case RUN, DEGUG, TOKENS, AST, HELP, VERSION, ABOUT, CODE, TIME, CORELIB:
		return cmd
	default:
		return cmd
	}
}

func errorNoArgsGivenTo(cmd string) error {
	return fmt.Errorf("no arguments given to the command '%v'", cmd)
}

func printVersion() {
	fmt.Printf("\n\n\n   %v\n   %v\n\n\n\n", vida.Name(), vida.Version())
}

func printHelp() {
	clear()
	printVersion()
	fmt.Println("   CLI Tool")
	fmt.Println()
	fmt.Println("   Usage: vida [command] [...arguments]")
	fmt.Println()
	fmt.Println("   Where [command]:")
	fmt.Println()
	fmt.Printf("   %-11v compile and run a script\n", RUN)
	fmt.Printf("   %-11v run focused or all scripts in the current working directory\n", TEST)
	fmt.Printf("   %-11v compile and run a script step by step\n", DEGUG)
	fmt.Printf("   %-11v compile and run a script measuring their runtime\n", TIME)
	fmt.Printf("   %-11v show the token list\n", TOKENS)
	fmt.Printf("   %-11v show the syntax tree\n", AST)
	fmt.Printf("   %-11v show this message\n", HELP)
	fmt.Printf("   %-11v show the language version\n", VERSION)
	fmt.Printf("   %-11v compile and show the compiled code\n", CODE)
	fmt.Printf("   %-11v show information about the corelib\n", CORELIB)
	fmt.Printf("   %-11v show the description of Vida\n", ABOUT)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()
}

func printAbout() {
	clear()
	fmt.Println(vida.About())
}

func printCoreLib() {
	clear()
	printVersion()
	vida.PrintCoreLibInformation()
}

func clear() {
	fmt.Printf("\u001B[H")
	fmt.Printf("\u001B[2J")
}
