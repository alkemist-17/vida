package vida

import (
	"fmt"
	"math"
	"strconv"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Integer int64

func (i Integer) Boolean() Bool {
	return True
}

func (i Integer) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.SUB):
		return -i, nil
	case uint64(token.NOT):
		return False, nil
	case uint64(token.ADD):
		return i, nil
	case uint64(token.TILDE):
		return Integer(^uint32(i)), nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (l Integer) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case Integer:
		switch op {
		case uint64(token.ADD):
			return l + r, nil
		case uint64(token.SUB):
			return l - r, nil
		case uint64(token.MUL):
			return l * r, nil
		case uint64(token.DIV):
			if r == 0 {
				return Nil, verror.ErrDivisionByZero
			}
			return l / r, nil
		case uint64(token.REM):
			if r == 0 {
				return Nil, verror.ErrDivisionByZero
			}
			return l % r, nil
		case uint64(token.LT):
			return Bool(l < r), nil
		case uint64(token.LE):
			return Bool(l <= r), nil
		case uint64(token.GT):
			return Bool(l > r), nil
		case uint64(token.GE):
			return Bool(l >= r), nil
		case uint64(token.BXOR):
			return Integer(uint32(l) ^ uint32(r)), nil
		case uint64(token.BOR):
			return Integer(uint32(l) | uint32(r)), nil
		case uint64(token.BAND):
			return Integer(uint32(l) & uint32(r)), nil
		case uint64(token.BSHL):
			return Integer(uint32(l) << uint32(r)), nil
		case uint64(token.BSHR):
			return Integer(uint32(l) >> uint32(r)), nil
		case uint64(token.POW):
			return Integer(math.Pow(float64(l), float64(r))), nil
		}
	case Float:
		switch op {
		case uint64(token.ADD):
			return Float(Float(l) + r), nil
		case uint64(token.SUB):
			return Float(Float(l) - r), nil
		case uint64(token.MUL):
			return Float(Float(l) * r), nil
		case uint64(token.DIV):
			return Float(Float(l) / r), nil
		case uint64(token.REM):
			return Float(math.Remainder(float64(l), float64(r))), nil
		case uint64(token.LT):
			return Bool(Float(l) < r), nil
		case uint64(token.LE):
			return Bool(Float(l) <= r), nil
		case uint64(token.GT):
			return Bool(Float(l) > r), nil
		case uint64(token.GE):
			return Bool(Float(l) >= r), nil
		case uint64(token.POW):
			return Float(math.Pow(float64(l), float64(r))), nil
		}
	}
	switch op {
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.OR):
		return l, nil
	case uint64(token.IN):
		return IsMemberOf(l, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (i Integer) Get(ctx *Context, index Value) Value {
	return Nil
}

func (i Integer) Set(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (i Integer) Equals(other Value) Bool {
	if val, ok := other.(Integer); ok {
		return i == val
	}
	if val, ok := other.(Float); ok {
		return i == Integer(val)
	}
	return false
}

func (i Integer) IsIterable() Bool {
	return true
}

func (i Integer) IsCallable() Bool {
	return false
}

func (i Integer) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (i Integer) Iterator() Value {
	if i < 0 {
		i = -i
	}
	return &IntegerIterator{Init: -1, End: i}
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) ObjectKey() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) LookUp(ctx *Context, message Value) Value {
	return Nil
}

func (i Integer) Type() string {
	return "int"
}

func (i Integer) Clone() Value {
	return i
}

type Float float64

func (f Float) Boolean() Bool {
	return True
}

func (f Float) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.SUB):
		return -f, nil
	case uint64(token.NOT):
		return False, nil
	case uint64(token.ADD):
		return f, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (f Float) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case Float:
		switch op {
		case uint64(token.ADD):
			return f + r, nil
		case uint64(token.SUB):
			return f - r, nil
		case uint64(token.MUL):
			return f * r, nil
		case uint64(token.DIV):
			return f / r, nil
		case uint64(token.REM):
			return Float(math.Remainder(float64(f), float64(r))), nil
		case uint64(token.LT):
			return Bool(f < r), nil
		case uint64(token.LE):
			return Bool(f <= r), nil
		case uint64(token.GT):
			return Bool(f > r), nil
		case uint64(token.GE):
			return Bool(f >= r), nil
		case uint64(token.POW):
			return Float(math.Pow(float64(f), float64(r))), nil
		}
	case Integer:
		switch op {
		case uint64(token.ADD):
			return f + Float(r), nil
		case uint64(token.SUB):
			return f - Float(r), nil
		case uint64(token.MUL):
			return f * Float(r), nil
		case uint64(token.DIV):
			return f / Float(r), nil
		case uint64(token.REM):
			return Float(math.Remainder(float64(f), float64(r))), nil
		case uint64(token.LT):
			return Bool(f < Float(r)), nil
		case uint64(token.LE):
			return Bool(f <= Float(r)), nil
		case uint64(token.GT):
			return Bool(f > Float(r)), nil
		case uint64(token.GE):
			return Bool(f >= Float(r)), nil
		case uint64(token.POW):
			return Float(math.Pow(float64(f), float64(r))), nil
		}
	}
	switch op {
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.OR):
		return f, nil
	case uint64(token.IN):
		return IsMemberOf(f, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (f Float) Get(ctx *Context, index Value) Value {
	return Nil
}

func (f Float) Set(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (f Float) Equals(other Value) Bool {
	if val, ok := other.(Float); ok {
		return f == val
	}
	if val, ok := other.(Integer); ok {
		return f == Float(val)
	}
	return false
}

func (f Float) IsIterable() Bool {
	return false
}

func (f Float) IsCallable() Bool {
	return false
}

func (f Float) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (f Float) Iterator() Value {
	return Nil
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'g', -1, 64)
}

func (f Float) ObjectKey() string {
	return fmt.Sprintf("%vf", strconv.FormatFloat(float64(f), 'g', -1, 64))
}

func (f Float) LookUp(ctx *Context, message Value) Value {
	return Nil
}

func (f Float) Type() string {
	return "float"
}

func (f Float) Clone() Value {
	return f
}
