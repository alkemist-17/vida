package vida

import (
	"fmt"
	"time"

	"github.com/alkemist-17/vida/ast"
	"github.com/alkemist-17/vida/lexer"
	"github.com/alkemist-17/vida/token"
)

type Interpreter struct {
	parser   *parser
	compiler *compiler
	vm       *VM
}

func NewInterpreter(path string, extensionlibloader map[string]func() Value) (*Interpreter, error) {
	src, err := readScript(path)
	if err != nil {
		return nil, err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return nil, err
	}
	c := newMainCompiler(rAst, path)
	script, err := c.compileScript()
	if err != nil {
		return nil, err
	}
	mainThread, err := newMainThread(script, extensionlibloader)
	if err != nil {
		return nil, err
	}
	vm := &VM{mainThread}
	(*(script.Store))[globalStateIndex] = &GlobalState{Main: mainThread, Current: mainThread, VM: vm}
	return &Interpreter{
		parser:   p,
		compiler: c,
		vm:       vm,
	}, nil
}

func NewDebugger(path string, extensionlibloader map[string]func() Value) (*Interpreter, error) {
	src, err := readScript(path)
	if err != nil {
		return nil, err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return nil, err
	}
	fmt.Println(ast.PrintAST(rAst))
	fmt.Print("\n\nPress 'Enter' to continue => ")
	fmt.Scanf(" ")
	c := newMainCompiler(rAst, path)
	script, err := c.compileScript()
	if err != nil {
		return nil, err
	}
	fmt.Println(PrintBytecode(script, script.MainFunction.CoreFn.ScriptName))
	fmt.Print("\n\nPress 'Enter' to continue => ")
	fmt.Scanf(" ")
	mainThread, err := newMainThread(script, extensionlibloader)
	if err != nil {
		return nil, err
	}
	vm := &VM{mainThread}
	(*(script.Store))[globalStateIndex] = &GlobalState{Main: mainThread, Current: mainThread, VM: vm}
	return &Interpreter{
		parser:   p,
		compiler: c,
		vm:       vm,
	}, nil
}

func PrintAST(path string) error {
	src, err := readScript(path)
	if err != nil {
		return err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return err
	}
	fmt.Println(ast.PrintAST(rAst))
	return nil
}

func PrintTokens(path string) error {
	src, err := readScript(path)
	if err != nil {
		return err
	}
	l := lexer.New(src, path)
	hadError := false
	fmt.Printf("%4v %-15v %-2v\n\n", "line", "token", "repr")
	for {
		line, tok, lit := l.Next()
		if l.LexicalError.Message != "" {
			hadError = true
			break
		}
		fmt.Printf("%4v %-15v %-2v\n", line, tok, lit)
		if tok == token.EOF {
			fmt.Println()
			break
		}
	}
	if hadError {
		return l.LexicalError
	}
	return nil
}

func PrintMachineCode(path string) error {
	src, err := readScript(path)
	if err != nil {
		return err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return err
	}
	c := newMainCompiler(rAst, path)
	script, err := c.compileScript()
	if err != nil {
		return err
	}
	fmt.Println(PrintBytecode(script, script.MainFunction.CoreFn.ScriptName))
	return nil
}

func (i *Interpreter) Run() (Result, error) {
	return i.vm.run()
}

func (i *Interpreter) MeasureRunTime() (Result, error) {
	init := time.Now()
	r, err := i.vm.run()
	end := time.Since(init)
	fmt.Printf("\n\nThe interpreter has finished.\n\n")
	fmt.Printf("Time = %vs\n", end.Seconds())
	fmt.Printf("Time = %v\n\n", end)
	return r, err
}

func (i *Interpreter) Debug() (Result, error) {
	return i.vm.debug()
}

func (i *Interpreter) PrintCallStack() {
	i.vm.printCallStack()
}
