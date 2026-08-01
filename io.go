package vida

import (
	"os"
)

func loadFoundationFile() Value {
	m := &Object{Value: make(map[string]Value, 22)}

	// File operations
	m.Value["open"] = NativeFunction(fileOpen)
	m.Value["create"] = NativeFunction(fileCreate)
	m.Value["readFile"] = NativeFunction(ioReadFile)
	m.Value["readTextFile"] = NativeFunction(ioReadTextFile)
	m.Value["writeFile"] = NativeFunction(ioWriteFile)
	m.Value["writeTextFile"] = NativeFunction(ioWriteTextFile)
	m.Value["appendFile"] = NativeFunction(ioAppendFile)

	// Streams
	m.Value["stdin"] = newVidaFile(os.Stdin, "[stdin]", os.O_RDONLY)
	m.Value["stdout"] = newVidaFile(os.Stdout, "[stdout]", os.O_WRONLY)
	m.Value["stderr"] = newVidaFile(os.Stderr, "[stderr]", os.O_WRONLY)

	// Constants
	m.Value["O_RDONLY"] = Integer(os.O_RDONLY)
	m.Value["O_WRONLY"] = Integer(os.O_WRONLY)
	m.Value["O_RDWR"] = Integer(os.O_RDWR)
	m.Value["O_CREATE"] = Integer(os.O_CREATE)
	m.Value["O_APPEND"] = Integer(os.O_APPEND)
	m.Value["O_TRUNC"] = Integer(os.O_TRUNC)
	m.Value["O_EXCL"] = Integer(os.O_EXCL)
	m.Value["SEEK_SET"] = Integer(0)
	m.Value["SEEK_CUR"] = Integer(1)
	m.Value["SEEK_END"] = Integer(2)
	m.Value["PERM_DEFAULT"] = Integer(0o644)
	m.Value["PERM_DIR"] = Integer(0o755)
	return m
}

// fileOpen supports both:
//
//	io.open("f.txt", "r")          ← mode string
//	io.open("f.txt", io.O_RDONLY)  ← flag integer
func fileOpen(_ *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return argError("open", "path (string) and optional mode")
	}
	path, ok := args[0].(*String)
	if !ok {
		return argError("open", "first argument must be a string path")
	}

	flag := os.O_RDONLY
	perm := os.FileMode(0o644)

	if len(args) > 1 {
		switch mode := args[1].(type) {
		case *String:
			f, p, err := parseModeString(mode.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			flag = f
			perm = p
		case Integer:
			flag = int(mode)
			if len(args) > 2 {
				if p, ok := args[2].(Integer); ok {
					perm = os.FileMode(p)
				}
			}
		default:
			return argError("open", "mode must be a string or integer")
		}
	}

	f, err := os.OpenFile(path.Value, flag, perm)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return newVidaFile(f, path.Value, flag), nil
}

func fileCreate(_ *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return argError("create", "path (string)")
	}
	path, ok := args[0].(*String)
	if !ok {
		return argError("create", "argument must be a string path")
	}
	f, err := os.Create(path.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return newVidaFile(f, path.Value, os.O_RDWR|os.O_CREATE|os.O_TRUNC), nil
}

// parseModeString translates Python-style mode strings to OS flags.
//
//	"r"   → O_RDONLY
//	"w"   → O_WRONLY | O_CREATE | O_TRUNC
//	"a"   → O_WRONLY | O_CREATE | O_APPEND
//	"r+"  → O_RDWR
//	"w+"  → O_RDWR | O_CREATE | O_TRUNC
//	"a+"  → O_RDWR | O_CREATE | O_APPEND
func parseModeString(mode string) (int, os.FileMode, error) {
	perm := os.FileMode(0o644)
	switch mode {
	case "r":
		return os.O_RDONLY, perm, nil
	case "w":
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC, perm, nil
	case "a":
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND, perm, nil
	case "r+":
		return os.O_RDWR, perm, nil
	case "w+":
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC, perm, nil
	case "a+":
		return os.O_RDWR | os.O_CREATE | os.O_APPEND, perm, nil
	default:
		return 0, 0, &VidaError{Message: &String{
			Value: `invalid mode "` + mode + `": expected one of r, w, a, r+, w+, a+`,
		}}
	}
}

func ioReadFile(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "readFile")
	if err != nil {
		return err, nil
	}
	data, e := os.ReadFile(path)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return &Bytes{Value: data}, nil
}

func ioReadTextFile(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "readTextFile")
	if err != nil {
		return err, nil
	}
	data, e := os.ReadFile(path)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return &String{Value: string(data)}, nil
}

func ioWriteFile(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "writeFile")
	if err != nil {
		return err, nil
	}
	if len(args) < 2 {
		return argError("writeFile", "path (string) and data (bytes)")
	}
	data, ok := args[1].(*Bytes)
	if !ok {
		return argError("writeFile", "second argument must be bytes")
	}
	perm := os.FileMode(0o644)
	if len(args) > 2 {
		if p, ok := args[2].(Integer); ok {
			perm = os.FileMode(p)
		}
	}
	if err := os.WriteFile(path, data.Value, perm); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(len(data.Value)), nil
}

func ioWriteTextFile(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "writeTextFile")
	if err != nil {
		return err, nil
	}
	if len(args) < 2 {
		return argError("writeTextFile", "path (string) and content (string)")
	}
	content, ok := args[1].(*String)
	if !ok {
		return argError("writeTextFile", "second argument must be a string")
	}
	perm := os.FileMode(0o644)
	if len(args) > 2 {
		if p, ok := args[2].(Integer); ok {
			perm = os.FileMode(p)
		}
	}
	if err := os.WriteFile(path, []byte(content.Value), perm); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Integer(len(content.Value)), nil
}

func ioAppendFile(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "appendFile")
	if err != nil {
		return err, nil
	}
	if len(args) < 2 {
		return argError("appendFile", "path (string) and data (bytes|string)")
	}
	var data []byte
	switch v := args[1].(type) {
	case *Bytes:
		data = v.Value
	case *String:
		data = []byte(v.Value)
	default:
		return argError("appendFile", "second argument must be bytes or string")
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	defer f.Close()
	n, e := f.Write(data)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return Integer(n), nil
}
