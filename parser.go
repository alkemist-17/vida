package vida

import (
	"fmt"
	"strconv"

	"github.com/alkemist-17/vida/ast"
	"github.com/alkemist-17/vida/token"
)

const (
	spreadFirst = iota + 1
	spreadLast
)

const (
	sliceHasEnd = 1 << iota
	sliceInitialMode
	sliceHasStart
)

type parser struct {
	current token.TokenInfo
	next    token.TokenInfo
	lexer   *Lexer
	ast     *ast.Ast
	err     *VidaRunTimeError
	ok      bool
}

func newParser(src []byte, scriptID string) *parser {
	p := &parser{
		lexer: NewLexer(src, scriptID),
		ok:    true,
		ast:   &ast.Ast{},
	}
	p.advance()
	p.advance()
	return p
}

func (p *parser) parse() (*ast.Ast, error) {
	for p.ok {
		switch p.current.Token {
		case token.IDENTIFIER:
			p.ast.Statement = append(p.ast.Statement, p.mutationOrCall(&p.ast.Statement))
		case token.LET:
			p.ast.Statement = append(p.ast.Statement, p.moduleVariableDecl())
		case token.VAR:
			p.ast.Statement = append(p.ast.Statement, p.localVariableDecl())
		case token.IF:
			p.ast.Statement = append(p.ast.Statement, p.branchStmt(false))
		case token.FOR:
			p.ast.Statement = append(p.ast.Statement, p.forLoop())
		case token.WHILE:
			p.ast.Statement = append(p.ast.Statement, p.whileLoop())
		case token.LCURLY:
			p.ast.Statement = append(p.ast.Statement, p.block(false))
			p.advance()
		case token.COMMENT, token.NOOP:
			for p.current.Token == token.COMMENT || p.current.Token == token.NOOP {
				p.advance()
			}
		case token.EXPORT:
			p.ast.Statement = append(p.ast.Statement, p.export())
			if p.ok {
				return p.ast, nil
			}
			return nil, p.err
		case token.EOF:
			p.ast.Statement = append(p.ast.Statement, &ast.Ret{Expr: &ast.Nil{}, Line: p.current.Line})
			return p.ast, nil
		default:
			if p.current.Token == token.UNEXPECTED {
				p.err = p.lexer.LexicalError
			} else {
				p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected a high level statement", SyntaxErrType, p.current.Line)
			}
			return nil, p.err
		}
	}
	return nil, p.err
}

func (p *parser) mutationOrCall(statements *[]ast.Node) ast.Node {
	if p.next.Token == token.DOT ||
		p.next.Token == token.STATIC_CALL ||
		p.next.Token == token.LBRACKET ||
		p.next.Token == token.LPAREN {
		return p.mutateDataStructureOrCallStmt(statements)
	}
	line := p.current.Line
	i := p.current.Lit
	p.advance()
	p.expect(token.ASSIGN)
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	return &ast.Mut{Identifier: i, Expr: e, Line: line}
}

func (p *parser) localVariableDecl() ast.Node {
	l := p.current.Line
	isRecursive := false
	p.advance()
	if p.current.Token == token.REC {
		isRecursive = true
		p.advance()
	}
	p.expect(token.IDENTIFIER)
	i := p.current.Lit
	p.advance()
	if p.current.Token == token.COMMA {
		identifiers := make([]string, 0, 5)
		identifiers = append(identifiers, i)
		for p.current.Token == token.COMMA {
			p.advance()
			p.expect(token.IDENTIFIER)
			identifiers = append(identifiers, p.current.Lit)
			p.advance()
		}
		p.expect(token.ASSIGN)
		p.advance()
		e := p.expression(token.LowestPrec)
		p.advance()
		return &ast.MultipleVar{Identifiers: identifiers, Expr: e, IsRecursive: isRecursive, Line: l}
	}
	p.expect(token.ASSIGN)
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	return &ast.Var{Identifier: i, Expr: e, IsRecursive: isRecursive, Line: l}
}

func (p *parser) moduleVariableDecl() ast.Node {
	l := p.current.Line
	p.advance()
	p.expect(token.IDENTIFIER)
	i := p.current.Lit
	p.advance()
	if p.current.Token == token.COMMA {
		identifiers := make([]string, 0, 5)
		identifiers = append(identifiers, i)
		for p.current.Token == token.COMMA {
			p.advance()
			p.expect(token.IDENTIFIER)
			identifiers = append(identifiers, p.current.Lit)
			p.advance()
		}
		p.expect(token.ASSIGN)
		p.advance()
		e := p.expression(token.LowestPrec)
		p.advance()
		return &ast.MultipleLet{Identifiers: identifiers, Expr: e, Line: l}
	}
	p.expect(token.ASSIGN)
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	return &ast.Let{Identifier: i, Expr: e, Line: l}
}

func (p *parser) block(isInsideLoop bool) ast.Node {
	block := &ast.Block{}
	p.advance()
	for p.current.Token != token.RCURLY {
		switch p.current.Token {
		case token.IDENTIFIER:
			block.Statement = append(block.Statement, p.mutationOrCall(&block.Statement))
		case token.LET:
			block.Statement = append(block.Statement, p.moduleVariableDecl())
		case token.VAR:
			block.Statement = append(block.Statement, p.localVariableDecl())
		case token.IF:
			block.Statement = append(block.Statement, p.branchStmt(isInsideLoop))
		case token.FOR:
			block.Statement = append(block.Statement, p.forLoop())
		case token.WHILE:
			block.Statement = append(block.Statement, p.whileLoop())
		case token.RET:
			block.Statement = append(block.Statement, p.ret())
		case token.BREAK:
			if isInsideLoop {
				block.Statement = append(block.Statement, p.breakStmt())
			} else {
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was found a break outside of a loop", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return block
			}
		case token.CONTINUE:
			if isInsideLoop {
				block.Statement = append(block.Statement, p.continueStmt())
			} else {
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was found a continue outside of a loop", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return block
			}
		case token.LCURLY:
			block.Statement = append(block.Statement, p.block(isInsideLoop))
			p.advance()
		case token.COMMENT, token.NOOP:
			for p.current.Token == token.COMMENT || p.current.Token == token.NOOP {
				p.advance()
			}
		default:
			if p.ok {
				p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected a block statement", SyntaxErrType, p.current.Line)
				p.ok = false
			}
			return block
		}
	}
	return block
}

func (p *parser) mutateDataStructureOrCallStmt(statements *[]ast.Node) ast.Node {
	*statements = append(*statements, &ast.ReferenceStmt{Value: p.current.Lit, Line: p.current.Line})
	var idx ast.Node
	l := p.current.Line
Loop:
	for p.next.Token == token.LBRACKET ||
		p.next.Token == token.DOT ||
		p.next.Token == token.STATIC_CALL ||
		p.next.Token == token.LPAREN {
		p.advance()
		switch p.current.Token {
		case token.LBRACKET:
			p.advance()
			idx = p.expression(token.LowestPrec)
			p.advance()
			p.expect(token.RBRACKET)
			if p.next.Token == token.ASSIGN {
				goto assignment
			}
			*statements = append(*statements, &ast.IGetStmt{Index: idx, Line: l})
		case token.DOT:
			p.advance()
			p.expect(token.IDENTIFIER)
			idx = &ast.Property{Value: p.current.Lit}
			if p.next.Token == token.ASSIGN {
				goto assignment
			}
			if p.next.Token == token.LPAREN {
				p.advance()
				p.expect(token.LPAREN)
				p.advance()
				var args []ast.Node
				var ellipsis int
				if p.current.Token != token.RPAREN && p.current.Token != token.EOF {
					if p.current.Token == token.ELLIPSIS {
						p.advance()
						ellipsis = spreadFirst
						args = append(args, p.expression(token.LowestPrec))
						p.advance()
						goto afterMethodCall
					}
					args = append(args, p.expression(token.LowestPrec))
					p.advance()
					for p.current.Token != token.RPAREN && p.current.Token != token.EOF {
						p.expect(token.COMMA)
						p.advance()
						if p.current.Token == token.ELLIPSIS {
							p.advance()
							ellipsis = spreadLast
							args = append(args, p.expression(token.LowestPrec))
							p.advance()
							goto afterMethodCall
						}
						args = append(args, p.expression(token.LowestPrec))
						p.advance()
					}
				}
			afterMethodCall:
				p.expect(token.RPAREN)
				if p.next.Token != token.LBRACKET &&
					p.next.Token != token.DOT &&
					p.next.Token != token.LPAREN &&
					p.next.Token != token.STATIC_CALL {
					p.advance()
					return &ast.MethodCallStmt{Args: args, Prop: idx, Ellipsis: ellipsis, Line: l}
				}
				*statements = append(*statements, &ast.MethodCallStmt{Args: args, Prop: idx, Ellipsis: ellipsis, Line: l})
				continue
			}
			*statements = append(*statements, &ast.SelectStmt{Selector: idx, Line: l})
		case token.LPAREN:
			var args []ast.Node
			var ellipsis int
			p.advance()
			if p.current.Token != token.RPAREN && p.current.Token != token.EOF {
				if p.current.Token == token.ELLIPSIS {
					p.advance()
					ellipsis = spreadFirst
					args = append(args, p.expression(token.LowestPrec))
					p.advance()
					goto afterParen
				}
				args = append(args, p.expression(token.LowestPrec))
				p.advance()
				for p.current.Token != token.RPAREN && p.current.Token != token.EOF {
					p.expect(token.COMMA)
					p.advance()
					if p.current.Token == token.ELLIPSIS {
						p.advance()
						ellipsis = spreadLast
						args = append(args, p.expression(token.LowestPrec))
						p.advance()
						goto afterParen
					}
					args = append(args, p.expression(token.LowestPrec))
					p.advance()
				}
			}
		afterParen:
			p.expect(token.RPAREN)
			if p.next.Token != token.LBRACKET &&
				p.next.Token != token.DOT &&
				p.next.Token != token.STATIC_CALL &&
				p.next.Token != token.LPAREN {
				p.advance()
				return &ast.CallStmt{Args: args, Ellipsis: ellipsis, Line: l}
			}
			*statements = append(*statements, &ast.CallStmt{Args: args, Ellipsis: ellipsis, Line: l})
		case token.STATIC_CALL:
			p.advance()
			p.expect(token.IDENTIFIER)
			idx = &ast.Property{Value: p.current.Lit}
			if p.next.Token != token.LPAREN {
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected a function call after selector '::'", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return &ast.Nil{}
			}
			*statements = append(*statements, &ast.SelectStmt{Selector: idx})
		default:
			break Loop
		}
	}
assignment:
	p.advance()
	p.expect(token.ASSIGN)
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	return &ast.ISet{Index: idx, Expr: e, Line: l}
}

func (p *parser) forLoop() ast.Node {
	line := p.current.Line
	p.advance()
	if p.current.Token == token.IN {
		p.advance()
		e := p.expression(token.LowestPrec)
		p.advance()
		p.expect(token.LCURLY)
		b := p.block(true)
		p.advance()
		id := "*_"
		return &ast.IFor{Key: id, Value: id, Expr: e, Block: b, Line: line}
	}
	p.expect(token.IDENTIFIER)
	id := p.current.Lit
	p.advance()
	if p.current.Token == token.COMMA {
		return p.iterforLoop(id)
	}
	var init, end, step ast.Node
	p.expect(token.IN)
	p.advance()
	init = p.expression(token.LowestPrec)
	p.advance()
	if p.current.Token == token.COMMA {
		p.expect(token.COMMA)
		p.advance()
		end = p.expression(token.LowestPrec)
		p.advance()
		step = &ast.Integer{Value: 1}
		if p.current.Token == token.COMMA {
			p.expect(token.COMMA)
			p.advance()
			step = p.expression(token.LowestPrec)
			p.advance()
		}
		p.expect(token.LCURLY)
		block := p.block(true)
		p.advance()
		return &ast.For{Init: init, End: end, Id: id, Step: step, Block: block, Line: line}
	}
	p.expect(token.LCURLY)
	block := p.block(true)
	p.advance()
	return &ast.For{Init: &ast.Integer{Value: 0}, End: init, Id: id, Step: &ast.Integer{Value: 1}, Block: block, Line: line}
}

func (p *parser) iterforLoop(key string) ast.Node {
	line := p.current.Line
	p.advance()
	p.expect(token.IDENTIFIER)
	v := p.current.Lit
	p.advance()
	p.expect(token.IN)
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	p.expect(token.LCURLY)
	b := p.block(true)
	p.advance()
	return &ast.IFor{Key: key, Value: v, Expr: e, Block: b, Line: line}
}

func (p *parser) branchStmt(isInsideLoop bool) ast.Node {
	l := p.current.Line
	p.advance()
	c := p.expression(token.LowestPrec)
	p.advance()
	p.expect(token.LCURLY)
	b := p.block(isInsideLoop)
	p.advance()
	branch := &ast.Branch{If: &ast.If{Condition: c, Block: b}, Line: l}
	for p.current.Token == token.ELSE && p.next.Token == token.IF {
		l = p.current.Line
		p.advance()
		p.advance()
		c := p.expression(token.LowestPrec)
		p.advance()
		p.expect(token.LCURLY)
		b := p.block(isInsideLoop)
		p.advance()
		branch.Elifs = append(branch.Elifs, &ast.If{Condition: c, Block: b})
	}
	if p.current.Token == token.ELSE {
		p.advance()
		b := p.block(isInsideLoop)
		p.advance()
		branch.Else = &ast.Else{Block: b}
	}
	return branch
}

func (p *parser) whileLoop() ast.Node {
	l := p.current.Line
	p.advance()
	c := p.expression(token.LowestPrec)
	p.advance()
	p.expect(token.LCURLY)
	b := p.block(true)
	p.advance()
	return &ast.While{Condition: c, Block: b, Line: l}
}

func (p *parser) breakStmt() ast.Node {
	p.advance()
	return &ast.Break{}
}

func (p *parser) continueStmt() ast.Node {
	p.advance()
	return &ast.Continue{}
}

func (p *parser) expression(precedence int) ast.Node {
	line := p.current.Line
	e := p.prefix()
	for p.next.Token.IsBinaryOperator() && p.next.Token.Precedence() > precedence {
		p.advance()
		op := p.current.Token
		p.advance()
		nextPrecedence := op.Precedence()
		if op.IsRightAssociative() {
			nextPrecedence--
		}
		r := p.expression(nextPrecedence)
		e = &ast.BinaryExpr{Op: op, Lhs: e, Rhs: r, Line: line}
	}
	return e
}

func (p *parser) prefix() ast.Node {
	switch p.current.Token {
	case token.NOT, token.SUB, token.ADD, token.TILDE:
		t := p.current.Token
		p.advance()
		e := p.expression(token.PrefixPrec)
		return &ast.PrefixExpr{Op: t, Expr: e, Line: p.current.Line}
	}
	return p.primary()
}

func (p *parser) primary() ast.Node {
	e := p.operand()
Loop:
	for p.next.Token == token.LBRACKET ||
		p.next.Token == token.DOT ||
		p.next.Token == token.STATIC_CALL ||
		p.next.Token == token.LPAREN {
		p.advance()
		switch p.current.Token {
		case token.LBRACKET:
			e = p.indexOrSlice(e)
		case token.DOT:
			p.advance()
			switch p.current.Token {
			case token.IDENTIFIER:
				e = p.selector(e)
			default:
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected an identifier after selector", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return &ast.Nil{}
			}
		case token.LPAREN:
			e = p.callExpr(e)
		case token.STATIC_CALL:
			p.advance()
			p.expect(token.IDENTIFIER)
			prop := &ast.Property{Value: p.current.Lit}
			if p.next.Token != token.LPAREN {
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected a function call after selector '::'", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return &ast.Nil{}
			}
			e = &ast.Select{Selectable: e, Selector: prop}
		default:
			break Loop
		}
	}
	return e
}

func (p *parser) operand() ast.Node {
	switch p.current.Token {
	case token.INTEGER:
		if i, err := strconv.ParseUint(p.current.Lit, 0, 64); err == nil {
			return &ast.Integer{Value: int64(i)}
		} else {
			if p.ok {
				p.err = NewRuntimeError(p.lexer.ScriptID, "an integer literal could not be processed", SyntaxErrType, p.current.Line)
				p.ok = false
			}
			return &ast.Nil{}
		}
	case token.FLOAT:
		if f, err := strconv.ParseFloat(p.current.Lit, 64); err == nil {
			return &ast.Float{Value: f}
		}
		if p.ok {
			p.err = NewRuntimeError(p.lexer.ScriptID, "a float literal could not be processed", SyntaxErrType, p.current.Line)
			p.ok = false
		}
		return &ast.Nil{}
	case token.STRING:
		s, e := strconv.Unquote(p.current.Lit)
		if e != nil {
			if p.ok {
				p.err = NewRuntimeError(p.lexer.ScriptID, "a string literal could not be processed", SyntaxErrType, p.current.Line)
				p.ok = false
			}
			return &ast.Nil{}
		}
		return &ast.String{Value: s}
	case token.TRUE:
		return &ast.Boolean{Value: true}
	case token.FALSE:
		return &ast.Boolean{Value: false}
	case token.NIL:
		return &ast.Nil{}
	case token.IDENTIFIER:
		return &ast.Reference{Value: p.current.Lit, Line: p.current.Line}
	case token.LBRACKET:
		xs := &ast.Array{}
		p.advance()
		for p.current.Token != token.RBRACKET && p.current.Token != token.EOF {
			e := p.expression(token.LowestPrec)
			p.advance()
			xs.ExprList = append(xs.ExprList, e)
			for p.current.Token == token.COMMA {
				p.advance()
				if p.current.Token == token.RBRACKET {
					p.expect(token.RBRACKET)
					return xs
				}
				e := p.expression(token.LowestPrec)
				p.advance()
				xs.ExprList = append(xs.ExprList, e)
			}
			goto endList
		}
	endList:
		p.expect(token.RBRACKET)
		return xs
	case token.LCURLY:
		obj := &ast.Object{Line: p.current.Line}
		p.advance()
	loop:
		for p.current.Token != token.RCURLY {
			var k ast.Node
			switch p.current.Token {
			case token.STRING:
				p.current.Token = token.IDENTIFIER
				p.current.Lit = p.current.Lit[1 : len(p.current.Lit)-1]
				fallthrough
			case token.IDENTIFIER:
				p.expect(token.IDENTIFIER)
				k = &ast.Property{Value: p.current.Lit}
			default:
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected an identifier or string as object property", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return &ast.Nil{}
			}
			p.advance()
			if p.current.Token == token.COMMA {
				p.advance()
			}
			switch p.current.Token {
			case token.IDENTIFIER, token.STRING:
				obj.Pairs = append(obj.Pairs, &ast.Pair{Key: k, Value: &ast.Nil{}})
			case token.ASSIGN:
				p.expect(token.ASSIGN)
				p.advance()
				v := p.expression(token.LowestPrec)
				p.advance()
				obj.Pairs = append(obj.Pairs, &ast.Pair{Key: k, Value: v})
				if p.current.Token == token.COMMA {
					p.advance()
				}
			case token.RCURLY:
				obj.Pairs = append(obj.Pairs, &ast.Pair{Key: k, Value: &ast.Nil{}})
				break loop
			default:
				if p.ok {
					p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected an identifier, comma or an equals symbol after object property", SyntaxErrType, p.current.Line)
					p.ok = false
				}
				return &ast.Nil{}
			}
		}
		p.expect(token.RCURLY)
		return obj
	case token.LPAREN:
		p.advance()
		if p.current.Token == token.RPAREN {
			if p.ok {
				p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected an expression after left parenthesis", SyntaxErrType, p.current.Line)
				p.ok = false
			}
			return &ast.Nil{}
		}
		e := p.expression(token.LowestPrec)
		p.advance()
		p.expect(token.RPAREN)
		return e
	case token.FUN:
		f := &ast.Fun{}
		p.advance()
		if p.current.Token != token.ARROW && p.current.Token != token.LCURLY && p.current.Token != token.EOF {
			p.expect(token.IDENTIFIER)
			f.Args = append(f.Args, p.current.Lit)
			p.advance()
		}
		for p.current.Token != token.ARROW && p.current.Token != token.LCURLY && p.current.Token != token.EOF {
			if p.current.Token == token.ELLIPSIS {
				f.IsVar = true
				p.advance()
				goto endParams
			}
			p.expect(token.COMMA)
			p.advance()
			p.expect(token.IDENTIFIER)
			f.Args = append(f.Args, p.current.Lit)
			p.advance()
		}
	endParams:
		if p.current.Token == token.ARROW {
			l := p.current.Line
			p.advance()
			e := p.expression(token.LowestPrec)
			b := &ast.Block{}
			b.Statement = append(b.Statement, &ast.Ret{Expr: e, Line: l})
			f.Body = b
			return f
		}
		p.expect(token.LCURLY)
		block := p.block(false)
		block.(*ast.Block).Statement = append(block.(*ast.Block).Statement, &ast.Ret{Expr: &ast.Nil{}})
		f.Body = block
		return f
	case token.IMPORT:
		i := &ast.Import{Line: p.current.Line}
		p.advance()
		p.expect(token.LPAREN)
		p.advance()
		p.expect(token.STRING)
		s, _ := strconv.Unquote(p.current.Lit)
		i.Path = s + VidaFileExtension
		p.advance()
		p.expect(token.RPAREN)
		return i
	case token.ENUM:
		e := &ast.Enum{}
		p.advance()
		p.expect(token.LCURLY)
		p.advance()
		p.expect(token.IDENTIFIER)
		e.Variants = append(e.Variants, p.current.Lit)
		p.advance()
		for p.current.Token != token.RCURLY {
			p.expect(token.IDENTIFIER)
			e.Variants = append(e.Variants, p.current.Lit)
			p.advance()
		}
		p.expect(token.RCURLY)
		return e
	default:
		if p.ok {
			if p.lexer.LexicalError == nil {
				p.err = NewRuntimeError(p.lexer.ScriptID, "it was expected a valid expression", SyntaxErrType, p.current.Line)
			} else {
				p.err = NewRuntimeError(p.lexer.ScriptID, p.lexer.LexicalError.Error(), SyntaxErrType, p.current.Line)
			}
			p.ok = false
		}
		return &ast.Nil{}
	}
}

func (p *parser) ret() ast.Node {
	l := p.current.Line
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	return &ast.Ret{Expr: e, Line: l}
}

func (p *parser) export() ast.Node {
	l := p.current.Line
	p.advance()
	e := p.expression(token.LowestPrec)
	p.advance()
	return &ast.Export{Expr: e, Line: l}
}

func (p *parser) callExpr(e ast.Node) ast.Node {
	var args []ast.Node
	var ellipsis int
	p.advance()
	if p.current.Token != token.RPAREN && p.current.Token != token.EOF {
		if p.current.Token == token.ELLIPSIS {
			p.advance()
			ellipsis = spreadFirst
			args = append(args, p.expression(token.LowestPrec))
			p.advance()
			goto afterParen
		}
		args = append(args, p.expression(token.LowestPrec))
		p.advance()
		for p.current.Token != token.RPAREN && p.current.Token != token.EOF {
			p.expect(token.COMMA)
			p.advance()
			if p.current.Token == token.ELLIPSIS {
				p.advance()
				ellipsis = spreadLast
				args = append(args, p.expression(token.LowestPrec))
				p.advance()
				goto afterParen
			}
			args = append(args, p.expression(token.LowestPrec))
			p.advance()
		}
	}
afterParen:
	p.expect(token.RPAREN)
	return &ast.CallExpr{Fun: e, Args: args, Ellipsis: ellipsis}
}

func (p *parser) indexOrSlice(e ast.Node) ast.Node {
	p.advance()
	var index [2]ast.Node
	mode := sliceInitialMode
	if p.current.Token != token.DOUBLE_DOT {
		mode |= sliceHasStart
		index[0] = p.expression(token.LowestPrec)
		p.advance()
	}
	var hasDots bool
	if p.current.Token == token.DOUBLE_DOT {
		hasDots = true
		p.advance()
		if p.current.Token != token.RBRACKET && p.current.Token != token.EOF {
			mode |= sliceHasEnd
			index[1] = p.expression(token.LowestPrec)
			p.advance()
		}
	}
	p.expect(token.RBRACKET)
	if hasDots {
		return &ast.Slice{
			Value: e,
			First: index[0],
			Last:  index[1],
			Mode:  mode,
		}
	}
	return &ast.IGet{
		Indexable: e,
		Index:     index[0],
		Line:      p.current.Line,
	}
}

func (p *parser) selector(e ast.Node) ast.Node {
	if p.next.Token == token.LPAREN {
		prop := &ast.Property{Value: p.current.Lit}
		p.advance()
		p.expect(token.LPAREN)
		p.advance()
		var args []ast.Node
		var ellipsis int
		if p.current.Token != token.RPAREN && p.current.Token != token.EOF {
			if p.current.Token == token.ELLIPSIS {
				p.advance()
				ellipsis = spreadFirst
				args = append(args, p.expression(token.LowestPrec))
				p.advance()
				goto afterMethodCall
			}
			args = append(args, p.expression(token.LowestPrec))
			p.advance()
			for p.current.Token != token.RPAREN && p.current.Token != token.EOF {
				p.expect(token.COMMA)
				p.advance()
				if p.current.Token == token.ELLIPSIS {
					p.advance()
					ellipsis = spreadLast
					args = append(args, p.expression(token.LowestPrec))
					p.advance()
					goto afterMethodCall
				}
				args = append(args, p.expression(token.LowestPrec))
				p.advance()
			}
		}
	afterMethodCall:
		p.expect(token.RPAREN)
		return &ast.MethodCallExpr{Args: args, Obj: e, Prop: prop, Ellipsis: ellipsis}
	}
	return &ast.Select{Selectable: e, Selector: &ast.Property{Value: p.current.Lit}}
}

func (p *parser) expect(tok token.Token) {
	if p.current.Token != tok && p.ok {
		p.ok = false
		message := fmt.Sprintf("expected symbol '%v', but got '%v'", tok, p.current.Lit)
		p.err = NewRuntimeError(p.lexer.ScriptID, message, SyntaxErrType, p.current.Line)
	}
}

func (p *parser) advance() token.Token {
	p.current.Line, p.current.Token, p.current.Lit = p.next.Line, p.next.Token, p.next.Lit
	p.next.Line, p.next.Token, p.next.Lit = p.lexer.Next()
	for p.next.Token == token.COMMENT {
		p.next.Line, p.next.Token, p.next.Lit = p.lexer.Next()
	}
	return p.current.Token
}
