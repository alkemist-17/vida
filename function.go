package vida

import (
	"encoding/json"
	"fmt"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type freeVarsInfo struct {
	Id      string
	Index   int
	IsLocal Bool
}

type CoreFunction struct {
	ReferenceSemanticsImpl
	Code          []uint64
	FreeVarsInfo  []freeVarsInfo
	ScriptID      string
	FreeVarsCount int
	Arity         int
	IsVarArg      bool
}

type coreFNConfigType = uint8

const (
	cZT coreFNConfigType = iota
	cZF
	cNT
	cNF
)

func (c *CoreFunction) getConfigType() coreFNConfigType {
	switch c.IsVarArg {
	case true:
		if c.Arity == 0 {
			return cZT
		}
		return cNT
	default:
		if c.Arity == 0 {
			return cZF
		}
		return cNF
	}
}

func (c *CoreFunction) Boolean() Bool {
	return true
}

func (c *CoreFunction) Equals(other Value) Bool {
	f, ok := other.(*CoreFunction)
	return Bool(ok && c == f)
}

func (c *CoreFunction) Type() string {
	return "coreFunction"
}

func (f CoreFunction) String() string {
	return fmt.Sprintf("CoreFunction(arity = %v, isVar = %v, freeVarsInfo = %v)", f.Arity, f.IsVarArg, f.FreeVarsCount)
}

func (f *CoreFunction) Clone() Value {
	return f
}

type Function struct {
	ReferenceSemanticsImpl
	FreeVarStore []Value
	CoreFn       *CoreFunction
}

func (f *Function) Boolean() Bool {
	return true
}

func (f *Function) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (f *Function) Binop(ctx *Context, op uint64, r Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return f, nil
	case uint64(token.AND):
		return r, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, f, r)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (f *Function) Equals(other Value) Bool {
	of, ok := other.(*Function)
	return Bool(ok && f == of)
}

func (f *Function) IsCallable() Bool {
	return true
}

func (f *Function) Type() string {
	return "function"
}

func (f *Function) Clone() Value {
	return f
}

func (f Function) String() string {
	return fmt.Sprintf("Function(%p)", f.CoreFn)
}

func (f *Function) ObjectKey() string {
	return fmt.Sprintf("Function(%p)", f.CoreFn)
}

type NativeFunction func(ctx *Context, args ...Value) (Value, error)

func (nativeFn NativeFunction) Boolean() Bool {
	return True
}

func (nativeFn NativeFunction) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (nativeFn NativeFunction) Binop(ctx *Context, op uint64, r Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return nativeFn, nil
	case uint64(token.AND):
		return r, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, nativeFn, r)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (nativeFn NativeFunction) Get(ctx *Context, index Value) (Value, error) {
	return Nil, verror.ErrValueNotIndexable
}

func (nativeFn NativeFunction) Set(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (nativeFn NativeFunction) Equals(other Value) Bool {
	return false
}

func (nativeFn NativeFunction) IsIterable() Bool {
	return false
}

func (nativeFn NativeFunction) IsCallable() Bool {
	return true
}

func (nativeFn NativeFunction) Call(ctx *Context, args ...Value) (Value, error) {
	return nativeFn(ctx, args...)
}

func (nativeFn NativeFunction) Iterator() Value {
	return Nil
}

func (nativeFn NativeFunction) String() string {
	return "NativeFunction"
}

func (nativeFn NativeFunction) ObjectKey() string {
	return "NativeFunction"
}

func (nativeFn NativeFunction) GetVTable() Value {
	return Nil
}

func (nativeFn NativeFunction) Clone() Value {
	return nativeFn
}

func (nativeFn NativeFunction) Type() string {
	return "NativeFunction"
}

func (nativeFn NativeFunction) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}
