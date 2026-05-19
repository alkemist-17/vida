package vida

import (
	"fmt"
	"maps"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

const (
	TNil      uint8 = 0
	TBool     uint8 = 1
	TInt      uint8 = 2
	TFloat    uint8 = 3
	TString   uint8 = 4
	TArray    uint8 = 5
	TObject   uint8 = 6
	TFunction uint8 = 7
	TCoreFn   uint8 = 8
	TGFn      uint8 = 9
	TBytes    uint8 = 10
	TEnum     uint8 = 11
	TError    uint8 = 12
	TThread   uint8 = 13
	TTime     uint8 = 14
	TExtern   uint8 = 255
)

type TValue struct {
	extern Extern
	ival   int64
	ptr    unsafe.Pointer
	ttype  uint8
}

func (v TValue) TType() uint8 {
	return v.ttype
}

func (v TValue) Int() int64 {
	return v.ival
}

func (v TValue) Float() float64 {
	return math.Float64frombits(uint64(v.ival))
}

func (v TValue) Bool() bool {
	return v.ival != 0
}

func (v TValue) Str() *TypeString {
	return (*TypeString)(v.ptr)
}

func (v TValue) Arr() *TypeArray {
	return (*TypeArray)(v.ptr)
}

func (v TValue) Obj() *TypeObject {
	return (*TypeObject)(v.ptr)
}

func (v TValue) Fn() *Function {
	return (*Function)(v.ptr)
}

func (v TValue) CoreFn() *CoreFunction {
	return (*CoreFunction)(v.ptr)
}

func (v TValue) GFunction() GoFunction {
	return (*GFnWrapper)(v.ptr).Fn
}

func (v TValue) BBytes() *Bytes {
	return (*Bytes)(v.ptr)
}

func (v TValue) Enm() *Enum {
	return (*Enum)(v.ptr)
}

func (v TValue) Err() *VError {
	return (*VError)(v.ptr)
}

func (v TValue) Time() time.Time {
	return (*VTime)(v.ptr).Time
}

func (v TValue) Thread() *Thread {
	return (*Thread)(v.ptr)
}

func (v TValue) Ext() Extern {
	return v.extern
}

func (v TValue) Boolean() bool {
	switch v.ttype {
	case TNil, TError:
		return false
	case TBool:
		return v.ival != 0
	case TExtern:
		return v.extern.Boolean()
	default:
		return true
	}
}

func (v TValue) ObjectKey() string {
	switch v.ttype {
	case TNil:
		return "nil"
	case TBool:
		if v.ival != 0 {
			return "true"
		}
		return "false"
	case TInt:
		return strconv.FormatInt(v.ival, 10)
	case TFloat:
		return fmt.Sprintf("%vf", strconv.FormatFloat(v.Float(), 'g', -1, 64))
	case TString:
		return v.Str().Value
	case TArray:
		return fmt.Sprintf("Array(%p)", v.ptr)
	case TObject:
		return fmt.Sprintf("Object(%p)", v.ptr)
	case TFunction:
		return fmt.Sprintf("Function(%p)", v.ptr)
	case TCoreFn:
		return "CoreFn"
	case TGFn:
		return "GFn"
	case TBytes:
		return fmt.Sprintf("Bytes(%p)", v.ptr)
	case TEnum:
		return fmt.Sprintf("Enum(%p)", v.ptr)
	case TError:
		return fmt.Sprintf("Error(%v)", v.Err().Message.ObjectKey())
	case TTime:
		return fmt.Sprintf("Time(%p)", v.ptr)
	case TThread:
		return fmt.Sprintf("Thread(%p)", v.ptr)
	default:
		return v.extern.ObjectKey()
	}
}

func (v TValue) Type() string {
	switch v.ttype {
	case TNil:
		return "nil"
	case TBool:
		return "bool"
	case TInt:
		return "int"
	case TFloat:
		return "float"
	case TString:
		return "string"
	case TArray:
		return "array"
	case TObject:
		return "object"
	case TFunction:
		return "function"
	case TCoreFn:
		return "corefn"
	case TGFn:
		return "gfn"
	case TBytes:
		return "bytes"
	case TEnum:
		return "enum"
	case TError:
		return "error"
	case TTime:
		return "time"
	case TThread:
		return "thread"
	default:
		return v.extern.ObjectKey()
	}
}

/*
switch v.ttype {
	case TNil:
		return
	case TBool:
		return
	case TInt:
		return
	case TFloat:
		return
	case TString:
		return
	case TArray:
		return
	case TObject:
		return
	case TFunction:
		return
	case TCoreFn:
		return
	case TGFn:
		return
	case TBytes:
		return
	case TEnum:
		return
	case TError:
		return
	case TTime:
		return
	case TThread:
		return
	default:
		return
	}
*/

func (v TValue) Binop(op uint64, r TValue) (TValue, error) {
	switch v.ttype {
	case TNil:
		switch op {
		case uint64(token.AND):
			return v, nil
		case uint64(token.OR):
			return r, nil
		case uint64(token.IN):
			return IsMemberOfWithTValue(v, r)
		}
	case TBool:
		switch op {
		case uint64(token.AND):
			if v.ival != 0 {
				return r, nil
			}
			return v, nil
		case uint64(token.OR):
			if v.ival != 0 {
				return v, nil
			}
			return r, nil
		case uint64(token.IN):
			return IsMemberOfWithTValue(v, r)
		}
	case TInt:
		switch r.ttype {
		case TInt:
			ll := v.ival
			rr := r.ival
			switch op {
			case uint64(token.ADD):
				return IntVal(ll + rr), nil
			case uint64(token.SUB):
				return IntVal(ll - rr), nil
			case uint64(token.MUL):
				return IntVal(ll * rr), nil
			case uint64(token.DIV):
				if rr == 0 {
					return NilVal(), verror.ErrDivisionByZero
				}
				return IntVal(ll / rr), nil
			case uint64(token.REM):
				if rr == 0 {
					return NilVal(), verror.ErrDivisionByZero
				}
				return IntVal(ll % rr), nil
			case uint64(token.LT):
				return BoolVal(ll < rr), nil
			case uint64(token.LE):
				return BoolVal(ll <= rr), nil
			case uint64(token.GT):
				return BoolVal(ll > rr), nil
			case uint64(token.GE):
				return BoolVal(ll >= rr), nil
			case uint64(token.BXOR):
				return IntVal(int64(uint32(ll) ^ uint32(rr))), nil
			case uint64(token.BOR):
				return IntVal(int64(uint32(ll) | uint32(rr))), nil
			case uint64(token.BAND):
				return IntVal(int64(uint32(ll) & uint32(rr))), nil
			case uint64(token.BSHL):
				return IntVal(int64(uint32(ll) << uint32(rr))), nil
			case uint64(token.BSHR):
				return IntVal(int64(uint32(ll) >> uint32(rr))), nil
			}
		case TFloat:
			ll := v.ival
			rr := r.Float()
			switch op {
			case uint64(token.ADD):
				return FloatVal(float64(ll) + rr), nil
			case uint64(token.SUB):
				return FloatVal(float64(ll) - rr), nil
			case uint64(token.MUL):
				return FloatVal(float64(ll) * rr), nil
			case uint64(token.DIV):
				return FloatVal(float64(ll) / rr), nil
			case uint64(token.REM):
				return FloatVal(math.Remainder(float64(ll), rr)), nil
			case uint64(token.LT):
				return BoolVal(float64(ll) < rr), nil
			case uint64(token.LE):
				return BoolVal(float64(ll) <= rr), nil
			case uint64(token.GT):
				return BoolVal(float64(ll) > rr), nil
			case uint64(token.GE):
				return BoolVal(float64(ll) >= rr), nil
			}
		}
	case TFloat:
		switch r.ttype {
		case TFloat:
			ll := v.Float()
			rr := r.Float()
			switch op {
			case uint64(token.ADD):
				return FloatVal(ll + rr), nil
			case uint64(token.SUB):
				return FloatVal(ll - rr), nil
			case uint64(token.MUL):
				return FloatVal(ll * rr), nil
			case uint64(token.DIV):
				return FloatVal(ll / rr), nil
			case uint64(token.REM):
				return FloatVal(math.Remainder(ll, rr)), nil
			case uint64(token.LT):
				return BoolVal(ll < rr), nil
			case uint64(token.LE):
				return BoolVal(ll <= rr), nil
			case uint64(token.GT):
				return BoolVal(ll > rr), nil
			case uint64(token.GE):
				return BoolVal(ll >= rr), nil
			}
		case TInt:
			ll := v.Float()
			rr := float64(r.ival)
			switch op {
			case uint64(token.ADD):
				return FloatVal(ll + rr), nil
			case uint64(token.SUB):
				return FloatVal(ll - rr), nil
			case uint64(token.MUL):
				return FloatVal(ll * rr), nil
			case uint64(token.DIV):
				return FloatVal(ll / rr), nil
			case uint64(token.REM):
				return FloatVal(math.Remainder(ll, rr)), nil
			case uint64(token.LT):
				return BoolVal(ll < rr), nil
			case uint64(token.LE):
				return BoolVal(ll <= rr), nil
			case uint64(token.GT):
				return BoolVal(ll > rr), nil
			case uint64(token.GE):
				return BoolVal(ll >= rr), nil
			}
		}
	case TString:
		if r.ttype == TString {
			ll := v.String()
			rr := r.String()
			switch op {
			case uint64(token.ADD):
				if len(ll)+len(rr) >= verror.MaxMemSize {
					return NilVal(), verror.ErrMaxMemSize
				}
				var sb strings.Builder
				fmt.Fprint(&sb, ll, rr)
				return StringVal(sb.String()), nil
			case uint64(token.AND):
				return r, nil
			case uint64(token.OR):
				return v, nil
			case uint64(token.LT):
				return BoolVal(ll < rr), nil
			case uint64(token.LE):
				return BoolVal(ll <= rr), nil
			case uint64(token.GT):
				return BoolVal(ll > rr), nil
			case uint64(token.GE):
				return BoolVal(ll >= rr), nil
			case uint64(token.IN):
				return BoolVal(strings.Contains(ll, rr)), nil
			}
		}
	case TArray:
		if r.ttype == TArray {
			xs := v.Arr()
			rr := r.Arr()
			switch op {
			case uint64(token.ADD):
				rLen := len(rr.Value)
				if rLen == 0 {
					return v, nil
				}
				lLen := len(xs.Value)
				if rLen+lLen >= verror.MaxMemSize {
					return NilVal(), verror.ErrMaxMemSize
				}
				values := make([]TValue, lLen+rLen)
				copy(values[:lLen], xs.Value)
				copy(values[lLen:], rr.Value)
				return ArrayVal(&TypeArray{Value: values}), nil
			}
		}
	case TObject:
		if r.ttype == TObject {
			ll := v.Obj()
			rr := r.Obj()
			switch op {
			case uint64(token.ADD):
				pairs := make(map[string]TValue, len(ll.Value)+len(rr.Value))
				maps.Copy(pairs, ll.Value)
				maps.Copy(pairs, rr.Value)
				return ObjectVal(&TypeObject{Value: pairs}), nil
			case uint64(token.SUB):
				pairs := make(map[string]TValue)
				for k, v := range ll.Value {
					if _, contains := rr.Value[k]; !contains {
						pairs[k] = v
					}
				}
				return ObjectVal(&TypeObject{Value: pairs}), nil
			case uint64(token.BAND):
				pairs := make(map[string]TValue)
				for k := range ll.Value {
					if x, contains := rr.Value[k]; contains {
						pairs[k] = x
					}
				}
				return ObjectVal(&TypeObject{Value: pairs}), nil
			case uint64(token.BOR):
				pairs := make(map[string]TValue, len(ll.Value)+len(rr.Value))
				maps.Copy(pairs, ll.Value)
				maps.Copy(pairs, rr.Value)
				return ObjectVal(&TypeObject{Value: pairs}), nil
			}
		}
	case TBytes:
		if r.ttype == TBytes {
			ll := v.BBytes()
			rr := r.BBytes()
			switch op {
			case uint64(token.ADD):
				rLen := len(rr.Value)
				if rLen == 0 {
					return v, nil
				}
				lLen := len(ll.Value)
				if rLen+lLen >= verror.MaxMemSize {
					return NilVal(), verror.ErrMaxMemSize
				}
				values := make([]byte, lLen+rLen)
				copy(values[:lLen], ll.Value)
				copy(values[lLen:], rr.Value)
				return BytesVal(&Bytes{Value: values}), nil
			}
		}
	default:
		switch op {
		case uint64(token.OR):
			return v, nil
		case uint64(token.AND):
			return r, nil
		case uint64(token.IN):
			return IsMemberOfWithTValue(v, r)
		}
	}
	return NilVal(), verror.ErrBinaryOpNotDefined
}

func (v TValue) Prefix(op uint64) (TValue, error) {
	switch v.ttype {
	case TNil, TError:
		switch op {
		case uint64(token.NOT):
			return BoolVal(true), nil
		}
	case TBool:
		switch op {
		case uint64(token.NOT):
			return BoolVal(!(v.ival != 0)), nil
		}
	case TInt:
		switch op {
		case uint64(token.SUB):
			return IntVal(-v.ival), nil
		case uint64(token.NOT):
			return BoolVal(false), nil
		case uint64(token.ADD):
			return IntVal(v.ival), nil
		case uint64(token.TILDE):
			return IntVal(int64(^uint32(v.ival))), nil
		}
	case TFloat:
		switch op {
		case uint64(token.SUB):
			return FloatVal(-v.Float()), nil
		case uint64(token.NOT):
			return BoolVal(false), nil
		case uint64(token.ADD):
			return FloatVal(v.Float()), nil
		}
	case TString, TArray, TObject, TFunction, TCoreFn, TGFn, TBytes, TEnum, TTime, TThread:
		switch op {
		case uint64(token.NOT):
			return BoolVal(false), nil
		}
	default:
		return v.extern.Prefix(op)
	}
	return NilVal(), verror.ErrBinaryOpNotDefined
}

func (v TValue) String() string {
	switch v.ttype {
	case TNil:
		return "nil"
	case TBool:
		if v.ival == 0 {
			return "true"
		}
		return "false"
	case TInt:
		return strconv.FormatInt(v.ival, 10)
	case TFloat:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case TString:
		return v.Str().Value
	case TArray:
		return v.Arr().stringify(make(map[uintptr]bool))
	case TObject:
		return v.Obj().stringify(make(map[uintptr]bool))
	case TCoreFn:
		cfn := v.CoreFn()
		return fmt.Sprintf("CoreFn(arity = %v, isVar = %v, free = %v)", cfn.Arity, cfn.IsVar, cfn.Free)
	case TFunction:
		return fmt.Sprintf("Function(%p)", v.Fn())
	case TGFn:
		return "GFn"
	case TError:
		return fmt.Sprintf("Error(%v)", v.Err().Message.String())
	case TEnum:
		e := v.Enm()
		if len(e.Pairs) == 0 {
			return "enum{}"
		}
		var r []string
		for k, v := range e.Pairs {
			r = append(r, fmt.Sprintf("%v: %v", k, v))
		}
		return fmt.Sprintf("enum{%v}", strings.Join(r, ", "))
	case TBytes:
		return fmt.Sprintf("bytes[% x]", v.BBytes().Value)
	case TThread:
		th := v.Thread()
		return fmt.Sprintf("Thread(%p) State(%v)", th, th.State.String())
	case TTime:
		return v.Time().Format(time.RFC3339)
	default:
		return v.extern.String()
	}
}

func (v TValue) Equals(other TValue) bool {
	switch v.ttype {
	case TNil:
		return other.ttype == TNil
	case TBool:
		if other.ttype == TBool {
			return v.ival == other.ival
		}
	case TInt:
		if other.ttype == TInt {
			return v.ival == other.ival
		}
	case TFloat:
		if other.ttype == TFloat {
			return v.Float() == other.Float()
		}
	case TString:
		if other.ttype == TString {
			return v.Str().Value == other.Str().Value
		}
	case TArray:
		if other.ttype == TArray {
			return v.Arr() == other.Arr()
		}
	case TObject:
		if other.ttype == TObject {
			return v.Obj() == other.Obj()
		}
	case TFunction:
		if other.ttype == TFunction {
			return v.Fn() == other.Fn()
		}
	case TError:
		if other.ttype == TError {
			return v.Err() == other.Err()
		}
	case TEnum:
		if other.ttype == TEnum {
			return v.Enm() == other.Enm()
		}
	case TBytes:
		if other.ttype == TBytes {
			return v.BBytes() == other.BBytes()
		}
	case TThread:
		if other.ttype == TThread {
			return v.Thread() == other.Thread()
		}
	case TTime:
		if other.ttype == TTime {
			return v.Time().Equal(other.Time())
		}
	case TExtern:
		if other.ttype == TExtern {
			return v.extern.Equals(other)
		}
	}
	return false
}

type GoFunction func(...TValue) (TValue, error)

type GFnWrapper struct {
	Fn GoFunction
}

type VError struct {
	Message TValue
}

type VTime struct {
	Time time.Time
}

func NilVal() TValue {
	return TValue{}
}

func BoolVal(b bool) TValue {
	if b {
		return TValue{ttype: TBool, ival: 1}
	}
	return TValue{ttype: TBool}
}

func IntVal(n int64) TValue {
	return TValue{ttype: TInt, ival: n}
}

func FloatVal(f float64) TValue {
	return TValue{ttype: TFloat, ival: int64(math.Float64bits(f))}
}

func StringVal(s string) TValue {
	return TValue{ttype: TString, ptr: unsafe.Pointer(&TypeString{Value: s})}
}

func ArrayVal(a *TypeArray) TValue {
	return TValue{ttype: TArray, ptr: unsafe.Pointer(a)}
}

func ObjectVal(o *TypeObject) TValue {
	return TValue{ttype: TObject, ptr: unsafe.Pointer(o)}
}

func FunctionVal(f *Function) TValue {
	return TValue{ttype: TFunction, ptr: unsafe.Pointer(f)}
}

func GFnVal(fn func(...TValue) (TValue, error)) TValue {
	return TValue{ttype: TGFn, ptr: unsafe.Pointer(&GFnWrapper{fn})}
}

func BytesVal(b *Bytes) TValue {
	return TValue{ttype: TBytes, ptr: unsafe.Pointer(b)}
}

func EnumVal(e *Enum) TValue {
	return TValue{ttype: TEnum, ptr: unsafe.Pointer(e)}
}

func ErrorVal(msg TValue) TValue {
	return TValue{ttype: TError, ptr: unsafe.Pointer(&VError{msg})}
}

func TimeVal(t time.Time) TValue {
	return TValue{ttype: TTime, ptr: unsafe.Pointer(&VTime{t})}
}

func ExtVal(e Extern) TValue {
	return TValue{ttype: TExtern, extern: e}
}

type Extern interface {
	Boolean() bool
	Prefix(op uint64) (TValue, error)
	Binop(op uint64, rhs TValue) (TValue, error)
	Equals(other TValue) bool
	IGet(index TValue) (TValue, error)
	ISet(index, val TValue) error
	IsIterable() bool
	Iterator() TValue
	IsCallable() bool
	Call(args ...TValue) (TValue, error)
	String() string
	TypeName() string
	ObjectKey() string
	Clone() TValue
}

type ExternalDefaults struct{}

func (e *ExternalDefaults) Boolean() bool {
	return true
}

func (e *ExternalDefaults) Prefix(op uint64) (TValue, error) {
	return NilVal(), verror.ErrPrefixOpNotDefined
}

func (e *ExternalDefaults) Binop(op uint64, rhs TValue) (TValue, error) {
	return NilVal(), verror.ErrBinaryOpNotDefined
}

func (e *ExternalDefaults) IGet(index TValue) (TValue, error) {
	return NilVal(), verror.ErrValueNotIndexable
}

func (e *ExternalDefaults) ISet(index, val TValue) error {
	return verror.ErrValueNotIndexable
}

func (e *ExternalDefaults) Equals(other TValue) bool {
	return false
}

func (e *ExternalDefaults) IsIterable() bool {
	return false
}

func (e *ExternalDefaults) Iterator() TValue {
	return NilVal()
}

func (e *ExternalDefaults) IsCallable() bool {
	return false
}

func (e *ExternalDefaults) Call(args ...TValue) (TValue, error) {
	return NilVal(), verror.ErrNotImplemented
}

func (e *ExternalDefaults) ObjectKey() string {
	return "undefined"
}

func (e *ExternalDefaults) Clone() TValue {
	return NilVal()
}

type TypeArray struct {
	Value []TValue
}

func (xs *TypeArray) stringify(visited map[uintptr]bool) string {
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
		r = append(r, stringWithVisitedTValue(v, visited))
	}
	return fmt.Sprintf("[%v]", strings.Join(r, ", "))
}

type TypeObject struct {
	Value map[string]TValue
}

func (o *TypeObject) stringify(visited map[uintptr]bool) string {
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
		if k != __meta {
			r = append(r, fmt.Sprintf("%v: %v", k, stringWithVisitedTValue(v, visited)))
		}
	}
	return fmt.Sprintf("{%v}", strings.Join(r, ", "))
}

type TypeString struct {
	Value string
}
