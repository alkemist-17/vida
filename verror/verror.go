package verror

import (
	"errors"
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
		return fmt.Sprintf("\n\n\t[%v]\n\t%v\n\tStart search around line %v\n\tBecause %v\n\n\n", e.ErrType, e.ScriptID, e.Line, e.Message)
	case AssertionErrType:
		return fmt.Sprintf("\n\n\t[%v]\n\t%v\n\tStart search around line %v\n\tBecause %v\n\n\n", e.ErrType, e.ScriptID, e.Line, e.Message)
	default:
		return fmt.Sprintf("\n\n\t[%v Error]\n\t%v\n\tStart search around line %v\n\tBecause %v\n\n\n", e.ErrType, e.ScriptID, e.Line, e.Message)
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
}

func NewStackFrameInfo(scriptID string, line uint) StackFrameInfo {
	return StackFrameInfo{
		ScriptID: scriptID,
		Line:     line,
	}
}

func (sfi StackFrameInfo) Error() string {
	return fmt.Sprintf("\tScript    : %v\n\tNear line : %v\n", sfi.ScriptID, sfi.Line)
}

var (
	ErrStringLimit                      = errors.New("strings max size has been reached")
	ErrOpNotDefinedForIterators         = errors.New("operation not defined for iterators")
	ErrValueNotIndexable                = errors.New("value is not indexable")
	ErrPrefixOpNotDefined               = errors.New("prefix operation not defined")
	ErrBinaryOpNotDefined               = errors.New("binary operation not defined")
	ErrDivisionByZero                   = errors.New("division by zero not defined")
	ErrExpectedInteger                  = errors.New("expected a value of type integer")
	ErrExpectedIntegerDifferentFromZero = errors.New("expected an integer value different from zero")
	ErrValueNotIterable                 = errors.New("value is not iterable")
	ErrValueNotCallable                 = errors.New("value is not callable")
	ErrStackOverflow                    = errors.New("stack overflow")
	ErrArity                            = errors.New("given arguments count is different from arity definition")
	ErrNotEnoughArgs                    = errors.New("not given enough arguments to the function")
	ErrVariadicArgs                     = errors.New("expected an array for variradic arguments or array length overflows the current stack")
	ErrSlice                            = errors.New("could not process the slice")
	ErrValueIsConstant                  = errors.New("value is constant")
	ErrMaxMemSize                       = errors.New("max memory size reached")
	ErrNotImplemented                   = errors.New("not implemented functionality for this value")
	ErrNotThread                        = errors.New("value is not a thread value")
	ErrResumingNotSuspendedThread       = errors.New("cannot run a completed, running or waiting thread")
	ErrNotAFunction                     = errors.New("threads must be build from function values")
	ErrSuspendingMainThread             = errors.New("cannot suspend the main thread")
	ErrSuspendingRunningThread          = errors.New("cannot suspend a running thread")
	ErrClosingAThread                   = errors.New("cannot complete a running, waiting or completed thread")
	ErrStartThreadSignal                = errors.New("start thread signal")
	ErrResumeThreadSignal               = errors.New("resume thread signal")
	ErrSuspendThreadSignal              = errors.New("suspend thread signal")
	ErrRecyclingThread                  = errors.New("cannot recycle an active thread")
	ErrSoringMixedTypes                 = errors.New("cannot sort mixed value types")
	ErrParallelArgs                     = errors.New("arguments for parallel tasks must be non empty arrays")
	ErrParallelFn                       = errors.New("first argument of a parallel argument must be a function")
	ErrNonNegativeIntegerTimeout        = errors.New("timeout must be a non negative integer milliseconds")
	ErrNonEmptyTaskArray                = errors.New("parallel tasks must be inside a non empty array")
	ErrInvalidJSON                      = errors.New("invalid json")
)
