package vida

func loadObjectLib() Value {
	m := &Object{Value: make(map[string]Value, 16)}
	m.Value["inject"] = NativeFunction(objectInjectProperties)
	m.Value["override"] = NativeFunction(objectInjectAndOverrideProperties)
	m.Value["extract"] = NativeFunction(objectExtractProperties)
	m.Value["conforms"] = NativeFunction(objectCheckProperties)
	m.Value["implements"] = NativeFunction(objectCheckProperties)
	m.Value["set"] = NativeFunction(objectSetValue)
	m.Value["get"] = NativeFunction(objectGetValue)
	m.Value["has"] = NativeFunction(objectHasValue)
	m.Value["del"] = NativeFunction(objectDeleteProperty)
	m.Value["keys"] = NativeFunction(objectGetKeys)
	m.Value["values"] = NativeFunction(objectGetValues)
	m.Value["isEmpty"] = NativeFunction(objectIsEmpty)
	m.Value["isObject"] = NativeFunction(objectIsObject)
	m.Value["isCallable"] = NativeFunction(objectIsCallable)
	m.Value["clear"] = NativeFunction(objectClear)
	m.Value["getset"] = NativeFunction(objectGetOrSet)
	return m
}

func objectInjectProperties(ctx *Context, args ...Value) (Value, error) {
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
	return Nil, nil
}

func objectInjectAndOverrideProperties(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k, x := range other.Value {
						self.Value[k] = x
					}
				}
			}
			return self, nil
		}
	}
	return Nil, nil
}

func objectCheckProperties(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			set := make(map[string]bool)
			for _, v := range args[1:] {
				if other, ok := v.(*Object); ok && other != self {
					for k := range other.Value {
						set[k] = false
					}
				}
			}
			objectrecursiveMetaSearch(set, self)
			for _, v := range set {
				if !v {
					return False, nil
				}
			}
			return True, nil
		}
	}
	return Nil, nil
}

func objectExtractProperties(ctx *Context, args ...Value) (Value, error) {
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
	return Nil, nil
}

func objectDeleteProperty(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, prop := range args[1:] {
				delete(self.Value, prop.ObjectKey())
			}
			return self, nil
		}
	}
	return Nil, nil
}

func objectGetValue(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if val, ok := self.Value[args[1].ObjectKey()]; ok {
				return val, nil
			}
		}
	}
	return Nil, nil
}

func objectSetValue(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 2 && (l-1)%2 == 0 {
		if self, ok := args[0].(*Object); ok {
			for i := 1; i < l; i += 2 {
				self.Value[args[i].ObjectKey()] = args[i+1]
			}
			return self, nil
		}
	}
	return Nil, nil
}

func objectGetOrSet(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if self, ok := args[0].(*Object); ok {
			if val, ok := self.Value[args[1].ObjectKey()]; ok {
				return val, nil
			}
			if l > 2 {
				self.Value[args[1].ObjectKey()] = args[2]
				return self, nil
			}
		}
	}
	return Nil, nil
}

func objectHasValue(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			for _, val := range args[1:] {
				if _, exists := self.Value[val.ObjectKey()]; !exists {
					return False, nil
				}
			}
			return True, nil
		}
	}
	return Nil, nil
}

func objectGetKeys(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			lobj := len(self.Value)
			keys := make([]Value, int(lobj))
			var idx int
			for k := range self.Value {
				keys[idx] = &String{Value: k}
				idx++
			}
			return &Array{Value: keys}, nil
		}
	}
	return Nil, nil
}

func objectGetValues(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			lobj := len(self.Value)
			values := make([]Value, int(lobj))
			var idx int
			for _, v := range self.Value {
				values[idx] = v
				idx++
			}
			return &Array{Value: values}, nil
		}
	}
	return Nil, nil
}

func objectrecursiveMetaSearch(set map[string]bool, self *Object) {
	if self == nil {
		return
	}
	for k := range self.Value {
		if _, isPresent := set[k]; isPresent {
			set[k] = true
		}
	}
}

func objectIsEmpty(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if self, ok := args[0].(*Object); ok {
			return Bool(len(self.Value) == 0), nil
		}
	}
	return Nil, nil
}

func objectIsObject(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		_, ok := args[0].(*Object)
		return Bool(ok), nil
	}
	return Nil, nil
}

func objectIsCallable(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if o, ok := args[0].(*Object); ok {
			return o.IsCallable(), nil
		}
	}
	return False, nil
}

func objectClear(ctx *Context, args ...Value) (Value, error) {
	for _, val := range args {
		if o, ok := val.(*Object); ok {
			for k := range o.Value {
				delete(o.Value, k)
			}
		}
	}
	return Nil, nil
}
