package vida

import (
	"fmt"
	"maps"
	"math"
	"strconv"
	"strings"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Value interface {
	Boolean() Bool
	Prefix(uint64) (Value, error)
	Binop(uint64, Value) (Value, error)
	IGet(Value) (Value, error)
	ISet(Value, Value) error
	Equals(Value) Bool
	IsIterable() Bool
	Iterator() Value
	IsCallable() Bool
	Call(args ...Value) (Value, error)
	String() string
	Type() string
	Clone() Value
	ObjectKey() string
}

type Nil struct {
	ValueSemanticsImpl
}

func (n Nil) Boolean() Bool {
	return Bool(false)
}

func (n Nil) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(true), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (n Nil) Binop(op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return NilValue, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(n, rhs)
	default:
		return NilValue, verror.ErrBinaryOpNotDefined
	}
}

func (n Nil) IGet(index Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (n Nil) ISet(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (n Nil) Equals(other Value) Bool {
	_, ok := other.(Nil)
	return Bool(ok)
}

func (n Nil) IsIterable() Bool {
	return false
}

func (n Nil) IsCallable() Bool {
	return false
}

func (n Nil) Iterator() Value {
	return NilValue
}

func (n Nil) String() string {
	return "nil"
}

func (n Nil) ObjectKey() string {
	return "nil"
}

func (n Nil) Type() string {
	return "nil"
}

func (n Nil) Clone() Value {
	return n
}

type Bool bool

func (b Bool) Boolean() Bool {
	return b
}

func (b Bool) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return !b, nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
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
		return NilValue, verror.ErrBinaryOpNotDefined
	}
}

func (b Bool) IGet(index Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (b Bool) ISet(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (b Bool) Equals(other Value) Bool {
	if val, ok := other.(Bool); ok {
		return b == val
	}
	return false
}

func (b Bool) IsIterable() Bool {
	return false
}

func (b Bool) IsCallable() Bool {
	return false
}

func (b Bool) Call(args ...Value) (Value, error) {
	return NilValue, verror.ErrNotImplemented
}

func (b Bool) Iterator() Value {
	return NilValue
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

func (b Bool) Type() string {
	return "bool"
}

func (b Bool) Clone() Value {
	return b
}

type String struct {
	ReferenceSemanticsImpl
	Runes []rune
	Value string
}

func (s *String) Boolean() Bool {
	return Bool(true)
}

func (s *String) Binop(op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *String:
		switch op {
		case uint64(token.ADD):
			if len(s.Value)+len(r.Value) >= verror.MaxMemSize {
				return NilValue, verror.ErrMaxMemSize
			}
			str := &String{Value: s.Value + r.Value}
			return str, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return s, nil
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
	default:
		switch op {
		case uint64(token.OR):
			return s, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.IN):
			return IsMemberOf(s, rhs)
		}
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (s *String) IGet(index Value) (Value, error) {
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
	return NilValue, verror.ErrValueNotIndexable
}

func (s *String) ISet(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (s *String) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (s *String) Equals(other Value) Bool {
	if val, ok := other.(*String); ok {
		return s.Value == val.Value
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

func (s *String) Type() string {
	return "string"
}

func (s *String) Clone() Value {
	return s
}

type Integer int64

func (i Integer) Boolean() Bool {
	return Bool(true)
}

func (i Integer) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.SUB):
		return -i, nil
	case uint64(token.NOT):
		return Bool(false), nil
	case uint64(token.ADD):
		return i, nil
	case uint64(token.TILDE):
		return Integer(^uint32(i)), nil
	}
	return NilValue, verror.ErrPrefixOpNotDefined
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
				return NilValue, verror.ErrDivisionByZero
			}
			return l / r, nil
		case uint64(token.REM):
			if r == 0 {
				return NilValue, verror.ErrDivisionByZero
			}
			return l % r, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return l, nil
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
		case uint64(token.IN):
			return IsMemberOf(l, rhs)
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
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return l, nil
		case uint64(token.LT):
			return Bool(Float(l) < r), nil
		case uint64(token.LE):
			return Bool(Float(l) <= r), nil
		case uint64(token.GT):
			return Bool(Float(l) > r), nil
		case uint64(token.GE):
			return Bool(Float(l) >= r), nil
		case uint64(token.IN):
			return IsMemberOf(l, rhs)
		}
	default:
		switch op {
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return l, nil
		case uint64(token.IN):
			return IsMemberOf(l, rhs)
		}
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (i Integer) IGet(index Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (i Integer) ISet(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (i Integer) Equals(other Value) Bool {
	if val, ok := other.(Integer); ok {
		return i == val
	}
	return false
}

func (i Integer) IsIterable() Bool {
	return true
}

func (i Integer) IsCallable() Bool {
	return false
}

func (i Integer) Call(args ...Value) (Value, error) {
	return NilValue, verror.ErrNotImplemented
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

func (i Integer) Type() string {
	return "int"
}

func (i Integer) Clone() Value {
	return i
}

type Float float64

func (f Float) Boolean() Bool {
	return Bool(true)
}

func (f Float) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.SUB):
		return -f, nil
	case uint64(token.NOT):
		return Bool(false), nil
	case uint64(token.ADD):
		return f, nil
	}
	return NilValue, verror.ErrPrefixOpNotDefined
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
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return f, nil
		case uint64(token.LT):
			return Bool(f < r), nil
		case uint64(token.LE):
			return Bool(f <= r), nil
		case uint64(token.GT):
			return Bool(f > r), nil
		case uint64(token.GE):
			return Bool(f >= r), nil
		case uint64(token.IN):
			return IsMemberOf(f, rhs)
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
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return f, nil
		case uint64(token.LT):
			return Bool(f < Float(r)), nil
		case uint64(token.LE):
			return Bool(f <= Float(r)), nil
		case uint64(token.GT):
			return Bool(f > Float(r)), nil
		case uint64(token.GE):
			return Bool(f >= Float(r)), nil
		case uint64(token.IN):
			return IsMemberOf(f, rhs)
		}
	default:
		switch op {
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return f, nil
		case uint64(token.IN):
			return IsMemberOf(f, rhs)
		}
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (f Float) IGet(index Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (f Float) ISet(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (f Float) Equals(other Value) Bool {
	if val, ok := other.(Float); ok {
		return f == val
	}
	return false
}

func (f Float) IsIterable() Bool {
	return false
}

func (f Float) IsCallable() Bool {
	return false
}

func (f Float) Call(args ...Value) (Value, error) {
	return NilValue, verror.ErrNotImplemented
}

func (f Float) Iterator() Value {
	return NilValue
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'g', -1, 64)
}

func (f Float) ObjectKey() string {
	return fmt.Sprintf("%vf", strconv.FormatFloat(float64(f), 'g', -1, 64))
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
	return Bool(true)
}

func (xs *Array) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
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
				return NilValue, verror.ErrMaxMemSize
			}
			values := make([]Value, lLen+rLen)
			copy(values[:lLen], xs.Value)
			copy(values[lLen:], r.Value)
			return &Array{Value: values}, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return xs, nil
		case uint64(token.IN):
			return IsMemberOf(xs, rhs)
		}
	default:
		switch op {
		case uint64(token.OR):
			return xs, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.IN):
			return IsMemberOf(xs, rhs)
		}
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (xs *Array) IGet(index Value) (Value, error) {
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
	return NilValue, verror.ErrValueNotIndexable
}

func (xs *Array) ISet(index, val Value) error {
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
	if val, ok := other.(*Array); ok {
		return xs == val
	}
	return false
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

func (xs Array) String() string {
	if len(xs.Value) == 0 {
		return "[]"
	}
	var r []string
	for _, v := range xs.Value {
		r = append(r, v.String())
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

type Object struct {
	ReferenceSemanticsImpl
	Value map[string]Value
}

func (o *Object) Boolean() Bool {
	return true
}

func (o *Object) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (o *Object) Binop(op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *Object:
		switch op {
		case uint64(token.ADD):
			pairs := make(map[string]Value)
			maps.Copy(pairs, o.Value)
			maps.Copy(pairs, r.Value)
			return &Object{Value: pairs}, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return o, nil
		case uint64(token.IN):
			return IsMemberOf(o, rhs)
		}
	default:
		switch op {
		case uint64(token.OR):
			return o, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.IN):
			return IsMemberOf(o, rhs)
		}
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (o *Object) IGet(index Value) (Value, error) {
	if meta, ok := o.Value[__meta].(*Object); ok {
		if get, ok := meta.Value[__get]; ok {
			switch val := get.(type) {
			case *Function:
				return o.execute(val, index)
			case *Object:
				return val.IGet(index)
			default:
				return val, nil
			}
		}
	}
	if val, ok := o.Value[index.ObjectKey()]; ok {
		return val, nil
	} else if proto, ok := o.Value[__proto].(*Object); ok {
		return proto.IGet(index)
	}
	return NilValue, nil
}

func (o *Object) ISet(index, val Value) error {
	if meta, ok := o.Value[__meta].(*Object); ok {
		if set, ok := meta.Value[__set]; ok {
			switch setvalue := set.(type) {
			case *Function:
				_, err := o.execute(setvalue, index, val)
				return err
			case *Object:
				return setvalue.ISet(index, val)
			default:
				return nil
			}
		}
	}
	o.Value[index.ObjectKey()] = val
	return nil
}

func (o *Object) Equals(other Value) Bool {
	if val, ok := other.(*Object); ok {
		return o == val
	}
	return false
}

func (o *Object) IsIterable() Bool {
	return true
}

func (o *Object) IsCallable() Bool {
	if meta, ok := o.Value[__meta].(*Object); ok {
		if _, ok := meta.Value[__call]; ok {
			return true
		}
	}
	return false
}

func (o *Object) Call(args ...Value) (Value, error) {
	switch fn := o.Value[__meta].(*Object).Value[__call].(type) {
	case *Function:
		if fn.CoreFn.IsVar {
			a := make([]Value, len(args))
			copy(a, args)
			return o.execute(fn, &Array{Value: a})
		}
		return o.execute(fn, args...)
	default:
		return fn, nil
	}
}

func (o *Object) Iterator() Value {
	return newObjectIterator(o)
}

func (o *Object) execute(fn *Function, args ...Value) (Value, error) {
	vm := ((*clbu)[globalStateIndex].(*GlobalState)).Pool.getVM()
	vm.Thread.Script.MainFunction = fn
	_, err := vm.runMetaFunction(fn, o, args...)
	((*clbu)[globalStateIndex].(*GlobalState)).Pool.Key--
	if err != nil {
		return NilValue, err
	}
	return vm.Channel, nil
}

func (o *Object) String() string {
	if meta, ok := o.Value[__meta].(*Object); ok {
		if str, ok := meta.Value[__str]; ok {
			switch fn := str.(type) {
			case *Function:
				val, _ := o.execute(fn)
				return val.String()
			default:
				return fn.String()
			}
		}
	}
	if len(o.Value) == 0 {
		return "{}"
	}
	var r []string
	for k, v := range o.Value {
		if k != __proto && k != __meta {
			r = append(r, fmt.Sprintf("%v: %v", k, v))
		}
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
	m := make(map[string]Value)
	for k, v := range o.Value {
		m[k] = v.Clone()
	}
	return &Object{Value: m}
}

type freeInfo struct {
	Index   int
	IsLocal Bool
	Id      string
}

type CoreFunction struct {
	ReferenceSemanticsImpl
	Code       []uint64
	Info       []freeInfo
	Free       int
	Arity      int
	IsVar      bool
	ScriptName string
}

func (c *CoreFunction) Boolean() Bool {
	return true
}

func (c *CoreFunction) Prefix(uint64) (Value, error) {
	return NilValue, verror.ErrPrefixOpNotDefined
}

func (c *CoreFunction) Binop(uint64, Value) (Value, error) {
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (c *CoreFunction) IGet(Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (c *CoreFunction) ISet(Value, Value) error {
	return verror.ErrValueNotIndexable
}

func (c *CoreFunction) Equals(other Value) Bool {
	if f, ok := other.(*CoreFunction); ok {
		return c == f
	}
	return false
}

func (c *CoreFunction) IsIterable() Bool {
	return false
}

func (c *CoreFunction) IsCallable() Bool {
	return false
}

func (c *CoreFunction) Iterator() Value {
	return NilValue
}

func (c *CoreFunction) Type() string {
	return "corefunction"
}

func (f CoreFunction) String() string {
	return fmt.Sprintf("CoreFunction(arity = %v, isVar = %v, free = %v)", f.Arity, f.IsVar, f.Free)
}

func (f *CoreFunction) Clone() Value {
	return f
}

type Function struct {
	ReferenceSemanticsImpl
	Free   []Value
	CoreFn *CoreFunction
}

func (f *Function) Boolean() Bool {
	return true
}

func (f *Function) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
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
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (f *Function) IGet(Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (f *Function) ISet(Value, Value) error {
	return verror.ErrValueNotIndexable
}

func (f *Function) Equals(other Value) Bool {
	if o, ok := other.(*Function); ok {
		return f == o
	}
	return false
}

func (f *Function) IsIterable() Bool {
	return false
}

func (f *Function) IsCallable() Bool {
	return true
}

func (f *Function) Iterator() Value {
	return NilValue
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

type GFn func(args ...Value) (Value, error)

func (gfn GFn) Boolean() Bool {
	return Bool(true)
}

func (gfn GFn) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (gfn GFn) Binop(op uint64, r Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return gfn, nil
	case uint64(token.AND):
		return r, nil
	case uint64(token.IN):
		return IsMemberOf(gfn, r)
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (gfn GFn) IGet(index Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (gfn GFn) ISet(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (gfn GFn) Equals(other Value) Bool {
	return false
}

func (gfn GFn) IsIterable() Bool {
	return false
}

func (gfn GFn) IsCallable() Bool {
	return true
}

func (gfn GFn) Call(args ...Value) (Value, error) {
	return gfn(args...)
}

func (gfn GFn) Iterator() Value {
	return NilValue
}

func (gfn GFn) String() string {
	return "GFn"
}

func (gfn GFn) ObjectKey() string {
	return "GFn"
}

func (gFn GFn) Clone() Value {
	return gFn
}

func (gfn GFn) Type() string {
	return "gfn"
}

type Error struct {
	ValueSemanticsImpl
	Message Value
}

func (e Error) Boolean() Bool {
	return false
}

func (e Error) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(true), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (e Error) Binop(op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return e, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(e, rhs)
	default:
		return NilValue, verror.ErrBinaryOpNotDefined
	}
}

func (e Error) IGet(index Value) (Value, error) {
	if val, ok := index.(*String); ok && val.Value == "message" {
		return e.Message, nil
	}
	return NilValue, nil
}

func (e Error) ISet(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (e Error) Equals(other Value) Bool {
	v, ok := other.(Error)
	return Bool(ok) && e.Message.Equals(v.Message)
}

func (e Error) IsIterable() Bool {
	return false
}

func (e Error) IsCallable() Bool {
	return false
}

func (e Error) Iterator() Value {
	return NilValue
}

func (e Error) String() string {
	return fmt.Sprintf("Error(message: %v)", e.Message.String())
}

func (e Error) ObjectKey() string {
	return fmt.Sprintf("Error(message: %v)", e.Message.ObjectKey())
}

func (e Error) Type() string {
	return "error"
}

func (e Error) Clone() Value {
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
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
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
		return NilValue, verror.ErrBinaryOpNotDefined
	}
}

func (e *Enum) IGet(index Value) (Value, error) {
	if val, ok := e.Pairs[index.String()]; ok {
		return val, nil
	}
	return NilValue, nil
}

func (e *Enum) ISet(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (e *Enum) Equals(other Value) Bool {
	if val, ok := other.(*Enum); ok {
		return Bool(val == other)
	}
	return false
}

func (e *Enum) IsIterable() Bool {
	return false
}

func (e *Enum) Iterator() Value {
	return NilValue
}

func (e *Enum) IsCallable() Bool {
	return false
}

func (e *Enum) Call(args ...Value) (Value, error) {
	return NilValue, verror.ErrNotImplemented
}

func (e Enum) String() string {
	if len(e.Pairs) == 0 {
		return "enum{ }"
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

func (e *Enum) Type() string {
	return "enum"
}

func (e *Enum) Clone() Value {
	return e
}

type Bytes struct {
	ReferenceSemanticsImpl
	Value []byte
}

func (b *Bytes) Boolean() Bool {
	return Bool(true)
}

func (b *Bytes) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
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
				return NilValue, verror.ErrMaxMemSize
			}
			values := make([]byte, lLen+rLen)
			copy(values[:lLen], b.Value)
			copy(values[lLen:], r.Value)
			return &Bytes{Value: values}, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.OR):
			return b, nil
		case uint64(token.IN):
			return IsMemberOf(b, r)
		}
	default:
		switch op {
		case uint64(token.OR):
			return b, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.IN):
			return IsMemberOf(b, r)
		}
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (b *Bytes) IGet(index Value) (Value, error) {
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
	return NilValue, verror.ErrValueNotIndexable
}

func (b *Bytes) ISet(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (b *Bytes) Equals(other Value) Bool {
	if val, ok := other.(*Bytes); ok {
		return b == val
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
	return NilValue, verror.ErrPrefixOpNotDefined
}

func (i ValueSemanticsImpl) Binop(uint64, Value) (Value, error) {
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (i ValueSemanticsImpl) IGet(Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (i ValueSemanticsImpl) ISet(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (i ValueSemanticsImpl) Equals(Value) Bool {
	return false
}

func (i ValueSemanticsImpl) IsIterable() Bool {
	return false
}

func (i ValueSemanticsImpl) Iterator() Value {
	return NilValue
}

func (i ValueSemanticsImpl) IsCallable() Bool {
	return false
}

func (i ValueSemanticsImpl) Call(args ...Value) (Value, error) {
	return NilValue, verror.ErrNotImplemented
}

func (i ValueSemanticsImpl) String() string {
	return ""
}

func (i ValueSemanticsImpl) Type() string {
	return ""
}

func (i ValueSemanticsImpl) Clone() Value {
	return NilValue
}

func (i ValueSemanticsImpl) ObjectKey() string {
	return ""
}

type ReferenceSemanticsImpl struct{}

func (i *ReferenceSemanticsImpl) Boolean() Bool {
	return false
}

func (i *ReferenceSemanticsImpl) Prefix(uint64) (Value, error) {
	return NilValue, verror.ErrPrefixOpNotDefined
}

func (i *ReferenceSemanticsImpl) Binop(uint64, Value) (Value, error) {
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (i *ReferenceSemanticsImpl) IGet(Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (i *ReferenceSemanticsImpl) ISet(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (i *ReferenceSemanticsImpl) Equals(Value) Bool {
	return false
}

func (i *ReferenceSemanticsImpl) IsIterable() Bool {
	return false
}

func (i *ReferenceSemanticsImpl) Iterator() Value {
	return NilValue
}

func (i *ReferenceSemanticsImpl) IsCallable() Bool {
	return false
}

func (i *ReferenceSemanticsImpl) Call(args ...Value) (Value, error) {
	return NilValue, verror.ErrNotImplemented
}

func (i ReferenceSemanticsImpl) String() string {
	return ""
}

func (i *ReferenceSemanticsImpl) Type() string {
	return ""
}

func (i *ReferenceSemanticsImpl) Clone() Value {
	return NilValue
}

func (i *ReferenceSemanticsImpl) ObjectKey() string {
	return ""
}
