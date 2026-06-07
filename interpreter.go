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

func NewInterpreter(path string, extensionsLoader ExtensionsLoader) (*Interpreter, error) {
	threadPoolIsDown = true
	src, err := LoadScriptFromFile(path)
	if err != nil {
		return nil, err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return nil, err
	}
	c := newCompiler(rAst, path)
	script, err := c.compileScript()
	if err != nil {
		return nil, err
	}
	mainThread, err := newMainThread(script, extensionsLoader)
	if err != nil {
		return nil, err
	}
	vm := &VM{mainThread}
	(*(script.GlobalStore))[globalStateIndex] = &GlobalState{Main: mainThread, Current: mainThread, VM: vm}
	return &Interpreter{
		parser:   p,
		compiler: c,
		vm:       vm,
	}, nil
}

func NewDebugger(path string, extensionsLoader ExtensionsLoader) (*Interpreter, error) {
	src, err := LoadScriptFromFile(path)
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
	c := newCompiler(rAst, path)
	script, err := c.compileScript()
	if err != nil {
		return nil, err
	}
	fmt.Println(PrintBytecode(script, script.MainFunction.CoreFn.ScriptID))
	fmt.Print("\n\nPress 'Enter' to continue => ")
	fmt.Scanf(" ")
	mainThread, err := newMainThread(script, extensionsLoader)
	if err != nil {
		return nil, err
	}
	vm := &VM{mainThread}
	(*(script.GlobalStore))[globalStateIndex] = &GlobalState{Main: mainThread, Current: mainThread, VM: vm}
	return &Interpreter{
		parser:   p,
		compiler: c,
		vm:       vm,
	}, nil
}

func PrintAST(path string, colorized bool) error {
	src, err := LoadScriptFromFile(path)
	if err != nil {
		return err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return err
	}
	if colorized {
		fmt.Println(ast.PrintASTColor(rAst))
	} else {
		fmt.Println(ast.PrintAST(rAst))
	}
	return nil
}

func PrintTokens(path string) error {
	src, err := LoadScriptFromFile(path)
	if err != nil {
		return err
	}
	l := lexer.New(src, path)
	hadError := false
	fmt.Printf("%5v   %-15v   %-2v\n\n", "Line", "Token", "Value")
	for {
		line, tok, lit := l.Next()
		if l.LexicalError.Message != EmptyString {
			hadError = true
			break
		}
		fmt.Printf("%5v   %-15v   %-2v\n", line, tok, lit)
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
	src, err := LoadScriptFromFile(path)
	if err != nil {
		return err
	}
	p := newParser(src, path)
	rAst, err := p.parse()
	if err != nil {
		return err
	}
	c := newCompiler(rAst, path)
	script, err := c.compileScript()
	if err != nil {
		return err
	}
	fmt.Println(PrintBytecode(script, script.MainFunction.CoreFn.ScriptID))
	return nil
}

func (i *Interpreter) Run() (Result, error) {
	return i.vm.run()
}

func (i *Interpreter) MeasureRunTime() (Result, error) {
	init := time.Now()
	r, err := i.vm.run()
	end := time.Since(init)
	fmt.Printf("\n\tThe interpreter has finished its work\n\n\n\n")
	fmt.Printf("\tTime Sec : %vs\n", end.Seconds())
	fmt.Printf("\tTime End : %v\n\n\n\n", end)
	return r, err
}

func (i *Interpreter) Debug() (Result, error) {
	return i.vm.debug()
}

func (i *Interpreter) PrintCallStack() {
	i.vm.printCallStack()
}
