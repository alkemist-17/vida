package vida

import (
	"slices"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationArray() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["concat"] = GFn(arrayConcat)
	m.Value["clear"] = GFn(arrayClear)
	m.Value["index"] = GFn(arrayIndex)
	m.Value["insert"] = GFn(arrayInsert)
	m.Value["reverse"] = GFn(arrayReverse)
	m.Value["reversed"] = GFn(arrayReversed)
	m.Value["pop"] = GFn(arrayPop)
	m.Value["sortI"] = GFn(arraySortInts)
	m.Value["sortF"] = GFn(arraySortFloats)
	m.Value["sortS"] = GFn(arraySortStrings)
	m.Value["sortBy"] = GFn(arraySortObjects)
	m.Value["toObject"] = GFn(arrayToObject)
	m.Value["new"] = GFn(coreMakeArray)
	m.Value["isArray"] = GFn(arrayIsArray)
	m.Value["isEmpty"] = GFn(arrayIsEmpty)
	m.Value["entries"] = GFn(arrayPairs)
	m.Value["compact"] = GFn(arrayCompact)
	m.Value["compacted"] = GFn(arrayCompacted)
	m.Value["chunk"] = GFn(arrayChunk)
	m.Value["clip"] = GFn(arrayClip)
	m.Value["del"] = GFn(arrayDelete)
	m.Value["replace"] = GFn(arrayReplace)
	return m
}

func arrayConcat(args ...Value) (Value, error) {
	if len(args) > 1 {
		var size int
		var arr []Value
		for _, v := range args {
			if xs, ok := v.(*Array); ok {
				size += len(xs.Value)
				if size >= verror.MaxMemSize {
					return NilValue, verror.ErrMaxMemSize
				}
				arr = append(arr, xs.Value...)
			}
		}
		return &Array{Value: arr}, nil
	}
	return NilValue, nil
}

func arrayClear(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = make([]Value, 0)
			return xs, nil
		}
	}
	return NilValue, nil
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
	return NilValue, nil
}

func arrayInsert(args ...Value) (Value, error) {
	if len(args) > 2 {
		if xs, ok := args[0].(*Array); ok {
			if idx, ok := args[1].(Integer); ok {
				if idx >= 0 && idx <= Integer(len(xs.Value)) {
					if xs.Equals(args[2]) {
						args[2] = args[2].Clone()
					}
					xs.Value = slices.Insert(xs.Value, int(idx), args[2])
					return xs, nil
				}
			}
		}
	}
	return NilValue, nil
}

func arrayReverse(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			slices.Reverse(xs.Value)
			return xs, nil
		}
	}
	return NilValue, nil
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
	return NilValue, nil
}

func arrayPop(args ...Value) (Value, error) {
	if len(args) == 1 {
		if xs, ok := args[0].(*Array); ok {
			var val Value
			l := len(xs.Value) - 1
			val, xs.Value = xs.Value[l], xs.Value[:l]
			return val, nil
		}
	}
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if i, ok := args[1].(Integer); ok {
				if 0 <= i && i < Integer(len(xs.Value)) {
					var val Value
					val, xs.Value = xs.Value[i], append(xs.Value[:i], xs.Value[i+1:]...)
					return val, nil
				}
			}
		}
	}
	return NilValue, nil
}

func arraySortInts(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			for _, v := range xs.Value {
				if v.Type() != Integer(0).Type() {
					return NilValue, verror.ErrSoringMixedTypes
				}
			}
			slices.SortFunc(xs.Value, func(l, r Value) int {
				if l.(Integer) < r.(Integer) {
					return -1
				} else if l.(Integer) > r.(Integer) {
					return 1
				} else {
					return 0
				}
			})
			return xs, nil
		}
	}
	return NilValue, nil
}

func arraySortFloats(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			for _, v := range xs.Value {
				if v.Type() != Float(0).Type() {
					return NilValue, verror.ErrSoringMixedTypes
				}
			}
			slices.SortFunc(xs.Value, func(l, r Value) int {
				if l.(Float) < r.(Float) {
					return -1
				} else if l.(Float) > r.(Float) {
					return 1
				} else {
					return 0
				}
			})
			return xs, nil
		}
	}
	return NilValue, nil
}

func arraySortStrings(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			for _, v := range xs.Value {
				if v.Type() != (&String{}).Type() {
					return NilValue, verror.ErrSoringMixedTypes
				}
			}
			slices.SortFunc(xs.Value, func(l, r Value) int {
				if l.(*String).Value < r.(*String).Value {
					return -1
				} else if l.(*String).Value > r.(*String).Value {
					return 1
				} else {
					return 0
				}
			})
			return xs, nil
		}
	}
	return NilValue, nil
}

func arrayToObject(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			o := &Object{Value: make(map[string]Value)}
			for i, v := range xs.Value {
				o.Value[Integer(i).String()] = v
			}
			return o, nil
		}
	}
	return NilValue, nil
}

func arraySortObjects(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if fn, ok := args[1].(*Function); ok {
				for _, v := range xs.Value {
					if v.Type() != (&Object{}).Type() {
						return NilValue, verror.ErrSoringMixedTypes
					}
				}
				if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
					((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
				}
				A := make([]Value, 2)
				slices.SortFunc(xs.Value, func(l, r Value) int {
					th := ((*clbu)[globalStateIndex].(*GlobalState)).Pool.getThread()
					th.State = Ready
					th.Script.MainFunction = fn
					A[0], A[1] = l, r
					_, err := coRunThread(th)
					vm := (*clbu)[globalStateIndex].(*GlobalState).VM
					if err != nil {
						switch err {
						case verror.ErrResumeThreadSignal:
							_, threadError := vm.runThread(vm.fp, vm.Frame.ip, false, A...)
							((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
							if threadError != nil {
								return 0
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
							_, threadError := vm.runThread(vm.fp, 0, true, A...)
							((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
							if threadError != nil {
								return 0
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
	return NilValue, nil
}

func arrayIsArray(args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(*Array)
		return Bool(ok), nil
	}
	return NilValue, nil
}

func arrayIsEmpty(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			return Bool(len(xs.Value) == 0), nil
		}
	}
	return NilValue, nil
}

func arrayPairs(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			entries := make([]Value, len(xs.Value))
			for i, v := range xs.Value {
				A := make([]Value, 2)
				A[0] = Integer(i)
				A[1] = v
				entries[i] = &Array{Value: A}
			}
			return &Array{Value: entries}, nil
		}
	}
	return NilValue, nil
}

func arrayCompact(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = slices.Compact(xs.Value)
			return xs, nil
		}
	}
	return NilValue, nil
}

func arrayCompacted(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			cloned := xs.Clone().(*Array)
			cloned.Value = slices.Compact(cloned.Value)
			return cloned, nil
		}
	}
	return NilValue, nil
}

func arrayChunk(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if n, ok := args[1].(Integer); ok && n > 1 {
				var container []Value
				for v := range slices.Chunk(xs.Value, int(n)) {
					container = append(container, &Array{Value: v})
				}
				return &Array{Value: container}, nil
			}
		}
	}
	return NilValue, nil
}

func arrayClip(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			xs.Value = slices.Clip(xs.Value)
			return xs, nil
		}
	}
	return NilValue, nil
}

func arrayDelete(args ...Value) (Value, error) {
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
					xs.Value = slices.Delete(xs.Value, ll, rr)
					return xs, nil
				}
			}
		}
	}
	return NilValue, nil
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
	return NilValue, nil
}
