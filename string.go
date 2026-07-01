package vida

import (
	"encoding/json"
	"strings"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type String struct {
	ReferenceSemanticsImpl
	Runes []rune
	Value string
}

func (s *String) Boolean() Bool {
	return True
}

func (s *String) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *String:
		switch op {
		case uint64(token.ADD):
			if len(s.Value)+len(r.Value) >= verror.MaxMemSize {
				return Nil, verror.ErrMaxMemSize
			}
			str := &String{Value: s.Value + r.Value}
			return str, nil
		case uint64(token.LT):
			return Bool(s.Value < r.Value), nil
		case uint64(token.LE):
			return Bool(s.Value <= r.Value), nil
		case uint64(token.GT):
			return Bool(s.Value > r.Value), nil
		case uint64(token.GE):
			return Bool(s.Value >= r.Value), nil
		case uint64(token.IN):
			return Bool(strings.Contains(r.Value, s.Value)), nil
		}
	}
	switch op {
	case uint64(token.OR):
		return s, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, s, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (s *String) Get(ctx *Context, index Value) Value {
	switch r := index.(type) {
	case Integer:
		if s.Runes == nil {
			s.Runes = []rune(s.Value)
		}
		l := Integer(len(s.Runes))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			sr := s.Runes[r : r+Integer(1)]
			return &String{Value: string(sr), Runes: sr}
		}
	}
	return Nil
}

func (s *String) Set(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (s *String) Prefix(ctx *Context, op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return False, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (s *String) Equals(ctx *Context, other Value) Bool {
	if val, ok := other.(*String); ok {
		return s.Value == val.Value
	}
	if val, ok := other.(*Bytes); ok {
		return s.Value == string(val.Value)
	}
	return false
}

func (s *String) IsIterable() Bool {
	return true
}

func (s *String) IsCallable() Bool {
	return false
}

func (s *String) Iterator() Value {
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	return &StringIterator{Runes: s.Runes, Init: -1, End: len(s.Runes)}
}

func (s String) String() string {
	return s.Value
}

func (s *String) ObjectKey() string {
	return s.Value
}

func (s *String) GetVTable(ctx *Context) Value {
	if ctx.vtables[stringT] == nil {
		ctx.loadStringVT()
	}
	return ctx.vtables[stringT]
}

func (s *String) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[stringT] == nil {
		ctx.loadStringVT()
	}
	if vtable, ok := ctx.vtables[stringT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (s *String) Type() string {
	return stringT
}

func (s *String) Clone() Value {
	return s
}

func (s *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}
