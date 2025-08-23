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
	m.Value["sort"] = GFn(arraySortObjects)
	m.Value["toMap"] = GFn(arraySet)
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
			return NilValue, verror.ErrMaxMemSize
		}
		var arr []Value
		for _, v := range args {
			if xs, ok := v.(*Array); ok {
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
			val := Integer(0)
			for _, v := range xs.Value {
				if v.Type() != val.Type() {
					return NilValue, verror.ErrSoringMixedTypes
				}
			}
			slices.SortFunc(xs.Value, func(l, r Value) int {
				if Integer(l.(Integer)) < Integer(r.(Integer)) {
					return -1
				} else if Integer(l.(Integer)) > Integer(r.(Integer)) {
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
			val := Float(0)
			for _, v := range xs.Value {
				if v.Type() != val.Type() {
					return NilValue, verror.ErrSoringMixedTypes
				}
			}
			slices.SortFunc(xs.Value, func(l, r Value) int {
				if Float(l.(Float)) < Float(r.(Float)) {
					return -1
				} else if Float(l.(Float)) > Float(r.(Float)) {
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
			val := &String{}
			for _, v := range xs.Value {
				if v.Type() != val.Type() {
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

func arraySet(args ...Value) (Value, error) {
	if len(args) > 0 {
		if xs, ok := args[0].(*Array); ok {
			set := &Object{Value: make(map[string]Value)}
			for i := 0; i < len(xs.Value); i++ {
				set.Value[xs.Value[i].ObjectKey()] = Bool(true)
			}
			return set, nil
		}
	}
	return NilValue, nil
}

func arraySortObjects(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if fn, ok := args[1].(*Function); ok {
				val := &Object{}
				for _, v := range xs.Value {
					if v.Type() != val.Type() {
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
					_, err := gfnRunThread(th)
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
