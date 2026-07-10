package verror

import (
	"fmt"
)

const (
	FileErrType      = "File"
	LexicalErrType   = "Lexical"
	SyntaxErrType    = "Syntax"
	BuildErrType     = "Build"
	RunTimeErrType   = "Runtime"
	AssertionErrType = "Assertion Failure"
	ExceptionErrType = "Exception"
	MaxMemSize       = 0x7FFF_FFFF
)

type VidaError struct {
	ScriptID string
	Message  string
	ErrType  string
	Line     uint
}

func (e *VidaError) Error() string {
	switch e.ErrType {
	case ExceptionErrType:
		return fmt.Sprintf("\n\n\n\tScript    : [%v]\n\t%v\n\t≈ at line : %v\n\tReason    : %v\n\n\n", e.ErrType, e.ScriptID, e.Line, e.Message)
	case AssertionErrType:
		return fmt.Sprintf("\n\n\n\tScript    : [%v]\n\t%v\n\t≈ at line : %v\n\tReason    : %v\n\n\n", e.ErrType, e.ScriptID, e.Line, e.Message)
	default:
		return fmt.Sprintf("\n\n\n\t[%v Error]\n\tScript    : %v\n\t≈ at line : %v\n\tReason    : %v\n\n\n", e.ErrType, e.ScriptID, e.Line, e.Message)
	}
}

func New(scriptID string, message string, errorType string, line uint) *VidaError {
	return &VidaError{
		ScriptID: scriptID,
		Line:     line,
		Message:  message,
		ErrType:  errorType,
	}
}

type StackFrameInfo struct {
	ScriptID string
	Line     uint
	Frame    uint
}

func NewStackFrameInfo(scriptID string, line, frame uint) StackFrameInfo {
	return StackFrameInfo{
		ScriptID: scriptID,
		Line:     line,
		Frame:    frame,
	}
}

func (sfi StackFrameInfo) Error() string {
	return fmt.Sprintf("\tScript    : %v\n\t≈ at line : %v\n\tFrame     : %v\n", sfi.ScriptID, sfi.Line, sfi.Frame)
}

type internalBasicError string

func (ibe internalBasicError) Error() string {
	return string(ibe)
}

const (
	ErrStringLimit                      = internalBasicError("max size for strings has been reached")
	ErrOpNotDefinedForIterators         = internalBasicError("operation not defined for iterators")
	ErrValueNotIndexable                = internalBasicError("value not indexable")
	ErrPrefixOpNotDefined               = internalBasicError("prefix operation not defined")
	ErrBinaryOpNotDefined               = internalBasicError("binary operation not defined")
	ErrDivisionByZero                   = internalBasicError("division by zero")
	ErrExpectedInteger                  = internalBasicError("expected a value of type integer")
	ErrExpectedIntegerDifferentFromZero = internalBasicError("expected an integer value different from zero")
	ErrValueNotIterable                 = internalBasicError("value is not iterable")
	ErrValueNotCallable                 = internalBasicError("value is not callable")
	ErrStackOverflow                    = internalBasicError("stack overflow")
	ErrArity                            = internalBasicError("number of arguments different from function arity")
	ErrNotEnoughArgs                    = internalBasicError("not enough arguments passed to the function")
	ErrVariadicArgs                     = internalBasicError("expected an array for variadic arguments or the given array overflows the stack")
	ErrSlice                            = internalBasicError("could not process the slice")
	ErrView                             = internalBasicError("could not process the view")
	ErrValueIsConstant                  = internalBasicError("value is constant")
	ErrMaxMemSize                       = internalBasicError("max memory size reached")
	ErrNotImplemented                   = internalBasicError("not implemented functionality for this value")
	ErrNotThread                        = internalBasicError("value is not a thread")
	ErrResumingNotSuspendedThread       = internalBasicError("cannot run a completed, running or waiting thread")
	ErrNotAFunction                     = internalBasicError("threads must be build from functions")
	ErrSuspendingMainThread             = internalBasicError("cannot suspend the main thread")
	ErrSuspendingRunningThread          = internalBasicError("cannot suspend a running thread")
	ErrClosingAThread                   = internalBasicError("cannot complete a running, waiting or completed thread")
	ErrStartThreadSignal                = internalBasicError("start-thread-signal")
	ErrResumeThreadSignal               = internalBasicError("resume-thread-signal")
	ErrSuspendThreadSignal              = internalBasicError("suspend-thread-signal")
	ErrRecyclingThread                  = internalBasicError("cannot recycle an active thread")
	ErrSoringMixedTypes                 = internalBasicError("cannot sort mixed data types")
	ErrParallelArgs                     = internalBasicError("arguments for parallel tasks must be non empty arrays")
	ErrParallelFn                       = internalBasicError("first argument of a parallel argument must be a function")
	ErrNonNegativeIntegerTimeout        = internalBasicError("timeout must be a non negative integer milliseconds")
	ErrNonEmptyTaskArray                = internalBasicError("parallel tasks must be inside a non empty array")
	ErrInvalidJSON                      = internalBasicError("invalid json")
	ErrExpectedString                   = internalBasicError("expected argument of type string")
	ErrExpectedIntegerArg               = internalBasicError("expected argument of type integer")
	ErrExpectedBool                     = internalBasicError("expected argument of type bool")
	ErrInvalidNumberOfArguments         = internalBasicError("invalid number of arguments")
	ErrInvalidTypeOfArgument            = internalBasicError("invalid type of argument")
)

// ErrOperatorOverrideNotCallable reports that a vtable entry was found under
// a binary operator's message name (e.g. "add", "lt" — see
// tokenBinopToString) but the value stored there is not a function — so it
// cannot actually be used as an operator override. This is kept distinct
// from ErrBinaryOpNotDefined on purpose: "no override exists" and "an
// override exists but is broken" are different problems for a developer to
// diagnose, and collapsing them into the same generic message hides a very
// findable mistake (usually a plain value assigned to a vtable slot where a
// function was meant to go).
//
// Note: this does not apply to equality ("eq"). Object.Equals has its own
// separate, inline override-lookup logic and can't surface this error at
// all, since the Value.Equals contract returns only a Bool, with no error
// channel — a non-callable "eq" override is currently treated the same as
// no override at all (falls back to pointer identity) rather than erroring.
func ErrOperatorOverrideNotCallable(operator string) error {
	return fmt.Errorf("operator override '%v' is defined on the vtable but is not callable: expected a function or native function", operator)
}
