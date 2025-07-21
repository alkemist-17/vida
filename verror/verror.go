package verror

import (
	"errors"
	"fmt"
)

const (
	FileErrType        = "File"
	LexicalErrType     = "Lexical"
	SyntaxErrType      = "Syntax"
	CompilationErrType = "Compilation"
	RunTimeErrType     = "Runtime"
	AssertionErrType   = "Assertion Failure"
	ExceptionErrType   = "Exception"
	MaxMemSize         = 0x7FFF_FFFF
)

type VidaError struct {
	ScriptName string
	Message    string
	ErrType    string
	Line       uint
}

func (e VidaError) Error() string {
	switch e.ErrType {
	case ExceptionErrType, AssertionErrType:
		return fmt.Sprintf("\n\n  [%v]\n   Script    : %v\n   Near line : %v\n   Message   : %v\n\n", e.ErrType, e.ScriptName, e.Line, e.Message)
	default:
		if e.Line == 0 {
			return fmt.Sprintf("\n\n  [%v Error]\n   Script  : %v\n   Message : %v\n\n", e.ErrType, e.ScriptName, e.Message)
		}
		return fmt.Sprintf("\n\n  [%v Error]\n   Script    : %v\n   Near line : %v\n   Message   : %v\n\n", e.ErrType, e.ScriptName, e.Line, e.Message)
	}
}

func New(scriptName string, message string, errorType string, line uint) VidaError {
	return VidaError{
		ScriptName: scriptName,
		Line:       line,
		Message:    message,
		ErrType:    errorType,
	}
}

type StackFrameInfo struct {
	ScriptName string
	Line       uint
}

func NewStackFrameInfo(scriptName string, line uint) StackFrameInfo {
	return StackFrameInfo{
		ScriptName: scriptName,
		Line:       line,
	}
}

func (sfi StackFrameInfo) Error() string {
	return fmt.Sprintf("   Script    : %v\n   Near line : %v\n", sfi.ScriptName, sfi.Line)
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
	ErrVariadicArgs                     = errors.New("expected an array for variradic arguments")
	ErrSlice                            = errors.New("could not process the slice")
	ErrValueIsConstant                  = errors.New("value is constant")
	ErrMaxMemSize                       = errors.New("max memory size")
	ErrNotImplemented                   = errors.New("not implemented functionality for this value")
	ErrNotThread                        = errors.New("value is not a thread")
	ErrResumingNotSuspendedThread       = errors.New("cannot run a closed, running or waiting thread")
	ErrNotAFunction                     = errors.New("threads must be build from a function")
	ErrStackSize                        = errors.New("thread stack size out of limits")
	ErrSuspendingMainThread             = errors.New("cannot suspend the main thread")
	ErrClosingAThread                   = errors.New("cannot close a closed, running or waiting thread")
	ErrStartThreadSignal                = errors.New("start thread signal")
	ErrResumeThreadSignal               = errors.New("resume thread signal")
	ErrSuspendThreadSignal              = errors.New("suspend thread signal")
	ErrRecyclingThread                  = errors.New("cannot recycle a non closed thread")
)
