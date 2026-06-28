package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/verror"
)

type Iterator interface {
	Next() bool
	Key(*Context) Value
	Value(*Context) Value
}

type ArrayIterator struct {
	ReferenceSemanticsImpl
	Array []Value
	Init  int
	End   int
}

func (it *ArrayIterator) Next() bool {
	it.Init++
	return it.Init < it.End
}

func (it *ArrayIterator) Key(ctx *Context) Value {
	return Integer(it.Init)
}

func (it *ArrayIterator) Value(ctx *Context) Value {
	return it.Array[it.Init]
}

func (it *ArrayIterator) Boolean() Bool {
	return true
}

func (it *ArrayIterator) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrOpNotDefinedForIterators
}

func (it *ArrayIterator) Get(*Context, Value) Value {
	return Nil
}

func (it *ArrayIterator) Set(Value, Value) error {
	return verror.ErrOpNotDefinedForIterators
}

func (it *ArrayIterator) Equals(Value) Bool {
	return false
}

func (it *ArrayIterator) IsIterable() Bool {
	return false
}

func (it *ArrayIterator) IsCallable() Bool {
	return false
}

func (it *ArrayIterator) Iterator() Value {
	return Nil
}

func (it ArrayIterator) String(ctx *Context) string {
	return fmt.Sprintf("ArrayIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *ArrayIterator) Clone() Value {
	return it
}

func (it *ArrayIterator) Type(ctx *Context) string {
	return "ArrayIterator"
}

type ObjectIterator struct {
	ReferenceSemanticsImpl
	Keys []string
	Obj  map[string]Value
	Init int
	End  int
}

func newObjectIterator(o *Object) *ObjectIterator {
	var keys []string
	for k := range o.Value {
		keys = append(keys, k)
	}
	it := &ObjectIterator{
		Obj:  o.Value,
		Init: -1,
		End:  len(keys),
		Keys: keys,
	}
	return it
}

func (it *ObjectIterator) Next() bool {
	it.Init++
	return it.Init < it.End
}

func (it *ObjectIterator) Key(ctx *Context) Value {
	return &String{Value: it.Keys[it.Init]}
}

func (it *ObjectIterator) Value(ctx *Context) Value {
	return it.Obj[it.Keys[it.Init]]
}

func (it *ObjectIterator) Boolean() Bool {
	return true
}

func (it *ObjectIterator) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrOpNotDefinedForIterators
}

func (it *ObjectIterator) Get(*Context, Value) Value {
	return Nil
}

func (it *ObjectIterator) Set(Value, Value) error {
	return verror.ErrOpNotDefinedForIterators
}

func (it *ObjectIterator) Equals(Value) Bool {
	return false
}

func (it *ObjectIterator) IsIterable() Bool {
	return false
}

func (it *ObjectIterator) IsCallable() Bool {
	return false
}

func (it *ObjectIterator) Iterator() Value {
	return Nil
}

func (it ObjectIterator) String(ctx *Context) string {
	return fmt.Sprintf("ObjectIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *ObjectIterator) Clone() Value {
	return it
}

func (it *ObjectIterator) Type(ctx *Context) string {
	return "ObjectIterator"
}

type StringIterator struct {
	ReferenceSemanticsImpl
	Runes []rune
	Init  int
	End   int
}

func (it *StringIterator) Next() bool {
	it.Init++
	return it.Init < it.End
}

func (it *StringIterator) Key(ctx *Context) Value {
	return Integer(it.Init)
}

func (it *StringIterator) Value(ctx *Context) Value {
	return &String{Value: string(it.Runes[it.Init]), Runes: it.Runes[it.Init : it.Init+1]}
}

func (it *StringIterator) Boolean() Bool {
	return true
}

func (it *StringIterator) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrOpNotDefinedForIterators
}

func (it *StringIterator) Get(*Context, Value) Value {
	return Nil
}

func (it *StringIterator) Set(Value, Value) error {
	return verror.ErrOpNotDefinedForIterators
}

func (it *StringIterator) Equals(Value) Bool {
	return false
}

func (it *StringIterator) IsIterable() Bool {
	return false
}

func (it *StringIterator) IsCallable() Bool {
	return false
}

func (it *StringIterator) Iterator() Value {
	return Nil
}

func (it StringIterator) String(ctx *Context) string {
	return fmt.Sprintf("StringIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *StringIterator) Clone() Value {
	return it
}

func (it *StringIterator) Type(ctx *Context) string {
	return "StringIterator"
}

type IntegerIterator struct {
	ReferenceSemanticsImpl
	Init Integer
	End  Integer
}

func (it *IntegerIterator) Next() bool {
	it.Init++
	return it.Init < it.End
}

func (it *IntegerIterator) Key(ctx *Context) Value {
	return it.Init
}

func (it *IntegerIterator) Value(ctx *Context) Value {
	return it.Init
}

func (it *IntegerIterator) Boolean() Bool {
	return true
}

func (it *IntegerIterator) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrOpNotDefinedForIterators
}

func (it *IntegerIterator) Get(*Context, Value) Value {
	return Nil
}

func (it *IntegerIterator) Set(Value, Value) error {
	return verror.ErrOpNotDefinedForIterators
}

func (it *IntegerIterator) Equals(Value) Bool {
	return false
}

func (it *IntegerIterator) IsIterable() Bool {
	return false
}

func (it *IntegerIterator) IsCallable() Bool {
	return false
}

func (it *IntegerIterator) Iterator() Value {
	return Nil
}

func (it IntegerIterator) String(ctx *Context) string {
	return fmt.Sprintf("IntIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *IntegerIterator) Clone() Value {
	return it
}

func (it *IntegerIterator) Type(ctx *Context) string {
	return "IntIterator"
}

type BytesIterator struct {
	ReferenceSemanticsImpl
	Bytes []byte
	Init  int
	End   int
}

func (bi *BytesIterator) Next() bool {
	bi.Init++
	return bi.Init < bi.End
}

func (bi *BytesIterator) Key(ctx *Context) Value {
	return Integer(bi.Init)
}

func (bi *BytesIterator) Value(ctx *Context) Value {
	return Integer(bi.Bytes[bi.Init])
}

func (bi *BytesIterator) Boolean() Bool {
	return true
}

func (bi *BytesIterator) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrOpNotDefinedForIterators
}

func (bi *BytesIterator) Get(*Context, Value) Value {
	return Nil
}

func (bi *BytesIterator) Set(Value, Value) error {
	return verror.ErrOpNotDefinedForIterators
}

func (bi *BytesIterator) Equals(Value) Bool {
	return false
}

func (bi *BytesIterator) IsIterable() Bool {
	return false
}

func (bi *BytesIterator) IsCallable() Bool {
	return false
}

func (bi *BytesIterator) Iterator() Value {
	return Nil
}

func (bi BytesIterator) String(ctx *Context) string {
	return fmt.Sprintf("BytesIterator[i = %v, e = %v]", bi.Init, bi.End)
}

func (bi *BytesIterator) Clone() Value {
	return bi
}

func (bi *BytesIterator) Type(ctx *Context) string {
	return "BytesIterator"
}
