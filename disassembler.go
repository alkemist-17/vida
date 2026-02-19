package vida

import (
	"fmt"
	"strings"

	"github.com/alkemist-17/vida/token"
)

func PrintBytecode(script *Script, name string) string {
	clear()
	fmt.Println("Machine Code for", name)
	var sb strings.Builder
	sb.WriteString(printHeader())
	var s string
	for i := 1; i < len(script.MainFunction.CoreFn.Code); i++ {
		s = printInstr(script.MainFunction.CoreFn.Code[i], uint64(i), false)
		sb.WriteString(s)
	}
	for idx, v := range *script.Konstants {
		if f, ok := v.(*CoreFunction); ok {
			fmt.Fprintf(&sb, "\n\n\n\nFunction %v/%v/%v", idx, f.Arity, f.Free)
			var s string
			for i := 0; i < len(f.Code); i++ {
				s = printInstr(f.Code[i], uint64(i), false)
				sb.WriteString(s)
			}
		}
	}
	sb.WriteString(printKonstants(*script.Konstants))
	return sb.String()
}

func printHeader() string {
	var sb strings.Builder
	sb.WriteRune(10)
	sb.WriteRune(10)
	sb.WriteRune(10)
	fmt.Fprintf(&sb, "%v\n%v", Name(), Version())
	sb.WriteRune(10)
	sb.WriteRune(10)
	sb.WriteRune(10)
	sb.WriteRune(10)
	sb.WriteString("Main\n")
	return sb.String()
}

func printKonstants(konst []Value) string {
	var sb strings.Builder
	sb.WriteString("\n\n\n\nKonstants\n")
	for i, v := range konst {
		fmt.Fprintf(&sb, "  %4v  [%4v]  %v\n", i+1, i, v)
	}
	sb.WriteRune(10)
	sb.WriteRune(10)
	return sb.String()
}

func printInstr(instr, ip uint64, isRunningDebug bool) string {
	var sb strings.Builder
	var op, A, B, P uint64
	op = instr >> shift56
	A = instr >> shift16 & clean16
	B = instr & clean16
	P = instr >> shift32 & clean24
	if !isRunningDebug {
		sb.WriteRune(10)
		fmt.Fprintf(&sb, "  [%3v]  ", ip)
		fmt.Fprintf(&sb, "%7v", opcodes[op])
	} else {
		sb.WriteString(opcodes[op])
	}
	switch op {
	case end:
		return sb.String()
	case array, slice, iForSet, check, load:
		fmt.Fprintf(&sb, " %3v %3v %3v", P, A, B)
	case forSet, forLoop, iForLoop, fun, ret:
		fmt.Fprintf(&sb, " %3v %3v", A, B)
	case prefix:
		fmt.Fprintf(&sb, " %3v %3v %3v", token.Token(P).String(), A, B)
	case binopG, binop, binopK, binopQ:
		fmt.Fprintf(&sb, " %3v %3v %3v %3v", token.Token(P>>shift16).String(), P&clean16, A, B)
	case call, store:
		fmt.Fprintf(&sb, " %3v %3v %3v %3v", P>>shift16, P&clean16, A, B)
	case object, jump:
		fmt.Fprintf(&sb, " %3v", B)
	case iSet, iGet:
		fmt.Fprintf(&sb, " %3v %3v %3v %3v %3v", (P>>shift16)>>shift4, P>>shift16&clean8, P&clean16, A, B)
	case eq:
		var op token.Token
		var s byte = byte(P >> shift16)
		if s>>shift4 == 0 {
			op = token.EQ
		} else {
			op = token.NEQ
		}
		l := s >> shift2 & clean2bits
		r := s & clean2bits
		fmt.Fprintf(&sb, " %3v %3v %3v %3v %3v %3v", op.String(), l, r, P&clean16, A, B)
	}
	return sb.String()
}
