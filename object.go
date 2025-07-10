package vida

import (
	"fmt"
	"math/rand/v2"
)

func loadObjectLib() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["inject"] = GFn(injectProps)
	m.Value["extract"] = GFn(extractProps)
	m.Value["override"] = GFn(injectAndOverrideProps)
	m.Value["conforms"] = GFn(checkProps)
	m.Value["del"] = GFn(deleteProperty)
	m.Value["setproto"] = GFn(setPrototype)
	m.Value["getproto"] = GFn(getPrototype)
	m.Value["setmeta"] = GFn(setMeta)
	m.Value["getmeta"] = GFn(getMeta)
	return m
}

func injectProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if o, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if m, ok := v.(*Object); ok && m != o {
					for k, x := range m.Value {
						if _, isPresent := o.Value[k]; !isPresent {
							o.Value[k] = x
						}
					}
				}
			}
			return o, nil
		}
	}
	return NilValue, nil
}

func injectAndOverrideProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if o, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if m, ok := v.(*Object); ok && m != o {
					for k, x := range m.Value {
						o.Value[k] = x
					}
				}
			}
			return o, nil
		}
	}
	return NilValue, nil
}

func checkProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if o, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if m, ok := v.(*Object); ok && m != o {
					for k := range m.Value {
						if _, isPresent := o.Value[k]; !isPresent {
							return Bool(false), nil
						}
					}
				}
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func extractProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if o, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if m, ok := v.(*Object); ok && m != o {
					for k := range m.Value {
						delete(o.Value, k)
					}
				}
			}
			return o, nil
		}
	}
	return NilValue, nil
}

func deleteProperty(args ...Value) (Value, error) {
	if len(args) >= 2 {
		if o, ok := args[0].(*Object); ok {
			delete(o.Value, args[1].String())
		}
	}
	return NilValue, nil
}

func setPrototype(args ...Value) (Value, error) {
	if len(__proto) == 0 {
		__proto = fmt.Sprint("__proto", rand.Uint64())
	}
	if len(args) >= 2 {
		if o, ok := args[0].(*Object); ok {
			if proto, ok := args[1].(*Object); ok {
				o.Value[__proto] = proto
				return o, nil
			}
		}
	}
	return NilValue, nil
}

func getPrototype(args ...Value) (Value, error) {
	if len(__proto) == 0 {
		__proto = fmt.Sprint("__proto", rand.Uint64())
	}
	if len(args) >= 0 {
		if o, ok := args[0].(*Object); ok {
			if proto, ok := o.Value[__proto]; ok {
				return proto, nil
			}
		}
	}
	return NilValue, nil
}

func setMeta(args ...Value) (Value, error) {
	if len(__meta) == 0 {
		((*clbu)[globalStateIndex].(*GlobalState)).Aux = newThread(nil, ((*clbu)[globalStateIndex].(*GlobalState)).Script, quarterStack)
		__meta = fmt.Sprint("__meta", rand.Uint64())
	}
	if len(args) >= 2 {
		if o, ok := args[0].(*Object); ok {
			if meta, ok := args[1].(*Object); ok {
				o.Value[__meta] = meta
				return o, nil
			}
		}
	}
	return NilValue, nil
}

func getMeta(args ...Value) (Value, error) {
	if len(__meta) == 0 {
		((*clbu)[globalStateIndex].(*GlobalState)).Aux = newThread(nil, ((*clbu)[globalStateIndex].(*GlobalState)).Script, quarterStack)
		__meta = fmt.Sprint("__meta", rand.Uint64())
	}
	if len(args) >= 0 {
		if o, ok := args[0].(*Object); ok {
			if meta, ok := o.Value[__meta]; ok {
				return meta, nil
			}
		}
	}
	return NilValue, nil
}
