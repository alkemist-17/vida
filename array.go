package vida

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"unsafe"

	"github.com/alkemist-17/vida/token"
)

type Array struct {
	ReferenceSemanticsImpl
	Value []Value
}

func (xs *Array) Boolean() Bool {
	return True
}

func (xs *Array) Prefix(ctx *Context, op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return False, nil
	}
	return Nil, ErrPrefixOpNotDefined
}

func (xs *Array) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *Array:
		switch op {
		case uint64(token.ADD):
			rLen := len(r.Value)
			if rLen == 0 {
				return xs, nil
			}
			lLen := len(xs.Value)
			if rLen+lLen >= MaxMemSize {
				return Nil, ErrMaxMemSize
			}
			values := make([]Value, lLen+rLen)
			copy(values[:lLen], xs.Value)
			copy(values[lLen:], r.Value)
			return &Array{Value: values}, nil
		case uint64(token.IN):
			return isMemberOf(ctx, xs, rhs)
		}
	}
	switch op {
	case uint64(token.OR):
		return xs, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return isMemberOf(ctx, xs, rhs)
	}
	return Nil, ErrBinaryOpNotDefined
}

func (xs *Array) Get(ctx *Context, index Value) Value {
	switch r := index.(type) {
	case Integer:
		l := Integer(len(xs.Value))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			return xs.Value[r]
		}
	}
	return Nil
}

func (xs *Array) Set(index, val Value) error {
	switch r := index.(type) {
	case Integer:
		l := Integer(len(xs.Value))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			xs.Value[r] = val
			return nil
		}
	}
	return ErrValueNotIndexable
}

func (xs *Array) Equals(ctx *Context, other Value) Bool {
	val, isArray := other.(*Array)
	return Bool(isArray && xs == val)
}

func (xs *Array) IsIterable() Bool {
	return true
}

func (xs *Array) IsCallable() Bool {
	return false
}

func (xs *Array) Iterator(ctx *Context) Value {
	return &ArrayIterator{Array: xs.Value, Init: -1, End: len(xs.Value)}
}

func (xs *Array) String() string {
	return xs.stringify(make(map[uintptr]bool))
}

func (xs *Array) stringify(visited map[uintptr]bool) string {
	if len(xs.Value) == 0 {
		return "[]"
	}

	ptr := reflect.ValueOf(xs).Pointer()

	if visited[ptr] {
		return "[...]"
	}

	visited[ptr] = true
	defer delete(visited, ptr)

	var r []string
	for _, v := range xs.Value {
		r = append(r, stringWithVisited(v, visited))
	}
	return fmt.Sprintf("[%v]", strings.Join(r, ",  "))
}

func (xs *Array) ObjectKey() string {
	return fmt.Sprintf("array[%p]", xs)
}

func (xs *Array) Type() string {
	return arrayT
}

func (xs *Array) Clone() Value {
	c := make([]Value, len(xs.Value))
	for i, v := range xs.Value {
		c[i] = v.Clone()
	}
	return &Array{Value: c}
}

func (xs *Array) GetVTable(ctx *Context) Value {
	if ctx.vtables[arrayT] == nil {
		ctx.loadArrayVT()
	}
	return ctx.vtables[arrayT]
}

func (xs *Array) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[arrayT] == nil {
		ctx.loadArrayVT()
	}
	if vtable, ok := ctx.vtables[arrayT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (xs *Array) MarshalJSON() ([]byte, error) {
	return json.Marshal(xs.Value)
}

func loadFoundationArray() Value {
	m := &Object{Value: make(map[string]Value, 26)}
	m.Value["shuffled"] = NativeFunction(randShuffled)
	m.Value["randomElement"] = NativeFunction(arrayRandomElement)
	m.Value["concat"] = NativeFunction(arrayConcat)
	m.Value["clear"] = NativeFunction(arrayClear)
	m.Value["index"] = NativeFunction(arrayIndex)
	m.Value["insert"] = NativeFunction(arrayInsert)
	m.Value["reverse"] = NativeFunction(arrayReverse)
	m.Value["reversed"] = NativeFunction(arrayReversed)
	m.Value["pop"] = NativeFunction(arrayPop)
	m.Value["sort"] = NativeFunction(arraySort)
	m.Value["sortBy"] = NativeFunction(arraySortWithCompareVidaFunction)
	m.Value["repeat"] = NativeFunction(arrayRepeat)
	m.Value["toObject"] = NativeFunction(arrayToObject)
	m.Value["new"] = NativeFunction(coreNewArray)
	m.Value["isArray"] = NativeFunction(arrayIsArray)
	m.Value["isEmpty"] = NativeFunction(arrayIsEmpty)
	m.Value["pairs"] = NativeFunction(arrayPairs)
	m.Value["compact"] = NativeFunction(arrayCompact)
	m.Value["compacted"] = NativeFunction(arrayCompacted)
	m.Value["chunk"] = NativeFunction(arrayChunk)
	m.Value["clip"] = NativeFunction(arrayClip)
	m.Value["replace"] = NativeFunction(arrayReplace)
	m.Value["cap"] = NativeFunction(arrayCap)
	m.Value["view"] = NativeFunction(collectionProcessView)
	m.Value["grow"] = NativeFunction(arrayGrow)
	m.Value["overlaps"] = NativeFunction(arrayOverlaps)
	return m
}

func arrayConcat(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("concat", fmt.Sprintf("expected at least 2 array arguments, got %d", len(args)))
	}
	var size int
	for i, v := range args {
		xs, ok := v.(*Array)
		if !ok {
			return argError("concat", fmt.Sprintf("argument %d must be an array, got %s", i+1, v.Type()))
		}
		size += len(xs.Value)
	}
	if size < 0 || size >= MaxMemSize {
		return Nil, ErrMaxMemSize
	}
	result := make([]Value, 0, size)
	for _, v := range args {
		xs := v.(*Array)
		result = append(result, xs.Value...)
	}
	return &Array{Value: result}, nil
}

func arrayRandomElement(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("randomElement", "expected an array, string, or bytes argument")
	}
	switch val := args[0].(type) {
	case *Array:
		if len(val.Value) == 0 {
			return argError("randomElement", "cannot pick a random element from an empty array")
		}
		return val.Value[rand.Int()%len(val.Value)], nil
	case *String:
		if val.Runes == nil {
			val.Runes = []rune(val.Value)
		}
		if len(val.Runes) == 0 {
			return argError("randomElement", "cannot pick a random element from an empty string")
		}
		return &String{Value: string(val.Runes[rand.Int()%len(val.Runes)])}, nil
	case *Bytes:
		if len(val.Value) == 0 {
			return argError("randomElement", "cannot pick a random element from empty bytes")
		}
		return Integer(val.Value[rand.Int()%len(val.Value)]), nil
	default:
		return argError("randomElement", fmt.Sprintf("expected an array, string, or bytes argument, got %s", val.Type()))
	}
}

func arrayClear(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("clear", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("clear", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	xs.Value = xs.Value[:0]
	return xs, nil
}

func arrayCap(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("cap", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("cap", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	return Integer(cap(xs.Value)), nil
}

func arrayOverlaps(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("overlaps", fmt.Sprintf("expected 2 array arguments, got %d", len(args)))
	}
	a, okA := args[0].(*Array)
	if !okA {
		return argError("overlaps", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	b, okB := args[1].(*Array)
	if !okB {
		return argError("overlaps", fmt.Sprintf("argument 2 must be an array, got %s", args[1].Type()))
	}
	return Bool(overlapsBackingArray(a.Value, b.Value)), nil
}

func arrayGrow(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("grow", fmt.Sprintf("expected 2 arguments (array, size), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("grow", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	size, oksize := args[1].(Integer)
	if !oksize {
		return argError("grow", fmt.Sprintf("argument 2 (size) must be an integer, got %s", args[1].Type()))
	}
	if size < 0 || size >= MaxMemSize {
		return argError("grow", fmt.Sprintf("size must be between 0 and %d, got %d", MaxMemSize, size))
	}
	xs.Value = slices.Grow(xs.Value, int(size))
	return xs, nil
}

func arrayIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("index", fmt.Sprintf("expected 2 arguments (array, value), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("index", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	for i, v := range xs.Value {
		if v.Equals(ctx, args[1]) {
			return Integer(i), nil
		}
	}
	return Nil, nil // not found is a valid result, not an error
}

func arrayInsert(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("insert", fmt.Sprintf("expected at least 3 arguments (array, index, value...), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("insert", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	idx, ok := args[1].(Integer)
	if !ok {
		return argError("insert", fmt.Sprintf("argument 2 (index) must be an integer, got %s", args[1].Type()))
	}
	if idx < 0 {
		idx = idx + Integer(len(xs.Value))
	}
	if idx < 0 || idx > Integer(len(xs.Value)) {
		return argError("insert", fmt.Sprintf("index %d out of range for array of length %d", args[1], len(xs.Value)))
	}
	xs.Value = slices.Insert(xs.Value, int(idx), args[2:]...)
	return xs, nil
}

func arrayReverse(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("reverse", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("reverse", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	slices.Reverse(xs.Value)
	return xs, nil
}

func arrayRemove(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 2 && len(args) != 3 {
		return argError("remove", fmt.Sprintf("expected 2 arguments (array, index) or 3 arguments (array, start, end), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("remove", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	if len(xs.Value) == 0 {
		return argError("remove", "cannot remove from an empty array")
	}
	switch len(args) {
	case 2:
		i, ok := args[1].(Integer)
		if !ok {
			return argError("remove", fmt.Sprintf("argument 2 (index) must be an integer, got %s", args[1].Type()))
		}
		if i < 0 {
			i = i + Integer(len(xs.Value))
		}
		if i < 0 || i >= Integer(len(xs.Value)) {
			return argError("remove", fmt.Sprintf("index %d out of range for array of length %d", args[1], len(xs.Value)))
		}
		val := xs.Value[i]
		xs.Value = slices.Delete(xs.Value, int(i), int(i+1))
		return val, nil
	default: // case 3
		i, ok := args[1].(Integer)
		if !ok {
			return argError("remove", fmt.Sprintf("argument 2 (start) must be an integer, got %s", args[1].Type()))
		}
		j, ok := args[2].(Integer)
		if !ok {
			return argError("remove", fmt.Sprintf("argument 3 (end) must be an integer, got %s", args[2].Type()))
		}
		if i < 0 {
			i = i + Integer(len(xs.Value))
		}
		if j < 0 {
			j = j + Integer(len(xs.Value))
		}
		if i < 0 || i >= Integer(len(xs.Value)) || j < 0 || j > Integer(len(xs.Value)) || i >= j {
			return argError("remove", fmt.Sprintf("invalid range [%d:%d] for array of length %d", args[1], args[2], len(xs.Value)))
		}
		val := make([]Value, len(xs.Value[i:j]))
		copy(val, xs.Value[i:j])
		xs.Value = slices.Delete(xs.Value, int(i), int(j))
		return &Array{Value: val}, nil
	}
}

func arrayReversed(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("reversed", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("reversed", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	vals := make([]Value, len(xs.Value))
	copy(vals, xs.Value)
	slices.Reverse(vals)
	return &Array{Value: vals}, nil
}

func arrayPop(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 1 {
		return argError("pop", fmt.Sprintf("expected 1 argument (array), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("pop", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	if len(xs.Value) > 0 {
		lastIndex := len(xs.Value) - 1
		val := xs.Value[lastIndex]
		xs.Value = xs.Value[:lastIndex]
		return val, nil
	}
	return argError("pop", "cannot pop from an empty array")
}

func arrayContains(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok && len(xs.Value) > 0 {
			val := args[1]
			return Bool(slices.ContainsFunc(xs.Value, func(v Value) bool {
				return bool(v.Equals(ctx, val))
			})), nil
		}
	}
	return False, nil
}

func arrayToObject(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("toObject", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("toObject", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	o := &Object{Value: make(map[string]Value, len(xs.Value))}
	for i, v := range xs.Value {
		o.Value[Integer(i).ObjectKey()] = v
	}
	return o, nil
}

func arrayIsArray(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isArray", "expected 1 argument")
	}
	_, ok := args[0].(*Array)
	return Bool(ok), nil
}

func arrayIsEmpty(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isEmpty", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("isEmpty", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	return Bool(len(xs.Value) == 0), nil
}

func arrayPairs(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("pairs", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("pairs", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	entries := make([]Value, len(xs.Value))
	for i, v := range xs.Value {
		pair := &Array{Value: make([]Value, 0, 2)}
		pair.Value = append(pair.Value, Integer(i), v)
		entries[i] = pair
	}
	return &Array{Value: entries}, nil
}

func arrayCompact(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("compact", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("compact", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	xs.Value = slices.Compact(xs.Value)
	return xs, nil
}

func arrayCompacted(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("compacted", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("compacted", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	cloned := xs.Clone().(*Array)
	cloned.Value = slices.Compact(cloned.Value)
	return cloned, nil
}

func arrayChunk(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("chunk", fmt.Sprintf("expected 2 arguments (array, size), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("chunk", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	if len(xs.Value) == 0 {
		return &Array{}, nil
	}
	n, ok := args[1].(Integer)
	if !ok {
		return argError("chunk", fmt.Sprintf("argument 2 (size) must be an integer, got %s", args[1].Type()))
	}
	if n < 1 {
		return argError("chunk", fmt.Sprintf("chunk size must be at least 1, got %d", n))
	}
	count := (len(xs.Value) + int(n) - 1) / int(n)
	container := make([]Value, 0, count)
	for v := range slices.Chunk(xs.Value, int(n)) {
		container = append(container, &Array{Value: v})
	}
	return &Array{Value: container}, nil
}

func arrayClip(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("clip", "expected an array argument")
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("clip", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}
	xs.Value = slices.Clip(xs.Value)
	return xs, nil
}

func arrayReplace(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("replace", fmt.Sprintf("expected at least 3 arguments (array, start, end, value...), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("replace", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	i, iok := args[1].(Integer)
	if !iok {
		return argError("replace", fmt.Sprintf("argument 2 (start) must be an integer, got %s", args[1].Type()))
	}
	j, jok := args[2].(Integer)
	if !jok {
		return argError("replace", fmt.Sprintf("argument 3 (end) must be an integer, got %s", args[2].Type()))
	}
	ll, rr := int(i), int(j)
	xsLen := len(xs.Value)
	if ll < 0 {
		ll += xsLen
	}
	if rr < 0 {
		rr += xsLen
	}
	if ll < 0 || ll > xsLen || rr < 0 || rr > xsLen || ll >= rr {
		return argError("replace", fmt.Sprintf("invalid range [%d:%d] for array of length %d", i, j, xsLen))
	}
	xs.Value = slices.Replace(xs.Value, ll, rr, args[3:]...)
	return xs, nil
}

func overlapsBackingArray[T any](a, b []T) bool {
	ptrA := unsafe.SliceData(a)
	ptrB := unsafe.SliceData(b)

	if ptrA == nil || ptrB == nil {
		return false
	}

	size := unsafe.Sizeof(a[0])
	if size == 0 {
		return false
	}

	addrA := uintptr(unsafe.Pointer(ptrA))
	addrB := uintptr(unsafe.Pointer(ptrB))

	endA := addrA + uintptr(cap(a))*size
	endB := addrB + uintptr(cap(b))*size

	return (addrA >= addrB && addrA < endB) || (addrB >= addrA && addrB < endA)
}

type ord interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func genericSortBy[T ord](xs *[]Value, extract func(Value) (T, error)) error {
	for i, v := range *xs {
		if _, err := extract(v); err != nil {
			return fmt.Errorf("sort: element %d %s", i, err.Error())
		}

	}
	slices.SortFunc(*xs, func(a, b Value) int {
		ka, _ := extract(a)
		kb, _ := extract(b)
		return cmp.Compare(ka, kb)
	})
	return nil
}

func extractInteger(v Value) (int64, error) {
	i, ok := v.(Integer)
	if !ok {
		return 0, fmt.Errorf("expected an integer, got %s", v.Type())
	}
	return int64(i), nil
}

func extractFloat(v Value) (float64, error) {
	f, ok := v.(Float)
	if !ok {
		return 0, fmt.Errorf("expected a float, got %s", v.Type())
	}
	return float64(f), nil
}

func extractString(v Value) (string, error) {
	s, ok := v.(*String)
	if !ok {
		return EmptyString, fmt.Errorf("expected a string, got %s", v.Type())
	}
	return s.Value, nil
}

func arraySort(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("sort", "expected an array argument")
	}

	xs, ok := args[0].(*Array)
	if !ok {
		return argError("sort", fmt.Sprintf("expected an array argument, got %s", args[0].Type()))
	}

	if len(xs.Value) <= 1 {
		return xs, nil
	}

	var sample Value = xs.Value[0]

	switch sample.(type) {
	case Integer:
		if err := genericSortBy(&xs.Value, extractInteger); err != nil {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	case Float:
		if err := genericSortBy(&xs.Value, extractFloat); err != nil {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	case *String:
		if err := genericSortBy(&xs.Value, extractString); err != nil {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	default:
		return argError("sort", fmt.Sprintf("unsupported element type %q: sort only supports arrays of all-integer, all-float, or all-string elements (use sortBy for custom comparisons)", sample.Type()))
	}

	return xs, nil
}

func arrayRepeat(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("repeat", fmt.Sprintf("expected 2 arguments (array, count), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("repeat", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	t, ok := args[1].(Integer)
	if !ok {
		return argError("repeat", fmt.Sprintf("argument 2 (count) must be an integer, got %s", args[1].Type()))
	}
	if t < 0 || t >= MaxMemSize {
		return argError("repeat", fmt.Sprintf("count must be between 0 and %d, got %d", MaxMemSize, t))
	}
	return &Array{Value: slices.Repeat(xs.Value, int(t))}, nil
}

func arraySortWithCompareVidaFunction(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("sortBy", fmt.Sprintf("expected 2 arguments (array, compareFn), got %d", len(args)))
	}
	xs, ok := args[0].(*Array)
	if !ok {
		return argError("sortBy", fmt.Sprintf("argument 1 must be an array, got %s", args[0].Type()))
	}
	compareFn, ok := args[1].(*Function)
	if !ok {
		return argError("sortBy", fmt.Sprintf("argument 2 (compareFn) must be a function, got %s", args[1].Type()))
	}
	th := ctx.getInternalThread(compareFn)
	vm := &VM{th, ctx}
	slices.SortFunc(xs.Value, func(a, b Value) int {
		if err := vm.runThread(0, 0, true, a, b); err == nil {
			v := vm.Channel
			vm.Reset(compareFn)
			switch t := v.(type) {
			case Integer:
				return int(t)
			case Bool:
				var r int
				if t {
					r = -1
				} else {
					r = 1
				}
				return r
			}
		}
		return 0
	})
	ctx.releaseInternalThread()
	return xs, nil
}

func collectionProcessView(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		return args[0], nil
	case 2:
		val := args[0]
		startIdx, okStartIdx := args[1].(Integer)
		var endIdx Integer
		if okStartIdx {
			var hasStart, hasEnd = true, false
			return processView(val, startIdx, endIdx, hasStart, hasEnd)
		}
	case 3:
		val := args[0]
		startIdx, okStartIdx := args[1].(Integer)
		endIdx, okEndIdx := args[2].(Integer)
		if okStartIdx && okEndIdx {
			var hasStart, hasEnd = true, true
			return processView(val, startIdx, endIdx, hasStart, hasEnd)
		}
	}
	return Nil, ErrView
}

func processView(val Value, startIdx, endIdx Integer, hasStart, hasEnd bool) (Value, error) {
	resolve := func(length Integer) (Integer, Integer, bool) {
		s := Integer(0)
		if hasStart {
			s = startIdx
		}
		e := length
		if hasEnd {
			e = endIdx
		}
		return sliceBounds(s, e, length)
	}
	switch v := val.(type) {
	case *Array:
		length := Integer(len(v.Value))
		s, e, empty := resolve(length)
		if empty {
			return &Array{}, nil
		}
		return &Array{Value: v.Value[s:e]}, nil
	case *String:
		if v.Runes == nil {
			v.Runes = []rune(v.Value)
		}
		length := Integer(len(v.Runes))
		s, e, empty := resolve(length)
		if empty {
			return &String{}, nil
		}
		return &String{Value: string(v.Runes[s:e])}, nil

	case *Bytes:
		length := Integer(len(v.Value))
		s, e, empty := resolve(length)
		if empty {
			return &Bytes{}, nil
		}
		return &Bytes{Value: v.Value[s:e]}, nil
	}
	return Nil, ErrView
}
