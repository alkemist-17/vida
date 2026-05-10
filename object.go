package vida

import (
	cryptoRand "crypto/rand"
	"maps"
)

func loadObjectLib() Value {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
		((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
	}
	if __proto == initProtName {
		__proto = cryptoRand.Text()
	}
	m := &Object{Value: make(map[string]Value, 18)}
	m.Value["inject"] = GFn(objectInjectProperties)
	m.Value["extends"] = GFn(objectSetPrototype)
	m.Value["extract"] = GFn(objectExtractProperties)
	m.Value["override"] = GFn(objectInjectAndOverrideProperties)
	m.Value["conforms"] = GFn(objectCheckProperties)
	m.Value["setproto"] = GFn(objectSetPrototype)
	m.Value["getproto"] = GFn(objectGetPrototype)
	m.Value["hasproto"] = GFn(objectHasPrototype)
	m.Value["delproto"] = GFn(objectDelPrototype)
	m.Value["getOrInsert"] = GFn(objectGetOrInsert)
	m.Value["get"] = GFn(objectGetValue)
	m.Value["set"] = GFn(objectSetValue)
	m.Value["has"] = GFn(objectHasValue)
	m.Value["del"] = GFn(objectDeleteProperty)
	m.Value["keys"] = GFn(objectGetKeys)
	m.Value["values"] = GFn(objectGetValues)
	m.Value["isEmpty"] = GFn(objectIsEmpty)
	m.Value["clear"] = GFn(objectClear)
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
						if k != __proto {
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
			if proto, ok := self.Value[__proto].(*Object); ok {
				if __delete, ok := proto.Value[__del]; ok {
					return __delete, nil
				}
			}
			for _, prop := range args[1:] {
				delete(self.Value, prop.ObjectKey())
			}
			return self, nil
		}
	}
	return NilValue, nil
}

func objectSetPrototype(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if proto, ok := args[1].(*Object); ok {
				if proto, ok := self.Value[__proto].(*Object); ok {
					if v, ok := proto.Value[__setproto]; ok {
						return v, nil
					}
				}
				self.Value[__proto] = proto
				return self, nil
			}
		}
	}
	return NilValue, nil
}

func objectGetPrototype(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if proto, ok := self.Value[__proto].(*Object); ok {
				if v, ok := proto.Value[__getproto]; ok {
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

func objectSetValue(args ...Value) (Value, error) {
	l := len(args)
	if l > 2 && (l-1)%2 == 0 {
		if self, ok := args[0].(*Object); ok {
			for i := 1; i < l; i += 2 {
				self.Value[args[i].ObjectKey()] = args[i+1]
			}
			return self, nil
		}
	}
	return NilValue, nil
}

func objectGetOrInsert(args ...Value) (Value, error) {
	if len(args) > 2 {
		if self, ok := args[0].(*Object); ok {
			if val, ok := self.Value[args[1].ObjectKey()]; ok {
				return val, nil
			}
			self.Value[args[1].ObjectKey()] = args[2]
			return self, nil
		}
	}
	return NilValue, nil
}

func objectHasValue(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, val := range args[1:] {
				if _, exists := self.Value[val.ObjectKey()]; !exists {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
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
			keys := make([]Value, int(lobj))
			var idx int
			for k := range self.Value {
				if k != __proto {
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
			values := make([]Value, int(lobj))
			var idx int
			for k, v := range self.Value {
				if k != __proto {
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
		if k != __proto {
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

func objectIsEmpty(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			l := len(self.Value)
			if _, ok := self.Value[__proto]; ok {
				l--
			}
			return Bool(l == 0), nil
		}
	}
	return NilValue, nil
}

func objectClear(args ...Value) (Value, error) {
	for _, val := range args {
		if o, ok := val.(*Object); ok {
			for k := range o.Value {
				delete(o.Value, k)
			}
		}
	}
	return NilValue, nil
}

func objectHasPrototype(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if _, ok := self.Value[__proto].(*Object); ok {
				return Bool(true), nil
			}
			return Bool(false), nil
		}
	}
	return NilValue, nil
}

func objectDelPrototype(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if proto, ok := self.Value[__proto].(*Object); ok {
				if dp, ok := proto.Value[__delproto]; ok {
					return dp, nil
				}
				delete(self.Value, __proto)
				return self, nil
			}
			return self, nil
		}
	}
	return NilValue, nil
}
