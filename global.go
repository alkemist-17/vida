package vida

import (
	"fmt"
	"os"

	"github.com/alkemist-17/vida/verror"
)

var NilValue = NilVal()

type LibsLoader map[string]func() Value

type ErrorInfo map[string]map[int]uint

var extensionlibsLoader LibsLoader

var __meta string = inititalMetaName

const (
	inititalMetaName = "$$__meta__$$"

	globalStateIndex = 0

	maxMetaSearch = 10_000

	errorMessageFieldName = "message"

	DefaultInputPrompt = "Input > "

	foundationInterfaceName = "std/"

	__getmeta = "__getmeta"

	__setmeta = "__setmeta"

	__call = "__call"

	__str = "__str"

	__type = "__type"

	__get = "__get"

	__set = "__set"

	__add = "__add"

	__sub = "__sub"

	__mul = "__mul"

	__div = "__div"

	__rem = "__rem"

	__eq = "__eq"

	__le = "__le"

	__lt = "__lt"

	__ge = "__ge"

	__gt = "__gt"

	__umin = "__umin"

	__uplus = "__uplus"
)

var clbu *[]Value

func stringWithVisitedTValue(v Value, visited map[uintptr]bool) string {
	switch v.ttype {
	case TArray:
		return v.Arr().stringify(visited)
	case TObject:
		return v.Obj().stringify(visited)
	default:
		return v.String()
	}
}

type GlobalState struct {
	*VM
	Main    *Thread
	Current *Thread
}

var coreLibNames = []string{
	"--G--",
	"print",
	"len",
	"append",
	"newArray",
	"load",
	"type",
	"assert",
	"format",
	"input",
	"clone",
	"error",
	"isError",
}

func loadCoreLib(store *[]Value) *[]Value {
	*store = append(*store,
		NilValue,
		GFnVal(corePrint),
		GFnVal(coreLen),
		GFnVal(coreAppend),
		GFnVal(coreMakeArray),
		GFnVal(coreLoadLib),
		GFnVal(coreType),
		GFnVal(coreAssert),
		GFnVal(coreFormat),
		GFnVal(coreReadLine),
		GFnVal(coreClone),
		GFnVal(coreError),
		GFnVal(coreIsError),
	)
	return store
}

func corePrint(args ...Value) (Value, error) {
	var s []any
	for _, v := range args {
		s = append(s, v)
	}
	fmt.Fprintln(os.Stdout, s...)
	return NilValue, nil
}

func coreLen(args ...Value) (Value, error) {
	return NilValue, nil
}

func coreType(args ...Value) (Value, error) {
	return NilValue, nil
}

func coreFormat(args ...Value) (Value, error) {
	return NilValue, nil
}

func coreAssert(args ...Value) (Value, error) {
	argsLength := len(args)
	if argsLength == 1 {
		if args[0].Boolean() {
			return BoolVal(true), nil
		}
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n\n\t[%v]\n\tMessage : %v\n\n", verror.AssertionErrType, "Generic Assertion Failure Message"))
		return NilValue, err
	}
	if argsLength > 1 {
		if args[0].Boolean() {
			return BoolVal(true), nil
		}
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n\n\t[%v]\n\tMessage : %v\n\n", verror.AssertionErrType, args[1].String()))
		return NilValue, err
	}
	err := fmt.Errorf("%s", fmt.Sprintf("\n\n\n\t[%v]\n\tMessage : %v\n\n", verror.AssertionErrType, "Generic Assertion Failure Message"))
	return NilValue, err
}

func coreAppend(args ...Value) (Value, error) {
	return NilValue, nil
}

func coreMakeArray(args ...Value) (Value, error) {
	return ArrayVal(&Array{}), nil
}

func coreReadLine(args ...Value) (Value, error) {
	return NilValue, nil
}

func coreClone(args ...Value) (Value, error) {
	if len(args) > 0 {
		return args[0].Clone(), nil
	}
	return NilValue, nil
}

func coreLoadLib(args ...Value) (Value, error) {
	return NilValue, nil
}

func coreError(args ...Value) (Value, error) {
	return ErrorVal(IntVal(42)), nil
}

func coreIsError(args ...Value) (Value, error) {
	return BoolVal(false), nil
}

func IsMemberOfWithTValue(args ...Value) (Value, error) {
	return BoolVal(false), nil
}

func pauseExecution(message string) {
	fmt.Printf("\n\n\n\t\tExecution Paused")
	fmt.Printf("\n\t\t%v", message)
	fmt.Printf("\n\n\n")
	fmt.Scanf(" ")
}
