package vida

import (
	"fmt"
	"maps"
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
	m.Value["get"] = GFn(getValue)
	m.Value["set"] = GFn(setValue)
	m.Value["has"] = GFn(hasValue)
	m.Value["keys"] = GFn(getKeys)
	m.Value["values"] = GFn(getValues)
	return m
}

func injectProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k, x := range other.Value {
						if _, isPresent := self.Value[k]; !isPresent {
							self.Value[k] = x
						}
					}
				}
			}
			return self, nil
		}
	}
	return NilValue, nil
}

func injectAndOverrideProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					maps.Copy(self.Value, other.Value)
				}
			}
			return self, nil
		}
	}
	return NilValue, nil
}

func checkProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			set := make(map[string]bool)
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k := range other.Value {
						if k != __proto && k != __meta {
							set[k] = false
						}
					}
				}
			}
			recursiveProtoCheck(set, self)
			for _, v := range set {
				if v == false {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func extractProps(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k := range other.Value {
						delete(self.Value, k)
					}
				}
			}
			return self, nil
		}
	}
	return NilValue, nil
}

func deleteProperty(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if meta, ok := self.Value[__meta].(*Object); ok {
				if __delete, ok := meta.Value[__del]; ok {
					return __delete, nil
				}
			}
			for _, prop := range args[1:] {
				delete(self.Value, prop.ObjectKey())
			}
		}
	}
	return NilValue, nil
}

func setPrototype(args ...Value) (Value, error) {
	if len(__proto) == 0 {
		__proto = fmt.Sprint(__proto, rand.Uint64())
	}
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if proto, ok := args[1].(*Object); ok {
				if setproto, ok := proto.Value[__setproto]; ok {
					return setproto, nil
				}
				self.Value[__proto] = proto
				return self, nil
			}
		}
	}
	return NilValue, nil
}

func getPrototype(args ...Value) (Value, error) {
	if len(__proto) == 0 {
		__proto = fmt.Sprint(__proto, rand.Uint64())
	}
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if proto, ok := self.Value[__proto].(*Object); ok {
				if getproto, ok := proto.Value[__getproto]; ok {
					return getproto, nil
				}
				return proto, nil
			}
		}
	}
	return NilValue, nil
}

func setMeta(args ...Value) (Value, error) {
	if len(__meta) == 0 {
		__meta = fmt.Sprint(__meta, rand.Uint64())
		if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
			((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
		}
	}
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if meta, ok := args[1].(*Object); ok {
				if metaobject, ok := meta.Value[__setmeta]; ok {
					return metaobject, nil
				}
				self.Value[__meta] = meta
				return self, nil
			}
		}
	}
	return NilValue, nil
}

func getMeta(args ...Value) (Value, error) {
	if len(__meta) == 0 {
		__meta = fmt.Sprint(__meta, rand.Uint64())
	}
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if meta, ok := self.Value[__meta].(*Object); ok {
				if metaobject, ok := meta.Value[__getmeta]; ok {
					return metaobject, nil
				}
				return meta, nil
			}
		}
	}
	return NilValue, nil
}

func getValue(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if val, ok := self.Value[args[1].ObjectKey()]; ok {
				return val, nil
			}
		}
	}
	return NilValue, nil
}

func setValue(args ...Value) (Value, error) {
	if len(args) > 2 {
		if self, ok := args[0].(*Object); ok {
			self.Value[args[1].ObjectKey()] = args[2]
		}
	}
	return NilValue, nil
}

func hasValue(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			item := args[1].ObjectKey()
			for k := range self.Value {
				if item == k {
					return Bool(true), nil
				}
			}
			return Bool(false), nil
		}
	}
	return NilValue, nil
}

func getKeys(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			lobj := len(self.Value)
			if _, ok := self.Value[__proto]; ok {
				lobj--
			}
			if _, ok := self.Value[__meta]; ok {
				lobj--
			}
			keys := make([]Value, int(lobj))
			var idx int
			for k := range self.Value {
				if k != __proto && k != __meta {
					keys[idx] = &String{Value: k}
					idx++
				}
			}
			return &Array{Value: keys}, nil
		}
	}
	return NilValue, nil
}

func getValues(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			lobj := len(self.Value)
			if _, ok := self.Value[__proto]; ok {
				lobj--
			}
			if _, ok := self.Value[__meta]; ok {
				lobj--
			}
			values := make([]Value, int(lobj))
			var idx int
			for k, v := range self.Value {
				if k != __proto && k != __meta {
					values[idx] = v
					idx++
				}
			}
			return &Array{Value: values}, nil
		}
	}
	return NilValue, nil
}

func recursiveProtoCheck(set map[string]bool, self *Object) {
	if self == nil {
		return
	}
	for k := range self.Value {
		if k != __proto && k != __meta {
			if _, isPresent := set[k]; isPresent {
				set[k] = true
			}
		}
	}
	if p, ok := self.Value[__proto]; ok {
		proto := p.(*Object)
		recursiveProtoCheck(set, proto)
	}
}
