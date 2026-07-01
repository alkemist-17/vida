package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type VidaError struct {
	ReferenceSemanticsImpl
	Message Value
}

func (e *VidaError) Boolean() Bool {
	return false
}

func (e *VidaError) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return True, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (e *VidaError) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return e, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, e, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (e *VidaError) Get(ctx *Context, index Value) Value {
	if val, ok := index.(*String); ok && val.Value == errorMessageFieldName {
		return e.Message
	}
	return Nil
}

func (e *VidaError) Set(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (e *VidaError) Equals(ctx *Context, other Value) Bool {
	v, isError := other.(*VidaError)
	return Bool(isError) && e == v && e.Message.Equals(ctx, v.Message)
}

func (e *VidaError) IsIterable() Bool {
	return false
}

func (e *VidaError) IsCallable() Bool {
	return false
}

func (e *VidaError) Iterator() Value {
	return Nil
}

func (e *VidaError) String() string {
	return fmt.Sprintf("error[message: %v]", e.Message.String())
}

func (e *VidaError) ObjectKey() string {
	return fmt.Sprintf("error[message: %v]", e.Message.ObjectKey())
}

func (e *VidaError) Type() string {
	return errorT
}

func (e *VidaError) GetVTable(ctx *Context) Value {
	if ctx.vtables[errorT] == nil {
		ctx.loadErrorVT()
	}
	return ctx.vtables[errorT]
}

func (e *VidaError) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[errorT] == nil {
		ctx.loadErrorVT()
	}
	if vtable, ok := ctx.vtables[errorT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (e *VidaError) Clone() Value {
	return e
}
