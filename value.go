package vida

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Value interface {
	Boolean() Bool
	Prefix(uint64) (Value, error)
	Binop(uint64, Value) (Value, error)
	Get(Value) (Value, error)
	Set(Value, Value) error
	Equals(Value) Bool
	IsIterable() Bool
	Iterator() Value
	IsCallable() Bool
	Call(ctx *Context, args ...Value) (Value, error)
	String() string
	Type() string
	Clone() Value
	ObjectKey() string
	GetVTable() Value
}

type NilValue struct {
	ValueSemanticsImpl
}

func (n NilValue) Boolean() Bool {
	return False
}

func (n NilValue) Prefix(op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return True, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (n NilValue) Binop(op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return Nil, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(n, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (n NilValue) Equals(other Value) Bool {
	_, ok := other.(NilValue)
	return Bool(ok)
}

func (n NilValue) String() string {
	return "nil"
}

func (n NilValue) ObjectKey() string {
	return "nil"
}

func (n NilValue) Type() string {
	return "nil"
}

func (n NilValue) Clone() Value {
	return n
}

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

func (b Bool) Binop(op uint64, rhs Value) (Value, error) {
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

func (b Bool) Get(index Value) (Value, error) {
	return Nil, verror.ErrValueNotIndexable
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

func (b Bool) String() string {
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

func (b Bool) GetVTable() Value {
	return Nil
}

func (b Bool) Type() string {
	return "bool"
}

func (b Bool) Clone() Value {
	return b
}

type String struct {
	ReferenceSemanticsImpl
	Runes  []rune
	VTable Value
	Value  string
}

func (s *String) Boolean() Bool {
	return True
}

func (s *String) Binop(op uint64, rhs Value) (Value, error) {
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
		return IsMemberOf(s, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (s *String) Get(index Value) (Value, error) {
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
			return &String{Value: string(sr), Runes: sr}, nil
		}
	}
	return Nil, verror.ErrValueNotIndexable
}

func (s *String) Set(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (s *String) Prefix(op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return False, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (s *String) Equals(other Value) Bool {
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

func (s *String) GetVTable() Value {
	return s.VTable
}

func (s *String) Type() string {
	return "string"
}

func (s *String) Clone() Value {
	return s
}

func (s *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}

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

func (l Integer) Binop(op uint64, rhs Value) (Value, error) {
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

func (i Integer) Get(index Value) (Value, error) {
	return Nil, verror.ErrValueNotIndexable
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

func (i Integer) GetVTable() Value {
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

func (f Float) Binop(op uint64, rhs Value) (Value, error) {
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

func (f Float) Get(index Value) (Value, error) {
	return Nil, verror.ErrValueNotIndexable
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

func (f Float) GetVTable() Value {
	return Nil
}

func (f Float) Type() string {
	return "float"
}

func (f Float) Clone() Value {
	return f
}

type Array struct {
	ReferenceSemanticsImpl
	Value []Value
}

func (xs *Array) Boolean() Bool {
	return True
}

func (xs *Array) Prefix(op uint64) (Value, error) {
	if op == uint64(token.NOT) {
		return False, nil
	}
	return Nil, verror.ErrPrefixOpNotDefined
}

func (xs *Array) Binop(op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *Array:
		switch op {
		case uint64(token.ADD):
			rLen := len(r.Value)
			if rLen == 0 {
				return xs, nil
			}
			lLen := len(xs.Value)
			if rLen+lLen >= verror.MaxMemSize {
				return Nil, verror.ErrMaxMemSize
			}
			values := make([]Value, lLen+rLen)
			copy(values[:lLen], xs.Value)
			copy(values[lLen:], r.Value)
			return &Array{Value: values}, nil
		case uint64(token.IN):
			return IsMemberOf(xs, rhs)
		}
	}
	switch op {
	case uint64(token.OR):
		return xs, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(xs, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (xs *Array) Get(index Value) (Value, error) {
	switch r := index.(type) {
	case Integer:
		l := Integer(len(xs.Value))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			return xs.Value[r], nil
		}
	}
	return Nil, verror.ErrValueNotIndexable
}

func (xs *Array) Set(index, val Value) error {
	switch r := index.(type) {
	case Integer:
		l := Integer(len(xs.Value))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			xs.Value[r] = val
			return nil
		}
	}
	return verror.ErrValueNotIndexable
}

func (xs *Array) Equals(other Value) Bool {
	val, isArray := other.(*Array)
	return Bool(isArray && xs == val)
}

func (xs *Array) IsIterable() Bool {
	return true
}

func (xs *Array) IsCallable() Bool {
	return false
}

func (xs *Array) Iterator() Value {
	return &ArrayIterator{Array: xs.Value, Init: -1, End: len(xs.Value)}
}

func (xs *Array) String() string {
	return xs.stringify(make(map[uintptr]bool))
}

func (xs *Array) stringify(visited map[uintptr]bool) string {
	if len(xs.Value) == 0 {
		return "[]"
	}

	ptr := reflect.ValueOf(xs).Pointer()

	if visited[ptr] {
		return "[...]"
	}

	visited[ptr] = true
	defer delete(visited, ptr)

	var r []string
	for _, v := range xs.Value {
		r = append(r, stringWithVisited(v, visited))
	}
	return fmt.Sprintf("[%v]", strings.Join(r, ", "))
}

func (xs *Array) ObjectKey() string {
	return fmt.Sprintf("Array(%p)", xs)
}

func (xs *Array) Type() string {
	return "array"
}

func (xs *Array) Clone() Value {
	c := make([]Value, len(xs.Value))
	for i, v := range xs.Value {
		c[i] = v.Clone()
	}
	return &Array{Value: c}
}

func (xs *Array) MarshalJSON() ([]byte, error) {
	return json.Marshal(xs.Value)
}

type Object struct {
	ReferenceSemanticsImpl
	Value map[string]Value
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

func (o *Object) Binop(op uint64, rhs Value) (Value, error) {
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

func (o *Object) Get(index Value) (Value, error) {
	if val, ok := o.Value[index.ObjectKey()]; ok {
		return val, nil
	}
	return Nil, nil
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

func (o *Object) Type() string {
	return "object"
}

func (o *Object) Clone() Value {
	m := make(map[string]Value, len(o.Value))
	for k, v := range o.Value {
		m[k] = v.Clone()
	}
	return &Object{Value: m}
}

func (o *Object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Value)
}

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

func (f *Function) Binop(op uint64, r Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return f, nil
	case uint64(token.AND):
		return r, nil
	case uint64(token.IN):
		return IsMemberOf(f, r)
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

func (nativeFn NativeFunction) Binop(op uint64, r Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return nativeFn, nil
	case uint64(token.AND):
		return r, nil
	case uint64(token.IN):
		return IsMemberOf(nativeFn, r)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (nativeFn NativeFunction) Get(index Value) (Value, error) {
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

type VidaError struct {
	ReferenceSemanticsImpl
	Message Value
}

func (e *VidaError) Boolean() Bool {
	return false
}

func (e *VidaError) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return True, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (e *VidaError) Binop(op uint64, rhs Value) (Value, error) {
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

func (e *VidaError) Get(index Value) (Value, error) {
	if val, ok := index.(*String); ok && val.Value == errorMessageFieldName {
		return e.Message, nil
	}
	return Nil, nil
}

func (e *VidaError) Set(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (e *VidaError) Equals(other Value) Bool {
	v, isError := other.(*VidaError)
	return Bool(isError) && e.Message.Equals(v.Message)
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
	return fmt.Sprintf("Error(%v)", e.Message.String())
}

func (e *VidaError) ObjectKey() string {
	return fmt.Sprintf("Error(%v)", e.Message.ObjectKey())
}

func (e *VidaError) Type() string {
	return "error"
}

func (e *VidaError) Clone() Value {
	return e
}

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

func (e *Enum) Binop(op uint64, rhs Value) (Value, error) {
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

func (e *Enum) Get(index Value) (Value, error) {
	if val, ok := e.Pairs[index.String()]; ok {
		return val, nil
	}
	return Nil, nil
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

func (e *Enum) GetVTable() Value {
	return Nil
}

func (e *Enum) Type() string {
	return "enum"
}

func (e *Enum) Clone() Value {
	return e
}

func (e *Enum) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

type Bytes struct {
	ReferenceSemanticsImpl
	Value []byte
}

func (b *Bytes) Boolean() Bool {
	return True
}

func (b *Bytes) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (b *Bytes) Binop(op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *Bytes:
		switch op {
		case uint64(token.ADD):
			rLen := len(r.Value)
			if rLen == 0 {
				return b, nil
			}
			lLen := len(b.Value)
			if rLen+lLen >= verror.MaxMemSize {
				return Nil, verror.ErrMaxMemSize
			}
			values := make([]byte, lLen+rLen)
			copy(values[:lLen], b.Value)
			copy(values[lLen:], r.Value)
			return &Bytes{Value: values}, nil
		}
	}
	switch op {
	case uint64(token.OR):
		return b, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(b, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (b *Bytes) Get(index Value) (Value, error) {
	switch r := index.(type) {
	case Integer:
		l := Integer(len(b.Value))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			return Integer(b.Value[r]), nil
		}
	}
	return Nil, verror.ErrValueNotIndexable
}

func (b *Bytes) Set(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (b *Bytes) Equals(other Value) Bool {
	if val, ok := other.(*Bytes); ok {
		return b == val
	}
	if val, ok := other.(*String); ok {
		return string(b.Value) == val.Value
	}
	return false
}

func (b *Bytes) IsIterable() Bool {
	return true
}

func (b *Bytes) IsCallable() Bool {
	return false
}

func (b *Bytes) Iterator() Value {
	return &BytesIterator{Bytes: b.Value, Init: -1, End: len(b.Value)}
}

func (b Bytes) String() string {
	return fmt.Sprintf("bytes[% x]", b.Value)
}

func (b *Bytes) ObjectKey() string {
	return fmt.Sprintf("Bytes(%p)", b)
}

func (b *Bytes) Type() string {
	return "bytes"
}

func (b *Bytes) Clone() Value {
	return &Bytes{Value: b.Value}
}

type ValueSemanticsImpl struct{}

func (i ValueSemanticsImpl) Boolean() Bool {
	return false
}

func (i ValueSemanticsImpl) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrPrefixOpNotDefined
}

func (i ValueSemanticsImpl) Binop(uint64, Value) (Value, error) {
	return Nil, verror.ErrBinaryOpNotDefined
}

func (i ValueSemanticsImpl) Get(Value) (Value, error) {
	return Nil, verror.ErrValueNotIndexable
}

func (i ValueSemanticsImpl) Set(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (i ValueSemanticsImpl) Equals(Value) Bool {
	return false
}

func (i ValueSemanticsImpl) IsIterable() Bool {
	return false
}

func (i ValueSemanticsImpl) Iterator() Value {
	return Nil
}

func (i ValueSemanticsImpl) IsCallable() Bool {
	return false
}

func (i ValueSemanticsImpl) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (i ValueSemanticsImpl) String() string {
	return EmptyString
}

func (i ValueSemanticsImpl) Type() string {
	return EmptyString
}

func (i ValueSemanticsImpl) Clone() Value {
	return Nil
}

func (i ValueSemanticsImpl) ObjectKey() string {
	return EmptyString
}

func (i ValueSemanticsImpl) GetVTable() Value {
	return Nil
}

func (i ValueSemanticsImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

type ReferenceSemanticsImpl struct{}

func (i *ReferenceSemanticsImpl) Boolean() Bool {
	return false
}

func (i *ReferenceSemanticsImpl) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrPrefixOpNotDefined
}

func (i *ReferenceSemanticsImpl) Binop(uint64, Value) (Value, error) {
	return Nil, verror.ErrBinaryOpNotDefined
}

func (i *ReferenceSemanticsImpl) Get(Value) (Value, error) {
	return Nil, verror.ErrValueNotIndexable
}

func (i *ReferenceSemanticsImpl) Set(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (i *ReferenceSemanticsImpl) Equals(Value) Bool {
	return false
}

func (i *ReferenceSemanticsImpl) IsIterable() Bool {
	return false
}

func (i *ReferenceSemanticsImpl) Iterator() Value {
	return Nil
}

func (i *ReferenceSemanticsImpl) IsCallable() Bool {
	return false
}

func (i *ReferenceSemanticsImpl) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, verror.ErrNotImplemented
}

func (i ReferenceSemanticsImpl) String() string {
	return EmptyString
}

func (i *ReferenceSemanticsImpl) Type() string {
	return EmptyString
}

func (i *ReferenceSemanticsImpl) Clone() Value {
	return Nil
}

func (i *ReferenceSemanticsImpl) ObjectKey() string {
	return EmptyString
}

func (i *ReferenceSemanticsImpl) GetVTable() Value {
	return Nil
}

func (i *ReferenceSemanticsImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}
