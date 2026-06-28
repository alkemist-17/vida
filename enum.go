package vida

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Enum struct {
	Pairs map[string]Integer
}

func (e *Enum) Boolean() Bool {
	return true
}

func (e *Enum) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (e *Enum) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return e, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(e, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (e *Enum) Get(ctx *Context, index Value) Value {
	if val, ok := e.Pairs[index.String()]; ok {
		return val
	}
	return Nil
}

func (e *Enum) Set(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (e *Enum) Equals(other Value) Bool {
	val, isEnum := other.(*Enum)
	return Bool(isEnum && val == other)
}

func (e *Enum) IsIterable() Bool {
	return false
}

func (e *Enum) Iterator() Value {
	return Nil
}

func (e *Enum) IsCallable() Bool {
	return false
}

func (e *Enum) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (e Enum) String() string {
	if len(e.Pairs) == 0 {
		return "enum{}"
	}
	var r []string
	for k, v := range e.Pairs {
		r = append(r, fmt.Sprintf("%v: %v", k, v))
	}
	return fmt.Sprintf("enum{%v}", strings.Join(r, ", "))
}

func (e *Enum) ObjectKey() string {
	return fmt.Sprintf("Enum(%p)", e)
}

func (e *Enum) LookUp(ctx *Context, message Value) Value {
	return Nil
}

func (e *Enum) Type(ctx *Context) string {
	return "enum"
}

func (e *Enum) Clone() Value {
	return e
}

func (e *Enum) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}
