package vida

import (
	"fmt"
	"maps"
	"math/rand/v2"
)

func loadObjectLib() Value {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
		((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
	}
	__meta = fmt.Sprint(__meta, rand.Uint64())
	__proto = fmt.Sprint(__proto, rand.Uint64())
	m := &Object{Value: make(map[string]Value)}
	m.Value["inject"] = GFn(objectInjectProperties)
	m.Value["extract"] = GFn(objectExtractProperties)
	m.Value["override"] = GFn(objectInjectAndOverrideProperties)
	m.Value["conforms"] = GFn(objectCheckProperties)
	m.Value["del"] = GFn(objectDeleteProperty)
	m.Value["setproto"] = GFn(objectSetPrototype)
	m.Value["getproto"] = GFn(objectGetPrototype)
	m.Value["setmeta"] = GFn(objectSetMetaObject)
	m.Value["getmeta"] = GFn(objectgetMetaObject)
	m.Value["search"] = GFn(objectSearchValueInProtoChain)
	m.Value["get"] = GFn(objectGetValue)
	m.Value["set"] = GFn(objectSetValue)
	m.Value["has"] = GFn(objectHasValue)
	m.Value["keys"] = GFn(objectGetKeys)
	m.Value["values"] = GFn(objectGetValues)
	return m
}

func objectInjectProperties(args ...Value) (Value, error) {
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

func objectInjectAndOverrideProperties(args ...Value) (Value, error) {
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

func objectCheckProperties(args ...Value) (Value, error) {
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
			objectrecursiveProtoCheck(set, self)
			for _, v := range set {
				if !v {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func objectExtractProperties(args ...Value) (Value, error) {
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

func objectDeleteProperty(args ...Value) (Value, error) {
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

func objectSetPrototype(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if possibleNewProto, ok := args[1].(*Object); ok {
				if meta, ok := self.Value[__meta].(*Object); ok {
					if v, ok := meta.Value[__setproto]; ok {
						return v, nil
					}
				}
				self.Value[__proto] = possibleNewProto
				return self, nil
			}
		}
	}
	return NilValue, nil
}

func objectGetPrototype(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if meta, ok := self.Value[__meta].(*Object); ok {
				if v, ok := meta.Value[__getproto]; ok {
					return v, nil
				}
			}
			if proto, ok := self.Value[__proto]; ok {
				return proto, nil
			}
		}
	}
	return NilValue, nil
}

func objectSetMetaObject(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if possibleNewMeta, ok := args[1].(*Object); ok {
				if meta, ok := self.Value[__meta].(*Object); ok {
					if v, ok := meta.Value[__setmeta]; ok {
						return v, nil
					}
				}
				self.Value[__meta] = possibleNewMeta
				return self, nil
			}
		}
	}
	return NilValue, nil
}

func objectgetMetaObject(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if meta, ok := self.Value[__meta].(*Object); ok {
				if v, ok := meta.Value[__getmeta]; ok {
					return v, nil
				}
				return meta, nil
			}
		}
	}
	return NilValue, nil
}

func objectGetValue(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if val, ok := self.Value[args[1].ObjectKey()]; ok {
				return val, nil
			}
		}
	}
	return NilValue, nil
}

func objectSearchValueInProtoChain(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if proto, ok := self.Value[__proto]; ok {
				return proto.IGet(args[1])
			}
		}
	}
	return NilValue, nil
}

func objectSetValue(args ...Value) (Value, error) {
	if len(args) > 2 {
		if self, ok := args[0].(*Object); ok {
			self.Value[args[1].ObjectKey()] = args[2]
		}
	}
	return NilValue, nil
}

func objectHasValue(args ...Value) (Value, error) {
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

func objectGetKeys(args ...Value) (Value, error) {
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

func objectGetValues(args ...Value) (Value, error) {
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

func objectrecursiveProtoCheck(set map[string]bool, self *Object) {
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
		objectrecursiveProtoCheck(set, proto)
	}
}
