package vida

import (
	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Bool bool

func (b Bool) Boolean() Bool {
	return b
}

func (b Bool) Prefix(op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return !b, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (b Bool) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		if b {
			return rhs, nil
		}
		return b, nil
	case uint64(token.OR):
		if b {
			return b, nil
		}
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(b, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (b Bool) Get(ctx *Context, index Value) Value {
	return Nil
}

func (b Bool) Set(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (b Bool) Equals(other Value) Bool {
	val, isBool := other.(Bool)
	return Bool(isBool && b == val)
}

func (b Bool) IsIterable() Bool {
	return false
}

func (b Bool) IsCallable() Bool {
	return false
}

func (b Bool) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (b Bool) Iterator() Value {
	return Nil
}

func (b Bool) String(ctx *Context) string {
	if b {
		return "true"
	}
	return "false"
}

func (b Bool) ObjectKey() string {
	if b {
		return "true"
	}
	return "false"
}

func (b Bool) LookUp(*Context, Value) Value {
	return Nil
}

func (b Bool) Type(ctx *Context) string {
	return "bool"
}

func (b Bool) Clone() Value {
	return b
}
