package vida

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alkemist-17/vida/verror"
)

var NilValue = Nil{}

type LibsLoader map[string]func() Value

type ErrorInfo map[string]map[int]uint

var extensionlibsLoader LibsLoader

var __proto string

var __meta string

const (
	globalStateIndex = 0

	DefaultInputPrompt = "Input > "

	foundationInterfaceName = "std/"

	__str = "__str"

	__call = "__call"

	__get = "__get"

	__set = "__set"

	__getmeta = "__getmeta"

	__getproto = "__getproto"

	__setmeta = "__setmeta"

	__setproto = "__setproto"

	__del = "__del"

	__add = "__add"

	__sub = "__sub"

	__mul = "__mul"

	__div = "__div"

	__rem = "__rem"

	__eq = "__eq"

	__neq = "__neq"

	__le = "__le"

	__lt = "__lt"

	__ge = "__ge"

	__gt = "__gt"
)

var clbu *[]Value

type metaobjectThreadPool struct {
	VM  map[int]*VM
	Key int
}

func newThreadPool() *metaobjectThreadPool {
	return &metaobjectThreadPool{
		VM: make(map[int]*VM),
	}
}

func (tp *metaobjectThreadPool) getVM() *VM {
	if vm, ok := tp.VM[tp.Key]; ok {
		tp.Key++
		return vm
	}
	vm := &VM{newThread(nil, ((*clbu)[globalStateIndex].(*GlobalState)).Script, defaultMetaStackSize)}
	tp.VM[tp.Key] = vm
	tp.Key++
	return vm
}

type GlobalState struct {
	*VM
	Main    *Thread
	Current *Thread
	Pool    *metaobjectThreadPool
}

var coreLibNames = []string{
	"--G--",
	"print",
	"len",
	"append",
	"array",
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
		GFn(gfnPrint),
		GFn(gfnLen),
		GFn(gfnAppend),
		GFn(gfnMakeArray),
		GFn(gfnLoadLib),
		GFn(gfnType),
		GFn(gfnAssert),
		GFn(gfnFormat),
		GFn(gfnReadLine),
		GFn(gfnClone),
		GFn(gfnError),
		GFn(gfnIsError),
	)
	return store
}

func gfnPrint(args ...Value) (Value, error) {
	var s []any
	for _, v := range args {
		s = append(s, v)
	}
	fmt.Fprintln(os.Stdout, s...)
	return NilValue, nil
}

func gfnLen(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *Array:
			return Integer(len(v.Value)), nil
		case *Object:
			lobj := len(v.Value)
			if lobj == 0 {
				return Integer(lobj), nil
			}
			if _, ok := v.Value[__proto]; ok {
				lobj--
			}
			if _, ok := v.Value[__meta]; ok {
				lobj--
			}
			return Integer(lobj), nil
		case *String:
			if v.Runes == nil {
				v.Runes = []rune(v.Value)
			}
			return Integer(len(v.Runes)), nil
		case *Bytes:
			return Integer(len(v.Value)), nil
		}
	}
	return NilValue, nil
}

func gfnType(args ...Value) (Value, error) {
	if len(args) > 0 {
		return &String{Value: args[0].Type()}, nil
	}
	return NilValue, nil
}

func gfnFormat(args ...Value) (Value, error) {
	if len(args) > 1 {
		switch v := args[0].(type) {
		case *String:
			s, e := FormatValue(v.Value, args[1:]...)
			return &String{Value: s}, e
		}
	}
	return NilValue, nil
}

func gfnAssert(args ...Value) (Value, error) {
	argsLength := len(args)
	if argsLength == 1 {
		if args[0].Boolean() {
			return NilValue, nil
		}
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n  [%v]\n   Message : %v\n\n", verror.AssertionErrType, "Generic Assertion Failure Message"))
		return NilValue, err
	}
	if argsLength > 1 {
		if args[0].Boolean() {
			return NilValue, nil
		}
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n  [%v]\n   Message : %v\n\n", verror.AssertionErrType, args[1].String()))
		return NilValue, err

	}
	return NilValue, nil
}

func gfnAppend(args ...Value) (Value, error) {
	if len(args) >= 2 {
		switch v := args[0].(type) {
		case *Array:
			v.Value = append(v.Value, args[1:]...)
			return v, nil
		case *Bytes:
			for _, val := range args[1:] {
				if i, ok := val.(Integer); ok {
					v.Value = append(v.Value, byte(i))
				}
			}
			return v, nil
		}
	}
	return NilValue, nil
}

func gfnMakeArray(args ...Value) (Value, error) {
	largs := len(args)
	if largs > 0 {
		switch v := args[0].(type) {
		case Integer:
			var init Value = NilValue
			if largs > 1 {
				init = args[1]
			}
			if v >= 0 && v < verror.MaxMemSize {
				arr := make([]Value, v)
				for i := range v {
					arr[i] = init
				}
				return &Array{Value: arr}, nil
			}
		case *Object:
			if f, ok := v.Value["from"]; ok && f.Type() == "int" {
				if t, ok := v.Value["to"]; ok && t.Type() == "int" {
					from := f.(Integer)
					to := t.(Integer)
					if from < to {
						l := to - from
						xs := make([]Value, l)
						for i := range l {
							xs[i] = Integer(from)
							from++
						}
						return &Array{Value: xs}, nil
					}
				}
			}
		}
	}
	return &Array{}, nil
}

func gfnReadLine(args ...Value) (Value, error) {
	if len(args) > 0 {
		fmt.Print(args[0])
	} else {
		fmt.Print(DefaultInputPrompt)
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		return &String{Value: scanner.Text()}, nil
	}
	if err := scanner.Err(); err != nil {
		return NilValue, err
	}
	return NilValue, nil
}

func gfnClone(args ...Value) (Value, error) {
	if len(args) > 0 {
		return args[0].Clone(), nil
	}
	return NilValue, nil
}

func gfnLoadLib(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			if strings.HasPrefix(v.Value, foundationInterfaceName) {
				switch v.Value[len(foundationInterfaceName):] {
				case "text":
					return loadFoundationText(), nil
				case "math":
					return loadFoundationMath(), nil
				case "object":
					return loadObjectLib(), nil
				case "bin":
					return loadFoundationBinary(), nil
				case "time":
					return loadFoundationTime(), nil
				case "cast":
					return loadFoundationCasting(), nil
				case "rand":
					return loadFoundationRandom(), nil
				case "io":
					return loadFoundationIO(), nil
				case "os":
					return loadFoundationOS(), nil
				case "exception":
					return loadFoundationException(), nil
				case "net":
					return loadFoundationNetworkIO(), nil
				case "co":
					return loadFoundationCoroutine(), nil
				case "core":
					return loadFoundationCorelib(), nil
				}
			} else if l, isPresent := extensionlibsLoader[v.Value]; isPresent {
				return l(), nil
			}
		}
	}
	return NilValue, nil
}

func gfnError(args ...Value) (Value, error) {
	if len(args) > 0 {
		return Error{Message: args[0]}, nil
	}
	return Error{Message: NilValue}, nil
}

func gfnIsError(args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(Error)
		return Bool(ok), nil
	}
	return Bool(false), nil
}

func gfnCopy(args ...Value) (Value, error) {
	if len(args) > 1 {
		switch dst := args[0].(type) {
		case *Array:
			switch src := args[1].(type) {
			case *Array:
				copy(dst.Value, src.Value)
				return dst, nil
			case *Bytes:
				l := len(dst.Value)
				b := len(src.Value)
				if b < l {
					l = b
				}
				for i := 0; i < l; i++ {
					dst.Value[i] = Integer(src.Value[i])
				}
				return dst, nil
			case *String:
				l := len(dst.Value)
				b := len(src.Value)
				if b < l {
					l = b
				}
				if src.Runes == nil {
					src.Runes = []rune(src.Value)
				}
				for i := 0; i < l; i++ {
					dst.Value[i] = &String{Value: string(src.Runes[i])}
				}
				return dst, nil
			}
		case *Bytes:
			switch src := args[1].(type) {
			case *Array:
				l := len(dst.Value)
				b := len(src.Value)
				if b < l {
					l = b
				}
				var i int
				var j int
				for i = 0; i < l; i++ {
					if v, ok := src.Value[i].(Integer); ok {
						dst.Value[j] = byte(v)
						j++
					}
				}
				return dst, nil
			case *Bytes:
				copy(dst.Value, src.Value)
				return dst, nil
			case *String:
				copy(dst.Value, []byte(src.Value))
				return dst, nil
			}
		}
	}
	return NilValue, nil
}

func DeepEqual(args ...Value) (Value, error) {
	if len(args) > 1 {
		return Bool(reflect.DeepEqual(args[0], args[1])), nil
	}
	return NilValue, nil
}

func loadFoundationCorelib() Value {
	m := &Object{Value: make(map[string]Value)}
	for i := 0; i < len((*clbu)); i++ {
		m.Value[coreLibNames[i]] = (*clbu)[i]
	}
	return m
}

func StringLength(input *String) Integer {
	if input.Runes == nil {
		input.Runes = []rune(input.Value)
	}
	return Integer(len(input.Runes))
}

func IsMemberOf(args ...Value) (Value, error) {
	if len(args) > 1 {
		switch collection := args[1].(type) {
		case *Array:
			item := args[0]
			for _, v := range collection.Value {
				if item.Equals(v) {
					return Bool(true), nil
				}
			}
			return Bool(false), nil
		case *Object:
			item := args[0]
			for k := range collection.Value {
				if item.Equals(&String{Value: k}) {
					return Bool(true), nil
				}
			}
			return Bool(false), nil
		case *String:
			item := args[0]
			for _, char := range collection.Runes {
				if item.Equals(&String{Value: string(char)}) {
					return Bool(true), nil
				}
			}
			return Bool(false), nil
		case *Bytes:
			item := args[0]
			for _, b := range collection.Value {
				if item.Equals(Integer(b)) {
					return Bool(true), nil
				}
			}
			return Bool(false), nil
		}
	}
	return NilValue, nil
}

var coreLibDescription = []string{
	``,
	`
	Print one or more values.
	Commas between values are optional.
	Examples: print(v0 v1 v2), print(a, b, c) -> nil
	`,
	`
	Return an integer representing the length of arrays, 
	objects, bytes or strings. In case of a string value, 
	the function returns the number of unicode codepoints.
	Example: len(value) -> int
	`,
	`
	Add one of more values at the end of an array.
	Return the array passed as first argument.
	Examples: let xs be an array, then 
	append(xs, value), append(xs a b c) -> xs
	If the array is an array of bytes, only convert
	integer values to uint8 bits values.
	`,
	`
	Create an array. 
	Receive 0, 1 or 2 arguments. 
	Whith zero arguments, return an empty array. 
	With 1 argument n of type intenger,
	return an array of n elements all initialized to nil.
	With 2 argumeents (n, m), with n of type integer,
	and m of type T, return an array of n elements all 
	initialized to the m value.
	Examples: 
		array() -> [],
		array(10) -> [nil, ... , nil],
		array(n, v) -> [v, v, ... , v]
	`,
	`
	Load a library from the package lib.
	Those librarires are written in go, and they are
	intended to extend the functionality of the language.
	Receive an argument s of type string, and return an object
	containing the lib functionality.
	If the library denoted by s does not exist, return nil.
	Example: load("math"), load("random")
	`,
	`
	Return the type of a value as string.
	Example: type(123) -> "int".
	A suggested convention to avoid type name clashes,
	is to name types other than the built-in ones,
	with this pattern {lib name} + . + {type name}.
	Example: type(43) -> "int" (Built-in type)
	type(file) -> "io.file" (From the io library) 
	`,
	`
	Make an assertion about an expression.
	If the expression represents a false value, then
	It fails and returns a run time error.
	Otherwise, it returns a nil value.
	Example: assert(false), assert(true), assert(not nil)
	`,
	`
	Return a string with the given format. String interpolation
	can be done very well with it.
	The most common verb formats are: %v, %T, %f, %d, %b, %x
	Example: format("This is the number %v", 15)
	`,
	`
	Show a prompt and wait for the input from the user.
	It is a blocking function.
	If no prompt is given, it shows a default one.
	Return a string representing the user input.
	Example: input("Write something here") -> string
	`,
	`
	Make a copy of value-semantics values or a deep copy 
	of reference-semantics values.
	Example: clone(someValue)
	`,
	`
	Create an error value. An error value may be used to signal
	some behavior considered an error. The boolean value of an
	error value is always false. When an argument is given, it will
	be the printable message for the client of the functionality
	with the unexpected behavior.
	Example: 
		ret error(message)
	        let result = f()
		if not result {handle the error} or
		if result {handle the returned value}
	`,
	`
	Help to explicitly check for an error value.
	Example: if isError(value) {handle the error here}
	`,
}

func PrintCoreLibInformation() {
	fmt.Printf("Core:\n\n")
	for i := 1; i < len(coreLibNames); i++ {
		fmt.Printf("  %v %v\n\n", coreLibNames[i], coreLibDescription[i])
	}
}

func pauseExecution(message string) {
	fmt.Printf("\n\n\n\t\tExecution Paused")
	fmt.Printf("\n\t\t%v", message)
	fmt.Printf("\n\n\n")
	fmt.Scanf(" ")
}
