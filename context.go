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
	extensionCache   map[string]*Object
	threadPool       *internalThreadPool
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
		return errors.New("error when running ctx.Run: source must be compiled first")
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
		if l.LexicalError != nil && l.LexicalError.Message != EmptyString {
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
		return errors.New("error when running ctx.PrintMachineCode: source must be compiled first")
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
	var ASTdescription string
	if withColors {
		ASTdescription = ast.PrintASTColor(scriptAST)
	} else {
		ASTdescription = ast.StringifyAST(scriptAST)
	}
	fmt.Println(ASTdescription)
	return nil
}

func (ctx *Context) RunDebugSession() (err error) {
	scriptAST, err := newParser(ctx.src, ctx.contextID).parse()
	if err != nil {
		return
	}
	fmt.Println(ast.StringifyAST(scriptAST))
	pressEnterToContinue()
	script, err := newCompiler(scriptAST, ctx.contextID, ctx.extensionsLoader).compileScript()
	if err != nil {
		return
	}
	ctx.script = script
	ctx.PrintMachineCode()
	pressEnterToContinue()
	ctx.vm = &VM{ctx.setMainThread(newThread(ctx.script)), ctx}
	return ctx.vm.debug()
}

func (ctx *Context) PrintCallStack() error {
	if ctx.script == nil && ctx.vm == nil {
		return errors.New("error when running ctx.PrintCallStack: source must be compiled first")
	}
	ctx.vm.printCallStack()
	return nil
}

func (ctx *Context) MeasureRunTime() (end time.Duration, err error) {
	if ctx.script == nil {
		return end, errors.New("error when running ctx.MeasureRunTime: source must be compiled first")
	}
	if ctx.vm != nil {
		init := time.Now()
		err = ctx.vm.run()
		end = time.Since(init)
		return end, err
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

func (ctx *Context) getInternalThread(fn *Function) *Thread {
	if ctx.threadPool == nil {
		ctx.threadPool = newInternalThreadPool()
	}
	return ctx.threadPool.get(fn, ctx.script)
}

func (ctx *Context) releaseInternalThread() {
	if ctx.threadPool != nil {
		ctx.threadPool.release()
	}
}
