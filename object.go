package vida

import (
	cryptoRand "crypto/rand"
)

func loadObjectLib() Value {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
		((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
	}
	if __meta == inititalMetaName {
		__meta = cryptoRand.Text()
	}
	m := &Object{Value: make(map[string]Value, 21)}
	m.Value["inject"] = GFn(objectInjectProperties)
	m.Value["override"] = GFn(objectInjectAndOverrideProperties)
	m.Value["extract"] = GFn(objectExtractProperties)
	m.Value["conforms"] = GFn(objectCheckProperties)
	m.Value["implements"] = GFn(objectCheckProperties)
	m.Value["extends"] = GFn(objectSetMeta)
	m.Value["setmeta"] = GFn(objectSetMeta)
	m.Value["getmeta"] = GFn(objectGetMeta)
	m.Value["hasmeta"] = GFn(objectHasMeta)
	m.Value["delmeta"] = GFn(objectDelMeta)
	m.Value["set"] = GFn(objectSetValue)
	m.Value["get"] = GFn(objectGetValue)
	m.Value["has"] = GFn(objectHasValue)
	m.Value["del"] = GFn(objectDeleteProperty)

	m.Value["keys"] = GFn(objectGetKeys)
	m.Value["values"] = GFn(objectGetValues)
	m.Value["isEmpty"] = GFn(objectIsEmpty)
	m.Value["isObject"] = GFn(objectIsObject)
	m.Value["isCallable"] = GFn(objectIsCallable)
	m.Value["clear"] = GFn(objectClear)
	m.Value["getset"] = GFn(objectGetOrSet)
	return m
}

func objectInjectProperties(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k, x := range other.Value {
						if _, isPresent := self.Value[k]; !isPresent && k != __meta {
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
					for k, x := range other.Value {
						if k != __meta {
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

func objectCheckProperties(args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			set := make(map[string]bool)
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k := range other.Value {
						if k != __meta {
							set[k] = false
						}
					}
				}
			}
			objectrecursiveMetaSearch(set, self)
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
						if k != __meta {
							delete(self.Value, k)
						}
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
			for _, prop := range args[1:] {
				delete(self.Value, prop.ObjectKey())
			}
			return self, nil
		}
	}
	return NilValue, nil
}

func objectSetMeta(args ...Value) (Value, error) {
	if len(args) > 1 {
		self, selfIsObj := args[0].(*Object)
		maybeMeta, metaIsObj := args[1].(*Object)
		if selfIsObj && metaIsObj {
			if meta, ok := self.Value[__meta].(*Object); ok {
				if v, ok := meta.Value[__setmeta]; ok {
					return v, nil
				}
			}
			self.Value[__meta] = maybeMeta
			return self, nil
		}
	}
	return NilValue, nil
}

func objectGetMeta(args ...Value) (Value, error) {
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

func objectGetOrSet(args ...Value) (Value, error) {
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
			if _, ok := self.Value[__meta]; ok {
				lobj--
			}
			keys := make([]Value, int(lobj))
			var idx int
			for k := range self.Value {
				if k != __meta {
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
			if _, ok := self.Value[__meta]; ok {
				lobj--
			}
			values := make([]Value, int(lobj))
			var idx int
			for k, v := range self.Value {
				if k != __meta {
					values[idx] = v
					idx++
				}
			}
			return &Array{Value: values}, nil
		}
	}
	return NilValue, nil
}

func objectrecursiveMetaSearch(set map[string]bool, self *Object) {
	if self == nil {
		return
	}
	for k := range self.Value {
		if k != __meta {
			if _, isPresent := set[k]; isPresent {
				set[k] = true
			}
		}
	}
	if meta, ok := self.Value[__meta].(*Object); ok {
		objectrecursiveMetaSearch(set, meta)
	}
}

func objectIsEmpty(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			l := len(self.Value)
			if _, ok := self.Value[__meta]; ok {
				l--
			}
			return Bool(l == 0), nil
		}
	}
	return NilValue, nil
}

func objectIsObject(args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(*Object)
		return Bool(ok), nil
	}
	return NilValue, nil
}

func objectIsCallable(args ...Value) (Value, error) {
	if len(args) > 0 {
		if o, ok := args[0].(*Object); ok {
			return o.IsCallable(), nil
		}
	}
	return Bool(false), nil
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

func objectHasMeta(args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			if _, ok := self.Value[__meta].(*Object); ok {
				return Bool(true), nil
			}
			return Bool(false), nil
		}
	}
	return NilValue, nil
}

func objectDelMeta(args ...Value) (Value, error) {
	if len(args) > 0 {
		for _, v := range args {
			if self, ok := v.(*Object); ok {
				if meta, ok := self.Value[__meta].(*Object); ok {
					if _, ok := meta.Value[__setmeta]; !ok {
						delete(self.Value, __meta)
					}
				}
			}
		}
	}
	return NilValue, nil
}
