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
			switch method := o.LookUp(ctx, tokenOPToString(token.ADD)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			default:
				pairs := make(map[string]Value, len(o.Value)+len(r.Value))
				maps.Copy(pairs, o.Value)
				maps.Copy(pairs, r.Value)
				return &Object{Value: pairs}, nil
			}
		case uint64(token.SUB):
			switch method := o.LookUp(ctx, tokenOPToString(token.SUB)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			default:
				pairs := make(map[string]Value)
				for k, v := range o.Value {
					if _, contains := r.Value[k]; !contains {
						pairs[k] = v
					}
				}
				return &Object{Value: pairs}, nil
			}
		case uint64(token.MUL):
			switch method := o.LookUp(ctx, tokenOPToString(token.MUL)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.DIV):
			switch method := o.LookUp(ctx, tokenOPToString(token.DIV)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.REM):
			switch method := o.LookUp(ctx, tokenOPToString(token.REM)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.POW):
			switch method := o.LookUp(ctx, tokenOPToString(token.POW)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.LT):
			switch method := o.LookUp(ctx, tokenOPToString(token.LT)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.LE):
			switch method := o.LookUp(ctx, tokenOPToString(token.LE)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.GT):
			switch method := o.LookUp(ctx, tokenOPToString(token.GT)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.GE):
			switch method := o.LookUp(ctx, tokenOPToString(token.GE)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.BAND):
			switch method := o.LookUp(ctx, tokenOPToString(token.BAND)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			default:
				pairs := make(map[string]Value)
				for k := range o.Value {
					if x, contains := r.Value[k]; contains {
						pairs[k] = x
					}
				}
				return &Object{Value: pairs}, nil
			}
		case uint64(token.BOR):
			switch method := o.LookUp(ctx, tokenOPToString(token.BOR)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			default:
				pairs := make(map[string]Value, len(o.Value)+len(r.Value))
				maps.Copy(pairs, o.Value)
				maps.Copy(pairs, r.Value)
				return &Object{Value: pairs}, nil
			}
		case uint64(token.BXOR):
			switch method := o.LookUp(ctx, tokenOPToString(token.BXOR)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.BSHL):
			switch method := o.LookUp(ctx, tokenOPToString(token.BSHL)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		case uint64(token.BSHR):
			switch method := o.LookUp(ctx, tokenOPToString(token.BSHR)).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
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
	default:
		switch op {
		case uint64(token.ADD), uint64(token.SUB), uint64(token.MUL),
			uint64(token.DIV), uint64(token.REM), uint64(token.POW),
			uint64(token.LT), uint64(token.LE), uint64(token.GT),
			uint64(token.GE), uint64(token.BAND), uint64(token.BOR),
			uint64(token.BXOR), uint64(token.BSHL), uint64(token.BSHR):
			switch method := o.LookUp(ctx, tokenOPToString(token.Token(op))).(type) {
			case *Function:
				return ctx.runFunctionInNewThread(method, o, rhs)
			case NativeFunction:
				return method.Call(ctx, o, r)
			}
		}
	}
	switch op {
	case uint64(token.OR):
		return o, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, o, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (o *Object) Get(ctx *Context, message Value) Value {
	current := o
	for current != nil {
		if val, ok := current.Value[message.ObjectKey()]; ok {
			return val
		}
		current = current.VTable
	}
	return Nil
}

func (o *Object) Set(index, val Value) error {
	o.Value[index.ObjectKey()] = val
	return nil
}

func (o *Object) Equals(ctx *Context, other Value) Bool {
	switch method := o.LookUp(ctx, tokenOPToString(token.EQ)).(type) {
	case *Function:
		if val, err := ctx.runFunctionInNewThread(method, o, other); err == nil {
			return val.Boolean()
		}
		return False
	case NativeFunction:
		if val, err := method.Call(ctx, o, other); err == nil {
			return val.Boolean()
		}
		return False
	default:
		val, isObject := other.(*Object)
		return Bool(isObject && o == val)
	}
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
		return "object[]"
	}

	ptr := reflect.ValueOf(o).Pointer()
	if visited[ptr] {
		return "object[...]"
	}

	visited[ptr] = true
	defer delete(visited, ptr)

	var r []string
	for k, v := range o.Value {
		r = append(r, fmt.Sprintf("%v: %v", k, stringWithVisited(v, visited)))
	}
	return fmt.Sprintf("object[%v]", strings.Join(r, ", "))
}

func (o *Object) ObjectKey() string {
	return fmt.Sprintf("object[%p]", o)
}

func (o *Object) GetVTable(ctx *Context) Value {
	if o.VTable != nil {
		return o.VTable
	}
	if ctx.vtables[objectT] == nil {
		ctx.loadObjectVT()
	}
	return ctx.vtables[objectT]
}

func (o *Object) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[objectT] == nil {
		ctx.loadObjectVT()
	}
	if o.VTable != nil {
		if val := o.VTable.Get(ctx, message); !val.Equals(ctx, Nil) {
			return val
		}
	}
	return ctx.vtables[objectT].Get(ctx, message)
}

func (o *Object) Type() string {
	return objectT
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
	m.Value["set"] = NativeFunction(objectCircumventSetValue)
	m.Value["get"] = NativeFunction(objectCircumventGetValue)
	m.Value["has"] = NativeFunction(objectCircumventHasValue)
	m.Value["del"] = NativeFunction(objectCircumventDeleteProperty)
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

func objectCircumventDeleteProperty(ctx *Context, args ...Value) (Value, error) {
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

func objectCircumventGetValue(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if self, ok := args[0].(*Object); ok {
			if val, ok := self.Value[args[1].ObjectKey()]; ok {
				return val, nil
			}
		}
	}
	return Nil, nil
}

func objectCircumventSetValue(ctx *Context, args ...Value) (Value, error) {
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

func objectCircumventHasValue(ctx *Context, args ...Value) (Value, error) {
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
