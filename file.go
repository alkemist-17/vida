package vida

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

const (
	fileHandlerName     = "handler"
	argIsNotFileHandler = "argument is not a FileHandler value"
	fileAlreadyClosed   = "file is already closed"
	noStringFormat      = "no string format given"
	noOrClosedFH        = argIsNotFileHandler + " or " + fileAlreadyClosed
	expectedBytes       = "expected a value of type bytes"
)

func generateFileHandlerObject(file *os.File) Value {
	o := &Object{Value: make(map[string]Value, 7)}
	o.Value[fileHandlerName] = &FileHandler{Handler: file}
	o.Value["close"] = fileClose()
	o.Value["isClosed"] = fileIsClosed()
	o.Value["name"] = fileName()
	o.Value["write"] = fileWrite()
	o.Value["lines"] = fileReadLines()
	o.Value["read"] = fileRead()
	return o
}

// File API
func fileOpen(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 1 {
		if fname, ok := args[0].(*String); ok {
			file, err := os.OpenFile(fname.Value, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				file.Close()
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return generateFileHandlerObject(file), nil
		}
		return Nil, nil
	}
	if len(args) > 1 {
		if path, ok := args[0].(*String); ok {
			if mode, ok := args[1].(Integer); ok {
				file, err := os.OpenFile(path.Value, int(mode), 0666)
				if err != nil {
					file.Close()
					return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
				}
				return generateFileHandlerObject(file), nil
			}
		}
		return Nil, nil
	}
	return Nil, nil
}

func fileCreate(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if fname, ok := args[0].(*String); ok {
			file, err := os.Create(fname.Value)
			if err != nil {
				file.Close()
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return generateFileHandlerObject(file), nil
		}
		return Nil, nil
	}
	return Nil, nil
}

func fileExists(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			_, err := os.Stat(path.Value)
			if errors.Is(err, os.ErrNotExist) {
				return False, nil
			}
			return True, nil
		}
		return Nil, nil
	}
	return Nil, nil
}

func fileRemove(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			err := os.Remove(path.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return True, nil
		}
		return Nil, nil
	}
	return Nil, nil
}

func fileSize(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			fileInfo, err := os.Stat(path.Value)
			if errors.Is(err, os.ErrNotExist) {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return Integer(fileInfo.Size()), nil
		}
		return Nil, nil
	}
	return Nil, nil
}

func fileIsFile(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if _, ok := args[0].(*FileHandler); ok {
			return Bool(ok), nil
		}
		return False, nil
	}
	return Nil, nil
}

func fileCreateTemp(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if dir, ok := args[0].(*String); ok {
			if pattern, ok := args[1].(*String); ok {
				f, err := os.CreateTemp(dir.Value, pattern.Value)
				if err != nil {
					f.Close()
					return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
				}
				return generateFileHandlerObject(f), nil
			}
		}
	}
	return Nil, nil
}

func fileGetTempDir(ctx *Context, args ...Value) (Value, error) {
	return &String{Value: os.TempDir(), VTable: ctx.initialVTables[stringVT]}, nil
}

// FileHandler API
// Type FileHandler is a wrap over *os.File
type FileHandler struct {
	ReferenceSemanticsImpl
	Handler  *os.File
	IsClosed bool
}

// Implementation of the interface Value
func (file *FileHandler) Boolean() Bool {
	return Bool(!file.IsClosed)
}

func (file *FileHandler) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return !file.Boolean(), nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (file *FileHandler) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return Nil, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, file, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (file *FileHandler) Equals(other Value) Bool {
	if v, ok := other.(*FileHandler); ok {
		return v.Handler.Fd() == file.Handler.Fd()
	}
	return False
}

func (file *FileHandler) String() string {
	return fmt.Sprintf("file(%v)", file.Handler.Fd())
}

func (file *FileHandler) Type() string {
	return "file"
}

func (file *FileHandler) Clone() Value {
	return file
}

// FileHandler Methods
func fileClose() NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.Handler.Fd() == os.Stdout.Fd() ||
						file.Handler.Fd() == os.Stdin.Fd() ||
						file.Handler.Fd() == os.Stderr.Fd() {
						return &VidaError{Message: &String{Value: "cannot close file open system files", VTable: ctx.initialVTables[stringVT]}}, nil
					}
					if file.IsClosed {
						return &VidaError{Message: &String{Value: fileAlreadyClosed, VTable: ctx.initialVTables[stringVT]}}, nil
					}
					err := file.Handler.Close()
					file.IsClosed = true
					if err != nil {
						return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
					}
					return True, nil
				}
				return &VidaError{Message: &String{Value: argIsNotFileHandler, VTable: ctx.initialVTables[stringVT]}}, nil
			}
		}
		return Nil, nil
	}
}

func fileIsClosed() NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					return Bool(file.IsClosed), nil
				}
				return &VidaError{Message: &String{Value: argIsNotFileHandler, VTable: ctx.initialVTables[stringVT]}}, nil
			}
		}
		return Nil, nil
	}
}

func fileName() NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					return &String{Value: file.Handler.Name(), VTable: ctx.initialVTables[stringVT]}, nil
				}
				return &VidaError{Message: &String{Value: argIsNotFileHandler, VTable: ctx.initialVTables[stringVT]}}, nil
			}
		}
		return Nil, nil
	}
}

func fileReadLines() NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.IsClosed {
						return &VidaError{Message: &String{Value: fileAlreadyClosed, VTable: ctx.initialVTables[stringVT]}}, nil
					}
					scanner := bufio.NewScanner(file.Handler)
					var data []string
					for scanner.Scan() {
						data = append(data, scanner.Text())
					}
					if err := scanner.Err(); err != nil {
						file.IsClosed = true
						file.Handler.Close()
						return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
					}
					xs := &Array{}
					for _, v := range data {
						xs.Value = append(xs.Value, &String{Value: v, VTable: ctx.initialVTables[stringVT]})
					}
					return xs, nil
				}
				return &VidaError{Message: &String{Value: argIsNotFileHandler, VTable: ctx.initialVTables[stringVT]}}, nil
			}
		}
		return Nil, nil
	}
}

func fileRead() NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 1 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.IsClosed {
						return &VidaError{Message: &String{Value: fileAlreadyClosed, VTable: ctx.initialVTables[stringVT]}}, nil
					}
					if b, ok := args[1].(*Bytes); ok {
						n, err := file.Handler.Read(b.Value)
						if err != nil && !errors.Is(err, io.EOF) {
							file.Handler.Close()
							file.IsClosed = true
							return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
						}
						return Integer(n), nil
					}
					return &VidaError{Message: &String{Value: expectedBytes, VTable: ctx.initialVTables[stringVT]}}, nil
				}
				return &VidaError{Message: &String{Value: argIsNotFileHandler, VTable: ctx.initialVTables[stringVT]}}, nil
			}
		}
		return Nil, nil
	}
}

func fileWrite() NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 1 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.IsClosed {
						return &VidaError{Message: &String{Value: fileAlreadyClosed, VTable: ctx.initialVTables[stringVT]}}, nil
					}
					if data, ok := args[1].(*String); ok {
						i, err := file.Handler.WriteString(data.Value)
						if err != nil {
							file.IsClosed = true
							file.Handler.Close()
							return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
						}
						return Integer(i), nil
					} else if data, ok := args[1].(*Bytes); ok {
						i, err := file.Handler.Write(data.Value)
						if err != nil {
							file.IsClosed = true
							file.Handler.Close()
							return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
						}
						return Integer(i), nil
					} else {
						return &VidaError{Message: &String{Value: "expected data of type string", VTable: ctx.initialVTables[stringVT]}}, nil
					}
				}
				return &VidaError{Message: &String{Value: argIsNotFileHandler, VTable: ctx.initialVTables[stringVT]}}, nil
			}
		}
		return Nil, nil
	}
}
