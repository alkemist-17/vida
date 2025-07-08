package main

import (
	"errors"
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
	UNKNOWN = "unknown"
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
			printVersion()
		case ABOUT:
			printAbout()
		case CODE:
			printMachineCode(args)
		case CORELIB:
			printCoreLib()
		default:
			clear()
			printVersion()
			handleError(errors.New("unknown command\ntype in your cli 'vida help' for assistance"))
		}
	} else {
		printHelp()
	}
}

func runDebug(args []string) {
	clear()
	extensions := extension.LoadExtensions()
	if len(args) > 2 {
		printVersion()
		i, err := vida.NewDebugger(args[2], extensions)
		handleError(err)
		r, err := i.Debug()
		handleError(err)
		fmt.Println(r)
	} else {
		printVersion()
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
		printVersion()
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
		fmt.Println(r)
	} else {
		printVersion()
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
		printVersion()
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
		printVersion()
		handleError(errorNoArgsGivenTo(AST))
	}
}

func printMachineCode(args []string) {
	clear()
	printVersion()
	largs := len(args)
	if largs > 2 {
		for i := 2; i < largs; i++ {
			err := vida.PrintMachineCode(args[i])
			handleError(err)
		}
	} else {
		printVersion()
		handleError(errorNoArgsGivenTo(CODE))
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("\n\n%v\n\n\n", err)
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
		return UNKNOWN
	}
}

func errorNoArgsGivenTo(cmd string) error {
	return fmt.Errorf("no arguments given to the option %v", cmd)
}

func printVersion() {
	fmt.Printf("\n\n%v %v\n\n\n", vida.Name(), vida.Version())
}

func printHelp() {
	clear()
	printVersion()
	fmt.Println("CLI")
	fmt.Println()
	fmt.Println("Usage: vida [command] [...arguments]")
	fmt.Println()
	fmt.Println("Command list")
	fmt.Println()
	fmt.Printf("%-11v compile and run Vida script\n", RUN)
	fmt.Printf("%-11v compile and run Vida scripts step by step\n", DEGUG)
	fmt.Printf("%-11v compile and run Vida scripts measuring their runtime\n", TIME)
	fmt.Printf("%-11v show the token list\n", TOKENS)
	fmt.Printf("%-11v show the syntax tree\n", AST)
	fmt.Printf("%-11v show this message\n", HELP)
	fmt.Printf("%-11v show the language version\n", VERSION)
	fmt.Printf("%-11v compile and show the compiled code\n", CODE)
	fmt.Printf("%-11v show information about the Vida corelib\n", CORELIB)
	fmt.Printf("%-11v show some information about Vida\n", ABOUT)
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
