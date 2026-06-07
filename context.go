package vida

import (
	"errors"
	"fmt"
	"time"

	"github.com/alkemist-17/vida/ast"
	"github.com/alkemist-17/vida/lexer"
	"github.com/alkemist-17/vida/token"
)

type Context struct {
	src              []byte
	contextID        string
	mainThread       *Thread
	currentThread    *Thread
	script           *Script
	extensionsLoader ExtensionsLoader
	vm               *VM
}

func NewContext(src []byte, contextID string, extensionsLoader ExtensionsLoader) *Context {
	return &Context{
		src:              src,
		extensionsLoader: extensionsLoader,
		contextID:        contextID,
	}
}

func (ctx *Context) Compile() (err error) {
	ast, err := newParser(ctx.src, ctx.contextID).parse()
	if err != nil {
		return
	}
	script, err := newCompiler(ast, ctx.contextID, ctx.extensionsLoader).compileScript()
	if err != nil {
		return
	}
	ctx.script = script
	return err
}

func (ctx *Context) Run() error {
	if ctx.script == nil {
		return errors.New("first must compile the source")
	}
	if ctx.vm != nil {
		return ctx.vm.run()
	}
	ctx.vm = &VM{ctx.setMainThread(newThread(ctx.script)), ctx}
	return ctx.vm.run()
}

func (ctx *Context) CompileAndRun() (err error) {
	ast, err := newParser(ctx.src, ctx.contextID).parse()
	if err != nil {
		return
	}
	script, err := newCompiler(ast, ctx.contextID, ctx.extensionsLoader).compileScript()
	if err != nil {
		return
	}
	ctx.script = script
	ctx.vm = &VM{ctx.setMainThread(newThread(ctx.script)), ctx}
	return ctx.vm.run()
}

func (ctx *Context) PrintTokens() error {
	l := lexer.New(ctx.src, ctx.contextID)
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

func (ctx *Context) PrintMachineCode() error {
	if ctx.script == nil {
		return errors.New("first must compile the source")
	}
	fmt.Println(PrintBytecode(ctx.script, ctx.contextID))
	return nil
}

func (ctx *Context) PrintAST(withColors bool) error {
	p := newParser(ctx.src, ctx.contextID)
	scriptAST, err := p.parse()
	if err != nil {
		return err
	}
	if withColors {
		fmt.Println(ast.PrintASTColor(scriptAST))
	} else {
		fmt.Println(ast.StringifyAST(scriptAST))
	}
	return nil
}

func (ctx *Context) RunDebugSession() (err error) {
	scriptAST, err := newParser(ctx.src, ctx.contextID).parse()
	if err != nil {
		return
	}
	fmt.Println(ast.StringifyAST(scriptAST))
	fmt.Print("\n\nPress 'Enter' to continue => ")
	fmt.Scanf(" ")
	script, err := newCompiler(scriptAST, ctx.contextID, ctx.extensionsLoader).compileScript()
	if err != nil {
		return
	}
	ctx.script = script
	ctx.PrintMachineCode()
	fmt.Print("\n\nPress 'Enter' to continue => ")
	fmt.Scanf(" ")
	ctx.vm = &VM{ctx.setMainThread(newThread(ctx.script)), ctx}
	return ctx.vm.debug()
}

func (ctx *Context) PrintCallStack() error {
	if ctx.vm == nil {
		return errors.New("first must compile the source")
	}
	ctx.vm.printCallStack()
	return nil
}

func (ctx *Context) MeasureRunTime() (end time.Duration, err error) {
	if ctx.script == nil {
		return end, errors.New("first must compile the source")
	}
	if ctx.vm != nil {
		return end, ctx.vm.run()
	}
	ctx.vm = &VM{ctx.setMainThread(newThread(ctx.script)), ctx}
	init := time.Now()
	err = ctx.vm.run()
	end = time.Since(init)
	return end, err
}

func (ctx *Context) SetContextID(contextId string) {
	ctx.contextID = contextId
}

func (ctx *Context) SetExtensionsLoader(extensionsLoader ExtensionsLoader) {
	ctx.extensionsLoader = extensionsLoader
}

func (ctx *Context) SetSource(src []byte) {
	ctx.src = src
}

func (ctx *Context) IsMainThreadRunning() bool {
	return ctx.mainThread == ctx.currentThread
}

func (ctx *Context) setMainThread(thread *Thread) *Thread {
	ctx.mainThread = thread
	ctx.currentThread = thread
	return thread
}
