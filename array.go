package vida

import (
	"cmp"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unsafe"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
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
	return Nil, verror.ErrPrefixOpNotDefined
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
			if rLen+lLen >= verror.MaxMemSize {
				return Nil, verror.ErrMaxMemSize
			}
			values := make([]Value, lLen+rLen)
			copy(values[:lLen], xs.Value)
			copy(values[lLen:], r.Value)
			return &Array{Value: values}, nil
		case uint64(token.IN):
			return IsMemberOf(ctx, xs, rhs)
		}
	}
	switch op {
	case uint64(token.OR):
		return xs, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, xs, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
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
	return verror.ErrValueNotIndexable
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

func (xs *Array) Iterator() Value {
	return &ArrayIterator{Array: xs.Value, Init: -1, End: len(xs.Value)}
}

func (xs *Array) String() string {
	return xs.stringify(make(map[uintptr]bool))
}

func (xs *Array) stringify(visited map[uintptr]bool) string {
	if len(xs.Value) == 0 {
		return "array[]"
	}

	ptr := reflect.ValueOf(xs).Pointer()

	if visited[ptr] {
		return "array[...]"
	}

	visited[ptr] = true
	defer delete(visited, ptr)

	var r []string
	for _, v := range xs.Value {
		r = append(r, stringWithVisited(v, visited))
	}
	return fmt.Sprintf("array[%v]", strings.Join(r, ",  "))
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
	m := &Object{Value: make(map[string]Value, 24)}
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
	m.Value["view"] = NativeFunction(arrayView)
	m.Value["grow"] = NativeFunction(arrayGrow)
	m.Value["overlaps"] = NativeFunction(arrayOverlaps)
	return m
}

func arrayConcat(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		var size int
		for _, v := range args {
			if xs, ok := v.(*Array); ok {
				size += len(xs.Value)
			}
		}
		if size < 0 || size >= verror.MaxMemSize {
			return Nil, verror.ErrMaxMemSize
		}
		result := make([]Value, 0, size)
		for _, v := range args {
			if xs, ok := v.(*Array); ok {
				result = append(result, xs.Value...)
			}
		}
		return &Array{Value: result}, nil
	}
	return Nil, nil
}

func arrayClear(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = xs.Value[:0]
			return xs, nil
		}
	}
	return Nil, nil
}

func arrayCap(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			return Integer(cap(xs.Value)), nil
		}
	}
	return Nil, nil
}

func arrayOverlaps(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		a, okA := args[0].(*Array)
		b, okB := args[1].(*Array)
		if okA && okB {
			return Bool(overlapsBackingArray(a.Value, b.Value)), nil
		}
	}
	return Nil, nil
}

func arrayView(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		xs, ok := args[0].(*Array)
		init, okI := args[1].(Integer)
		length, okE := args[2].(Integer)
		if ok && okI && okE {
			srclen := len(xs.Value)
			start, end, _ := sliceBounds(init, init+length, Integer(srclen))
			return &Array{Value: xs.Value[start:end]}, nil
		}
	}
	return Nil, nil
}

func arrayGrow(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		xs, ok := args[0].(*Array)
		size, oksize := args[1].(Integer)
		if ok && oksize && 0 <= size && size < verror.MaxMemSize {
			xs.Value = slices.Grow(xs.Value, int(size))
			return xs, nil
		}
	}
	return Nil, nil
}

func arrayIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			for i, v := range xs.Value {
				if v.Equals(ctx, args[1]) {
					return Integer(i), nil
				}
			}
		}
	}
	return Nil, nil
}

func arrayInsert(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		if xs, ok := args[0].(*Array); ok {
			if idx, ok := args[1].(Integer); ok {
				if idx >= 0 && idx <= Integer(len(xs.Value)) {
					xs.Value = slices.Insert(xs.Value, int(idx), args[2])
					return xs, nil
				}
			}
		}
	}
	return Nil, nil
}

func arrayReverse(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			slices.Reverse(xs.Value)
			return xs, nil
		}
	}
	return Nil, nil
}

func arrayDelete(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 2 {
		if xs, ok := args[0].(*Array); ok && len(xs.Value) > 0 {
			if i, ok := args[1].(Integer); ok {
				if 0 <= i && i < Integer(len(xs.Value)) {
					val := xs.Value[i]
					xs.Value = slices.Delete(xs.Value, int(i), int(i+1))
					return val, nil
				}
			}
		}
	} else if len(args) == 3 {
		if xs, ok := args[0].(*Array); ok && len(xs.Value) > 0 {
			if i, ok := args[1].(Integer); ok {
				if j, ok := args[2].(Integer); ok {
					if 0 <= i && i < Integer(len(xs.Value)) && 0 <= j && j <= Integer(len(xs.Value)) && i < j {
						val := make([]Value, len(xs.Value[i:j]))
						copy(val, xs.Value[i:j])
						xs.Value = slices.Delete(xs.Value, int(i), int(j))
						return &Array{Value: val}, nil
					}
				}
			}
		}
	}
	return Nil, nil
}

func arrayReversed(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			vals := make([]Value, len(xs.Value))
			copy(vals, xs.Value)
			slices.Reverse(vals)
			return &Array{Value: vals}, nil
		}
	}
	return Nil, nil
}

func arrayPop(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 1 {
		if xs, ok := args[0].(*Array); ok && len(xs.Value) > 0 {
			lastIndex := len(xs.Value) - 1
			val := xs.Value[lastIndex]
			xs.Value = xs.Value[:lastIndex]
			return val, nil
		}
	}
	return Nil, nil
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
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			o := &Object{Value: make(map[string]Value, len(xs.Value))}
			for i, v := range xs.Value {
				o.Value[Integer(i).ObjectKey()] = v
			}
			return o, nil
		}
	}
	return Nil, nil
}

func arrayIsArray(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(*Array)
		return Bool(ok), nil
	}
	return Nil, nil
}

func arrayIsEmpty(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			return Bool(len(xs.Value) == 0), nil
		}
	}
	return Nil, nil
}

func arrayPairs(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			entries := make([]Value, len(xs.Value))
			for i, v := range xs.Value {
				pair := &Array{Value: make([]Value, 0, 2)}
				pair.Value = append(pair.Value, Integer(i), v)
				entries[i] = pair
			}
			return &Array{Value: entries}, nil
		}
	}
	return Nil, nil
}

func arrayCompact(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = slices.Compact(xs.Value)
			return xs, nil
		}
	}
	return Nil, nil
}

func arrayCompacted(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			cloned := xs.Clone().(*Array)
			cloned.Value = slices.Compact(cloned.Value)
			return cloned, nil
		}
	}
	return Nil, nil
}

func arrayChunk(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if len(xs.Value) == 0 {
				return &Array{}, nil
			}
			if n, ok := args[1].(Integer); ok && n >= 1 {
				count := (len(xs.Value) + int(n) - 1) / int(n)
				container := make([]Value, 0, count)
				for v := range slices.Chunk(xs.Value, int(n)) {
					container = append(container, &Array{Value: v})
				}
				return &Array{Value: container}, nil
			}
		}
	}
	return Nil, nil
}

func arrayClip(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = slices.Clip(xs.Value)
			return xs, nil
		}
	}
	return Nil, nil
}

func arrayReplace(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		if xs, ok := args[0].(*Array); ok {
			i, iok := args[1].(Integer)
			j, jok := args[2].(Integer)
			ll, rr := int(i), int(j)
			if iok && jok {
				xsLen := len(xs.Value)
				if ll < 0 {
					ll += xsLen
				}
				if rr < 0 {
					rr += xsLen
				}
				if 0 <= ll && ll <= xsLen && 0 <= rr && rr <= xsLen && ll < rr {
					xs.Value = slices.Replace(xs.Value, ll, rr, args[3:]...)
					return xs, nil
				}
			}
		}
	}
	return Nil, nil
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
			return fmt.Errorf("sort: element[%d] type mismatch: %w", i, err)
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
		return 0, fmt.Errorf("expected Integer, got %T", v)
	}
	return int64(i), nil
}

func extractFloat(v Value) (float64, error) {
	f, ok := v.(Float)
	if !ok {
		return 0, fmt.Errorf("expected Float, got %T", v)
	}
	return float64(f), nil
}

func extractString(v Value) (string, error) {
	s, ok := v.(*String)
	if !ok {
		return "", fmt.Errorf("expected *String, got %T", v)
	}
	return s.Value, nil
}

func arraySort(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("sort: expected array argument")
	}

	xs, ok := args[0].(*Array)
	if !ok {
		return nil, fmt.Errorf("sort: expected array argument")
	}

	if len(xs.Value) <= 1 {
		return xs, nil
	}

	var sample Value = xs.Value[0]

	switch sample.(type) {
	case Integer:
		if err := genericSortBy(&xs.Value, extractInteger); err != nil {
			return nil, err
		}
	case Float:
		if err := genericSortBy(&xs.Value, extractFloat); err != nil {
			return nil, err
		}
	case *String:
		if err := genericSortBy(&xs.Value, extractString); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("std.array.sort: unsupported type for native sort: %v", sample.Type())
	}

	return xs, nil
}

func arrayRepeat(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if t, ok := args[1].(Integer); ok && t >= 0 && t < verror.MaxMemSize {
				return &Array{Value: slices.Repeat(xs.Value, int(t))}, nil
			}
		}
	}
	return Nil, nil
}

func arraySortWithCompareVidaFunction(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if compareFn, ok := args[1].(*Function); ok {
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
		}
	}
	return Nil, nil
}
