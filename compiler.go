package vida

import (
	"path/filepath"

	"github.com/alkemist-17/vida/ast"
	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type compiler struct {
	jumps            []int
	breakJumps       []int
	breakCount       []int
	continueJumps    []int
	continueCount    []int
	fn               []*CoreFunction
	errMsg           string
	mainScriptID     string
	currentFn        *CoreFunction
	ast              *ast.Ast
	script           *Script
	kb               *konstBuilder
	sb               *symbolBuilder
	scriptMap        map[string]int
	depMap           map[string]struct{}
	extensionsLoader ExtensionsLoader
	errorInfo        ErrorInfo
	lineErr          uint
	scope            int
	level            int
	rAlloc           int
	rDest            int
	fromRefStmt      bool
	mutLoc           bool
	hadError         bool
	isSubcompiler    bool
}

var dummy = struct{}{}

func newCompiler(ast *ast.Ast, scriptID string, ctx *Context, extensionsLoader ExtensionsLoader) *compiler {
	dm := make(map[string]struct{})
	dm[scriptID] = dummy
	ei := make(ErrorInfo)
	ei[scriptID] = make(map[int]uint)
	c := &compiler{
		ast:              ast,
		script:           newScript(scriptID, extensionsLoader),
		kb:               newKonstBuilder(ctx),
		sb:               newSymbolBuilder(0),
		scriptMap:        make(map[string]int),
		depMap:           dm,
		errorInfo:        ei,
		mainScriptID:     scriptID,
		extensionsLoader: extensionsLoader,
	}
	c.fn = append(c.fn, c.script.MainFunction.CoreFn)
	c.currentFn = c.script.MainFunction.CoreFn
	estInstr := max(len(ast.Statement)*4, 64)
	c.currentFn.Code = make([]uint64, 0, estInstr)
	return c
}

func newSubCompiler(ast *ast.Ast, scriptID string, kb *konstBuilder, store *[]Value, scriptMap map[string]int, depMap map[string]struct{}, ei ErrorInfo, initialIndex int, extensionsLoader ExtensionsLoader) *compiler {
	ei[scriptID] = make(map[int]uint)
	c := &compiler{
		ast:           ast,
		script:        newSubScript(scriptID, store, extensionsLoader),
		kb:            kb,
		sb:            newSymbolBuilder(initialIndex),
		isSubcompiler: true,
		scriptMap:     scriptMap,
		depMap:        depMap,
		errorInfo:     ei,
		mainScriptID:  scriptID,
	}
	c.fn = append(c.fn, c.script.MainFunction.CoreFn)
	c.currentFn = c.script.MainFunction.CoreFn
	return c
}

func (c *compiler) compileScript() (*Script, error) {
	c.appendHeader()
	var i int
	for i = range len(c.ast.Statement) {
		c.compileStmt(c.ast.Statement[i])
		if c.hadError {
			return nil, verror.New(c.script.MainFunction.CoreFn.ScriptID, c.errMsg, verror.BuildErrType, c.lineErr)
		}
	}
	c.script.Konstants = c.kb.Konstants
	c.script.ErrorInfo = c.errorInfo
	c.appendEnd()
	return c.script, nil
}

func (c *compiler) compileSubScript() (*Script, error) {
	for i := range len(c.ast.Statement) {
		c.compileStmt(c.ast.Statement[i])
		if c.hadError {
			return nil, verror.New(c.script.MainFunction.CoreFn.ScriptID, c.errMsg, verror.BuildErrType, c.lineErr)
		}
	}
	return c.script, nil
}

func (c *compiler) compileStmt(node ast.Node) {
	switch n := node.(type) {
	case *ast.Mut:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		to, sIdent := c.refScope(n.Identifier)
		switch sIdent {
		case rFree:
			from, sexpr := c.compileExpr(n.Expr, true)
			switch sexpr {
			case rGlob:
				c.emitStore(from, to, storeFromGlobal, storeFromFree)
			case rKonst:
				c.emitStore(from, to, storeFromKonst, storeFromFree)
			case rFree:
				if from != to {
					c.emitStore(from, to, storeFromFree, storeFromFree)
				}
			case rLoc:
				c.emitStore(from, to, storeFromLocal, storeFromFree)
			}
		case rLoc:
			c.mutLoc = true
			c.rDest = to
			from, sexpr := c.compileExpr(n.Expr, true)
			switch sexpr {
			case rGlob:
				c.emitLoad(from, to, loadFromGlobal)
			case rLoc:
				if from != to {
					c.emitLoad(from, to, loadFromLocal)
				}
			case rKonst:
				c.emitLoad(from, to, loadFromKonst)
			case rFree:
				c.emitLoad(from, to, loadFromFree)
			}
			c.mutLoc = false
		case rGlob:
			from, sexpr := c.compileExpr(n.Expr, true)
			switch sexpr {
			case rGlob:
				if from != to {
					c.emitStore(from, to, storeFromGlobal, storeFromGlobal)
				}
			case rKonst:
				c.emitStore(from, to, storeFromKonst, storeFromGlobal)
			case rFree:
				c.emitStore(from, to, storeFromFree, storeFromGlobal)
			case rLoc:
				c.emitStore(from, to, storeFromLocal, storeFromGlobal)
			}
		case rNotDefined:
			c.generateReferenceError(n.Identifier, n.Line)
		}
	case *ast.Let:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		to, isPresent := c.sb.addGlobal(n.Identifier)
		if isPresent {
			c.generateGlobalAlreadyDefinedError(n.Identifier, n.Line)
			return
		}
		if _, isLocal, _ := c.sb.isLocal(n.Identifier); isLocal {
			c.generateGlobalShadowedByLocalError(n.Identifier, n.Line)
			return
		}
		*c.script.GlobalStore = append(*c.script.GlobalStore, Nil)
		from, scope := c.compileExpr(n.Expr, true)
		switch scope {
		case rKonst:
			c.emitStore(from, to, storeFromKonst, storeFromGlobal)
		case rGlob:
			if from != to {
				c.emitStore(from, to, storeFromGlobal, storeFromGlobal)
			}
		case rFree:
			c.emitStore(from, to, storeFromFree, storeFromGlobal)
		case rLoc:
			c.emitStore(from, to, storeFromLocal, storeFromGlobal)
		}
	case *ast.MultipleLet:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		for _, id := range n.Identifiers {
			to, isPresent := c.sb.addGlobal(id)
			if isPresent {
				c.generateGlobalAlreadyDefinedError(id, n.Line)
				return
			}
			if _, isLocal, _ := c.sb.isLocal(id); isLocal {
				c.generateGlobalShadowedByLocalError(id, n.Line)
				return
			}
			*c.script.GlobalStore = append(*c.script.GlobalStore, Nil)
			from, scope := c.compileExpr(n.Expr, true)
			switch scope {
			case rKonst:
				c.emitStore(from, to, storeFromKonst, storeFromGlobal)
			case rGlob:
				if from != to {
					c.emitStore(from, to, storeFromGlobal, storeFromGlobal)
				}
			case rFree:
				c.emitStore(from, to, storeFromFree, storeFromGlobal)
			case rLoc:
				c.emitStore(from, to, storeFromLocal, storeFromGlobal)
			}
		}
	case *ast.Var:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		if _, isGlobal := c.sb.isGlobal(n.Identifier); isGlobal {
			c.generateGlobalShadowedByLocalError(n.Identifier, n.Line)
			return
		}
		if _, isLocal, k := c.sb.isLocal(n.Identifier); isLocal && c.scope == k.scope {
			c.generateLocalAlreadyDefinedError(n.Identifier, n.Line)
			return
		}
		to := c.rAlloc
		var from, scope int
		if n.IsRecursive {
			c.sb.addLocal(n.Identifier, c.level, c.scope, to)
			c.emitLoad(c.kb.NilIndex(), to, loadFromKonst)
			from, scope = c.compileExpr(n.Expr, true)
		} else {
			from, scope = c.compileExpr(n.Expr, true)
			c.sb.addLocal(n.Identifier, c.level, c.scope, to)
		}
		switch scope {
		case rKonst:
			c.emitLoad(from, to, loadFromKonst)
		case rGlob:
			c.emitLoad(from, to, loadFromGlobal)
		case rFree:
			c.emitLoad(from, to, loadFromFree)
		case rLoc:
			if from != to {
				c.emitLoad(from, to, loadFromLocal)
			}
		}
		c.rAlloc++
	case *ast.MultipleVar:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		for _, id := range n.Identifiers {
			if _, isGlobal := c.sb.isGlobal(id); isGlobal {
				c.generateGlobalShadowedByLocalError(id, n.Line)
				return
			}
			if _, isLocal, k := c.sb.isLocal(id); isLocal && c.scope == k.scope {
				c.generateLocalAlreadyDefinedError(id, n.Line)
				return
			}
			to := c.rAlloc
			var from, scope int
			if n.IsRecursive {
				c.sb.addLocal(id, c.level, c.scope, to)
				c.emitLoad(c.kb.NilIndex(), to, loadFromKonst)
				from, scope = c.compileExpr(n.Expr, true)
			} else {
				from, scope = c.compileExpr(n.Expr, true)
				c.sb.addLocal(id, c.level, c.scope, to)
			}
			switch scope {
			case rKonst:
				c.emitLoad(from, to, loadFromKonst)
			case rGlob:
				c.emitLoad(from, to, loadFromGlobal)
			case rFree:
				c.emitLoad(from, to, loadFromFree)
			case rLoc:
				if from != to {
					c.emitLoad(from, to, loadFromLocal)
				}
			}
			c.rAlloc++
		}
	case *ast.Branch:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		elifCount := len(n.Elifs)
		hasElif := elifCount != 0
		e, hasElse := n.Else.(*ast.Else)
		shouldJumpOutside := hasElif || hasElse
		c.compileConditional(n.If.(*ast.If), shouldJumpOutside)
		if hasElif {
			for i := 0; i < elifCount-1; i++ {
				c.compileConditional(n.Elifs[i].(*ast.If), hasElif)
			}
			c.compileConditional(n.Elifs[elifCount-1].(*ast.If), hasElse)
		}
		if hasElse {
			c.compileStmt(e.Block)
		}
		if shouldJumpOutside {
			addr := len(c.currentFn.Code)
			for _, v := range c.jumps {
				c.currentFn.Code[v] |= uint64(addr)
			}
			c.jumps = c.jumps[:0]
		}
	case *ast.For:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		c.startLoopScope()
		c.scope++
		ireg := c.rAlloc

		initIdx, initScope := c.compileExpr(n.Init, true)
		c.exprToReg(initIdx, initScope)

		c.rAlloc++
		endIdx, endScope := c.compileExpr(n.End, true)
		c.exprToReg(endIdx, endScope)

		c.rAlloc++
		stepIdx, stepScope := c.compileExpr(n.Step, true)
		c.exprToReg(stepIdx, stepScope)

		c.rAlloc++
		c.sb.addLocal(n.Id, c.level, c.scope, c.rAlloc)
		c.emitLoad(c.kb.IntegerIndex(0), c.rAlloc, loadFromKonst)

		c.rAlloc++
		c.emitForSet(ireg, 0)
		loop := len(c.currentFn.Code)

		c.compileStmt(n.Block)
		checkLoop := len(c.currentFn.Code)

		c.currentFn.Code[loop-1] |= uint64(checkLoop) << shift16
		c.emitForLoop(ireg, loop)
		c.cleanUpLoopScope(loop, false)

		c.rAlloc -= c.sb.clearLocals(c.level, c.scope)
		c.rAlloc -= 3
		c.scope--
	case *ast.IFor:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		c.startLoopScope()
		c.scope++
		ireg := c.rAlloc
		c.emitLoad(c.kb.IntegerIndex(0), ireg, loadFromKonst)

		c.rAlloc++
		c.sb.addLocal(n.Key, c.level, c.scope, c.rAlloc)
		c.emitLoad(c.kb.IntegerIndex(0), c.rAlloc, loadFromKonst)

		c.rAlloc++
		c.sb.addLocal(n.Value, c.level, c.scope, c.rAlloc)
		c.emitLoad(c.kb.IntegerIndex(0), c.rAlloc, loadFromKonst)

		c.rAlloc++
		i, s := c.compileExpr(n.Expr, true)
		c.exprToReg(i, s)

		c.emitIForSet(0, c.rAlloc, ireg)
		loop := len(c.currentFn.Code)

		c.compileStmt(n.Block)
		checkLoop := len(c.currentFn.Code)

		c.currentFn.Code[loop-1] |= uint64(checkLoop) << shift32
		c.emitIForLoop(ireg, loop)
		c.cleanUpLoopScope(loop, false)

		c.rAlloc -= c.sb.clearLocals(c.level, c.scope)
		c.rAlloc--
		c.scope--
	case *ast.While:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		c.startLoopScope()
		init := len(c.currentFn.Code)
		idx, scope := c.compileExpr(n.Condition, true)
		if scope == rKonst {
			switch v := (*c.kb.Konstants)[idx].(type) {
			case NilValue:
				c.skipBlock(n.Block)
				c.cleanUpLoopScope(init, true)
				return
			case Bool:
				if !v {
					c.skipBlock(n.Block)
					c.cleanUpLoopScope(init, true)
					return
				}
			}
			c.compileStmt(n.Block)
			c.emitJump(init)
			c.cleanUpLoopScope(init, true)
		} else {
			c.exprToReg(idx, scope)
			addr := len(c.currentFn.Code)
			c.emitCheck(0, c.rAlloc, 0)
			c.compileStmt(n.Block)
			c.emitJump(init)
			c.currentFn.Code[addr] |= uint64(len(c.currentFn.Code))
			c.cleanUpLoopScope(init, true)
		}
	case *ast.Break:
		c.breakJumps = append(c.breakJumps, len(c.currentFn.Code))
		c.breakCount[len(c.breakCount)-1]++
		c.emitJump(0)
	case *ast.Continue:
		c.continueJumps = append(c.continueJumps, len(c.currentFn.Code))
		c.continueCount[len(c.continueCount)-1]++
		c.emitJump(0)
	case *ast.ReferenceStmt:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		c.fromRefStmt = true
		i, s := c.refScope(n.Value)
		switch s {
		case rLoc:
			c.emitLoad(i, c.rAlloc, loadFromLocal)
		case rGlob:
			c.emitLoad(i, c.rAlloc, loadFromGlobal)
		case rFree:
			c.emitLoad(i, c.rAlloc, loadFromFree)
		case rNotDefined:
			c.generateReferenceError(n.Value, n.Line)
		}
		c.rAlloc++
	case *ast.IGetStmt:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		i := c.rAlloc
		if c.fromRefStmt {
			i -= 1
		} else {
			c.rAlloc++
		}
		j, t := c.compileExpr(n.Index, true)
		switch t {
		case rKonst:
			c.emitGet(i, j, i, storeFromKonst, storeFromLocal)
		case rLoc:
			c.emitGet(i, j, i, storeFromLocal, storeFromLocal)
		case rGlob:
			c.emitGet(i, j, i, storeFromGlobal, storeFromLocal)
		case rFree:
			c.emitGet(i, j, i, storeFromFree, storeFromLocal)
		}
		if !c.fromRefStmt {
			c.rAlloc--
		}
	case *ast.SelectStmt:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		i := c.rAlloc
		if c.fromRefStmt {
			i -= 1
		} else {
			c.rAlloc++
		}
		j, t := c.compileExpr(n.Selector, true)
		switch t {
		case rKonst:
			c.emitGet(i, j, i, storeFromKonst, storeFromLocal)
		case rLoc:
			c.emitGet(i, j, i, storeFromLocal, storeFromLocal)
		case rGlob:
			c.emitGet(i, j, i, storeFromGlobal, storeFromLocal)
		case rFree:
			c.emitGet(i, j, i, storeFromFree, storeFromLocal)
		}
		if !c.fromRefStmt {
			c.rAlloc--
		}
	case *ast.ISet:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		i := c.rAlloc
		if c.fromRefStmt {
			i -= 1
		} else {
			c.rAlloc++
		}
		j, t := c.compileExpr(n.Index, true)
		switch t {
		case rLoc:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitSet(i, j, k, storeFromLocal, storeFromLocal)
			case rKonst:
				c.emitSet(i, j, k, storeFromLocal, storeFromKonst)
			case rGlob:
				c.emitSet(i, j, k, storeFromLocal, storeFromGlobal)
			case rFree:
				c.emitSet(i, j, k, storeFromLocal, storeFromFree)
			}
			c.rAlloc--
		case rGlob:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitSet(i, j, k, storeFromGlobal, storeFromLocal)
			case rKonst:
				c.emitSet(i, j, k, storeFromGlobal, storeFromKonst)
			case rGlob:
				c.emitSet(i, j, k, storeFromGlobal, storeFromGlobal)
			case rFree:
				c.emitSet(i, j, k, storeFromGlobal, storeFromFree)
			}
			c.rAlloc--
		case rKonst:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitSet(i, j, k, storeFromKonst, storeFromLocal)
			case rKonst:
				c.emitSet(i, j, k, storeFromKonst, storeFromKonst)
			case rGlob:
				c.emitSet(i, j, k, storeFromKonst, storeFromGlobal)
			case rFree:
				c.emitSet(i, j, k, storeFromKonst, storeFromFree)
			}
			c.rAlloc--
		case rFree:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitSet(i, j, k, storeFromFree, storeFromLocal)
			case rKonst:
				c.emitSet(i, j, k, storeFromFree, storeFromKonst)
			case rGlob:
				c.emitSet(i, j, k, storeFromFree, storeFromGlobal)
			case rFree:
				c.emitSet(i, j, k, storeFromFree, storeFromGlobal)
			}
			c.rAlloc--
		}
		c.rAlloc--
		c.fromRefStmt = false
	case *ast.Block:
		c.scope++
		for i := range len(n.Statement) {
			c.compileStmt(n.Statement[i])
		}
		locals := c.sb.clearLocals(c.level, c.scope)
		c.rAlloc -= locals
		c.scope--
	case *ast.Ret:
		if c.level != 0 || c.isSubcompiler {
			i, s := c.compileExpr(n.Expr, true)
			switch s {
			case rLoc:
				c.emitRet(storeFromLocal, i)
			case rGlob:
				c.emitRet(storeFromGlobal, i)
			case rKonst:
				c.emitRet(storeFromKonst, i)
			case rFree:
				c.emitRet(storeFromFree, i)
			}
		}
	case *ast.CallStmt:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		callable := c.rAlloc
		if c.fromRefStmt {
			callable -= 1
		} else {
			c.rAlloc++
		}
		for _, v := range n.Args {
			i, s := c.compileExpr(v, true)
			c.exprToReg(i, s)
			c.rAlloc++
		}
		c.rAlloc = callable
		c.fromRefStmt = false
		c.emitCall(callable, len(n.Args), n.Ellipsis, 1)
	case *ast.MethodCallStmt:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		o := c.rAlloc
		if c.fromRefStmt {
			o -= 1
		} else {
			c.rAlloc++
		}
		c.emitLoad(o, c.rAlloc, loadFromLocal)
		c.rAlloc++
		j, _ := c.compileExpr(n.Prop, true)
		c.emitSend(o, j, o, storeFromKonst, storeFromLocal)
		for _, v := range n.Args {
			i, s := c.compileExpr(v, true)
			c.exprToReg(i, s)
			c.rAlloc++
		}
		c.rAlloc = o
		c.fromRefStmt = false
		c.emitCall(o, len(n.Args)+1, n.Ellipsis, 2)
	case *ast.StaticCallStmt:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		callable := c.rAlloc
		if c.fromRefStmt {
			callable -= 1
		} else {
			c.rAlloc++
		}
		j, t := c.compileExpr(n.Prop, true)
		switch t {
		case rKonst:
			c.emitGet(callable, j, callable, storeFromKonst, storeFromLocal)
		case rLoc:
			c.emitGet(callable, j, callable, storeFromLocal, storeFromLocal)
		case rGlob:
			c.emitGet(callable, j, callable, storeFromGlobal, storeFromLocal)
		case rFree:
			c.emitGet(callable, j, callable, storeFromFree, storeFromLocal)
		}
		if !c.fromRefStmt {
			c.rAlloc--
		}
		for _, v := range n.Args {
			i, s := c.compileExpr(v, true)
			c.exprToReg(i, s)
			c.rAlloc++
		}
		c.rAlloc = callable
		c.fromRefStmt = false
		c.emitCall(callable, len(n.Args), n.Ellipsis, 1)
	case *ast.Export:
		c.errorInfo[c.currentFn.ScriptID][len(c.currentFn.Code)] = n.Line
		if c.isSubcompiler {
			i, s := c.compileExpr(n.Expr, true)
			switch s {
			case rLoc:
				c.emitRet(storeFromLocal, i)
			case rGlob:
				c.emitRet(storeFromGlobal, i)
			case rKonst:
				c.emitRet(storeFromKonst, i)
			case rFree:
				c.emitRet(storeFromFree, i)
			}
		}
	}
}

func (c *compiler) compileExpr(node ast.Node, isRoot bool) (int, int) {
	switch n := node.(type) {
	case *ast.Integer:
		return c.kb.IntegerIndex(n.Value), rKonst
	case *ast.Float:
		return c.kb.FloatIndex(n.Value), rKonst
	case *ast.String:
		return c.kb.StringIndex(n.Value), rKonst
	case *ast.BinaryExpr:
		switch n.Op {
		case token.EQ, token.NEQ:
			return c.compileBinaryEq(n, isRoot)
		default:
			return c.compileBinaryExpr(n, isRoot)
		}
	case *ast.PrefixExpr:
		from, scope := c.compileExpr(n.Expr, false)
		if scope == rKonst {
			if val, err := (*c.kb.Konstants)[from].Prefix(uint64(n.Op)); err == nil {
				return c.integrateKonst(val)
			} else {
				c.hadError = true
				c.errMsg = "cannot perform prefix operation"
				c.lineErr = n.Line
			}
		}
		switch scope {
		case rGlob:
			c.emitLoad(from, c.rAlloc, loadFromGlobal)
			c.emitPrefix(c.rAlloc, c.rAlloc, n.Op)
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitPrefix(from, c.rDest, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitPrefix(from, c.rAlloc, n.Op)
			}
		case rKonst:
			c.emitLoad(from, c.rAlloc, loadFromKonst)
			c.emitPrefix(c.rAlloc, c.rAlloc, n.Op)
		case rFree:
			c.emitLoad(from, c.rAlloc, loadFromFree)
			c.emitPrefix(c.rAlloc, c.rAlloc, n.Op)
		}
		return c.rAlloc, rLoc
	case *ast.Boolean:
		return c.kb.BooleanIndex(n.Value), rKonst
	case *ast.Nil:
		return c.kb.NilIndex(), rKonst
	case *ast.Reference:
		i, s := c.refScope(n.Value)
		if s == rNotDefined {
			c.generateReferenceError(n.Value, n.Line)
		}
		return i, s
	case *ast.Array:
		var count int
		for _, v := range n.ExprList {
			i, s := c.compileExpr(v, false)
			switch s {
			case rLoc:
				if i != c.rAlloc {
					c.emitLoad(i, c.rAlloc, loadFromLocal)
				}
			case rKonst:
				c.emitLoad(i, c.rAlloc, loadFromKonst)
			case rGlob:
				c.emitLoad(i, c.rAlloc, loadFromGlobal)
			case rFree:
				c.emitLoad(i, c.rAlloc, loadFromFree)
			}
			c.rAlloc++
			count++
		}
		c.rAlloc -= count
		if c.mutLoc && isRoot {
			c.emitArray(count, c.rAlloc, c.rDest)
			return c.rDest, rLoc
		}
		c.emitArray(count, c.rAlloc, c.rAlloc)
		return c.rAlloc, rLoc
	case *ast.Object:
		o := c.rAlloc
		if c.mutLoc && isRoot {
			o = c.rDest
		}
		c.emitObject(o)
		for _, v := range n.Pairs {
			k, _ := c.compileExpr(v.Key, false)
			c.rAlloc++
			v, sv := c.compileExpr(v.Value, false)
			switch sv {
			case rKonst:
				c.emitSet(o, k, v, storeFromKonst, storeFromKonst)
			case rLoc:
				c.emitSet(o, k, v, storeFromKonst, storeFromLocal)
			case rGlob:
				c.emitSet(o, k, v, storeFromKonst, storeFromGlobal)
			case rFree:
				c.emitSet(o, k, v, storeFromKonst, storeFromFree)
			}
			c.rAlloc--
		}
		return o, rLoc
	case *ast.Property:
		return c.kb.StringIndex(n.Value), rKonst
	case *ast.ForState:
		return c.kb.IntegerIndex(0), rKonst
	case *ast.IGet:
		i, s := c.compileExpr(n.Indexable, false)
		dest := c.rAlloc
		switch s {
		case rLoc:
			c.rAlloc++
			j, t := c.compileExpr(n.Index, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromLocal, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromLocal, storeFromLocal)
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromGlobal, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromGlobal, storeFromLocal)
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromKonst, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromKonst, storeFromLocal)
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromFree, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromFree, storeFromLocal)
					c.rAlloc--
				}
			}
		case rGlob:
			c.rAlloc++
			j, t := c.compileExpr(n.Index, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromLocal, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromLocal, storeFromGlobal)
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromGlobal, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromGlobal, storeFromGlobal)
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromKonst, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromKonst, storeFromGlobal)
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromFree, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromFree, storeFromGlobal)
					c.rAlloc--
				}
			}
		case rFree:
			c.rAlloc++
			j, t := c.compileExpr(n.Index, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromLocal, storeFromFree)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromLocal, storeFromFree)
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromGlobal, storeFromFree)
					c.rAlloc--
				} else {
					c.emitGet(i, j, dest, storeFromGlobal, storeFromFree)
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromKonst, storeFromFree)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromKonst, storeFromFree)
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromFree, storeFromFree)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromFree, storeFromFree)
					c.rAlloc--
				}
			}
		}
		return dest, rLoc
	case *ast.Select:
		i, s := c.compileExpr(n.Selectable, false)
		dest := c.rAlloc
		switch s {
		case rLoc:
			c.rAlloc++
			j, t := c.compileExpr(n.Selector, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromLocal, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromLocal, storeFromLocal)
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromGlobal, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromGlobal, storeFromLocal)
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromKonst, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromKonst, storeFromLocal)
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromFree, storeFromLocal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromFree, storeFromLocal)
					c.rAlloc--
				}
			}
		case rGlob:
			c.rAlloc++
			j, t := c.compileExpr(n.Selector, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromLocal, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromLocal, storeFromGlobal)
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromGlobal, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromGlobal, storeFromGlobal)
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromKonst, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromKonst, storeFromGlobal)
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromFree, storeFromGlobal)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromFree, storeFromGlobal)
					c.rAlloc--
				}
			}
		case rFree:
			c.rAlloc++
			j, t := c.compileExpr(n.Selector, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromLocal, storeFromFree)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromLocal, storeFromFree)
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromGlobal, storeFromFree)
					c.rAlloc--
				} else {
					c.emitGet(i, j, dest, storeFromGlobal, storeFromFree)
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromKonst, storeFromFree)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromKonst, storeFromFree)
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitGet(i, j, c.rDest, storeFromFree, storeFromFree)
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitGet(i, j, dest, storeFromFree, storeFromFree)
					c.rAlloc--
				}
			}
		}
		return dest, rLoc
	case *ast.Slice:
		v, s := c.compileExpr(n.Value, false)
		c.exprToReg(v, s)
		switch n.Mode {
		case vcv:
			if c.mutLoc && isRoot {
				c.emitSlice(n.Mode, c.rAlloc, c.rDest)
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
			}
		case vce:
			c.rAlloc++
			v, s := c.compileExpr(n.Last, false)
			c.exprToReg(v, s)
			c.rAlloc--
			if c.mutLoc && isRoot {
				c.emitSlice(n.Mode, c.rAlloc, c.rDest)
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
			}
		case ecv:
			c.rAlloc++
			v, s := c.compileExpr(n.First, false)
			c.exprToReg(v, s)
			c.rAlloc--
			if c.mutLoc && isRoot {
				c.emitSlice(n.Mode, c.rAlloc, c.rDest)
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
			}
		case ece:
			c.rAlloc++
			f, sf := c.compileExpr(n.First, false)
			c.exprToReg(f, sf)
			c.rAlloc++
			l, sl := c.compileExpr(n.Last, false)
			c.exprToReg(l, sl)
			c.rAlloc -= 2
			if c.mutLoc && isRoot {
				c.emitSlice(n.Mode, c.rAlloc, c.rDest)
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
			}
		}
		return c.rAlloc, rLoc
	case *ast.Fun:
		fn := &CoreFunction{ScriptID: c.script.MainFunction.CoreFn.ScriptID}
		c.fn = append(c.fn, fn)
		c.emitFun(c.kb.FunctionIndex(fn), c.rAlloc)
		c.currentFn = fn
		reg := c.startFuncScope()
		for _, v := range n.Args {
			fn.Arity++
			c.sb.addLocal(v, c.level, c.scope, c.rAlloc)
			c.rAlloc++
		}
		if n.IsVar {
			fn.IsVarArg = true
			fn.Arity--
		}
		c.compileStmt(n.Body)
		c.leaveFuncScope()
		c.rAlloc = reg
		return c.rAlloc, rLoc
	case *ast.CallExpr:
		reg := c.rAlloc
		idx, s := c.compileExpr(n.Fun, false)
		c.exprToReg(idx, s)
		for _, v := range n.Args {
			c.rAlloc++
			i, s := c.compileExpr(v, false)
			c.exprToReg(i, s)
		}
		c.rAlloc = reg
		c.emitCall(reg, len(n.Args), n.Ellipsis, 1)
		return reg, rLoc
	case *ast.MethodCallExpr:
		o := c.rAlloc
		c.rAlloc++
		i, s := c.compileExpr(n.Obj, false)
		c.exprToReg(i, s)
		i = c.rAlloc
		c.rAlloc++
		j, _ := c.compileExpr(n.Prop, false)
		c.emitSend(i, j, o, storeFromKonst, storeFromLocal)
		for _, v := range n.Args {
			i, s := c.compileExpr(v, false)
			c.exprToReg(i, s)
			c.rAlloc++
		}
		c.rAlloc = o
		c.emitCall(o, len(n.Args)+1, n.Ellipsis, 2)
		return o, rLoc
	case *ast.StaticCallExpr:
		callable := c.rAlloc
		j, t := c.compileExpr(n.Prop, true)
		switch t {
		case rKonst:
			c.emitGet(callable, j, callable, storeFromKonst, storeFromLocal)
		case rLoc:
			c.emitGet(callable, j, callable, storeFromLocal, storeFromLocal)
		case rGlob:
			c.emitGet(callable, j, callable, storeFromGlobal, storeFromLocal)
		case rFree:
			c.emitGet(callable, j, callable, storeFromFree, storeFromLocal)
		}
		for _, v := range n.Args {
			c.rAlloc++
			i, s := c.compileExpr(v, false)
			c.exprToReg(i, s)
		}
		c.rAlloc = callable
		c.emitCall(callable, len(n.Args), n.Ellipsis, 1)
		return callable, rLoc
	case *ast.Import:
		var importFilePath string
		if filepath.IsAbs(n.Path) {
			importFilePath = n.Path
		} else {
			importFilePath = filepath.Join(filepath.Dir(c.mainScriptID), n.Path)
		}
		if _, isCycle := c.depMap[importFilePath]; isCycle {
			c.hadError = true
			c.errMsg = "import cycle detected"
			c.lineErr = n.Line
			return 0, rGlob
		} else {
			c.depMap[importFilePath] = dummy
		}
		if v, isPresent := c.scriptMap[importFilePath]; isPresent {
			delete(c.depMap, importFilePath)
			c.emitFun(v, c.rAlloc)
			c.emitCall(c.rAlloc, 0, 0, 1)
			return c.rAlloc, rLoc
		}
		src, err := LoadScriptFromFile(importFilePath)
		if err != nil {
			c.hadError = true
			c.errMsg = err.Error()
			c.lineErr = n.Line
			return 0, rGlob
		}
		p := newParser(src, importFilePath)
		scriptAST, err := p.parse()
		if err != nil {
			c.hadError = true
			c.errMsg = err.Error()
			c.lineErr = n.Line
			return 0, rGlob
		}
		subCompiler := newSubCompiler(scriptAST, importFilePath, c.kb, c.script.GlobalStore, c.scriptMap, c.depMap, c.errorInfo, len(*c.script.GlobalStore), c.extensionsLoader)
		m, err := subCompiler.compileSubScript()
		c.sb.index = len(*c.script.GlobalStore)
		if err != nil {
			c.hadError = true
			c.errMsg = err.Error()
			c.lineErr = n.Line
			return 0, rGlob
		}
		fnIndex := c.kb.FunctionIndex(m.MainFunction.CoreFn)
		c.scriptMap[importFilePath] = fnIndex
		delete(c.depMap, importFilePath)
		c.emitFun(fnIndex, c.rAlloc)
		c.emitCall(c.rAlloc, 0, 0, 1)
		return c.rAlloc, rLoc
	case *ast.Enum:
		e := &Enum{Pairs: map[string]Integer{}}
		for i, v := range n.Variants {
			e.Pairs[v] = Integer(i)
		}
		return c.kb.EnumIndex(e), rKonst
	default:
		return 0, rGlob
	}
}

func (c *compiler) compileConditional(n *ast.If, shouldJumpOutside bool) {
	idx, scope := c.compileExpr(n.Condition, false)
	if scope == rKonst {
		switch v := (*c.kb.Konstants)[idx].(type) {
		case NilValue:
			c.skipBlock(n.Block)
			return
		case Bool:
			if !v {
				c.skipBlock(n.Block)
				return
			}
		}
		c.compileBlockAndCheckJump(n.Block, shouldJumpOutside)
	} else {
		c.exprToReg(idx, scope)
		addr := len(c.currentFn.Code)
		c.emitCheck(0, c.rAlloc, 0)
		c.compileBlockAndCheckJump(n.Block, shouldJumpOutside)
		c.currentFn.Code[addr] |= uint64(len(c.currentFn.Code))
	}
}

func (c *compiler) skipBlock(block ast.Node) {
	addr := len(c.currentFn.Code)
	c.emitJump(0)
	c.compileStmt(block)
	c.currentFn.Code[addr] |= uint64(len(c.currentFn.Code))
}

func (c *compiler) compileBlockAndCheckJump(block ast.Node, shouldJumpOutside bool) {
	c.compileStmt(block)
	if shouldJumpOutside {
		c.jumps = append(c.jumps, len(c.currentFn.Code))
		c.emitJump(0)
	}
}

func (c *compiler) cleanUpLoopScope(init int, isWhileLoop bool) {
	hasBreaks := len(c.breakJumps)
	lastElem := len(c.breakCount) - 1
	count := c.breakCount[lastElem]
	if hasBreaks > 0 {
		for i := 1; i <= count; i++ {
			c.currentFn.Code[c.breakJumps[hasBreaks-i]] |= uint64(len(c.currentFn.Code))
		}
		c.breakJumps = c.breakJumps[:hasBreaks-count]
	}
	c.breakCount = c.breakCount[:lastElem]
	hasContinues := len(c.continueJumps)
	lastElem = len(c.continueCount) - 1
	count = c.continueCount[lastElem]
	if hasContinues > 0 {
		for i := 1; i <= count; i++ {
			if isWhileLoop {
				c.currentFn.Code[c.continueJumps[hasContinues-i]] |= uint64(init)
			} else {
				c.currentFn.Code[c.continueJumps[hasContinues-i]] |= uint64(len(c.currentFn.Code) - 1)
			}
		}
		c.continueJumps = c.continueJumps[:hasContinues-count]
	}
	c.continueCount = c.continueCount[:lastElem]
}

func (c *compiler) startLoopScope() {
	c.breakCount = append(c.breakCount, 0)
	c.continueCount = append(c.continueCount, 0)
}

func (c *compiler) startFuncScope() int {
	r := c.rAlloc
	c.rAlloc = 0
	c.level++
	return r
}

func (c *compiler) leaveFuncScope() {
	c.sb.clearLocals(c.level, c.scope)
	c.fn = c.fn[:c.level]
	c.level--
	c.currentFn = c.fn[c.level]
}

func (c *compiler) integrateKonst(val Value) (int, int) {
	switch e := val.(type) {
	case Integer:
		return c.kb.IntegerIndex(int64(e)), rKonst
	case Float:
		return c.kb.FloatIndex(float64(e)), rKonst
	case Bool:
		return c.kb.BooleanIndex(bool(e)), rKonst
	case *String:
		return c.kb.StringIndex(e.Value), rKonst
	default:
		return c.kb.NilIndex(), rKonst
	}
}

func (c *compiler) exprToReg(i, s int) {
	switch s {
	case rLoc:
		if i != c.rAlloc {
			c.emitLoad(i, c.rAlloc, loadFromLocal)
		}
	case rGlob:
		c.emitLoad(i, c.rAlloc, loadFromGlobal)
	case rKonst:
		c.emitLoad(i, c.rAlloc, loadFromKonst)
	case rFree:
		c.emitLoad(i, c.rAlloc, loadFromFree)
	}
}

func (c *compiler) compileBinaryExpr(n *ast.BinaryExpr, isRoot bool) (int, int) {
	lidx, lscope := c.compileExpr(n.Lhs, false)
	lreg := c.rAlloc
	switch lscope {
	case rKonst:
		ridx, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rKonst:
			if val, err := (*c.kb.Konstants)[lidx].Binop(c.kb.ctx, uint64(n.Op), (*c.kb.Konstants)[ridx]); err == nil {
				return c.integrateKonst(val)
			} else {
				c.hadError = true
				c.errMsg = "cannot perform binary operation"
				c.lineErr = n.Line
			}
		case rGlob:
			c.emitLoad(ridx, lreg, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinopQ(lidx, lreg, c.rDest, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitBinopQ(lidx, lreg, lreg, n.Op)
			}
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitBinopQ(lidx, ridx, c.rDest, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitBinopQ(lidx, ridx, lreg, n.Op)
			}
		case rFree:
			c.emitLoad(ridx, lreg, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinopQ(lidx, lreg, c.rDest, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitBinopQ(lidx, lreg, lreg, n.Op)
			}
		}
	case rGlob:
		ridx, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rGlob:
			if c.mutLoc && isRoot {
				c.emitBinopG(lidx, ridx, c.rDest, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitBinopG(lidx, ridx, lreg, n.Op)
			}
		case rKonst:
			c.emitLoad(lidx, lreg, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinopK(ridx, lreg, c.rDest, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitBinopK(ridx, lreg, lreg, n.Op)
			}
		case rLoc:
			c.emitLoad(ridx, lreg, loadFromLocal)
			c.rAlloc++
			c.emitLoad(lidx, c.rAlloc, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinop(c.rAlloc, lreg, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(c.rAlloc, lreg, lreg, n.Op)
				c.rAlloc--
			}
		case rFree:
			c.rAlloc++
			c.emitLoad(lidx, lreg, loadFromGlobal)
			c.emitLoad(ridx, c.rAlloc, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, c.rAlloc, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, c.rAlloc, lreg, n.Op)
				c.rAlloc--
			}
		}
	case rLoc:
		c.rAlloc++
		ridx, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitBinop(lidx, ridx, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lidx, ridx, lreg, n.Op)
				c.rAlloc--
			}
		case rGlob:
			c.emitLoad(ridx, c.rAlloc, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinop(lidx, c.rAlloc, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lidx, c.rAlloc, lreg, n.Op)
				c.rAlloc--
			}
		case rKonst:
			if c.mutLoc && isRoot {
				c.emitBinopK(ridx, lidx, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinopK(ridx, lidx, lreg, n.Op)
				c.rAlloc--
			}
		case rFree:
			c.emitLoad(ridx, c.rAlloc, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lidx, c.rAlloc, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lidx, c.rAlloc, lreg, n.Op)
				c.rAlloc--
			}
		}
	case rFree:
		c.rAlloc++
		ridx, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rLoc:
			c.emitLoad(lidx, lreg, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, ridx, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, ridx, lreg, n.Op)
				c.rAlloc--
			}
		case rGlob:
			c.emitLoad(lidx, lreg, loadFromFree)
			c.emitLoad(ridx, c.rAlloc, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, c.rAlloc, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, c.rAlloc, lreg, n.Op)
				c.rAlloc--
			}
		case rKonst:
			c.emitLoad(lidx, lreg, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinopK(ridx, lreg, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinopK(ridx, lreg, lreg, n.Op)
				c.rAlloc--
			}
		case rFree:
			c.emitLoad(lidx, lreg, loadFromFree)
			c.emitLoad(ridx, c.rAlloc, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, c.rAlloc, c.rDest, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, c.rAlloc, lreg, n.Op)
				c.rAlloc--
			}
		}
	}
	return lreg, rLoc
}

func (c *compiler) compileBinaryEq(n *ast.BinaryExpr, isRoot bool) (int, int) {
	i, iscope := c.compileExpr(n.Lhs, false)
	k := c.rAlloc
	switch iscope {
	case rKonst:
		j, jscope := c.compileExpr(n.Rhs, false)
		switch jscope {
		case rKonst:
			val := (*c.kb.Konstants)[i].Equals((*c.kb.Konstants)[j])
			if n.Op == token.NEQ {
				val = !val
			}
			return c.integrateKonst(val)
		case rGlob:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromKonst, loadFromGlobal, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromKonst, loadFromGlobal, n.Op)
			}
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromKonst, loadFromLocal, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromKonst, loadFromLocal, n.Op)
			}
		case rFree:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromKonst, loadFromFree, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromKonst, loadFromFree, n.Op)
			}
		}
	case rGlob:
		j, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rGlob:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromGlobal, loadFromGlobal, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromGlobal, loadFromGlobal, n.Op)
			}
		case rKonst:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, storeFromGlobal, storeFromKonst, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, storeFromGlobal, storeFromKonst, n.Op)
			}
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromGlobal, loadFromLocal, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, storeFromGlobal, storeFromLocal, n.Op)
			}
		case rFree:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromGlobal, loadFromFree, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromGlobal, loadFromFree, n.Op)
			}
		}
	case rLoc:
		c.rAlloc++
		j, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, storeFromLocal, storeFromLocal, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, storeFromLocal, storeFromLocal, n.Op)
				c.rAlloc--
			}
		case rGlob:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, storeFromLocal, storeFromGlobal, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, storeFromLocal, storeFromGlobal, n.Op)
				c.rAlloc--
			}
		case rKonst:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromLocal, loadFromKonst, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromLocal, loadFromKonst, n.Op)
				c.rAlloc--
			}
		case rFree:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromLocal, loadFromFree, n.Op)
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromLocal, loadFromFree, n.Op)
				c.rAlloc--
			}
		}
	case rFree:
		j, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromFree, loadFromLocal, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromFree, loadFromLocal, n.Op)
			}
		case rGlob:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, storeFromFree, storeFromGlobal, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, storeFromFree, storeFromGlobal, n.Op)
			}
		case rKonst:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, storeFromFree, storeFromKonst, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, storeFromFree, storeFromKonst, n.Op)
			}
		case rFree:
			if c.mutLoc && isRoot {
				c.emitSuperEq(i, j, c.rDest, loadFromFree, loadFromFree, n.Op)
				return c.rDest, rLoc
			} else {
				c.emitSuperEq(i, j, k, loadFromFree, loadFromFree, n.Op)
			}
		}
	}
	return k, rLoc
}
