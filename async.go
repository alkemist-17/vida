package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type AsyncState int

const (
	AsyncReady AsyncState = iota
	AsyncWaiting
	AsyncDone
)

func (state AsyncState) String() string {
	switch state {
	case AsyncReady:
		return "ready"
	case AsyncWaiting:
		return "waiting"
	default:
		return "done"
	}
}

type Async struct {
	ReferenceSemanticsImpl
	State AsyncState
}

func (async *Async) Boolean() Bool {
	return Bool(true)
}

func (async *Async) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(false), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (async *Async) Binop(op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return async, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return Bool(false), nil
	}
	return NilValue, verror.ErrBinaryOpNotDefined
}

func (async *Async) Equals(other Value) Bool {
	if val, ok := other.(*Async); ok {
		return async == val
	}
	return false
}

func (async *Async) String() string {
	return fmt.Sprintf("Async(%p) State(%v)", async, async.State.String())
}

func (async *Async) ObjectKey() string {
	return fmt.Sprintf("Async(%p)", async)
}

func (async *Async) Type() string {
	return "async"
}

func (async *Async) Clone() Value {
	return async
}
