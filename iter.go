package vida

import (
	"fmt"
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

func (it ArrayIterator) String() string {
	return fmt.Sprintf("ArrayIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *ArrayIterator) Clone() Value {
	return it
}

func (it *ArrayIterator) Type() string {
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

func (it ObjectIterator) String() string {
	return fmt.Sprintf("ObjectIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *ObjectIterator) Clone() Value {
	return it
}

func (it *ObjectIterator) Type() string {
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

func (it StringIterator) String() string {
	return fmt.Sprintf("StringIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *StringIterator) Clone() Value {
	return it
}

func (it *StringIterator) Type() string {
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

func (it IntegerIterator) String() string {
	return fmt.Sprintf("IntIterator[i = %v, e = %v]", it.Init, it.End)
}

func (it *IntegerIterator) Clone() Value {
	return it
}

func (it *IntegerIterator) Type() string {
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

func (bi BytesIterator) String() string {
	return fmt.Sprintf("BytesIterator[i = %v, e = %v]", bi.Init, bi.End)
}

func (bi *BytesIterator) Clone() Value {
	return bi
}

func (bi *BytesIterator) Type() string {
	return "BytesIterator"
}
