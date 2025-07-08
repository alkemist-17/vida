package vida

import (
	"fmt"
	"os"
)

func loadFoundationIO() Value {
	m := &Object{Value: make(map[string]Value)}
	// fmt
	m.Value["write"] = GFn(ioWrite)
	m.Value["fwrite"] = GFn(ioFWrite)
	m.Value["printf"] = GFn(ioPrintF)
	m.Value["fprintf"] = GFn(ioFPrintF)
	m.Value["errorf"] = GFn(ioErrorf)
	// file
	m.Value["open"] = GFn(openFile)
	m.Value["create"] = GFn(createFile)
	m.Value["exists"] = GFn(exists)
	m.Value["remove"] = GFn(remove)
	m.Value["size"] = GFn(fsize)
	m.Value["isFile"] = GFn(isFile)
	m.Value["createTemp"] = GFn(tempfile)
	m.Value["tempDir"] = &String{Value: os.TempDir()}
	m.Value["ok"] = Bool(true)
	m.Value["R"] = Integer(os.O_RDONLY)
	m.Value["W"] = Integer(os.O_WRONLY)
	m.Value["RW"] = Integer(os.O_RDWR)
	m.Value["A"] = Integer(os.O_APPEND)
	m.Value["C"] = Integer(os.O_CREATE)
	m.Value["T"] = Integer(os.O_TRUNC)
	// Streams
	m.Value["stdin"] = &FileHandler{Handler: os.Stdin}
	m.Value["stdout"] = &FileHandler{Handler: os.Stdout}
	m.Value["stderr"] = &FileHandler{Handler: os.Stderr}
	return m
}

// fmt API
func ioFWrite(args ...Value) (Value, error) {
	if len(args) > 1 {
		switch handler := args[0].(type) {
		case *Object:
			if fileHandler, ok := handler.Value[fileHandlerName].(*FileHandler); ok && !fileHandler.IsClosed {
				var s []any
				for _, v := range args[1:] {
					s = append(s, v)
				}
				n, err := fmt.Fprint(fileHandler.Handler, s...)
				if err != nil {
					fileHandler.IsClosed = true
					fileHandler.Handler.Close()
					return Error{Message: &String{Value: err.Error()}}, nil
				}
				return Integer(n), nil
			}
			return Error{Message: &String{Value: noOrClosedFH}}, nil
		case *FileHandler:
			if handler.IsClosed {
				return Error{Message: &String{Value: fileAlreadyClosed}}, nil
			}
			var s []any
			for _, v := range args[1:] {
				s = append(s, v)
			}
			n, err := fmt.Fprint(handler.Handler, s...)
			if err != nil {
				handler.IsClosed = true
				handler.Handler.Close()
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
	}
	return NilValue, nil
}

func ioFPrintF(args ...Value) (Value, error) {
	if len(args) > 2 {
		switch handler := args[0].(type) {
		case *Object:
			if fileHandler, ok := handler.Value[fileHandlerName].(*FileHandler); ok && !fileHandler.IsClosed {
				if formatstr, ok := args[1].(*String); ok {
					var s []any
					for _, v := range args[2:] {
						s = append(s, v)
					}
					n, err := fmt.Fprintf(fileHandler.Handler, formatstr.Value, s...)
					if err != nil {
						fileHandler.IsClosed = true
						fileHandler.Handler.Close()
						return Error{Message: &String{Value: err.Error()}}, nil
					}
					return Integer(n), nil
				}
				return Error{Message: &String{Value: noStringFormat}}, nil
			}
			return Error{Message: &String{Value: noOrClosedFH}}, nil
		case *FileHandler:
			if formatstr, ok := args[1].(*String); ok {
				if handler.IsClosed {
					return Error{Message: &String{Value: fileAlreadyClosed}}, nil
				}
				var s []any
				for _, v := range args[2:] {
					s = append(s, v)
				}
				n, err := fmt.Fprintf(handler.Handler, formatstr.Value, s...)
				if err != nil {
					handler.IsClosed = true
					handler.Handler.Close()
					return Error{Message: &String{Value: err.Error()}}, nil
				}
				return Integer(n), nil
			}
			return Error{Message: &String{Value: noStringFormat}}, nil
		}
	}
	return NilValue, nil
}

func ioWrite(args ...Value) (Value, error) {
	var s []any
	for _, v := range args {
		s = append(s, v)
	}
	fmt.Fprint(os.Stdout, s...)
	return NilValue, nil
}

func ioPrintF(args ...Value) (Value, error) {
	if len(args) > 1 {
		if formatstr, ok := args[0].(*String); ok {
			var s []any
			for _, v := range args[1:] {
				s = append(s, v)
			}
			n, err := fmt.Fprintf(os.Stdout, formatstr.Value, s...)
			if err != nil {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
		return Error{Message: &String{Value: noStringFormat}}, nil
	}
	return NilValue, nil
}

func ioErrorf(args ...Value) (Value, error) {
	if len(args) > 1 {
		if formatstr, ok := args[0].(*String); ok {
			var s []any
			for _, v := range args[1:] {
				s = append(s, v)
			}
			n, err := fmt.Fprintf(os.Stderr, formatstr.Value, s...)
			if err != nil {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
		return Error{Message: &String{Value: noStringFormat}}, nil
	}
	return NilValue, nil
}
