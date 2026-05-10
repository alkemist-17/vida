package vida

import (
	"bufio"
	"fmt"
	"math/rand/v2"
	"os"
	"reflect"
	"strings"

	"github.com/alkemist-17/vida/verror"
)

var NilValue = Nil{}

type LibsLoader map[string]func() Value

type ErrorInfo map[string]map[int]uint

var extensionlibsLoader LibsLoader

var __proto string = initProtName

const (
	initProtName = "$__proto$"

	globalStateIndex = 0

	errorMessageFieldName = "message"

	DefaultInputPrompt = "Input > "

	foundationInterfaceName = "std/"

	__str = "__str"

	__call = "__call"

	__type = "__type"

	__get = "__get"

	__set = "__set"

	__del = "__del"

	__getproto = "__getproto"

	__setproto = "__setproto"

	__delproto = "__delproto"

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

type threadPool struct {
	ThreadMap map[int]*Thread
	Key       int
}

func newThreadPool() *threadPool {
	return &threadPool{
		ThreadMap: make(map[int]*Thread),
	}
}

func (tp *threadPool) getThread() *Thread {
	if t, ok := tp.ThreadMap[tp.Key]; ok {
		tp.Key++
		return t
	}
	t := newThread(nil, ((*clbu)[globalStateIndex].(*GlobalState)).Script)
	tp.ThreadMap[tp.Key] = t
	tp.Key++
	return t
}

func (tp *threadPool) releaseThread() {
	tp.Key--
}

type GlobalState struct {
	*VM
	Main    *Thread
	Current *Thread
	Pool    *threadPool
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
		GFn(corePrint),
		GFn(coreLen),
		GFn(coreAppend),
		GFn(coreMakeArray),
		GFn(coreLoadLib),
		GFn(coreType),
		GFn(coreAssert),
		GFn(coreFormat),
		GFn(coreReadLine),
		GFn(coreClone),
		GFn(coreError),
		GFn(coreIsError),
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

func coreType(args ...Value) (Value, error) {
	if len(args) > 0 {
		return &String{Value: args[0].Type()}, nil
	}
	return NilValue, nil
}

func coreFormat(args ...Value) (Value, error) {
	if len(args) > 1 {
		switch v := args[0].(type) {
		case *String:
			s, e := FormatValue(v.Value, args[1:]...)
			return &String{Value: s}, e
		}
	}
	return NilValue, nil
}

func coreAssert(args ...Value) (Value, error) {
	argsLength := len(args)
	if argsLength == 1 {
		if args[0].Boolean() {
			return Bool(true), nil
		}
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n\n\t[%v]\n\tMessage : %v\n\n", verror.AssertionErrType, "Generic Assertion Failure Message"))
		return NilValue, err
	}
	if argsLength > 1 {
		if args[0].Boolean() {
			return Bool(true), nil
		}
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n\n\t[%v]\n\tMessage : %v\n\n", verror.AssertionErrType, args[1].String()))
		return NilValue, err
	}
	err := fmt.Errorf("%s", fmt.Sprintf("\n\n\n\t[%v]\n\tMessage : %v\n\n", verror.AssertionErrType, "Generic Assertion Failure Message"))
	return NilValue, err
}

func coreAppend(args ...Value) (Value, error) {
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

func coreMakeArray(args ...Value) (Value, error) {
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
			if from, ok := v.Value["from"].(Integer); ok {
				if to, ok := v.Value["to"].(Integer); ok {
					if step, ok := v.Value["step"].(Integer); ok && step > 0 {
						if from < to {
							var xs []Value
							for i := from; i <= to; i += step {
								xs = append(xs, i)
							}
							return &Array{Value: xs}, nil
						}
					} else {
						if from < to {
							l := to - from
							l++
							xs := make([]Value, l)
							for i := range l {
								xs[i] = Integer(from)
								from++
							}
							return &Array{Value: xs}, nil
						}
					}
				}
				goto common
			} else if size, ok := v.Value["len"].(Integer); ok && size > 0 && size < verror.MaxMemSize {
				if val, ok := v.Value["val"]; ok {
					if clone, ok := v.Value["clone"].(Bool); ok {
						A := make([]Value, size)
						if clone {
							for i := range size {
								A[i] = val.Clone()
							}
						} else {
							for i := range size {
								A[i] = val
							}
						}
						return &Array{Value: A}, nil
					}
				} else if random, ok := v.Value["random"].(*String); ok {
					A := make([]Value, size)
					switch random.Value {
					case (&String{}).Type():
						for i := range size {
							nanoid, _ := randNanoID(Integer(nanoIDMaxSize))
							A[i] = nanoid
						}
					case Integer(0).Type():
						for i := range size {
							n, _ := randN()
							A[i] = n
						}
					case Float(0).Type():
						for i := range size {
							A[i] = Float(rand.Float64())
						}
					case Bool(true).Type():
						for i := range size {
							n, _ := randN()
							if n.(Integer)%2 == 0 {
								A[i] = Bool(true)
							} else {
								A[i] = Bool(false)
							}
						}
					default:
						for i := range size {
							A[i] = NilValue
						}
					}
					return &Array{Value: A}, nil
				}
			}
		common:
			var i int
			it := v.Iterator().(Iterator)
			A := make([]Value, len(v.Value))
			for it.Next() {
				B := make([]Value, 2)
				B[0] = it.Key()
				B[1] = it.Value()
				A[i] = &Array{Value: B}
				i++
			}
			return &Array{Value: A}, nil
		case *String:
			var i int
			it := v.Iterator().(Iterator)
			A := make([]Value, StringLength(v))
			for it.Next() {
				A[i] = it.Value()
				i++
			}
			return &Array{Value: A}, nil
		case *Bytes:
			A := make([]Value, len(v.Value))
			for i, v := range v.Value {
				A[i] = Integer(v)
			}
			return &Array{Value: A}, nil
		case *Array:
			return v.Clone(), nil
		}
	}
	return &Array{}, nil
}

func coreReadLine(args ...Value) (Value, error) {
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

func coreClone(args ...Value) (Value, error) {
	if len(args) > 0 {
		return args[0].Clone(), nil
	}
	return NilValue, nil
}

func coreLoadLib(args ...Value) (Value, error) {
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
				case "array":
					return loadFoundationArray(), nil
				case "bytes":
					return loadFoundationBytes(), nil
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
				case "co":
					return loadFoundationCoroutine(), nil
				case "http":
					return loadFoundationHttpClient(), nil
				case "json":
					return loadFoundationJSON(), nil
				case "core":
					return loadFoundationCorelib(), nil
				case "task":
					return loadFoundationTask(), nil
				case "re":
					return loadFoundationRegexp(), nil
				case "color":
					return loadFoundationColor(), nil
				}
			} else if l, isPresent := extensionlibsLoader[v.Value]; isPresent {
				return l(), nil
			}
		}
	}
	return NilValue, nil
}

func coreError(args ...Value) (Value, error) {
	if len(args) > 0 {
		return VidaError{Message: args[0]}, nil
	}
	return VidaError{Message: NilValue}, nil
}

func coreIsError(args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(VidaError)
		return Bool(ok), nil
	}
	return Bool(false), nil
}

func coreCopy(args ...Value) (Value, error) {
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
	m := &Object{Value: make(map[string]Value, len((*clbu)))}
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

func pauseExecution(message string) {
	fmt.Printf("\n\n\n\t\tExecution Paused")
	fmt.Printf("\n\t\t%v", message)
	fmt.Printf("\n\n\n")
	fmt.Scanf(" ")
}
