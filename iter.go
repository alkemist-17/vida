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

func (it *ArrayIterator) IsIterable() Bool {
	return True
}

func (it *ArrayIterator) Iterator(ctx *Context) Value {
	return it
}

func (it *ArrayIterator) String() string {
	return fmt.Sprintf("Iterator[%p]", it)
}

func (it *ArrayIterator) ObjectKey() string {
	return it.String()
}

func (it *ArrayIterator) Clone() Value {
	return it
}

func (it *ArrayIterator) Type() string {
	return "iterator"
}

type ObjectIterator struct {
	ReferenceSemanticsImpl
	Keys []string
	Obj  map[string]Value
	Init int
	End  int
}

func newObjectIterator(ctx *Context, o *Object) Value {
	if ui, ok := hasUserIterator(ctx, o); ok {
		return ui
	}
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

func (it *ObjectIterator) IsIterable() Bool {
	return True
}

func (it *ObjectIterator) Iterator(ctx *Context) Value {
	return it
}

func (it *ObjectIterator) String() string {
	return fmt.Sprintf("Iterator[%p]", it)
}

func (it *ObjectIterator) ObjectKey() string {
	return it.String()
}

func (it *ObjectIterator) Clone() Value {
	return it
}

func (it *ObjectIterator) Type() string {
	return "iterator"
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

func (it *StringIterator) IsIterable() Bool {
	return True
}

func (it *StringIterator) Iterator(ctx *Context) Value {
	return it
}

func (it *StringIterator) String() string {
	return fmt.Sprintf("Iterator[%p]", it)
}

func (it *StringIterator) ObjectKey() string {
	return it.String()
}

func (it *StringIterator) Clone() Value {
	return it
}

func (it *StringIterator) Type() string {
	return "iterator"
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

func (it *IntegerIterator) IsIterable() Bool {
	return True
}

func (it *IntegerIterator) Iterator(ctx *Context) Value {
	return it
}

func (it *IntegerIterator) String() string {
	return fmt.Sprintf("Iterator[%p]", it)
}

func (it *IntegerIterator) ObjectKey() string {
	return it.String()
}

func (it *IntegerIterator) Clone() Value {
	return it
}

func (it *IntegerIterator) Type() string {
	return "iterator"
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

func (bi *BytesIterator) IsIterable() Bool {
	return True
}

func (bi *BytesIterator) Iterator(ctx *Context) Value {
	return bi
}

func (bi *BytesIterator) String() string {
	return fmt.Sprintf("Iterator[%p]", bi)
}

func (it *BytesIterator) ObjectKey() string {
	return it.String()
}

func (bi *BytesIterator) Clone() Value {
	return bi
}

func (bi *BytesIterator) Type() string {
	return "iterator"
}

type UserIterator struct {
	ReferenceSemanticsImpl
	obj     *Object
	ctx     *Context
	nextFn  Value
	keyFn   Value
	valueFn Value
}

// Creates a UserIterator from a Vida object that exposes
// callable next(), key(), and value() methods. All three are required;
// a missing method produces a descriptive error.
func newUserIterator(ctx *Context, obj *Object) (*UserIterator, error) {
	nextFn := obj.Get(ctx, &String{Value: "next"})
	if !nextFn.IsCallable() {
		return nil, fmt.Errorf("iterator object is missing a callable next() method")
	}
	keyFn := obj.Get(ctx, &String{Value: "key"})
	if !keyFn.IsCallable() {
		return nil, fmt.Errorf("iterator object is missing a callable key() method")
	}
	valueFn := obj.Get(ctx, &String{Value: "value"})
	if !valueFn.IsCallable() {
		return nil, fmt.Errorf("iterator object is missing a callable value() method")
	}
	return &UserIterator{
		obj:     obj,
		ctx:     ctx,
		nextFn:  nextFn,
		keyFn:   keyFn,
		valueFn: valueFn,
	}, nil
}

func (vi *UserIterator) Next() bool {
	result, err := callVidaMethod(vi.ctx, vi.obj, vi.nextFn)
	if err != nil {
		return false
	}
	return bool(result.Boolean())
}

func (vi *UserIterator) Key(ctx *Context) Value {
	result, err := callVidaMethod(ctx, vi.obj, vi.keyFn)
	if err != nil {
		return Nil
	}
	return result
}

func (vi *UserIterator) Value(ctx *Context) Value {
	result, err := callVidaMethod(ctx, vi.obj, vi.valueFn)
	if err != nil {
		return Nil
	}
	return result
}

func (vi *UserIterator) Boolean() Bool {
	return True
}

func (vi *UserIterator) IsIterable() Bool {
	return True
}

func (vi *UserIterator) Iterator(ctx *Context) Value {
	return vi
}

func (vi *UserIterator) String() string {
	return fmt.Sprintf("Iterator[%p]", vi)
}

func (it *UserIterator) ObjectKey() string {
	return it.String()
}

func (vi *UserIterator) Type() string {
	return "iterator"
}

func (vi *UserIterator) Clone() Value {
	return vi
}

// callVidaMethod invokes a method on self, passing self as the first
// argument (the receiver). It mirrors the dispatch pattern used
// throughout Object (see dispatchOperatorOverride).
func callVidaMethod(ctx *Context, self *Object, method Value) (Value, error) {
	switch fn := method.(type) {
	case *Function:
		return ctx.runFunctionInNewThread(fn, self)
	case NativeFunction:
		return fn.Call(ctx, self)
	default:
		return Nil, fmt.Errorf("value of type %s is not callable", method.Type())
	}
}

func hasUserIterator(ctx *Context, obj *Object) (Value, bool) {
	nextFn := obj.Get(ctx, &String{Value: "next"})
	keyFn := obj.Get(ctx, &String{Value: "key"})
	valueFn := obj.Get(ctx, &String{Value: "value"})
	hasNext := nextFn.IsCallable()
	hasKey := keyFn.IsCallable()
	hasValue := valueFn.IsCallable()

	if hasNext && hasKey && hasValue {
		return &UserIterator{
			obj:     obj,
			ctx:     ctx,
			nextFn:  nextFn,
			keyFn:   keyFn,
			valueFn: valueFn,
		}, true
	}
	return Nil, false
}
