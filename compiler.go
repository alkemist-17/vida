package vida

import (
	"github.com/alkemist-17/vida/ast"
	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type compiler struct {
	jumps         []int
	breakJumps    []int
	breakCount    []int
	continueJumps []int
	continueCount []int
	errMsg        string
	fn            []*CoreFunction
	currentFn     *CoreFunction
	ast           *ast.Ast
	script        *Script
	kb            *konstBuilder
	sb            *symbolBuilder
	scriptMap     map[string]int
	depMap        map[string]struct{}
	errorInfo     ErrorInfo
	lineErr       uint
	scope         int
	level         int
	rAlloc        int
	rDest         int
	fromRefStmt   bool
	mutLoc        bool
	hadError      bool
	isSubcompiler bool
}

var dummy = struct{}{}

func newMainCompiler(ast *ast.Ast, scriptName string) *compiler {
	dm := make(map[string]struct{})
	dm[scriptName] = dummy
	ei := make(ErrorInfo)
	ei[scriptName] = make(map[int]uint)
	c := &compiler{
		ast:       ast,
		script:    newMainScript(scriptName),
		kb:        newKonstBuilder(),
		sb:        newSymbolBuilder(0),
		scriptMap: make(map[string]int),
		depMap:    dm,
		errorInfo: ei,
	}
	c.fn = append(c.fn, c.script.MainFunction.CoreFn)
	c.currentFn = c.script.MainFunction.CoreFn
	return c
}

func newSubCompiler(ast *ast.Ast, scriptName string, kb *konstBuilder, store *[]Value, scriptMap map[string]int, depMap map[string]struct{}, ei ErrorInfo, initialIndex int) *compiler {
	ei[scriptName] = make(map[int]uint)
	c := &compiler{
		ast:           ast,
		script:        newScript(scriptName, store),
		kb:            kb,
		sb:            newSymbolBuilder(initialIndex),
		isSubcompiler: true,
		scriptMap:     scriptMap,
		depMap:        depMap,
		errorInfo:     ei,
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
			return nil, verror.New(c.script.MainFunction.CoreFn.ScriptName, c.errMsg, verror.CompilationErrType, c.lineErr)
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
			return nil, verror.New(c.script.MainFunction.CoreFn.ScriptName, c.errMsg, verror.CompilationErrType, c.lineErr)
		}
	}
	return c.script, nil
}

func (c *compiler) compileStmt(node ast.Node) {
	switch n := node.(type) {
	case *ast.Mut:
		to, sIdent := c.refScope(n.Indentifier)
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
			c.generateReferenceError(n.Indentifier, n.Line)
		}
	case *ast.Let:
		to, isPresent := c.sb.addGlobal(n.Indentifier)
		if !isPresent {
			*c.script.Store = append(*c.script.Store, NilValue)
		}
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
	case *ast.Var:
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
	case *ast.Branch:
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
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
		c.startLoopScope()
		init := len(c.currentFn.Code)
		idx, scope := c.compileExpr(n.Condition, true)
		if scope == rKonst {
			switch v := (*c.kb.Konstants)[idx].(type) {
			case Nil:
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
		i := c.rAlloc
		if c.fromRefStmt {
			i -= 1
		} else {
			c.rAlloc++
		}
		j, t := c.compileExpr(n.Index, true)
		switch t {
		case rKonst:
			c.emitIGet(i, j, i, storeFromKonst, storeFromLocal)
		case rLoc:
			c.emitIGet(i, j, i, storeFromLocal, storeFromLocal)
		case rGlob:
			c.emitIGet(i, j, i, storeFromGlobal, storeFromLocal)
		case rFree:
			c.emitIGet(i, j, i, storeFromFree, storeFromLocal)
		}
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		if !c.fromRefStmt {
			c.rAlloc--
		}
	case *ast.SelectStmt:
		i := c.rAlloc
		if c.fromRefStmt {
			i -= 1
		} else {
			c.rAlloc++
		}
		j, t := c.compileExpr(n.Selector, true)
		switch t {
		case rKonst:
			c.emitIGet(i, j, i, storeFromKonst, storeFromLocal)
		case rLoc:
			c.emitIGet(i, j, i, storeFromLocal, storeFromLocal)
		case rGlob:
			c.emitIGet(i, j, i, storeFromGlobal, storeFromLocal)
		case rFree:
			c.emitIGet(i, j, i, storeFromFree, storeFromLocal)
		}
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		if !c.fromRefStmt {
			c.rAlloc--
		}
	case *ast.ISet:
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
				c.emitISet(i, j, k, storeFromLocal, storeFromLocal)
			case rKonst:
				c.emitISet(i, j, k, storeFromLocal, storeFromKonst)
			case rGlob:
				c.emitISet(i, j, k, storeFromLocal, storeFromGlobal)
			case rFree:
				c.emitISet(i, j, k, storeFromLocal, storeFromFree)
			}
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			c.rAlloc--
		case rGlob:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitISet(i, j, k, storeFromGlobal, storeFromLocal)
			case rKonst:
				c.emitISet(i, j, k, storeFromGlobal, storeFromKonst)
			case rGlob:
				c.emitISet(i, j, k, storeFromGlobal, storeFromGlobal)
			case rFree:
				c.emitISet(i, j, k, storeFromGlobal, storeFromFree)
			}
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			c.rAlloc--
		case rKonst:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitISet(i, j, k, storeFromKonst, storeFromLocal)
			case rKonst:
				c.emitISet(i, j, k, storeFromKonst, storeFromKonst)
			case rGlob:
				c.emitISet(i, j, k, storeFromKonst, storeFromGlobal)
			case rFree:
				c.emitISet(i, j, k, storeFromKonst, storeFromFree)
			}
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			c.rAlloc--
		case rFree:
			c.rAlloc++
			k, u := c.compileExpr(n.Expr, true)
			switch u {
			case rLoc:
				c.emitISet(i, j, k, storeFromFree, storeFromLocal)
			case rKonst:
				c.emitISet(i, j, k, storeFromFree, storeFromKonst)
			case rGlob:
				c.emitISet(i, j, k, storeFromFree, storeFromGlobal)
			case rFree:
				c.emitISet(i, j, k, storeFromFree, storeFromGlobal)
			}
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
	case *ast.MethodCallStmt:
		o := c.rAlloc
		if c.fromRefStmt {
			o -= 1
		} else {
			c.rAlloc++
		}
		c.emitLoad(o, c.rAlloc, loadFromLocal)
		c.rAlloc++
		j, _ := c.compileExpr(n.Prop, true)
		c.emitIGet(o, j, o, storeFromKonst, storeFromLocal)
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		for _, v := range n.Args {
			i, s := c.compileExpr(v, true)
			c.exprToReg(i, s)
			c.rAlloc++
		}
		c.rAlloc = o
		c.fromRefStmt = false
		c.emitCall(o, len(n.Args)+1, n.Ellipsis, 2)
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
	case *ast.Export:
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
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitPrefix(from, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitPrefix(from, c.rAlloc, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case rKonst:
			c.emitLoad(from, c.rAlloc, loadFromKonst)
			c.emitPrefix(c.rAlloc, c.rAlloc, n.Op)
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		case rFree:
			c.emitLoad(from, c.rAlloc, loadFromFree)
			c.emitPrefix(c.rAlloc, c.rAlloc, n.Op)
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
				c.emitISet(o, k, v, storeFromKonst, storeFromKonst)
			case rLoc:
				c.emitISet(o, k, v, storeFromKonst, storeFromLocal)
			case rGlob:
				c.emitISet(o, k, v, storeFromKonst, storeFromGlobal)
			case rFree:
				c.emitISet(o, k, v, storeFromKonst, storeFromFree)
			}
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
					c.emitIGet(i, j, c.rDest, storeFromLocal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromLocal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromGlobal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromGlobal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromKonst, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromKonst, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromFree, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromFree, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			}
		case rGlob:
			c.rAlloc++
			j, t := c.compileExpr(n.Index, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromLocal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromLocal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromGlobal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromGlobal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromKonst, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromKonst, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromFree, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromFree, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			}
		case rFree:
			c.rAlloc++
			j, t := c.compileExpr(n.Index, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromLocal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromLocal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromGlobal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				} else {
					c.emitIGet(i, j, dest, storeFromGlobal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromKonst, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromKonst, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromFree, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromFree, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
					c.emitIGet(i, j, c.rDest, storeFromLocal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromLocal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromGlobal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromGlobal, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromKonst, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromKonst, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromFree, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromFree, storeFromLocal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			}
		case rGlob:
			c.rAlloc++
			j, t := c.compileExpr(n.Selector, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromLocal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromLocal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromGlobal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromGlobal, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromKonst, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromKonst, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromFree, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromFree, storeFromGlobal)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			}
		case rFree:
			c.rAlloc++
			j, t := c.compileExpr(n.Selector, false)
			switch t {
			case rLoc:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromLocal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromLocal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rGlob:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromGlobal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				} else {
					c.emitIGet(i, j, dest, storeFromGlobal, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rKonst:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromKonst, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromKonst, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
				}
			case rFree:
				if c.mutLoc && isRoot {
					c.emitIGet(i, j, c.rDest, storeFromFree, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
					c.rAlloc--
					return c.rDest, rLoc
				} else {
					c.emitIGet(i, j, dest, storeFromFree, storeFromFree)
					c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case vce:
			c.rAlloc++
			v, s := c.compileExpr(n.Last, false)
			c.exprToReg(v, s)
			c.rAlloc--
			if c.mutLoc && isRoot {
				c.emitSlice(n.Mode, c.rAlloc, c.rDest)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case ecv:
			c.rAlloc++
			v, s := c.compileExpr(n.First, false)
			c.exprToReg(v, s)
			c.rAlloc--
			if c.mutLoc && isRoot {
				c.emitSlice(n.Mode, c.rAlloc, c.rDest)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitSlice(n.Mode, c.rAlloc, c.rAlloc)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		}
		return c.rAlloc, rLoc
	case *ast.Fun:
		fn := &CoreFunction{ScriptName: c.script.MainFunction.CoreFn.ScriptName}
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
			fn.IsVar = true
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
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		return reg, rLoc
	case *ast.MethodCallExpr:
		o := c.rAlloc
		c.rAlloc++
		i, s := c.compileExpr(n.Obj, false)
		c.exprToReg(i, s)
		i = c.rAlloc
		c.rAlloc++
		j, _ := c.compileExpr(n.Prop, false)
		c.emitIGet(i, j, o, storeFromKonst, storeFromLocal)
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		for _, v := range n.Args {
			i, s := c.compileExpr(v, false)
			c.exprToReg(i, s)
			c.rAlloc++
		}
		c.rAlloc = o
		c.emitCall(o, len(n.Args)+1, n.Ellipsis, 2)
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		return o, rLoc
	case *ast.Import:
		if _, isCycle := c.depMap[n.Path]; isCycle {
			c.hadError = true
			c.errMsg = "import cycle detected"
			c.lineErr = n.Line
			return 0, rGlob
		} else {
			c.depMap[n.Path] = dummy
		}
		if v, isPresent := c.scriptMap[n.Path]; isPresent {
			delete(c.depMap, n.Path)
			c.emitFun(v, c.rAlloc)
			c.emitCall(c.rAlloc, 0, 0, 1)
			c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			return c.rAlloc, rLoc
		}
		src, err := readScript(n.Path)
		if err != nil {
			c.hadError = true
			c.errMsg = err.Error()
			c.lineErr = n.Line
			return 0, rGlob
		}
		p := newParser(src, n.Path)
		scriptAST, err := p.parse()
		if err != nil {
			c.hadError = true
			c.errMsg = err.Error()
			c.lineErr = n.Line
			return 0, rGlob
		}
		subCompiler := newSubCompiler(scriptAST, n.Path, c.kb, c.script.Store, c.scriptMap, c.depMap, c.errorInfo, len(*c.script.Store))
		m, err := subCompiler.compileSubScript()
		c.sb.index = len(*c.script.Store)
		if err != nil {
			c.hadError = true
			c.errMsg = err.Error()
			c.lineErr = n.Line
			return 0, rGlob
		}
		fnIndex := c.kb.FunctionIndex(m.MainFunction.CoreFn)
		c.scriptMap[n.Path] = fnIndex
		delete(c.depMap, n.Path)
		c.emitFun(fnIndex, c.rAlloc)
		c.emitCall(c.rAlloc, 0, 0, 1)
		c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
		return c.rAlloc, rLoc
	case *ast.Enum:
		e := &Enum{Pairs: map[string]Integer{}}
		if n.HasInitVal {
			for _, v := range n.Variants {
				if v == "_" {
					n.Init++
					continue
				}
				e.Pairs[v] = Integer(n.Init)
				n.Init++
			}
			return c.kb.EnumIndex(e), rKonst
		}
		for i, v := range n.Variants {
			if v == "_" {
				continue
			}
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
		case Nil:
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
			if val, err := (*c.kb.Konstants)[lidx].Binop(uint64(n.Op), (*c.kb.Konstants)[ridx]); err == nil {
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
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitBinopQ(lidx, lreg, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case rLoc:
			if c.mutLoc && isRoot {
				c.emitBinopQ(lidx, ridx, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitBinopQ(lidx, ridx, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case rFree:
			c.emitLoad(ridx, lreg, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinopQ(lidx, lreg, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitBinopQ(lidx, lreg, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		}
	case rGlob:
		ridx, rscope := c.compileExpr(n.Rhs, false)
		switch rscope {
		case rGlob:
			if c.mutLoc && isRoot {
				c.emitBinopG(lidx, ridx, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitBinopG(lidx, ridx, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case rKonst:
			c.emitLoad(lidx, lreg, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinopK(ridx, lreg, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				return c.rDest, rLoc
			} else {
				c.emitBinopK(ridx, lreg, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
			}
		case rLoc:
			c.emitLoad(ridx, lreg, loadFromLocal)
			c.rAlloc++
			c.emitLoad(lidx, c.rAlloc, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinop(c.rAlloc, lreg, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(c.rAlloc, lreg, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rFree:
			c.rAlloc++
			c.emitLoad(lidx, lreg, loadFromGlobal)
			c.emitLoad(ridx, c.rAlloc, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, c.rAlloc, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, c.rAlloc, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lidx, ridx, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rGlob:
			c.emitLoad(ridx, c.rAlloc, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinop(lidx, c.rAlloc, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lidx, c.rAlloc, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rKonst:
			if c.mutLoc && isRoot {
				c.emitBinopK(ridx, lidx, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinopK(ridx, lidx, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rFree:
			c.emitLoad(ridx, c.rAlloc, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lidx, c.rAlloc, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lidx, c.rAlloc, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, ridx, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rGlob:
			c.emitLoad(lidx, lreg, loadFromFree)
			c.emitLoad(ridx, c.rAlloc, loadFromGlobal)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, c.rAlloc, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, c.rAlloc, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rKonst:
			c.emitLoad(lidx, lreg, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinopK(ridx, lreg, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinopK(ridx, lreg, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
			}
		case rFree:
			c.emitLoad(lidx, lreg, loadFromFree)
			c.emitLoad(ridx, c.rAlloc, loadFromFree)
			if c.mutLoc && isRoot {
				c.emitBinop(lreg, c.rAlloc, c.rDest, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
				c.rAlloc--
				return c.rDest, rLoc
			} else {
				c.emitBinop(lreg, c.rAlloc, lreg, n.Op)
				c.errorInfo[c.currentFn.ScriptName][len(c.currentFn.Code)] = n.Line
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
