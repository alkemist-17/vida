package vida

import (
	"fmt"
	"strings"
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

type VidaIterator struct {
	ReferenceSemanticsImpl
	obj     *Object
	ctx     *Context
	nextFn  Value
	keyFn   Value
	valueFn Value
}

// NewVidaIterator creates a VidaIterator from a Vida object that exposes
// callable next(), key(), and value() methods. All three are required;
// a missing method produces a descriptive error.
func NewVidaIterator(ctx *Context, obj *Object) (*VidaIterator, error) {
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
	return &VidaIterator{
		obj:     obj,
		ctx:     ctx,
		nextFn:  nextFn,
		keyFn:   keyFn,
		valueFn: valueFn,
	}, nil
}

func (vi *VidaIterator) Next() bool {
	result, err := callVidaMethod(vi.ctx, vi.obj, vi.nextFn)
	if err != nil {
		return false
	}
	return bool(result.Boolean())
}

func (vi *VidaIterator) Key(ctx *Context) Value {
	result, err := callVidaMethod(ctx, vi.obj, vi.keyFn)
	if err != nil {
		return Nil
	}
	return result
}

func (vi *VidaIterator) Value(ctx *Context) Value {
	result, err := callVidaMethod(ctx, vi.obj, vi.valueFn)
	if err != nil {
		return Nil
	}
	return result
}

func (vi *VidaIterator) Boolean() Bool {
	return True
}

func (vi *VidaIterator) IsIterable() Bool {
	return True
}

func (vi *VidaIterator) Iterator() Value {
	return vi
}

func (vi *VidaIterator) String() string {
	return fmt.Sprintf("VidaIterator[%p]", vi)
}

func (vi *VidaIterator) Type() string {
	return "VidaIterator"
}

func (vi *VidaIterator) Clone() Value {
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

// resolveVidaIterator checks whether obj defines a custom iterator protocol
// and returns a ready-to-use VidaIterator if so.
//
// Priority:
//  1. iter() method — invoked to produce a fresh iterator object, which
//     must itself expose next/key/value.
//  2. next() + key() + value() methods — the object is its own iterator.
//  3. Neither — returns (nil, nil); the caller should fall back to the
//     default Object.Iterator() (property iteration).
//
// A partial protocol (e.g. next() defined but key() missing) is reported
// as an error rather than silently ignored, so typos surface early.
func resolveVidaIterator(ctx *Context, obj *Object) (*VidaIterator, error) {
	// Priority 1: iter() method
	iterFn := obj.Get(ctx, &String{Value: "iter"})
	if iterFn.IsCallable() {
		result, err := callVidaMethod(ctx, obj, iterFn)
		if err != nil {
			return nil, fmt.Errorf("iter() call failed: %w", err)
		}
		iterObj, ok := result.(*Object)
		if !ok {
			return nil, fmt.Errorf("iter() must return an object, got %s", result.Type())
		}
		return NewVidaIterator(ctx, iterObj)
	}

	// Priority 2: self-iterating (next + key + value)
	nextFn := obj.Get(ctx, &String{Value: "next"})
	keyFn := obj.Get(ctx, &String{Value: "key"})
	valueFn := obj.Get(ctx, &String{Value: "value"})
	hasNext := nextFn.IsCallable()
	hasKey := keyFn.IsCallable()
	hasValue := valueFn.IsCallable()

	if hasNext && hasKey && hasValue {
		return &VidaIterator{
			obj:     obj,
			ctx:     ctx,
			nextFn:  nextFn,
			keyFn:   keyFn,
			valueFn: valueFn,
		}, nil
	}

	// Partial protocol: next() is the strongest signal of iterator intent.
	// If it is present but key() or value() is missing, report an error
	// instead of silently falling back to property iteration.
	if hasNext {
		var missing []string
		if !hasKey {
			missing = append(missing, "key()")
		}
		if !hasValue {
			missing = append(missing, "value()")
		}
		return nil, fmt.Errorf("incomplete iterator protocol: has next() but missing %s",
			strings.Join(missing, ", "))
	}

	// No custom iterator defined
	return nil, nil
}
