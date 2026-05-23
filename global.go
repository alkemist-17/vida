package vida

import (
	"bufio"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/alkemist-17/vida/verror"
)

var NilValue = Nil{}

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

	__pow = "__pow"

	__eq = "__eq"

	__le = "__le"

	__lt = "__lt"

	__ge = "__ge"

	__gt = "__gt"

	__umin = "__umin"

	__uplus = "__uplus"
)

const (
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
	foundationCore      = "core"
	foundationTask      = "task"
	foundationRegex     = "re"
	foundationColor     = "color"
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
		GFn(coreNewArray),
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

func coreNewArray(args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return &Array{}, nil
	}

	switch v := args[0].(type) {

	// ── newArray(N) or newArray(N, initVal)
	case Integer:
		var init Value = NilValue
		if l > 1 {
			init = args[1]
		}
		if v >= 0 && v < verror.MaxMemSize {
			arr := make([]Value, v)
			for i := range v {
				arr[i] = init
			}
			return &Array{Value: arr}, nil
		}

	// ── newArray({ ... })
	case *Object:

		// ── { from, to } or { from, to, step } ──────────────────────────────
		// Ranges: integers or floats, positive or negative step.
		//   newArray({ from=1,  to=5  })           → [1,2,3,4,5]
		//   newArray({ from=10, to=1, step=-2 })   → [10,8,6,4,2]
		//   newArray({ from=0.0, to=1.0, step=0.25 }) → float range
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
						// Positive step: from must be < to
						if step > 0 {
							if from > to {
								return &Array{}, nil // silent no-op
							}
							var xs []Value
							for i := from; i <= to; i += step {
								xs = append(xs, i)
							}
							return &Array{Value: xs}, nil
						}
						// Negative step: countdown
						if step < 0 {
							if from < to {
								return &Array{}, nil // silent no-op
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

		// ── { linspace } ─────────────────────────────────────────────────────
		// Evenly spaced floats.
		//   newArray({ linspace={ from=0.0, to=1.0, n=5 } })       → [0.0,0.25,0.5,0.75,1.0]
		//   newArray({ linspace={ from=0.0, to=1.0, n=4, open=true } }) → [0.0,0.25,0.5,0.75]
		if ls, ok := v.Value["linspace"].(*Object); ok {
			if fromF, ok := ls.Value["from"].(Float); ok {
				if toF, ok := ls.Value["to"].(Float); ok {
					if n, ok := ls.Value["n"].(Integer); ok && n > 1 && n < verror.MaxMemSize {
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

		// ── { len, val?, clone?, cap?, clip? } ───────────────────────────────
		// Fill with a value; optionally clone each element independently.
		//   newArray({ len=5, val=0 })                  → [0,0,0,0,0]
		//   newArray({ len=3, val=[], clone=true })      → three distinct empty arrays
		//   newArray({ len=5, cap=100 })                 → len 5, capacity 100
		//   newArray({ len=5, val=0, clip=true })        → tight allocation
		if size, ok := v.Value["len"].(Integer); ok && size >= 0 && size < verror.MaxMemSize {
			capSize := size
			if c, ok := v.Value["cap"].(Integer); ok && c > size {
				capSize = c
			}
			if capSize >= verror.MaxMemSize {
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
				// ── { random, len } ──────────────────────────────────────────────────
				// Random arrays of typed values.
				//   newArray({ len=5, random="int" })     → random integers
				//   newArray({ len=5, random="float" })   → random [0.0,1.0)
				//   newArray({ len=5, random="bool" })    → random booleans
				//   newArray({ len=5, random="string" })  → random nano-IDs
				//   newArray({ len=5, random="byte" })    → random bytes (0-255 as Integer)
				A := make([]Value, size)
				switch random.Value {
				case "string":
					for i := range size {
						nanoid, _ := randNanoID(Integer(nanoIDMaxSize))
						A[i] = nanoid
					}
				case "int":
					for i := range size {
						n, _ := randN()
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
				case "byte":
					for i := range size {
						A[i] = Integer(rand.IntN(256))
					}
				default:
					for i := range size {
						A[i] = NilValue
					}
				}
				return &Array{Value: A}, nil
			} else {
				// default fill: NilValue
				for i := range size {
					A[i] = NilValue
				}
			}

			if cl, ok := v.Value["clip"].(Bool); ok && bool(cl) {
				A = slices.Clip(A)
			}

			return &Array{Value: A}, nil
		}

		// ── { seq, n } ───────────────────────────────────────────────────────
		// Named mathematical sequences.
		//   newArray({ seq="fibonacci",  n=10 }) → [0,1,1,2,3,5,8,13,21,34]
		//   newArray({ seq="primes",     n=8  }) → [2,3,5,7,11,13,17,19]
		//   newArray({ seq="squares",    n=6  }) → [0,1,4,9,16,25]
		//   newArray({ seq="cubes",      n=5  }) → [0,1,8,27,64]
		//   newArray({ seq="triangular", n=6  }) → [0,1,3,6,10,15]
		//   newArray({ seq="catalan",    n=7  }) → [1,1,2,5,14,42,132]
		//   newArray({ seq="powers2",    n=8  }) → [1,2,4,8,16,32,64,128]
		//   newArray({ seq="factorial",  n=7  }) → [1,1,2,6,24,120,720]
		if seqName, ok := v.Value["seq"].(*String); ok {
			if n, ok := v.Value["n"].(Integer); ok && n > 0 && n < verror.MaxMemSize {
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

		// ── { repeat, times } ────────────────────────────────────────────────
		// Repeat a sub-array N times.
		//   newArray({ repeat=[1,2,3], times=3 }) → [1,2,3,1,2,3,1,2,3]
		if src, ok := v.Value["repeat"].(*Array); ok {
			if times, ok := v.Value["times"].(Integer); ok && times > 0 {
				total := Integer(len(src.Value)) * times
				if total >= verror.MaxMemSize {
					return &Array{}, nil
				}
				A := make([]Value, 0, total)
				for range times {
					A = append(A, src.Value...)
				}
				return &Array{Value: A}, nil
			}
		}

		// ── { zip = [arr1, arr2] } ───────────────────────────────────────────
		// Interleave two arrays into pairs.
		//   newArray({ zip=[a, b] }) → [[a0,b0],[a1,b1],...]
		// Shorter array determines length; use { zip=[a,b], pad=nil } to fill gaps.
		if zipVal, ok := v.Value["zip"].(*Array); ok && len(zipVal.Value) == 2 {
			if arr1, ok := zipVal.Value[0].(*Array); ok {
				if arr2, ok := zipVal.Value[1].(*Array); ok {
					minLen := Integer(min(len(arr1.Value), len(arr2.Value)))
					padMode := false
					var padVal Value = NilValue
					if pad, hasPad := v.Value["pad"]; hasPad {
						padMode = true
						padVal = pad
					}
					maxLen := Integer(max(len(arr1.Value), len(arr2.Value)))
					resultLen := minLen
					if padMode {
						resultLen = maxLen
					}
					if resultLen >= verror.MaxMemSize {
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

		// ── { flatten } ──────────────────────────────────────────────────────
		// Flatten one level of nested arrays.
		//   newArray({ flatten=[[1,2],[3,4],[5]] }) → [1,2,3,4,5]
		if nested, ok := v.Value["flatten"].(*Array); ok {
			var A []Value
			for _, item := range nested.Value {
				if inner, ok := item.(*Array); ok {
					A = append(A, inner.Value...)
				} else {
					A = append(A, item)
				}
				if Integer(len(A)) >= verror.MaxMemSize {
					return &Array{}, nil
				}
			}
			return &Array{Value: A}, nil
		}

		// ── { keys }, { values }, { pairs } ─────────────────────────────────
		// Extract keys, values, or [key,val] pairs from an Object.
		//   newArray({ keys=obj })   → ["name","age","city"]
		//   newArray({ values=obj }) → ["Alice", 30, "NYC"]
		//   newArray({ pairs=obj })  → [["name","Alice"],["age",30],...]
		if obj, ok := v.Value["keys"].(*Object); ok {
			it := obj.Iterator().(Iterator)
			A := make([]Value, 0, len(obj.Value))
			for it.Next() {
				A = append(A, it.Key())
			}
			return &Array{Value: A}, nil
		}

		if obj, ok := v.Value["values"].(*Object); ok {
			it := obj.Iterator().(Iterator)
			A := make([]Value, 0, len(obj.Value))
			for it.Next() {
				A = append(A, it.Value())
			}
			return &Array{Value: A}, nil
		}

		if obj, ok := v.Value["pairs"].(*Object); ok {
			it := obj.Iterator().(Iterator)
			A := make([]Value, 0, len(obj.Value))
			for it.Next() {
				pair := &Array{Value: []Value{it.Key(), it.Value()}}
				A = append(A, pair)
			}
			return &Array{Value: A}, nil
		}

		// ── { grow } ─────────────────────────────────────────────────────────
		// Grow an existing array's capacity without changing length.
		//   newArray({ grow=myArray, by=50 })
		if arr, ok := v.Value["grow"].(*Array); ok {
			if by, ok := v.Value["by"].(Integer); ok && 0 < by && by < verror.MaxMemSize {
				clone := arr.Clone().(*Array)
				clone.Value = slices.Grow(clone.Value, int(by))
				return clone, nil
			}
		}

		// ── { clip } ─────────────────────────────────────────────────────────
		// Clip an array to its length, releasing excess capacity.
		//   newArray({ clip=myArray })
		if arr, ok := v.Value["clip"].(*Array); ok {
			clone := arr.Clone().(*Array)
			clone.Value = slices.Clip(clone.Value)
			return clone, nil
		}

	// ── newArray("hello") ────────────────────────────────────────────────────
	// Explode a string into an array of its runes (codepoints).
	//   newArray("hello") → ["h","e","l","l","o"]
	case *String:
		var i int
		it := v.Iterator().(Iterator)
		A := make([]Value, StringLength(v))
		for it.Next() {
			A[i] = it.Value()
			i++
		}
		return &Array{Value: A}, nil

	// ── newArray(bytes) ──────────────────────────────────────────────────────
	// Convert a Bytes value into an array of integers.
	//   newArray(b) → [72, 101, 108, ...]
	case *Bytes:
		A := make([]Value, len(v.Value))
		for i, b := range v.Value {
			A[i] = Integer(b)
		}
		return &Array{Value: A}, nil

	// ── newArray(float) ──────────────────────────────────────────────────────
	// Decompose a float into its IEEE 754 components.
	//   newArray(3.14) → [sign, exponent, mantissa]  all as Integer
	case Float:
		bits := math.Float64bits(float64(v))
		sign := Integer((bits >> 63) & 1)
		exponent := Integer((bits>>52)&0x7FF) - 1023
		mantissa := Integer(bits & 0x000FFFFFFFFFFFFF)
		A := []Value{sign, exponent, mantissa}
		return &Array{Value: A}, nil

	// ── newArray(arr) ────────────────────────────────────────────────────────
	// Clone an existing array (deep copy).
	//   newArray(arr) → a fresh independent copy
	case *Array:
		return v.Clone(), nil
	}

	// ── Object fallback: entries ─────────────────────────────────────────────
	// Any Object not matched above → array of [key, val] pairs.
common:
	if obj, ok := args[0].(*Object); ok {
		var i int
		it := obj.Iterator().(Iterator)
		A := make([]Value, len(obj.Value))
		for it.Next() {
			B := []Value{it.Key(), it.Value()}
			A[i] = &Array{Value: B}
			i++
		}
		return &Array{Value: A}, nil
	}

	return &Array{}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────

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
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if strings.HasPrefix(v.Value, foundationInterfaceName) {
				var module Value
				switch v.Value[len(foundationInterfaceName):] {
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
				case foundationCore:
					module = loadFoundationCorelib()
				case foundationTask:
					module = loadFoundationTask()
				case foundationRegex:
					module = loadFoundationRegexp()
				case foundationColor:
					module = loadFoundationColor()
				}
				return module, nil
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

func IsMemberOf(args ...Value) (Bool, error) {
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
	return Bool(false), nil
}

func pauseExecution(message string) {
	fmt.Printf("\n\n\n\t\tExecution Paused")
	fmt.Printf("\n\t\t%v", message)
	fmt.Printf("\n\n\n")
	fmt.Scanf(" ")
}
