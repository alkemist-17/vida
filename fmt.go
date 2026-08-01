package vida

import (
	"fmt"
	"os"
	"strings"
)

func loadFoundationFmt() Value {
	m := &Object{Value: make(map[string]Value, 8)}
	m.Value["print"] = NativeFunction(ioPrint)
	m.Value["printf"] = NativeFunction(ioPrintf)
	m.Value["eprint"] = NativeFunction(ioEprint)
	m.Value["eprintf"] = NativeFunction(ioEprintf)
	m.Value["fprint"] = NativeFunction(ioFprint)
	m.Value["fprintf"] = NativeFunction(ioFprintf)
	m.Value["sprint"] = NativeFunction(ioSprint)
	m.Value["sprintf"] = NativeFunction(ioSprintf)
	return m
}

func ioPrint(_ *Context, args ...Value) (Value, error) {
	VFprint(os.Stdout, args...)
	return Nil, nil
}

func ioPrintf(_ *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return argError("printf", "format string and arguments")
	}
	fmtStr, ok := args[0].(*String)
	if !ok {
		return argError("printf", "first argument must be a format string")
	}
	n, err := VFprintf(os.Stdout, fmtStr.Value, args[1:]...)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

func ioEprint(_ *Context, args ...Value) (Value, error) {
	VFprint(os.Stderr, args...)
	return Nil, nil
}

func ioEprintf(_ *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return argError("eprintf", "format string and arguments")
	}
	fmtStr, ok := args[0].(*String)
	if !ok {
		return argError("eprintf", "first argument must be a format string")
	}
	n, err := VFprintf(os.Stderr, fmtStr.Value, args[1:]...)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

func ioFprint(_ *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("fprint", "file and at least one value")
	}

	vf, err := extractFile(args[0])

	if err != nil {
		return err, nil
	}

	if e := vf.guardClosed(); e != nil {
		return e, nil
	}

	n, werr := VFprint(vf.Writer, args[1:]...)

	if werr != nil {
		return &VidaError{Message: &String{Value: werr.Error()}}, nil
	}

	return Integer(n), nil
}

func ioFprintf(_ *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("fprintf", "file, format string, and arguments")
	}
	vf, err := extractFile(args[0])
	if err != nil {
		return err, nil
	}
	if e := vf.guardClosed(); e != nil {
		return e, nil
	}

	fmtStr, ok := args[1].(*String)

	if !ok {
		return argError("fprintf", "second argument must be a format string")
	}

	n, werr := VFprintf(vf.Writer, fmtStr.Value, args[2:]...)

	if werr != nil {
		return &VidaError{Message: &String{Value: werr.Error()}}, nil
	}

	return Integer(n), nil
}

func ioSprint(_ *Context, args ...Value) (Value, error) {
	var sb strings.Builder
	VFprint(&sb, args...)
	return &String{Value: sb.String()}, nil
}

func ioSprintf(_ *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return argError("sprintf", "format string and arguments")
	}

	fmtStr, ok := args[0].(*String)

	if !ok {
		return argError("sprintf", "first argument must be a format string")
	}

	result, _ := VSprintf(fmtStr.Value, args[1:]...)

	return &String{Value: result}, nil
}

func extractFile(v Value) (*File, *VidaError) {
	if file, ok := v.(*File); ok {
		return file, nil
	}
	return nil, &VidaError{Message: &String{Value: "expected a File object"}}
}

func requireStringArg(args []Value, idx int, fnName string) (string, *VidaError) {
	if len(args) <= idx {
		return "", &VidaError{Message: &String{
			Value: fmt.Sprintf("%s: missing argument at position %d", fnName, idx),
		}}
	}
	s, ok := args[idx].(*String)
	if !ok {
		return "", &VidaError{Message: &String{
			Value: fmt.Sprintf("%s: argument %d must be a string", fnName, idx),
		}}
	}
	return s.Value, nil
}

func argError(fn, msg string) (*VidaError, error) {
	return &VidaError{Message: &String{
		Value: fmt.Sprintf("%s: %s", fn, msg),
	}}, nil
}

func osCwd(_ *Context, _ ...Value) (Value, error) {
	dir, err := os.Getwd()
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: dir}, nil
}

func osChdir(_ *Context, args ...Value) (Value, error) {
	dir, err := requireStringArg(args, 0, "chdir")
	if err != nil {
		return err, nil
	}
	if err := os.Chdir(dir); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}
