package vida

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"unsafe"

	"github.com/alkemist-17/vida/verror"
)

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
	m.Value["sortBy"] = NativeFunction(arraySortObjects)
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

func arrayConcat(args ...Value) (Value, error) {
	if len(args) > 1 {
		var size int
		for _, v := range args {
			if xs, ok := v.(*Array); ok {
				size += len(xs.Value)
			}
		}
		if size < 0 || size >= verror.MaxMemSize {
			return GlobalNil, verror.ErrMaxMemSize
		}
		result := make([]Value, 0, size)
		for _, v := range args {
			if xs, ok := v.(*Array); ok {
				result = append(result, xs.Value...)
			}
		}
		return &Array{Value: result}, nil
	}
	return GlobalNil, nil
}

func arrayClear(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = xs.Value[:0]
			return xs, nil
		}
	}
	return GlobalNil, nil
}

func arrayCap(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			return Integer(cap(xs.Value)), nil
		}
	}
	return GlobalNil, nil
}

func arrayOverlaps(args ...Value) (Value, error) {
	if len(args) > 1 {
		a, okA := args[0].(*Array)
		b, okB := args[1].(*Array)
		if okA && okB {
			return Bool(overlapsBackingArray(a.Value, b.Value)), nil
		}
	}
	return GlobalNil, nil
}

func arrayView(args ...Value) (Value, error) {
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
	return GlobalNil, nil
}

func arrayGrow(args ...Value) (Value, error) {
	if len(args) > 1 {
		xs, ok := args[0].(*Array)
		size, oksize := args[1].(Integer)
		if ok && oksize && 0 <= size && size < verror.MaxMemSize {
			xs.Value = slices.Grow(xs.Value, int(size))
			return xs, nil
		}
	}
	return GlobalNil, nil
}

func arrayIndex(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			for i, v := range xs.Value {
				if v.Equals(args[1]) {
					return Integer(i), nil
				}
			}
		}
	}
	return GlobalNil, nil
}

func arrayInsert(args ...Value) (Value, error) {
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
	return GlobalNil, nil
}

func arrayReverse(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			slices.Reverse(xs.Value)
			return xs, nil
		}
	}
	return GlobalNil, nil
}

func arrayReversed(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			vals := make([]Value, len(xs.Value))
			copy(vals, xs.Value)
			slices.Reverse(vals)
			return &Array{Value: vals}, nil
		}
	}
	return GlobalNil, nil
}

func arrayPop(args ...Value) (Value, error) {
	if len(args) == 1 {
		if xs, ok := args[0].(*Array); ok && len(xs.Value) > 0 {
			lastIndex := len(xs.Value) - 1
			val := xs.Value[lastIndex]
			xs.Value = xs.Value[:lastIndex]
			return val, nil
		}
	} else if len(args) == 2 {
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
	return GlobalNil, nil
}

func arrayToObject(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			o := &Object{Value: make(map[string]Value, len(xs.Value))}
			for i, v := range xs.Value {
				o.Value[Integer(i).ObjectKey()] = v
			}
			return o, nil
		}
	}
	return GlobalNil, nil
}

func arraySortObjects(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if fn, ok := args[1].(*Function); ok {
				if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
					((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
				}
				slices.SortFunc(xs.Value, func(l, r Value) int {
					th := ((*clbu)[globalStateIndex].(*GlobalState)).Pool.getThread()
					th.State = Ready
					th.Script.MainFunction = fn
					_, err := coRunThread(th)
					vm := (*clbu)[globalStateIndex].(*GlobalState).VM
					if err != nil {
						switch err {
						case verror.ErrResumeThreadSignal:
							_, threadError := vm.runThread(vm.fp, vm.Frame.ip, false, l, r)
							((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
							if threadError != nil {
								invoker := vm.Thread.Invoker
								invoker.State = Running
								vm.Thread.Invoker = nil
								(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
								vm.Thread = invoker
								fmt.Println("\n\nFATAL ERROR", threadError.Error())
								os.Exit(0)
							}
							switch vm.State {
							case Completed, Suspended:
								v := vm.Channel
								invoker := vm.Thread.Invoker
								invoker.State = Running
								vm.Thread.Invoker = nil
								(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
								vm.Thread = invoker
								if bval, ok := v.(Bool); ok {
									if bval {
										return -1
									}
									return 1
								}
							}
						case verror.ErrStartThreadSignal:
							_, threadError := vm.runThread(vm.fp, 0, true, l, r)
							((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
							if threadError != nil {
								invoker := vm.Thread.Invoker
								invoker.State = Running
								vm.Thread.Invoker = nil
								(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
								vm.Thread = invoker
								fmt.Println("\n\nFATAL ERROR", threadError.Error())
								os.Exit(0)
							}
							switch vm.State {
							case Completed, Suspended:
								v := vm.Channel
								invoker := vm.Thread.Invoker
								invoker.State = Running
								vm.Thread.Invoker = nil
								(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
								vm.Thread = invoker
								if bval, ok := v.(Bool); ok {
									if bval {
										return -1
									}
									return 1
								}
							}
						default:
							return 0
						}
					}
					return 0
				})
				return xs, nil
			}
		}
	}
	return GlobalNil, nil
}

func arrayIsArray(args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(*Array)
		return Bool(ok), nil
	}
	return GlobalNil, nil
}

func arrayIsEmpty(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			return Bool(len(xs.Value) == 0), nil
		}
	}
	return GlobalNil, nil
}

func arrayPairs(args ...Value) (Value, error) {
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
	return GlobalNil, nil
}

func arrayCompact(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = slices.Compact(xs.Value)
			return xs, nil
		}
	}
	return GlobalNil, nil
}

func arrayCompacted(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			cloned := xs.Clone().(*Array)
			cloned.Value = slices.Compact(cloned.Value)
			return cloned, nil
		}
	}
	return GlobalNil, nil
}

func arrayChunk(args ...Value) (Value, error) {
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
	return GlobalNil, nil
}

func arrayClip(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = slices.Clip(xs.Value)
			return xs, nil
		}
	}
	return GlobalNil, nil
}

func arrayReplace(args ...Value) (Value, error) {
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
	return GlobalNil, nil
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

func arraySort(args ...Value) (Value, error) {
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

	// Auto-detect type from first non-nil element
	var sample Value
	for _, v := range xs.Value {
		if v.Type() != GlobalNil.Type() {
			sample = v
			break
		}
	}

	// Dispatch to generic SortBy
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
		return nil, fmt.Errorf("sort: unsupported type for auto-sort: %T", sample)
	}

	return xs, nil
}

func arrayRepeat(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if t, ok := args[1].(Integer); ok && t >= 0 && t < verror.MaxMemSize {
				return &Array{Value: slices.Repeat(xs.Value, int(t))}, nil
			}
		}
	}
	return GlobalNil, nil
}
