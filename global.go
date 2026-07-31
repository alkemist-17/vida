package vida

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alkemist-17/vida/token"
)

var Nil = NilValue{}

const True = Bool(true)

const False = Bool(false)

type ExtensionsLoader map[string]func() Value

type MapScriptIPLine map[string]map[int]uint

const (
	VidaFileExtension       = ".vida"
	errorMessageFieldName   = "message"
	DefaultInputPrompt      = "Input > "
	foundationInterfaceName = "std."
	EmptyString             = ""
	DefaultValField         = "value"
	DefaultCloseMethodName  = "close"
)

const (
	foundationBuilders  = "builders"
	foundationText      = "text"
	foundationMath      = "math"
	foundationObj       = "object"
	foundationArray     = "array"
	foundationBytes     = "bytes"
	foundationTime      = "time"
	foundationCast      = "cast"
	foundationRand      = "rand"
	foundationIO        = "io"
	foundationOS        = "os"
	foundationException = "exception"
	foundationCO        = "co"
	foundationHttp      = "http"
	foundationJSON      = "json"
	foundationTask      = "task"
	foundationRegex     = "re"
	foundationStyle     = "style"
)

const (
	VIDAPATH        = "VIDAPATH"
	ProjectCellsDir = "cells"
	VidaPathDirName = "vida-cells"
	downloadTimeout = 15 * time.Second
)

const (
	assertionFailureDefaultMessage = "no assumption was given for the assertion"
	exceptionDefaultMessage        = "no reason was given for the raised exception"
)

func createNewMapScriptIPLine(scriptID string) MapScriptIPLine {
	mapScriptIPLine := make(MapScriptIPLine)
	mapScriptIPLine[scriptID] = make(map[int]uint)
	return mapScriptIPLine
}

func stringWithVisited(v Value, visited map[uintptr]bool) string {
	switch c := v.(type) {
	case *Array:
		return c.stringify(visited)
	case *Object:
		return c.stringify(visited)
	default:
		return v.String()
	}
}

var coreLibNames = []string{
	"print",
	"newArray",
	"load",
	"assert",
	"fmt",
	"input",
	"error",
}

func loadCoreLib(store *[]Value, extensionsLoader ExtensionsLoader) *[]Value {
	*store = append(*store,
		NativeFunction(corePrint),
		NativeFunction(coreNewArray),
		generateLoadFunction(extensionsLoader),
		NativeFunction(coreAssert),
		NativeFunction(coreFormat),
		NativeFunction(coreReadLine),
		NativeFunction(coreError),
	)
	return store
}

func corePrintOld(args ...Value) (Value, error) {
	VFprintln(os.Stdout, args...)
	return Nil, nil
}

func corePrint(ctx *Context, args ...Value) (Value, error) {
	for i, arg := range args {
		if i > 0 {
			io.WriteString(os.Stdout, " ")
		}
		io.WriteString(os.Stdout, resolveDisplay(ctx, arg))
	}
	io.WriteString(os.Stdout, "\n")
	return Nil, nil
}

func coreLen(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *Array:
			return Integer(len(v.Value)), nil
		case *Object:
			lobj := len(v.Value)
			if lobj == 0 {
				return Integer(lobj), nil
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
	return Nil, nil
}

func coreType(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		return &String{Value: args[0].Type()}, nil
	}
	return Nil, nil
}

func coreFormat(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		switch v := args[0].(type) {
		case *String:
			s, e := VSprintf(v.Value, args[1:]...)
			return &String{Value: s}, e
		}
	}
	return Nil, nil
}

// coreAssert checks an assumption that must be true.
//
// Signatures:
//
//	assert(condition)                    — halt on failure with default message
//	assert(condition, "message")         — halt on failure with custom message
//	assert(condition, handler)           — call handler on failure; handler decides
//	assert(condition, "message", handler) — call handler on failure; handler decides
//
// The handler receives a structured info object with the following fields:
//
//	{
//	    passed    = false,
//	    message   = "<assertion message>",
//	    file      = "<script path / context ID>",
//	    timestamp = "<RFC-3339 timestamp>"
//	}
//
// Handler return-value semantics ("handler decides"):
//
//	handler returns truthy  → assertion failure is handled; execution continues;
//	                          assert returns False.
//	handler returns falsy   → VM halts with the assertion error (same as no handler).
//	handler raises an error → the handler error is reported to stderr and the VM
//	                          halts with the ORIGINAL assertion error.
//	no handler provided     → VM halts (fully backward-compatible).
func coreAssert(ctx *Context, args ...Value) (Value, error) {
	argsLength := len(args)
	if argsLength == 0 {
		return Nil, fmt.Errorf("\n\n\t[%v]\n\tAssumption: %v",
			AssertionErrType, assertionFailureDefaultMessage)
	}

	assumption := args[0]

	// Fast path: assumption holds
	if assumption.Boolean() {
		return True, nil
	}

	// Parse optional message and handler
	message := assertionFailureDefaultMessage
	var handler Value

	switch {
	case argsLength == 2:
		// assert(cond, handler)  OR  assert(cond, "message")
		if args[1].IsCallable() {
			handler = args[1]
		} else {
			message = args[1].String()
		}
	case argsLength >= 3:
		// assert(cond, "message", handler)
		message = args[1].String()
		if args[2].IsCallable() {
			handler = args[2]
		}
	}

	// Build the assertion error (used if we halt)
	assertErr := fmt.Errorf("\n\n\n\n\t[%v]\n\tAssumption: %v", AssertionErrType, message)

	// No handler → original behaviour: halt immediately
	if handler == nil {
		return Nil, assertErr
	}

	// Build structured info object for the handler
	info := &Object{Value: map[string]Value{
		"passed":    False,
		"message":   &String{Value: message},
		"file":      &String{Value: ctx.contextID},
		"timestamp": &String{Value: time.Now().Format(time.RFC3339)},
	}}

	// ── Invoke the handler ──────────────────────────────────────────
	result, handlerErr := callAssertHandler(ctx, handler, info)

	if handlerErr != nil {
		// The handler itself failed. Report it, but do NOT let it
		// swallow the original assertion failure.
		fmt.Fprintf(os.Stderr, "\n\t⚠ Assertion handler error: %v\n", handlerErr)
		return Nil, assertErr
	}

	// Handler decides: truthy → continue; falsy → halt.
	if result != nil && result.Boolean() {
		return False, nil
	}

	return Nil, assertErr
}

func callAssertHandler(ctx *Context, handler Value, info *Object) (Value, error) {
	switch fn := handler.(type) {
	case *Function:
		return ctx.runFunctionInNewThread(fn, info)
	default:
		return handler.Call(ctx, info)
	}
}

func coreAppend(ctx *Context, args ...Value) (Value, error) {
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
	return Nil, nil
}

func coreNewArray(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return &Array{}, nil
	}

	switch v := args[0].(type) {
	case Integer:
		var init Value = Nil
		if l > 1 {
			init = args[1]
		}
		if v >= 0 && v < MaxMemSize {
			arr := make([]Value, v)
			for i := range v {
				arr[i] = init
			}
			return &Array{Value: arr}, nil
		}
	case *Object:
		if _, hasFrom := v.Value["from"]; hasFrom {
			if _, hasTo := v.Value["to"]; hasTo {

				// Float range
				if fromF, ok := v.Value["from"].(Float); ok {
					if toF, ok := v.Value["to"].(Float); ok {
						stepF := Float(1.0)
						if s, ok := v.Value["step"].(Float); ok {
							stepF = s
						}
						if stepF == 0 {
							return &Array{}, nil
						}
						var xs []Value
						if stepF > 0 {
							for i := fromF; i <= toF; i += stepF {
								xs = append(xs, i)
							}
						} else {
							for i := fromF; i >= toF; i += stepF {
								xs = append(xs, i)
							}
						}
						return &Array{Value: xs}, nil
					}
				}

				// Integer range
				if from, ok := v.Value["from"].(Integer); ok {
					if to, ok := v.Value["to"].(Integer); ok {
						step := Integer(1)
						if s, ok := v.Value["step"].(Integer); ok {
							step = s
						}
						if step == 0 {
							return &Array{}, nil
						}
						if step > 0 {
							if from > to {
								return &Array{}, nil
							}
							var xs []Value
							for i := from; i <= to; i += step {
								xs = append(xs, i)
							}
							return &Array{Value: xs}, nil
						}
						if step < 0 {
							if from < to {
								return &Array{}, nil
							}
							var xs []Value
							for i := from; i >= to; i += step {
								xs = append(xs, i)
							}
							return &Array{Value: xs}, nil
						}
					}
				}

				goto common
			}
		}
		if ls, ok := v.Value["linspace"].(*Object); ok {
			if fromF, ok := ls.Value["from"].(Float); ok {
				if toF, ok := ls.Value["to"].(Float); ok {
					if n, ok := ls.Value["n"].(Integer); ok && n > 1 && n < MaxMemSize {
						open := false
						if o, ok := ls.Value["open"].(Bool); ok {
							open = bool(o)
						}
						count := n
						if open {
							count = n
						}
						xs := make([]Value, count)
						steps := float64(n - 1)
						if open {
							steps = float64(n)
						}
						for i := range count {
							xs[i] = Float(float64(fromF) + float64(i)*(float64(toF)-float64(fromF))/steps)
						}
						return &Array{Value: xs}, nil
					}
				}
			}
		}
		if size, ok := v.Value["len"].(Integer); ok && size >= 0 && size < MaxMemSize {
			capSize := size
			if c, ok := v.Value["cap"].(Integer); ok && c > size {
				capSize = c
			}
			if capSize >= MaxMemSize {
				return &Array{}, nil
			}

			A := make([]Value, size, capSize)

			if val, ok := v.Value["val"]; ok {
				clone := false
				if cl, ok := v.Value["clone"].(Bool); ok {
					clone = bool(cl)
				}
				if clone {
					for i := range size {
						A[i] = val.Clone()
					}
				} else {
					for i := range size {
						A[i] = val
					}
				}
			} else if random, ok := v.Value["random"].(*String); ok {
				A := make([]Value, size)
				switch random.Value {
				case "string":
					for i := range size {
						nanoid, _ := randNanoID(ctx, Integer(nanoIDMaxSize))
						A[i] = nanoid
					}
				case "int":
					for i := range size {
						n, _ := randN(ctx)
						A[i] = n
					}
				case "float":
					for i := range size {
						A[i] = Float(rand.Float64())
					}
				case "bool":
					for i := range size {
						A[i] = Bool(rand.IntN(2) == 1)
					}
				case "bytes":
					for i := range size {
						A[i] = Integer(rand.IntN(256))
					}
				default:
					for i := range size {
						A[i] = Nil
					}
				}
				return &Array{Value: A}, nil
			} else {
				for i := range size {
					A[i] = Nil
				}
			}

			if cl, ok := v.Value["clip"].(Bool); ok && bool(cl) {
				A = slices.Clip(A)
			}

			return &Array{Value: A}, nil
		}
		if seqName, ok := v.Value["seq"].(*String); ok {
			if n, ok := v.Value["n"].(Integer); ok && n > 0 && n < MaxMemSize {
				switch seqName.Value {
				case "fibonacci":
					A := make([]Value, n)
					a, b := Integer(0), Integer(1)
					for i := range n {
						A[i] = a
						a, b = b, a+b
					}
					return &Array{Value: A}, nil
				case "primes":
					A := make([]Value, 0, n)
					candidate := Integer(2)
					for Integer(len(A)) < n {
						if isPrime(candidate) {
							A = append(A, candidate)
						}
						candidate++
					}
					return &Array{Value: A}, nil
				case "squares":
					A := make([]Value, n)
					for i := range n {
						A[i] = Integer(i * i)
					}
					return &Array{Value: A}, nil
				case "cubes":
					A := make([]Value, n)
					for i := range n {
						A[i] = Integer(i * i * i)
					}
					return &Array{Value: A}, nil
				case "triangular":
					A := make([]Value, n)
					for i := range n {
						A[i] = Integer(i * (i + 1) / 2)
					}
					return &Array{Value: A}, nil
				case "catalan":
					A := make([]Value, n)
					for i := range n {
						A[i] = catalanNumber(Integer(i))
					}
					return &Array{Value: A}, nil
				case "powers2":
					A := make([]Value, n)
					for i := range n {
						A[i] = Integer(1) << uint(i)
					}
					return &Array{Value: A}, nil
				case "factorial":
					A := make([]Value, n)
					f := Integer(1)
					for i := range n {
						if i > 0 {
							f *= Integer(i)
						}
						A[i] = f
					}
					return &Array{Value: A}, nil
				case "evens":
					A := make([]Value, n)
					for i := range n {
						A[i] = Integer(i * 2)
					}
					return &Array{Value: A}, nil
				case "odds":
					A := make([]Value, n)
					for i := range n {
						A[i] = Integer(i*2 + 1)
					}
					return &Array{Value: A}, nil
				}
			}
		}
		if src, ok := v.Value["repeat"].(*Array); ok {
			if times, ok := v.Value["times"].(Integer); ok && times > 0 {
				total := Integer(len(src.Value)) * times
				if total >= MaxMemSize {
					return &Array{}, nil
				}
				A := make([]Value, 0, total)
				for range times {
					A = append(A, src.Value...)
				}
				return &Array{Value: A}, nil
			}
		}
		if zipVal, ok := v.Value["zip"].(*Array); ok && len(zipVal.Value) == 2 {
			if arr1, ok := zipVal.Value[0].(*Array); ok {
				if arr2, ok := zipVal.Value[1].(*Array); ok {
					minLen := Integer(min(len(arr1.Value), len(arr2.Value)))
					padMode := false
					var padVal Value = Nil
					if pad, hasPad := v.Value["pad"]; hasPad {
						padMode = true
						padVal = pad
					}
					maxLen := Integer(max(len(arr1.Value), len(arr2.Value)))
					resultLen := minLen
					if padMode {
						resultLen = maxLen
					}
					if resultLen >= MaxMemSize {
						return &Array{}, nil
					}
					A := make([]Value, resultLen)
					for i := range resultLen {
						pair := make([]Value, 2)
						if int(i) < len(arr1.Value) {
							pair[0] = arr1.Value[i]
						} else {
							pair[0] = padVal
						}
						if int(i) < len(arr2.Value) {
							pair[1] = arr2.Value[i]
						} else {
							pair[1] = padVal
						}
						A[i] = &Array{Value: pair}
					}
					return &Array{Value: A}, nil
				}
			}
		}
		if nested, ok := v.Value["flatten"].(*Array); ok {
			var A []Value
			for _, item := range nested.Value {
				if inner, ok := item.(*Array); ok {
					A = append(A, inner.Value...)
				} else {
					A = append(A, item)
				}
				if Integer(len(A)) >= MaxMemSize {
					return &Array{}, nil
				}
			}
			return &Array{Value: A}, nil
		}
		if obj, ok := v.Value["keys"].(*Object); ok {
			it := obj.Iterator(ctx).(Iterator)
			A := make([]Value, 0, len(obj.Value))
			for it.Next() {
				A = append(A, it.Key(ctx))
			}
			return &Array{Value: A}, nil
		}

		if obj, ok := v.Value["values"].(*Object); ok {
			it := obj.Iterator(ctx).(Iterator)
			A := make([]Value, 0, len(obj.Value))
			for it.Next() {
				A = append(A, it.Value(ctx))
			}
			return &Array{Value: A}, nil
		}

		if obj, ok := v.Value["pairs"].(*Object); ok {
			it := obj.Iterator(ctx).(Iterator)
			A := make([]Value, 0, len(obj.Value))
			for it.Next() {
				pair := &Array{Value: []Value{it.Key(ctx), it.Value(ctx)}}
				A = append(A, pair)
			}
			return &Array{Value: A}, nil
		}
		if arr, ok := v.Value["grow"].(*Array); ok {
			if by, ok := v.Value["by"].(Integer); ok && 0 < by && by < MaxMemSize {
				clone := arr.Clone().(*Array)
				clone.Value = slices.Grow(clone.Value, int(by))
				return clone, nil
			}
		}
		if arr, ok := v.Value["clip"].(*Array); ok {
			clone := arr.Clone().(*Array)
			clone.Value = slices.Clip(clone.Value)
			return clone, nil
		}
	case *String:
		var i int
		it := v.Iterator(ctx).(Iterator)
		A := make([]Value, utf8.RuneCountInString(v.Value))
		for it.Next() {
			A[i] = it.Value(ctx)
			i++
		}
		return &Array{Value: A}, nil
	case *Bytes:
		A := make([]Value, len(v.Value))
		for i, b := range v.Value {
			A[i] = Integer(b)
		}
		return &Array{Value: A}, nil
	case Float:
		bits := math.Float64bits(float64(v))
		sign := Integer((bits >> 63) & 1)
		exponent := Integer((bits>>52)&0x7FF) - 1023
		mantissa := Integer(bits & 0x000FFFFFFFFFFFFF)
		A := []Value{sign, exponent, mantissa}
		return &Array{Value: A}, nil
	case *Array:
		return v.Clone(), nil
	}
common:
	if obj, ok := args[0].(*Object); ok {
		var i int
		it := obj.Iterator(ctx).(Iterator)
		A := make([]Value, len(obj.Value))
		for it.Next() {
			B := []Value{it.Key(ctx), it.Value(ctx)}
			A[i] = &Array{Value: B}
			i++
		}
		return &Array{Value: A}, nil
	}

	return &Array{}, nil
}

func isPrime(n Integer) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := Integer(3); i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// catalanNumber computes the nth Catalan number: C(n) = (2n)! / ((n+1)! * n!)
func catalanNumber(n Integer) Integer {
	if n == 0 {
		return Integer(1)
	}
	result := Integer(1)
	for i := range n {
		result = result * (2*n - i) / (i + 1)
	}
	return result / (n + 1)
}

func coreReadLine(ctx *Context, args ...Value) (Value, error) {
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
		return Nil, err
	}
	return Nil, nil
}

func coreClone(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		return args[0].Clone(), nil
	}
	return Nil, nil
}

func generateLoadFunction(extensionsLoader ExtensionsLoader) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if extensionName, ok := args[0].(*String); ok {
				if strings.HasPrefix(extensionName.Value, foundationInterfaceName) {
					if ctx.extensionsCache == nil {
						ctx.extensionsCache = make(map[string]*Object, 10)
					}
					if m, isPresent := ctx.extensionsCache[extensionName.Value]; isPresent {
						return m, nil
					}
					var module Value
					switch extensionName.Value[len(foundationInterfaceName):] {
					case foundationText:
						module = loadFoundationText()
					case foundationMath:
						module = loadFoundationMath()
					case foundationObj:
						module = loadObjectLib()
					case foundationArray:
						module = loadFoundationArray()
					case foundationBytes:
						module = loadFoundationBytes()
					case foundationTime:
						module = loadFoundationTime()
					case foundationCast:
						module = loadFoundationCasting()
					case foundationRand:
						module = loadFoundationRandom()
					case foundationIO:
						module = loadFoundationIO()
					case foundationOS:
						module = loadFoundationOS()
					case foundationException:
						module = loadFoundationException()
					case foundationCO:
						module = loadFoundationCoroutine()
					case foundationHttp:
						module = loadFoundationHttpClient()
					case foundationJSON:
						module = loadFoundationJSON()
					case foundationTask:
						module = loadFoundationTask()
					case foundationRegex:
						module = loadFoundationRegexp()
					case foundationStyle:
						module = loadFoundationStyle()
					case foundationBuilders:
						module = loadFoundationBuilders()
					default:
						module = Nil
						return &VidaError{Message: &String{Value: fmt.Sprintf("load function could not find the module '%v'", extensionName.Value)}}, nil
					}
					ctx.extensionsCache[extensionName.Value] = module.(*Object)
					return module, nil
				} else if extensionsLoader != nil {
					if ctx.extensionsCache == nil {
						ctx.extensionsCache = make(map[string]*Object, 10)
					}
					if m, isPresent := ctx.extensionsCache[extensionName.Value]; isPresent {
						return m, nil
					}
					if l, isPresent := extensionsLoader[extensionName.Value]; isPresent {
						module := l()
						ctx.extensionsCache[extensionName.Value] = module.(*Object)
						return module, nil
					}
				}
				return &VidaError{Message: &String{Value: fmt.Sprintf("load function could not find the module '%v'", extensionName.Value)}}, nil
			}
		}
		return &VidaError{Message: &String{Value: "load function should have one argument of type string"}}, nil
	}
}

func coreError(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		return &VidaError{Message: args[0]}, nil
	}
	return &VidaError{Message: Nil}, nil
}

func coreIsError(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(*VidaError)
		return Bool(ok), nil
	}
	return False, nil
}

func coreIsNil(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(NilValue)
		return Bool(ok), nil
	}
	return False, nil
}

func coreGetVTable(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		return args[0].GetVTable(ctx), nil
	}
	return Nil, nil
}

func coreExtendVTable(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 2:
		if extension, ok := args[1].(*Object); ok {
			if vt, hasVTable := args[0].GetVTable(ctx).(*Object); hasVTable {
				for k, x := range extension.Value {
					if _, isPresent := vt.Value[k]; !isPresent {
						vt.Value[k] = x
					}
				}
				return args[0], nil
			}
		}
	case 3:
		messageName, okMessage := args[1].(*String)
		if okMessage && bool(args[2].IsCallable()) {
			if vt, hasVTable := args[0].GetVTable(ctx).(*Object); hasVTable {
				if _, isPresent := vt.Value[messageName.Value]; !isPresent {
					vt.Value[messageName.Value] = args[2]
				}
				return args[0], nil
			}
		}
	}
	return &VidaError{Message: &String{Value: "extendvt expected three args: value, string and function or two args: value, object"}}, nil
}

func StringLength(input *String) Integer {
	if input.Runes == nil {
		input.Runes = []rune(input.Value)
	}
	return Integer(len(input.Runes))
}

func isMemberOf(ctx *Context, args ...Value) (Bool, error) {
	if len(args) > 1 {
		switch collection := args[1].(type) {
		case *Array:
			item := args[0]
			for _, v := range collection.Value {
				if item.Equals(ctx, v) {
					return True, nil
				}
			}
			return False, nil
		case *Object:
			item := args[0]
			for k := range collection.Value {
				if item.Equals(ctx, &String{Value: k}) {
					return True, nil
				}
			}
			return False, nil
		case *String:
			item := args[0]
			for _, char := range collection.Runes {
				if item.Equals(ctx, &String{Value: string(char)}) {
					return True, nil
				}
			}
			return False, nil
		case *Bytes:
			item := args[0]
			for _, b := range collection.Value {
				if item.Equals(ctx, Integer(b)) {
					return True, nil
				}
			}
			return False, nil
		}
	}
	return False, nil
}

func tokenBinopToString(t token.Token) *String {
	switch t {
	case token.ADD:
		return &String{Value: "add"}
	case token.SUB:
		return &String{Value: "sub"}
	case token.MUL:
		return &String{Value: "mul"}
	case token.DIV:
		return &String{Value: "div"}
	case token.REM:
		return &String{Value: "rem"}
	case token.POW:
		return &String{Value: "pow"}
	case token.EQ:
		return &String{Value: "eq"}
	case token.LT:
		return &String{Value: "lt"}
	case token.LE:
		return &String{Value: "le"}
	case token.GT:
		return &String{Value: "gt"}
	case token.GE:
		return &String{Value: "ge"}
	case token.BOR:
		return &String{Value: "bor"}
	case token.BAND:
		return &String{Value: "band"}
	case token.BXOR:
		return &String{Value: "bxor"}
	case token.BSHL:
		return &String{Value: "bshl"}
	case token.BSHR:
		return &String{Value: "bshr"}
	default:
		return &String{Value: EmptyString}
	}
}

func tokenPrefixToString(t token.Token) *String {
	switch t {
	case token.ADD:
		return &String{Value: "pos"}
	case token.SUB:
		return &String{Value: "neg"}
	case token.TILDE:
		return &String{Value: "tilde"}
	default:
		return &String{Value: EmptyString}
	}
}

func pressEnterToContinue() {
	fmt.Print("\n\nPress 'Enter' to continue  ")
	fmt.Scanf(" ")
}

// DownloadCellTo fetches rawURL and writes it verbatim to destPath,
// creating any needed directories. It does not check for an existing
// cached file — callers decide caching policy.
func DownloadCellTo(rawURL, destPath string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid module url %q: %v", rawURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported module scheme %q in %q", u.Scheme, rawURL)
	}

	client := http.Client{Timeout: downloadTimeout}
	resp, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("could not download module %q: %v", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not download module %q: server responded %v", rawURL, resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("could not create directory for %q: %v", rawURL, err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not save module %q: %v", rawURL, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		os.Remove(destPath)
		return fmt.Errorf("could not save module %q: %v", rawURL, err)
	}

	return nil
}

func resolveDisplay(ctx *Context, v Value) string {
	if v == nil {
		return Nil.String()
	}
	method := v.LookUp(ctx, &String{Value: "toString"})
	if method.Equals(ctx, Nil) {
		return v.String()
	}
	var result Value
	var err error
	switch fn := method.(type) {
	case *Function:
		result, err = ctx.runFunctionInNewThread(fn, v)
	case NativeFunction:
		result, err = fn.Call(ctx, v)
	default:
		return v.String()
	}
	if err != nil || result == nil {
		return v.String()
	}
	return result.String()
}
