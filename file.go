package vida

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alkemist-17/vida/token"
)

const (
	fileHandlerName     = "handler"
	argIsNotFileHandler = "argument is not a FileHandler value"
	fileAlreadyClosed   = "file is already closed"
	noStringFormat      = "no string format given"
	noOrClosedFH        = argIsNotFileHandler + " or " + fileAlreadyClosed
	expectedBytes       = "expected a value of type bytes"
)

// File API
func fileOpen(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 1 {
		if fname, ok := args[0].(*String); ok {
			file, err := os.OpenFile(fname.Value, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				file.Close()
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &FileHandler{Handler: file}, nil
		}
		return Nil, nil
	}
	if len(args) > 1 {
		if path, ok := args[0].(*String); ok {
			if mode, ok := args[1].(Integer); ok {
				file, err := os.OpenFile(path.Value, int(mode), 0666)
				if err != nil {
					file.Close()
					return &VidaError{Message: &String{Value: err.Error()}}, nil
				}
				return &FileHandler{Handler: file}, nil
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
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &FileHandler{Handler: file}, nil
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
				return &VidaError{Message: &String{Value: err.Error()}}, nil
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
				return &VidaError{Message: &String{Value: err.Error()}}, nil
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
					return &VidaError{Message: &String{Value: err.Error()}}, nil
				}
				return &FileHandler{Handler: f}, nil
			}
		}
	}
	return Nil, nil
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

func (file *FileHandler) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return !file.Boolean(), nil
	default:
		return Nil, ErrPrefixOpNotDefined
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
		return Nil, ErrBinaryOpNotDefined
	}
}

func (file *FileHandler) Equals(ctx *Context, other Value) Bool {
	if v, ok := other.(*FileHandler); ok {
		return v.Handler.Fd() == file.Handler.Fd()
	}
	return False
}

func (file *FileHandler) String() string {
	return fmt.Sprintf("file[%v]", file.Handler.Fd())
}

func (file *FileHandler) ObjectKey() string {
	return file.String()
}

func (file *FileHandler) Type() string {
	return fileHandlerT
}

func (file *FileHandler) Clone() Value {
	return file
}

func (file *FileHandler) GetVTable(ctx *Context) Value {
	if ctx.vtables[fileHandlerT] == nil {
		ctx.loadFileHandlerVT()
	}
	return ctx.vtables[fileHandlerT]
}

func (file *FileHandler) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[fileHandlerT] == nil {
		ctx.loadFileHandlerVT()
	}
	if vtable, ok := ctx.vtables[fileHandlerT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

// FileHandler Methods
func fileClose(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if file, ok := args[0].(*FileHandler); ok {
			if file.Handler.Fd() == os.Stdout.Fd() ||
				file.Handler.Fd() == os.Stdin.Fd() ||
				file.Handler.Fd() == os.Stderr.Fd() {
				return &VidaError{Message: &String{Value: "cannot close file open system files"}}, nil
			}
			if file.IsClosed {
				return &VidaError{Message: &String{Value: fileAlreadyClosed}}, nil
			}
			err := file.Handler.Close()
			file.IsClosed = true
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return True, nil
		}
		return &VidaError{Message: &String{Value: argIsNotFileHandler}}, nil
	}
	return Nil, nil
}

func fileIsClosed(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if file, ok := args[0].(*FileHandler); ok {
			return Bool(file.IsClosed), nil
		}
		return &VidaError{Message: &String{Value: argIsNotFileHandler}}, nil
	}
	return Nil, nil
}

func fileName(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if file, ok := args[0].(*FileHandler); ok {
			return &String{Value: file.Handler.Name()}, nil
		}
		return &VidaError{Message: &String{Value: argIsNotFileHandler}}, nil
	}
	return Nil, nil
}

func fileReadLines(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if file, ok := args[0].(*FileHandler); ok {
			if file.IsClosed {
				return &VidaError{Message: &String{Value: fileAlreadyClosed}}, nil
			}
			scanner := bufio.NewScanner(file.Handler)
			var data []string
			for scanner.Scan() {
				data = append(data, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				file.IsClosed = true
				file.Handler.Close()
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			xs := &Array{}
			for _, v := range data {
				xs.Value = append(xs.Value, &String{Value: v})
			}
			return xs, nil
		}
		return &VidaError{Message: &String{Value: argIsNotFileHandler}}, nil
	}
	return Nil, nil
}

func fileRead(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if file, ok := args[0].(*FileHandler); ok {
			if file.IsClosed {
				return &VidaError{Message: &String{Value: fileAlreadyClosed}}, nil
			}
			if b, ok := args[1].(*Bytes); ok {
				n, err := file.Handler.Read(b.Value)
				if err != nil && !errors.Is(err, io.EOF) {
					file.Handler.Close()
					file.IsClosed = true
					return &VidaError{Message: &String{Value: err.Error()}}, nil
				}
				return Integer(n), nil
			}
			return &VidaError{Message: &String{Value: expectedBytes}}, nil
		}
		return &VidaError{Message: &String{Value: argIsNotFileHandler}}, nil
	}
	return Nil, nil
}

func fileWrite(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if file, ok := args[0].(*FileHandler); ok {
			if file.IsClosed {
				return &VidaError{Message: &String{Value: fileAlreadyClosed}}, nil
			}
			if data, ok := args[1].(*String); ok {
				i, err := file.Handler.WriteString(data.Value)
				if err != nil {
					file.IsClosed = true
					file.Handler.Close()
					return &VidaError{Message: &String{Value: err.Error()}}, nil
				}
				return Integer(i), nil
			} else if data, ok := args[1].(*Bytes); ok {
				i, err := file.Handler.Write(data.Value)
				if err != nil {
					file.IsClosed = true
					file.Handler.Close()
					return &VidaError{Message: &String{Value: err.Error()}}, nil
				}
				return Integer(i), nil
			} else {
				return &VidaError{Message: &String{Value: "expected data of type string"}}, nil
			}
		}
		return &VidaError{Message: &String{Value: argIsNotFileHandler}}, nil
	}
	return Nil, nil
}
