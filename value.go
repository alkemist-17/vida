package vida

import (
	"encoding/json"

	"github.com/alkemist-17/vida/verror"
)

type Value interface {
	Boolean() Bool
	Prefix(uint64) (Value, error)
	Binop(ctx *Context, op uint64, other Value) (Value, error)
	Get(ctx *Context, message Value) Value
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
	LookUp(ctx *Context, message Value) Value
	GetVTable(ctx *Context) Value
}

// Value semantics default impl
type ValueSemanticsImpl struct{}

func (i ValueSemanticsImpl) Boolean() Bool {
	return True
}

func (i ValueSemanticsImpl) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrPrefixOpNotDefined
}

func (i ValueSemanticsImpl) Binop(*Context, uint64, Value) (Value, error) {
	return Nil, verror.ErrBinaryOpNotDefined
}

func (i ValueSemanticsImpl) Get(*Context, Value) Value {
	return Nil
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

func (i ValueSemanticsImpl) GetVTable(ctx *Context) Value {
	return Nil
}

func (i ValueSemanticsImpl) LookUp(*Context, Value) Value {
	return Nil
}

func (i ValueSemanticsImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

// Reference semantics default impl
type ReferenceSemanticsImpl struct{}

func (i *ReferenceSemanticsImpl) Boolean() Bool {
	return True
}

func (i *ReferenceSemanticsImpl) Prefix(uint64) (Value, error) {
	return Nil, verror.ErrPrefixOpNotDefined
}

func (i *ReferenceSemanticsImpl) Binop(*Context, uint64, Value) (Value, error) {
	return Nil, verror.ErrBinaryOpNotDefined
}

func (i *ReferenceSemanticsImpl) Get(*Context, Value) Value {
	return Nil
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

func (i *ReferenceSemanticsImpl) GetVTable(ctx *Context) Value {
	return Nil
}

func (i *ReferenceSemanticsImpl) LookUp(*Context, Value) Value {
	return Nil
}

func (i *ReferenceSemanticsImpl) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}
