package vida

import (
	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type NilValue struct {
	ValueSemanticsImpl
}

func (n NilValue) Boolean() Bool {
	return False
}

func (n NilValue) Prefix(ctx *Context, op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return True, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (n NilValue) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return Nil, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, n, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (n NilValue) Equals(ctx *Context, other Value) Bool {
	_, ok := other.(NilValue)
	return Bool(ok)
}

func (n NilValue) String() string {
	return nilT
}

func (n NilValue) ObjectKey() string {
	return nilT
}

func (n NilValue) Type() string {
	return nilT
}

func (n NilValue) Clone() Value {
	return n
}

func (n NilValue) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[nilT] == nil {
		ctx.loadNilVT()
	}
	if vtable, ok := ctx.vtables[nilT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}
