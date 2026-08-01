package vida

import (
	"io"
	"os"
	"path/filepath"
)

func loadFoundationFileSystem() Value {
	m := &Object{Value: make(map[string]Value, 17)}
	m.Value["exists"] = NativeFunction(fsExists)
	m.Value["remove"] = NativeFunction(fsRemove)
	m.Value["removeAll"] = NativeFunction(fsRemoveAll)
	m.Value["rename"] = NativeFunction(fsRename)
	m.Value["copy"] = NativeFunction(fsCopy)
	m.Value["mkdir"] = NativeFunction(fsMkdir)
	m.Value["mkdirAll"] = NativeFunction(fsMkdirAll)
	m.Value["readDir"] = NativeFunction(fsReadDir)
	m.Value["stat"] = NativeFunction(fsStat)
	m.Value["lstat"] = NativeFunction(fsLstat)
	m.Value["chmod"] = NativeFunction(fsChmod)
	m.Value["symlink"] = NativeFunction(fsSymlink)
	m.Value["readLink"] = NativeFunction(fsReadLink)
	m.Value["walk"] = NativeFunction(fsWalk)
	m.Value["glob"] = NativeFunction(fsGlob)
	m.Value["tempFile"] = NativeFunction(fsTempFile)
	m.Value["tempDir"] = NativeFunction(fsTempDir)
	return m
}

func fsExists(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "exists")
	if err != nil {
		return err, nil
	}
	if _, err := os.Lstat(path); err == nil {
		return True, nil
	}
	return False, nil
}

func fsRemove(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "remove")
	if err != nil {
		return err, nil
	}
	if err := os.Remove(path); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsRemoveAll(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "removeAll")
	if err != nil {
		return err, nil
	}
	if err := os.RemoveAll(path); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsRename(_ *Context, args ...Value) (Value, error) {
	oldPath, err := requireStringArg(args, 0, "rename")
	if err != nil {
		return err, nil
	}
	newPath, err := requireStringArg(args, 1, "rename")
	if err != nil {
		return err, nil
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsCopy(_ *Context, args ...Value) (Value, error) {
	src, err := requireStringArg(args, 0, "copy")
	if err != nil {
		return err, nil
	}
	dst, err := requireStringArg(args, 1, "copy")
	if err != nil {
		return err, nil
	}

	in, e := os.Open(src)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	defer in.Close()

	out, e := os.Create(dst)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	defer out.Close()

	n, e := io.Copy(out, in)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return Integer(n), nil
}

func fsMkdir(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "mkdir")
	if err != nil {
		return err, nil
	}
	perm := os.FileMode(0o755)
	if len(args) > 1 {
		if p, ok := args[1].(Integer); ok {
			perm = os.FileMode(p)
		}
	}
	if err := os.Mkdir(path, perm); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsMkdirAll(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "mkdirAll")
	if err != nil {
		return err, nil
	}
	perm := os.FileMode(0o755)
	if len(args) > 1 {
		if p, ok := args[1].(Integer); ok {
			perm = os.FileMode(p)
		}
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsReadDir(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "readDir")
	if err != nil {
		return err, nil
	}
	entries, e := os.ReadDir(path)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	arr := &Array{Value: make([]Value, 0, len(entries))}
	for _, e := range entries {
		entry := &Object{Value: map[string]Value{
			"name":  &String{Value: e.Name()},
			"isDir": boolToValue(e.IsDir()),
			"type":  &String{Value: e.Type().String()},
		}}
		arr.Value = append(arr.Value, entry)
	}
	return arr, nil
}

func fsStat(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "stat")
	if err != nil {
		return err, nil
	}
	info, e := os.Stat(path)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return statToValue(info), nil
}

func fsLstat(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "lstat")
	if err != nil {
		return err, nil
	}
	info, e := os.Lstat(path)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return statToValue(info), nil
}

func fsChmod(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "chmod")
	if err != nil {
		return err, nil
	}
	if len(args) < 2 {
		return argError("chmod", "path (string) and mode (integer)")
	}
	mode, ok := args[1].(Integer)
	if !ok {
		return argError("chmod", "mode must be an integer (e.g. 0o644)")
	}
	if err := os.Chmod(path, os.FileMode(mode)); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsSymlink(_ *Context, args ...Value) (Value, error) {
	target, err := requireStringArg(args, 0, "symlink")
	if err != nil {
		return err, nil
	}
	link, err := requireStringArg(args, 1, "symlink")
	if err != nil {
		return err, nil
	}
	if err := os.Symlink(target, link); err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Nil, nil
}

func fsReadLink(_ *Context, args ...Value) (Value, error) {
	path, err := requireStringArg(args, 0, "readLink")
	if err != nil {
		return err, nil
	}
	target, e := os.Readlink(path)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return &String{Value: target}, nil
}

func fsWalk(_ *Context, args ...Value) (Value, error) {
	// path, err := requireStringArg(args, 0, "walk")
	// if err != nil {
	// 	return err, nil
	// }
	// if len(args) < 2 {
	// 	return argError("walk", "path (string) and callback (function)")
	// }
	// cb, ok := args[1].(*Function)
	// if !ok {
	// 	return argError("walk", "second argument must be a function(path, info)")
	// }

	// var walkErr error
	// filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	info := &Object{Value: map[string]Value{
	// 		"path":  &String{Value: p},
	// 		"name":  &String{Value: d.Name()},
	// 		"isDir": boolToValue(d.IsDir()),
	// 	}}
	// 	_, callErr := args[0].(*Context).Call(cb, &String{Value: p}, info)
	// 	if callErr != nil {
	// 		walkErr = callErr
	// 		return callErr
	// 	}
	// 	return nil
	// })
	// if walkErr != nil {
	// 	return &VidaError{Message: &String{Value: walkErr.Error()}}, nil
	// }
	return Nil, nil
}

func fsGlob(_ *Context, args ...Value) (Value, error) {
	pattern, err := requireStringArg(args, 0, "glob")
	if err != nil {
		return err, nil
	}
	matches, e := filepath.Glob(pattern)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	arr := &Array{Value: make([]Value, 0, len(matches))}
	for _, m := range matches {
		arr.Value = append(arr.Value, &String{Value: m})
	}
	return arr, nil
}

func fsTempFile(_ *Context, args ...Value) (Value, error) {
	dir := ""
	pattern := "vida-*"
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			dir = s.Value
		}
	}
	if len(args) > 1 {
		if s, ok := args[1].(*String); ok {
			pattern = s.Value
		}
	}
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return newVidaFile(f, f.Name(), os.O_RDWR), nil
}

func fsTempDir(_ *Context, args ...Value) (Value, error) {
	dir := ""
	pattern := "vida-*"
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			dir = s.Value
		}
	}
	if len(args) > 1 {
		if s, ok := args[1].(*String); ok {
			pattern = s.Value
		}
	}
	name, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: name}, nil
}
