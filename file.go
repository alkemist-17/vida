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
	o := &Object{Value: make(map[string]Value)}
	o.Value[fileHandlerName] = &FileHandler{Handler: file}
	o.Value["close"] = GFn(fileClose())
	o.Value["isClosed"] = GFn(fileIsClosed())
	o.Value["name"] = GFn(fileName())
	o.Value["write"] = GFn(fileWrite())
	o.Value["lines"] = GFn(fileReadLines())
	o.Value["read"] = GFn(fileRead())
	return o
}

// File API
func fileOpen(args ...Value) (Value, error) {
	l := len(args)
	if l == 1 {
		if fname, ok := args[0].(*String); ok {
			file, err := os.OpenFile(fname.Value, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				file.Close()
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return generateFileHandlerObject(file), nil
		}
		return NilValue, nil
	}
	if len(args) > 1 {
		if path, ok := args[0].(*String); ok {
			if mode, ok := args[1].(Integer); ok {
				file, err := os.OpenFile(path.Value, int(mode), 0666)
				if err != nil {
					file.Close()
					return Error{Message: &String{Value: err.Error()}}, nil
				}
				return generateFileHandlerObject(file), nil
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func fileCreate(args ...Value) (Value, error) {
	if len(args) > 0 {
		if fname, ok := args[0].(*String); ok {
			file, err := os.Create(fname.Value)
			if err != nil {
				file.Close()
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return generateFileHandlerObject(file), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func fileExists(args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			_, err := os.Stat(path.Value)
			if errors.Is(err, os.ErrNotExist) {
				return Bool(false), nil
			}
			return Bool(true), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func fileRemove(args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			err := os.Remove(path.Value)
			if err != nil {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(true), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func fileSize(args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			fileInfo, err := os.Stat(path.Value)
			if errors.Is(err, os.ErrNotExist) {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(fileInfo.Size()), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func fileIsFile(args ...Value) (Value, error) {
	if len(args) > 0 {
		if _, ok := args[0].(*FileHandler); ok {
			return Bool(ok), nil
		}
		return Bool(false), nil
	}
	return NilValue, nil
}

func fileCreateTemp(args ...Value) (Value, error) {
	if len(args) > 1 {
		if dir, ok := args[0].(*String); ok {
			if pattern, ok := args[1].(*String); ok {
				f, err := os.CreateTemp(dir.Value, pattern.Value)
				if err != nil {
					f.Close()
					return Error{Message: &String{Value: err.Error()}}, nil
				}
				return generateFileHandlerObject(f), nil
			}
		}
	}
	return NilValue, nil
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
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (file *FileHandler) Binop(op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return NilValue, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(file, rhs)
	default:
		return NilValue, verror.ErrBinaryOpNotDefined
	}
}

func (file *FileHandler) IGet(Value) (Value, error) {
	return NilValue, verror.ErrValueNotIndexable
}

func (file *FileHandler) ISet(Value, Value) error {
	return verror.ErrValueIsConstant
}

func (file *FileHandler) Equals(other Value) Bool {
	if v, ok := other.(*FileHandler); ok {
		return v.Handler.Fd() == file.Handler.Fd()
	}
	return Bool(false)
}

func (file *FileHandler) IsIterable() Bool {
	return false
}

func (file *FileHandler) Iterator() Value {
	return NilValue
}

func (file *FileHandler) IsCallable() Bool {
	return false
}

func (file *FileHandler) String() string {
	if file.IsClosed {
		return "fileHandler(closed)"
	}
	return fmt.Sprintf("fileHandler(Fd=%v)", file.Handler.Fd())
}

func (file *FileHandler) Type() string {
	return "std/io/fileHandler"
}

func (file *FileHandler) Clone() Value {
	return file
}

// FileHandler Methods
func fileClose() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.Handler.Fd() == os.Stdout.Fd() ||
						file.Handler.Fd() == os.Stdin.Fd() ||
						file.Handler.Fd() == os.Stderr.Fd() {
						return Error{Message: &String{Value: "cannot close file open system files"}}, nil
					}
					if file.IsClosed {
						return Error{Message: &String{Value: fileAlreadyClosed}}, nil
					}
					err := file.Handler.Close()
					file.IsClosed = true
					if err != nil {
						return Error{Message: &String{Value: err.Error()}}, nil
					}
					return Bool(true), nil
				}
				return Error{Message: &String{Value: argIsNotFileHandler}}, nil
			}
		}
		return NilValue, nil
	}
}

func fileIsClosed() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					return Bool(file.IsClosed), nil
				}
				return Error{Message: &String{Value: argIsNotFileHandler}}, nil
			}
		}
		return NilValue, nil
	}
}

func fileName() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					return &String{Value: file.Handler.Name()}, nil
				}
				return Error{Message: &String{Value: argIsNotFileHandler}}, nil
			}
		}
		return NilValue, nil
	}
}

func fileReadLines() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 0 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.IsClosed {
						return Error{Message: &String{Value: fileAlreadyClosed}}, nil
					}
					scanner := bufio.NewScanner(file.Handler)
					var data []string
					for scanner.Scan() {
						data = append(data, scanner.Text())
					}
					if err := scanner.Err(); err != nil {
						file.IsClosed = true
						file.Handler.Close()
						return Error{Message: &String{Value: err.Error()}}, nil
					}
					xs := &Array{}
					for _, v := range data {
						xs.Value = append(xs.Value, &String{Value: v})
					}
					return xs, nil
				}
				return Error{Message: &String{Value: argIsNotFileHandler}}, nil
			}
		}
		return NilValue, nil
	}
}

func fileRead() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 1 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.IsClosed {
						return Error{Message: &String{Value: fileAlreadyClosed}}, nil
					}
					if b, ok := args[1].(*Bytes); ok {
						n, err := file.Handler.Read(b.Value)
						if err != nil && !errors.Is(err, io.EOF) {
							file.Handler.Close()
							file.IsClosed = true
							return Error{Message: &String{Value: err.Error()}}, nil
						}
						return Integer(n), nil
					}
					return Error{Message: &String{Value: expectedBytes}}, nil
				}
				return Error{Message: &String{Value: argIsNotFileHandler}}, nil
			}
		}
		return NilValue, nil
	}
}

func fileWrite() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 1 {
			if obj, ok := args[0].(*Object); ok {
				if file, ok := obj.Value[fileHandlerName].(*FileHandler); ok {
					if file.IsClosed {
						return Error{Message: &String{Value: fileAlreadyClosed}}, nil
					}
					if data, ok := args[1].(*String); ok {
						i, err := file.Handler.WriteString(data.Value)
						if err != nil {
							file.IsClosed = true
							file.Handler.Close()
							return Error{Message: &String{Value: err.Error()}}, nil
						}
						return Integer(i), nil
					} else if data, ok := args[1].(*Bytes); ok {
						i, err := file.Handler.Write(data.Value)
						if err != nil {
							file.IsClosed = true
							file.Handler.Close()
							return Error{Message: &String{Value: err.Error()}}, nil
						}
						return Integer(i), nil
					} else {
						return Error{Message: &String{Value: "expected data of type string"}}, nil
					}
				}
				return Error{Message: &String{Value: argIsNotFileHandler}}, nil
			}
		}
		return NilValue, nil
	}
}
