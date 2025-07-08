package ast

import (
	"fmt"
	"strings"
)

func PrintAST(node Node) string {
	var sb strings.Builder
	printAST(node, &sb, zeroLevel)
	return sb.String()
}

func buildIndent(sb *strings.Builder, level int) {
	for range level {
		sb.WriteRune(space)
	}
}

func printAST(node Node, sb *strings.Builder, level int) {
	switch n := node.(type) {
	case *Ast:
		sb.WriteRune(nl)
		sb.WriteRune(nl)
		sb.WriteString("AST")
		for i := range len(n.Statement) {
			printAST(n.Statement[i], sb, level+oneLevel)
		}
		sb.WriteRune(nl)
		sb.WriteRune(nl)
	case *Var:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Loc")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Identifier)
		sb.WriteRune(nl)
		printAST(n.Expr, sb, level+oneLevel)
	case *Mut:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Mut")
		buildIndent(sb, level+twoLevels)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Indentifier)
		sb.WriteRune(nl)
		printAST(n.Expr, sb, level+oneLevel)
	case *Let:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Let")
		buildIndent(sb, level+twoLevels)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Indentifier)
		sb.WriteRune(nl)
		printAST(n.Expr, sb, level+oneLevel)
	case *Reference:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Ref")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
	case *ReferenceStmt:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("RefStmt")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
	case *Identifier:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Id")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
	case *Boolean:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Bool")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		if n.Value {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case *Nil:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Nil")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString("nil")
	case *Integer:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Integer")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(fmt.Sprint(n.Value))
	case *Float:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Float")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(fmt.Sprint(n.Value))
	case *String:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("String")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
	case *Array:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Array")
		sb.WriteRune(nl)
		if len(n.ExprList) == 0 {
			buildIndent(sb, level+twoLevels)
			sb.WriteString("[]")
		} else {
			for _, v := range n.ExprList {
				printAST(v, sb, level+oneLevel)
				sb.WriteRune(nl)
			}
		}
	case *Object:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Object")
		sb.WriteRune(nl)
		if len(n.Pairs) == 0 {
			buildIndent(sb, level+twoLevels)
			sb.WriteString("{}")
		} else {
			for _, v := range n.Pairs {
				printAST(v, sb, level+oneLevel)
				sb.WriteRune(nl)
			}
		}
	case *Pair:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Pair")
		sb.WriteRune(nl)
		printAST(n.Key, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.Value, sb, level+oneLevel)
	case *Property:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Property")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
	case *PrefixExpr:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Prefix")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Op.String())
		sb.WriteRune(nl)
		printAST(n.Expr, sb, level+oneLevel)
	case *BinaryExpr:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Binary")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Op.String())
		sb.WriteRune(nl)
		printAST(n.Lhs, sb, level+twoLevels)
		sb.WriteRune(nl)
		printAST(n.Rhs, sb, level+twoLevels)
	case *IGet:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("IGet")
		sb.WriteRune(nl)
		printAST(n.Indexable, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.Index, sb, level+oneLevel)
	case *IGetStmt:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("IGetStmt")
		sb.WriteRune(nl)
		printAST(n.Index, sb, level+oneLevel)
	case *ISet:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("ISet")
		sb.WriteRune(nl)
		printAST(n.Index, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.Expr, sb, level+oneLevel)
	case *Slice:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Slice")
		sb.WriteRune(nl)
		printAST(n.Value, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.First, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.Last, sb, level+oneLevel)
	case *Select:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Select")
		sb.WriteRune(nl)
		printAST(n.Selectable, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.Selector, sb, level+oneLevel)
	case *SelectStmt:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("SelectStmt")
		sb.WriteRune(nl)
		printAST(n.Selector, sb, level+oneLevel)
	case *For:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("For")
		sb.WriteRune(nl)
		printAST(n.Init, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.End, sb, level+oneLevel)
		sb.WriteRune(nl)
		printAST(n.Step, sb, level+oneLevel)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Id)
		sb.WriteRune(nl)
		printAST(n.Block, sb, level+oneLevel)
	case *IFor:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("IFor")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Key)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
		sb.WriteRune(nl)
		printAST(n.Expr, sb, level+oneLevel)
		printAST(n.Block, sb, level+oneLevel)
	case *ForState:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("State")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Value)
	case *Branch:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Branch")
		printAST(n.If, sb, level+oneLevel)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		if len(n.Elifs) == 0 {
			sb.WriteString("No Elifs")
		} else {
			sb.WriteString("Elifs")
		}
		for _, v := range n.Elifs {
			printAST(v, sb, level+oneLevel)
		}
		printAST(n.Else, sb, level+oneLevel)
	case *If:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("If")
		sb.WriteRune(nl)
		printAST(n.Condition, sb, level+twoLevels)
		sb.WriteRune(nl)
		printAST(n.Block, sb, level+twoLevels)
	case *Else:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Else")
		printAST(n.Block, sb, level+oneLevel)
	case *While:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("While")
		sb.WriteRune(nl)
		printAST(n.Condition, sb, level+twoLevels)
		sb.WriteRune(nl)
		printAST(n.Block, sb, level+twoLevels)
	case *Break:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Break")
	case *Continue:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Continue")
	case *Fun:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Fun")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		for _, v := range n.Args {
			sb.WriteString(v)
			sb.WriteRune(nl)
			buildIndent(sb, level+twoLevels)
		}
		printAST(n.Body, sb, level+oneLevel)
	case *Ret:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Ret")
		sb.WriteRune(nl)
		printAST(n.Expr, sb, twoLevels+twoLevels+oneLevel)
	case *Export:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Export")
		sb.WriteRune(nl)
		printAST(n.Expr, sb, twoLevels+twoLevels+oneLevel)
	case *Import:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Import")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString(n.Path)
	case *CallExpr:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("CallExpr")
		sb.WriteRune(nl)
		printAST(n.Fun, sb, level+twoLevels)
		buildIndent(sb, level+twoLevels)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString("Args")
		sb.WriteRune(nl)
		for _, v := range n.Args {
			printAST(v, sb, level+twoLevels)
			sb.WriteRune(nl)
		}
	case *CallStmt:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("CallStmt")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString("Args")
		sb.WriteRune(nl)
		for _, v := range n.Args {
			printAST(v, sb, level+twoLevels)
			sb.WriteRune(nl)
		}
	case *MethodCallExpr:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("MethodCallExpr")
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString("Doc")
		sb.WriteRune(nl)
		printAST(n.Obj, sb, level+twoLevels)
		sb.WriteRune(nl)
		printAST(n.Prop, sb, level+oneLevel)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString("Args")
		sb.WriteRune(nl)
		for _, v := range n.Args {
			printAST(v, sb, level+twoLevels)
			sb.WriteRune(nl)
		}
	case *MethodCallStmt:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("MethodCallStmt")
		sb.WriteRune(nl)
		printAST(n.Prop, sb, level+oneLevel)
		sb.WriteRune(nl)
		buildIndent(sb, level+twoLevels)
		sb.WriteString("Args")
		sb.WriteRune(nl)
		for _, v := range n.Args {
			printAST(v, sb, level+twoLevels)
			sb.WriteRune(nl)
		}
	case *Block:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Block")
		for i := range len(n.Statement) {
			printAST(n.Statement[i], sb, level+twoLevels)
		}
	case *Enum:
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Enum")
		sb.WriteRune(nl)
		if n.HasForExpr {
			if n.Step != 0 {
				sb.WriteString(fmt.Sprintf("for %v = %v, %v", n.Identifier, n.Init, n.Step))
			} else {
				sb.WriteString(fmt.Sprintf("for %v = %v", n.Identifier, n.Init))
			}
		}
		for _, v := range n.Variants {
			buildIndent(sb, level+twoLevels)
			sb.WriteString(v)
			sb.WriteRune(nl)
		}
	default:
		sb.WriteRune(nl)
		buildIndent(sb, level+oneLevel)
		sb.WriteString("Nothing")
	}
}

const space rune = 32
const nl rune = 10
const oneLevel = 2
const twoLevels = 4
const zeroLevel = 0
