package ast

import (
	"fmt"
	"strings"
)

func StringifyAST(node Node) string {
	count := countNodes(node)
	var sb strings.Builder
	printNode(node, &sb, "", "", false)
	out := sb.String()
	nl := strings.Index(out, "\n")
	if nl >= 0 {
		countLine := fmt.Sprintf("[%d nodes]\n", count)
		out = out[:nl+1] + countLine + out[nl+1:]
	}
	return out
}

func PrintASTColor(node Node) string {
	count := countNodes(node)
	var sb strings.Builder
	printNode(node, &sb, "", "", true)
	out := sb.String()
	nl := strings.Index(out, "\n")
	if nl >= 0 {
		countLine := fmt.Sprintf("[%d nodes]\n", count)
		out = out[:nl+1] + countLine + out[nl+1:]
	}
	return out
}

const (
	colorReset   = "\x1b[0m"
	colorDecl    = "\x1b[38;5;108m"
	colorRef     = "\x1b[38;5;172m"
	colorLiteral = "\x1b[38;5;113m"
	colorColl    = "\x1b[38;5;104m"
	colorExpr    = "\x1b[38;5;251m"
	colorCall    = "\x1b[96m"
	colorFlow    = "\x1b[38;5;251m"
	colorFun     = "\x1b[38;5;138m"
	colorArgs    = "\x1b[38;5;216m"
)

func nodeColor(node Node) string {
	switch node.(type) {
	case *Var, *Let, *Mut, *MultipleLet, *MultipleVar, *MultipleMut:
		return colorDecl
	case *Reference, *ReferenceStmt, *Identifier:
		return colorRef
	case *Boolean, *Nil, *Integer, *Float, *String:
		return colorLiteral
	case *Array, *Object, *Pair, *Property:
		return colorColl
	case *BinaryExpr, *PrefixExpr, *IGet, *IGetStmt, *ISet, *Slice, *Select, *SelectStmt:
		return colorExpr
	case *CallExpr, *CallStmt, *MethodCallExpr, *MethodCallStmt:
		return colorCall
	case syntheticArgs, syntheticAssign:
		return colorArgs
	case *Block, *Branch, *If, *Else, *While, *For, *IFor, *ForState, *Break, *Continue:
		return colorFlow
	case *Fun, *Ret, *Export, *Import, *Enum:
		return colorFun
	default:
		return colorExpr
	}
}

const (
	branchMid  = "├─ "
	branchLast = "└─ "
	contMid    = "│  "
	contLast   = "   "
)

func printNode(node Node, sb *strings.Builder, own, cont string, color bool) {
	if node == nil {
		writeLine(sb, own, "·nil·", "", color)
		return
	}

	col := ""
	if color {
		col = nodeColor(node)
	}

	switch n := node.(type) {
	case syntheticArgs:
		if len(n) == 0 {
			writeLine(sb, own, "Args (none)", col, color)
			return
		}
		writeLine(sb, own, "Args", col, color)
		writeChildren(sb, cont, n, color)

	case *Ast:
		writeLine(sb, own, "AST", col, color)
		writeChildren(sb, cont, nodeSlice(n.Statement), color)

	case *Var:
		writeLine(sb, own, "Var "+n.Identifier, col, color)
		writeChild(sb, cont, n.Expr, color)

	case *Let:
		writeLine(sb, own, "Let "+n.Identifier, col, color)
		writeChild(sb, cont, n.Expr, color)

	case *MultipleLet:
		writeLine(sb, own, "Let "+multipleIdentifiers(n.Identifiers), col, color)
		writeChildren(sb, cont, n.Exprs, color)

	case *MultipleVar:
		writeLine(sb, own, "Var "+multipleIdentifiers(n.Identifiers), col, color)
		writeChildren(sb, cont, n.Exprs, color)

	case *Mut:
		writeLine(sb, own, "Mut "+n.Identifier, col, color)
		writeChild(sb, cont, n.Expr, color)

	case *MultipleMut:
		writeLine(sb, own, "Mut (multi)", col, color)
		pairs := make([]Node, len(n.Targets))
		for i, t := range n.Targets {
			pairs[i] = syntheticAssign{Target: targetToNode(t), Expr: n.Exprs[i]}
		}
		writeChildren(sb, cont, pairs, color)

	case syntheticAssign:
		writeLine(sb, own, "Assign", col, color)
		writeChildren(sb, cont, []Node{n.Target, n.Expr}, color)

	case *Reference:
		writeLine(sb, own, "Ref "+n.Value, col, color)

	case *ReferenceStmt:
		writeLine(sb, own, "RefStmt "+n.Value, col, color)

	case *Identifier:
		writeLine(sb, own, "Id "+n.Value, col, color)

	case *Boolean:
		writeLine(sb, own, fmt.Sprintf("Bool %v", n.Value), col, color)

	case *Nil:
		writeLine(sb, own, "Nil", col, color)

	case *Integer:
		writeLine(sb, own, fmt.Sprintf("Int %v", n.Value), col, color)

	case *Float:
		writeLine(sb, own, fmt.Sprintf("Float %v", n.Value), col, color)

	case *String:
		writeLine(sb, own, fmt.Sprintf("String %q", n.Value), col, color)

	case *Array:
		if len(n.ExprList) == 0 {
			writeLine(sb, own, "Array []", col, color)
		} else {
			writeLine(sb, own, "Array", col, color)
			writeChildren(sb, cont, n.ExprList, color)
		}

	case *Object:
		if len(n.Pairs) == 0 {
			writeLine(sb, own, "Object {}", col, color)
		} else {
			writeLine(sb, own, "Object", col, color)
			writeChildren(sb, cont, nodeSlice(n.Pairs), color)
		}

	case *Pair:
		writeLine(sb, own, "Pair", col, color)
		writeChildren(sb, cont, []Node{n.Key, n.Value}, color)

	case *Property:
		writeLine(sb, own, "Prop ."+n.Value, col, color)

	case *PrefixExpr:
		writeLine(sb, own, "Prefix "+n.Op.String(), col, color)
		writeChild(sb, cont, n.Expr, color)

	case *BinaryExpr:
		writeLine(sb, own, "Binary "+n.Op.String(), col, color)
		writeChildren(sb, cont, []Node{n.Lhs, n.Rhs}, color)

	case *IGet:
		writeLine(sb, own, "Get", col, color)
		writeChildren(sb, cont, []Node{n.Indexable, n.Index}, color)

	case *IGetStmt:
		writeLine(sb, own, "IGetStmt", col, color)
		writeChild(sb, cont, n.Index, color)

	case *ISet:
		writeLine(sb, own, "Set", col, color)
		writeChildren(sb, cont, []Node{n.Index, n.Expr}, color)

	case *Slice:
		writeLine(sb, own, "Slice", col, color)
		writeChildren(sb, cont, []Node{n.Value, n.First, n.Last}, color)

	case *Select:
		writeLine(sb, own, "Select", col, color)
		writeChildren(sb, cont, []Node{n.Selectable, n.Selector}, color)

	case *SelectStmt:
		writeLine(sb, own, "SelectStmt", col, color)
		writeChild(sb, cont, n.Selector, color)

	case *CallExpr:
		writeLine(sb, own, "CallExpr", col, color)
		writeChildren(sb, cont, []Node{n.Fun, syntheticArgs(n.Args)}, color)

	case *CallStmt:
		writeLine(sb, own, "CallStmt", col, color)
		writeChild(sb, cont, syntheticArgs(n.Args), color)

	case *MethodCallExpr:
		writeLine(sb, own, "MethodCallExpr", col, color)
		writeChildren(sb, cont, []Node{n.Obj, n.Prop, syntheticArgs(n.Args)}, color)

	case *MethodCallStmt:
		writeLine(sb, own, "MethodCallStmt", col, color)
		writeChildren(sb, cont, []Node{n.Prop, syntheticArgs(n.Args)}, color)

	case *Block:
		writeLine(sb, own, "Block", col, color)
		writeChildren(sb, cont, nodeSlice(n.Statement), color)

	case *Branch:
		writeLine(sb, own, "Branch", col, color)
		kids := []Node{n.If}
		for _, e := range n.Elifs {
			kids = append(kids, e)
		}
		kids = append(kids, n.Else)
		writeChildren(sb, cont, kids, color)

	case *If:
		writeLine(sb, own, "If", col, color)
		writeChildren(sb, cont, []Node{n.Condition, n.Block}, color)

	case *Else:
		writeLine(sb, own, "Else", col, color)
		writeChild(sb, cont, n.Block, color)

	case *While:
		writeLine(sb, own, "While", col, color)
		writeChildren(sb, cont, []Node{n.Condition, n.Block}, color)

	case *For:
		writeLine(sb, own, "For "+n.Id, col, color)
		writeChildren(sb, cont, []Node{n.Init, n.End, n.Step, n.Block}, color)

	case *IFor:
		label := n.Key
		if n.Value != "" {
			label += ", " + n.Value
		}
		writeLine(sb, own, "IFor "+label, col, color)
		writeChildren(sb, cont, []Node{n.Expr, n.Block}, color)

	case *ForState:
		writeLine(sb, own, "ForState "+n.Value, col, color)

	case *Break:
		writeLine(sb, own, "Break", col, color)

	case *Continue:
		writeLine(sb, own, "Continue", col, color)

	case *Fun:
		writeLine(sb, own, "Fun("+strings.Join(n.Args, ", ")+")", col, color)
		writeChild(sb, cont, n.Body, color)

	case *Ret:
		writeLine(sb, own, "Ret", col, color)
		writeChild(sb, cont, n.Expr, color)

	case *Export:
		writeLine(sb, own, "Export", col, color)
		writeChild(sb, cont, n.Expr, color)

	case *Import:
		writeLine(sb, own, fmt.Sprintf("Import %q", n.Path), col, color)

	case *Enum:
		writeLine(sb, own, "Enum", col, color)
		for i, v := range n.Variants {
			branch := branchMid
			if i == len(n.Variants)-1 {
				branch = branchLast
			}
			writeLine(sb, cont+branch, v, col, color)
		}

	default:
		writeLine(sb, own, fmt.Sprintf("·unknown(%T)·", node), "", color)
	}
}

func writeLine(sb *strings.Builder, own, label, col string, color bool) {
	sb.WriteString(own)
	if color && col != "" {
		sb.WriteString(col)
		sb.WriteString(label)
		sb.WriteString(colorReset)
	} else {
		sb.WriteString(label)
	}
	sb.WriteByte('\n')
}

func writeChild(sb *strings.Builder, cont string, child Node, color bool) {
	printNode(child, sb, cont+branchLast, cont+contLast, color)
}

func writeChildren(sb *strings.Builder, cont string, children []Node, color bool) {
	for i, child := range children {
		if i == len(children)-1 {
			printNode(child, sb, cont+branchLast, cont+contLast, color)
		} else {
			printNode(child, sb, cont+branchMid, cont+contMid, color)
		}
	}
}

func countNodes(node Node) int {
	if node == nil {
		return 0
	}
	count := 1
	switch n := node.(type) {
	case *Ast:
		for _, s := range n.Statement {
			count += countNodes(s)
		}
	case *Var:
		count += countNodes(n.Expr)
	case *Let:
		count += countNodes(n.Expr)
	case *MultipleLet:
		for _, e := range n.Exprs {
			count += countNodes(e)
		}
	case *MultipleVar:
		for _, e := range n.Exprs {
			count += countNodes(e)
		}
	case *Mut:
		count += countNodes(n.Expr)
	case *MultipleMut:
		for _, t := range n.Targets {
			if t.Identifier == "" {
				if t.Indexable != nil {
					count += countNodes(t.Indexable) + countNodes(t.Index)
				} else {
					count += countNodes(t.Selectable) + countNodes(t.Selector)
				}
			}
		}
		for _, e := range n.Exprs {
			count += countNodes(e)
		}
	case *PrefixExpr:
		count += countNodes(n.Expr)
	case *BinaryExpr:
		count += countNodes(n.Lhs) + countNodes(n.Rhs)
	case *IGet:
		count += countNodes(n.Indexable) + countNodes(n.Index)
	case *IGetStmt:
		count += countNodes(n.Index)
	case *ISet:
		count += countNodes(n.Index) + countNodes(n.Expr)
	case *Slice:
		count += countNodes(n.Value) + countNodes(n.First) + countNodes(n.Last)
	case *Select:
		count += countNodes(n.Selectable) + countNodes(n.Selector)
	case *SelectStmt:
		count += countNodes(n.Selector)
	case *Array:
		for _, e := range n.ExprList {
			count += countNodes(e)
		}
	case *Object:
		for _, p := range n.Pairs {
			count += countNodes(p)
		}
	case *Pair:
		count += countNodes(n.Key) + countNodes(n.Value)
	case *CallExpr:
		count += countNodes(n.Fun)
		for _, a := range n.Args {
			count += countNodes(a)
		}
	case *CallStmt:
		for _, a := range n.Args {
			count += countNodes(a)
		}
	case *MethodCallExpr:
		count += countNodes(n.Obj) + countNodes(n.Prop)
		for _, a := range n.Args {
			count += countNodes(a)
		}
	case *MethodCallStmt:
		count += countNodes(n.Prop)
		for _, a := range n.Args {
			count += countNodes(a)
		}
	case *Block:
		for _, s := range n.Statement {
			count += countNodes(s)
		}
	case *Branch:
		count += countNodes(n.If)
		for _, e := range n.Elifs {
			count += countNodes(e)
		}
		count += countNodes(n.Else)
	case *If:
		count += countNodes(n.Condition) + countNodes(n.Block)
	case *Else:
		count += countNodes(n.Block)
	case *While:
		count += countNodes(n.Condition) + countNodes(n.Block)
	case *For:
		count += countNodes(n.Init) + countNodes(n.End) + countNodes(n.Step) + countNodes(n.Block)
	case *IFor:
		count += countNodes(n.Expr) + countNodes(n.Block)
	case *Fun:
		count += countNodes(n.Body)
	case *Ret:
		count += countNodes(n.Expr)
	case *Export:
		count += countNodes(n.Expr)
	}
	return count
}

type syntheticArgs []Node

func (sa syntheticArgs) _node() {}

// syntheticAssign pairs one MultipleMut target with its corresponding
// RHS expression for display purposes only -- it never appears in a
// real AST, just in the printed tree.
type syntheticAssign struct {
	Target Node
	Expr   Node
}

func (sa syntheticAssign) _node() {}

// targetToNode renders a MutTarget by reusing the existing Reference,
// IGet, and Select printers rather than inventing new display logic --
// a target prints exactly like it would if it appeared as an expression.
func targetToNode(t *MutTarget) Node {
	if t.Identifier != "" {
		return &Reference{Value: t.Identifier, Line: t.Line}
	}
	if t.Indexable != nil {
		return &IGet{Indexable: t.Indexable, Index: t.Index, Line: t.Line}
	}
	return &Select{Selectable: t.Selectable, Selector: t.Selector}
}

func nodeSlice[T Node](in []T) []Node {
	out := make([]Node, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func multipleIdentifiers(identifiers []string) string {
	return strings.Join(identifiers, ",")
}
