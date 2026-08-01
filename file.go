package vida

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/alkemist-17/vida/token"
)

const (
	fileHandlerName = "__FILE__"
)

// File Handler API
type File struct {
	ReferenceSemanticsImpl
	Handler  *os.File
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	Path     string
	Flags    int
	IsClosed bool
}

func newVidaFile(f *os.File, path string, flags int) *File {
	return &File{
		Handler: f,
		Reader:  bufio.NewReaderSize(f, 64*1024), // 65k
		Writer:  bufio.NewWriterSize(f, 64*1024),
		Path:    path,
		Flags:   flags,
	}
}

func (file *File) Boolean() Bool {
	return True
}

func (file *File) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, ErrPrefixOpNotDefined
	}
}

func (file *File) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return Nil, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return isMemberOf(ctx, file, rhs)
	default:
		return Nil, ErrBinaryOpNotDefined
	}
}

func (file *File) Equals(ctx *Context, other Value) Bool {
	if v, ok := other.(*File); ok {
		return v.Handler.Fd() == file.Handler.Fd()
	}
	return False
}

func (file *File) String() string {
	return fmt.Sprintf("File[%v]", file.Handler.Fd())
}

func (file *File) ObjectKey() string {
	return file.String()
}

func (file *File) Type() string {
	return fileHandlerT
}

func (file *File) Clone() Value {
	return file
}

func (file *File) GetVTable(ctx *Context) Value {
	if ctx.vtables[fileHandlerT] == nil {
		ctx.loadFileHandlerVT()
	}
	return ctx.vtables[fileHandlerT]
}

func (file *File) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[fileHandlerT] == nil {
		ctx.loadFileHandlerVT()
	}
	if vtable, ok := ctx.vtables[fileHandlerT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (file *File) guardClosed() *VidaError {
	if file.IsClosed {
		return &VidaError{Message: &String{Value: "file is already closed: " + file.Path}}
	}
	return nil
}

func (file *File) methodRead(_ *Context, args ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	n := 4096
	if len(args) > 0 {
		if iv, ok := args[0].(Integer); ok {
			n = int(iv)
		}
	}
	buf := make([]byte, n)
	read, err := file.Reader.Read(buf)
	if err != nil && err != io.EOF {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &Bytes{Value: buf[:read]}, nil
}

func (file *File) methodReadAll(_ *Context, _ ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	data, err := io.ReadAll(file.Reader)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &Bytes{Value: data}, nil
}

func (file *File) methodReadLine(_ *Context, _ ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	line, err := file.Reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	if err == io.EOF && len(line) == 0 {
		return Nil, nil // EOF signal
	}
	return &String{Value: line}, nil
}

func (file *File) methodWrite(_ *Context, args ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	if len(args) == 0 {
		return argError("write", "at least 1 argument (bytes)")
	}
	b, ok := args[0].(*Bytes)
	if !ok {
		return argError("write", "argument must be bytes")
	}
	n, err := file.Writer.Write(b.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

func (file *File) methodWriteString(_ *Context, args ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	if len(args) == 0 {
		return argError("writeString", "at least 1 argument (string)")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("writeString", "argument must be a string")
	}
	n, err := file.Writer.WriteString(s.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

func (file *File) methodSeek(_ *Context, args ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	if len(args) < 1 {
		return argError("seek", "offset (integer) and optional whence")
	}
	offset, ok := args[0].(Integer)
	if !ok {
		return argError("seek", "offset must be an integer")
	}
	whence := 0
	if len(args) > 1 {
		if w, ok := args[1].(Integer); ok {
			whence = int(w)
		}
	}
	// Flush buffered writer before seeking
	file.Writer.Flush()
	pos, err := file.Handler.Seek(int64(offset), whence)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	// Reset buffered reader after seek
	file.Reader.Reset(file.Handler)
	return Integer(pos), nil
}

func (file *File) methodTell(_ *Context, _ ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	pos, err := file.Handler.Seek(0, io.SeekCurrent)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	// Adjust for buffered data not yet consumed
	return Integer(pos - int64(file.Reader.Buffered())), nil
}

func (file *File) methodFlush(_ *Context, _ ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	if err := file.Writer.Flush(); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func (file *File) methodSync(_ *Context, _ ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	file.Writer.Flush()
	if err := file.Handler.Sync(); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func (file *File) methodTruncate(_ *Context, args ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	size := int64(0)
	if len(args) > 0 {
		if s, ok := args[0].(Integer); ok {
			size = int64(s)
		}
	}
	if err := file.Handler.Truncate(size); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func (file *File) methodStat(_ *Context, _ ...Value) (Value, error) {
	if e := file.guardClosed(); e != nil {
		return e, nil
	}
	info, err := file.Handler.Stat()
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return statToValue(info), nil
}

func (file *File) methodClose(_ *Context, _ ...Value) (Value, error) {
	if file.IsClosed {
		return Nil, nil // idempotent close
	}
	file.Writer.Flush()
	err := file.Handler.Close()
	file.IsClosed = true
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func (file *File) methodName(_ *Context, _ ...Value) (Value, error) {
	return &String{Value: file.Path}, nil
}

func (file *File) methodIsClosed(_ *Context, _ ...Value) (Value, error) {
	if file.IsClosed {
		return True, nil
	}
	return False, nil
}

func statToValue(info os.FileInfo) Value {
	return &Object{Value: map[string]Value{
		"name":    &String{Value: info.Name()},
		"size":    Integer(info.Size()),
		"isDir":   boolToValue(info.IsDir()),
		"mode":    Integer(info.Mode().Perm()),
		"modTime": Integer(info.ModTime().UnixMilli()),
		"type":    &String{Value: info.Mode().Type().String()},
	}}
}

func boolToValue(b bool) Value {
	if b {
		return True
	}
	return False
}

func (vf *File) GetMethod(name string) (Value, bool) {
	switch name {
	case "read": // vt
		return NativeFunction(vf.methodRead), true
	case "readAll": // new
		return NativeFunction(vf.methodReadAll), true
	case "readLine": // new
		return NativeFunction(vf.methodReadLine), true
	case "write": // vt
		return NativeFunction(vf.methodWrite), true
	case "writeString": // new
		return NativeFunction(vf.methodWriteString), true
	case "seek": // new
		return NativeFunction(vf.methodSeek), true
	case "tell": // new
		return NativeFunction(vf.methodTell), true
	case "flush": // new
		return NativeFunction(vf.methodFlush), true
	case "sync": // new
		return NativeFunction(vf.methodSync), true
	case "truncate": // new
		return NativeFunction(vf.methodTruncate), true
	case "stat": // new
		return NativeFunction(vf.methodStat), true
	case "close": // vt
		return NativeFunction(vf.methodClose), true
	case "name": // vt
		return NativeFunction(vf.methodName), true
	case "isClosed": // vt
		return NativeFunction(vf.methodIsClosed), true
	}
	return nil, false
}

// ─────────────────────────────────────────────────────────────────────
// Lifecycle
// ─────────────────────────────────────────────────────────────────────

func fhClose(_ *Context, args ...Value) (Value, error) {
	f, e := fhReceiver(args, "close")
	if e != nil {
		return e, nil
	}
	if f.IsClosed {
		return Nil, nil // idempotent — safe to call twice
	}
	f.Writer.Flush()
	err := f.Handler.Close()
	f.IsClosed = true
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fhIsClosed(_ *Context, args ...Value) (Value, error) {
	f, e := fhReceiver(args, "isClosed")
	if e != nil {
		return e, nil
	}
	if f.IsClosed {
		return True, nil
	}
	return False, nil
}

func fhName(_ *Context, args ...Value) (Value, error) {
	f, e := fhReceiver(args, "name")
	if e != nil {
		return e, nil
	}
	return &String{Value: f.Path}, nil
}

// ─────────────────────────────────────────────────────────────────────
// Reading
// ─────────────────────────────────────────────────────────────────────

// f.read()        → read up to 4096 bytes (default)
// f.read(1024)    → read up to 1024 bytes
func fhRead(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "read")
	if e != nil {
		return e, nil
	}
	n := 4096
	if len(args) > 1 {
		if iv, ok := args[1].(Integer); ok && iv > 0 {
			n = int(iv)
		}
	}
	buf := make([]byte, n)
	read, err := f.Reader.Read(buf)
	if err != nil && err != io.EOF {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	if read == 0 && err == io.EOF {
		return Nil, nil // EOF signal — no more data
	}
	return &Bytes{Value: buf[:read]}, nil
}

// f.readAll() → read everything remaining, return bytes
func fhReadAll(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "readAll")
	if e != nil {
		return e, nil
	}
	data, err := io.ReadAll(f.Reader)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &Bytes{Value: data}, nil
}

// f.readLine() → next line including trailing \n, or Nil at EOF
func fhReadLine(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "readLine")
	if e != nil {
		return e, nil
	}
	line, err := f.Reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	if err == io.EOF && len(line) == 0 {
		return Nil, nil // clean EOF
	}
	return &String{Value: line}, nil
}

// f.lines() → array of all remaining lines (back-compat with old API)
func fhLines(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "lines")
	if e != nil {
		return e, nil
	}
	var lines []Value
	for {
		line, err := f.Reader.ReadString('\n')
		if len(line) > 0 {
			lines = append(lines, &String{Value: line})
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	}
	return &Array{Value: lines}, nil
}

// f.readBytes(n) → alias for read(n), explicit name for clarity
func fhReadBytes(_ *Context, args ...Value) (Value, error) {
	return fhRead(nil, args...)
}

// ─────────────────────────────────────────────────────────────────────
// Writing
// ─────────────────────────────────────────────────────────────────────

// f.write(bytes) → write raw bytes, returns count
func fhWrite(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "write")
	if e != nil {
		return e, nil
	}
	if len(args) < 2 {
		return argError("write", "data (bytes or string)")
	}
	var data []byte
	switch v := args[1].(type) {
	case *Bytes:
		data = v.Value
	case *String:
		data = []byte(v.Value)
	default:
		return argError("write", "argument must be bytes or string")
	}
	n, err := f.Writer.Write(data)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

// f.writeString("hello") → convenience, returns count
func fhWriteString(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "writeString")
	if e != nil {
		return e, nil
	}
	if len(args) < 2 {
		return argError("writeString", "a string argument")
	}
	s, ok := args[1].(*String)
	if !ok {
		return argError("writeString", "argument must be a string")
	}
	n, err := f.Writer.WriteString(s.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

// f.writeLine("hello") → writes string + "\n", returns count
func fhWriteLine(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "writeLine")
	if e != nil {
		return e, nil
	}
	if len(args) < 2 {
		return argError("writeLine", "a string argument")
	}
	s, ok := args[1].(*String)
	if !ok {
		return argError("writeLine", "argument must be a string")
	}
	n, err := f.Writer.WriteString(s.Value + "\n")
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(n), nil
}

// ─────────────────────────────────────────────────────────────────────
// Positioning
// ─────────────────────────────────────────────────────────────────────

// f.seek(offset)          → SEEK_SET
// f.seek(offset, whence)  → explicit whence (io.SEEK_SET / CUR / END)
func fhSeek(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "seek")
	if e != nil {
		return e, nil
	}
	if len(args) < 2 {
		return argError("seek", "offset (integer) and optional whence")
	}
	offset, ok := args[1].(Integer)
	if !ok {
		return argError("seek", "offset must be an integer")
	}
	whence := 0 // SEEK_SET
	if len(args) > 2 {
		if w, ok := args[2].(Integer); ok {
			whence = int(w)
		}
	}
	// Flush pending writes before repositioning
	if err := f.Writer.Flush(); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	pos, err := f.Handler.Seek(int64(offset), whence)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	// Discard buffered read data (it's stale after seek)
	f.Reader.Reset(f.Handler)
	return Integer(pos), nil
}

// f.tell() → current logical position accounting for buffer
func fhTell(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "tell")
	if e != nil {
		return e, nil
	}
	pos, err := f.Handler.Seek(0, io.SeekCurrent)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	// Subtract unread buffered bytes to get the logical position
	return Integer(pos - int64(f.Reader.Buffered())), nil
}

// ─────────────────────────────────────────────────────────────────────
// Buffering & Durability
// ─────────────────────────────────────────────────────────────────────

// f.flush() → flush the bufio.Writer to the OS
func fhFlush(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "flush")
	if e != nil {
		return e, nil
	}
	if err := f.Writer.Flush(); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

// f.sync() → flush + fsync (data hits physical disk)
func fhSync(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "sync")
	if e != nil {
		return e, nil
	}
	if err := f.Writer.Flush(); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	if err := f.Handler.Sync(); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

// ─────────────────────────────────────────────────────────────────────
// Metadata & Mutation
// ─────────────────────────────────────────────────────────────────────

// f.stat() → { name, size, isDir, mode, modTime, type }
func fhStat(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "stat")
	if e != nil {
		return e, nil
	}
	info, err := f.Handler.Stat()
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return statToValue(info), nil
}

// f.size() → integer byte size (convenience shortcut)
func fhSize(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "size")
	if e != nil {
		return e, nil
	}
	info, err := f.Handler.Stat()
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(info.Size()), nil
}

// f.truncate()     → truncate to 0
// f.truncate(1024) → truncate to 1024 bytes
func fhTruncate(_ *Context, args ...Value) (Value, error) {
	f, e := fhGuard(args, "truncate")
	if e != nil {
		return e, nil
	}
	size := int64(0)
	if len(args) > 1 {
		if s, ok := args[1].(Integer); ok {
			size = int64(s)
		}
	}
	if err := f.Handler.Truncate(size); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

// fhReceiver extracts and validates the *File from args[0].
// Returns the file and nil on success, or nil and a *VidaError on failure.
func fhReceiver(args []Value, method string) (*File, *VidaError) {
	if len(args) == 0 {
		return nil, &VidaError{Message: &String{
			Value: method + ": missing receiver (internal error)",
		}}
	}
	switch f := args[0].(type) {
	case *File:
		return f, nil
	case *Object:
		// Back-compat: old-style Object wrapping a FileHandler
		if fh, ok := f.Value[fileHandlerName].(*File); ok {
			return fh, nil
		}
	}
	return nil, &VidaError{Message: &String{
		Value: method + ": receiver is not a File",
	}}
}

// fhGuard combines receiver extraction + closed check.
func fhGuard(args []Value, method string) (*File, *VidaError) {
	f, e := fhReceiver(args, method)
	if e != nil {
		return nil, e
	}
	if f.IsClosed {
		return nil, &VidaError{Message: &String{
			Value: method + ": file is already closed (" + f.Path + ")",
		}}
	}
	return f, nil
}
