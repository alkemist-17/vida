package vida

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Object struct {
	ReferenceSemanticsImpl
	Value  map[string]Value
	VTable *Object
}

func (o *Object) Boolean() Bool {
	return true
}

func (o *Object) Prefix(op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return False, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (o *Object) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *Object:
		switch op {
		case uint64(token.ADD):
			pairs := make(map[string]Value, len(o.Value)+len(r.Value))
			maps.Copy(pairs, o.Value)
			maps.Copy(pairs, r.Value)
			return &Object{Value: pairs}, nil
		case uint64(token.SUB):
			pairs := make(map[string]Value)
			for k, v := range o.Value {
				if _, contains := r.Value[k]; !contains {
					pairs[k] = v
				}
			}
			return &Object{Value: pairs}, nil
		case uint64(token.BAND):
			pairs := make(map[string]Value)
			for k := range o.Value {
				if x, contains := r.Value[k]; contains {
					pairs[k] = x
				}
			}
			return &Object{Value: pairs}, nil
		case uint64(token.BOR):
			pairs := make(map[string]Value, len(o.Value)+len(r.Value))
			maps.Copy(pairs, o.Value)
			maps.Copy(pairs, r.Value)
			return &Object{Value: pairs}, nil
		case uint64(token.VTABLE):
			o.VTable = r
			return o, nil
		}
	case NilValue:
		switch op {
		case uint64(token.VTABLE):
			o.VTable = nil
			return o, nil
		}
	}
	switch op {
	case uint64(token.OR):
		return o, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(o, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (o *Object) Get(ctx *Context, index Value) Value {
	if val, ok := o.Value[index.ObjectKey()]; ok {
		return val
	}
	if o.VTable != nil {
		return o.VTable.Get(ctx, index)
	}
	return Nil
}

func (o *Object) Set(index, val Value) error {
	o.Value[index.ObjectKey()] = val
	return nil
}

func (o *Object) Equals(other Value) Bool {
	val, isObject := other.(*Object)
	return Bool(isObject && o == val)
}

func (o *Object) IsIterable() Bool {
	return true
}

func (o *Object) IsCallable() Bool {
	return false
}

func (o *Object) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (o *Object) Iterator() Value {
	return newObjectIterator(o)
}

func (o *Object) String() string {
	return o.stringify(make(map[uintptr]bool))
}

func (o *Object) stringify(visited map[uintptr]bool) string {
	if len(o.Value) == 0 {
		return "{}"
	}

	ptr := reflect.ValueOf(o).Pointer()
	if visited[ptr] {
		return "{...}"
	}

	visited[ptr] = true
	defer delete(visited, ptr)

	var r []string
	for k, v := range o.Value {
		r = append(r, fmt.Sprintf("%v: %v", k, stringWithVisited(v, visited)))
	}
	return fmt.Sprintf("{%v}", strings.Join(r, ", "))
}

func (o *Object) ObjectKey() string {
	return fmt.Sprintf("Object(%p)", o)
}

func (o *Object) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[objectVT] == nil {
		ctx.vtables[objectVT] = loadObjectVT()
	}
	if o.VTable != nil {
		if val := o.VTable.Get(ctx, message); !val.Equals(Nil) {
			return val
		}
	}
	return ctx.vtables[objectVT].Get(ctx, message)
}

func (o *Object) Type(ctx *Context) string {
	if o.VTable != nil {
		return o.VTable.Get(ctx, &String{Value: __type}).String()
	}
	return "object"
}

func (o *Object) Clone() Value {
	m := make(map[string]Value, len(o.Value))
	for k, v := range o.Value {
		m[k] = v.Clone()
	}
	return &Object{Value: m, VTable: o.VTable}
}

func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

func loadObjectLib() Value {
	m := &Object{Value: make(map[string]Value, 15)}
	m.Value["inject"] = NativeFunction(objectInjectProperties)
	m.Value["override"] = NativeFunction(objectInjectAndOverrideProperties)
	m.Value["extract"] = NativeFunction(objectExtractProperties)
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
