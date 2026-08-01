package vida

import "path/filepath"

func loadFoundationPath() Value {
	m := &Object{Value: make(map[string]Value, 10)}
	m.Value["join"] = NativeFunction(pathJoin)
	m.Value["resolve"] = NativeFunction(pathResolve)
	m.Value["normalize"] = NativeFunction(pathNormalize)
	m.Value["dirname"] = NativeFunction(pathDirname)
	m.Value["basename"] = NativeFunction(pathBasename)
	m.Value["extname"] = NativeFunction(pathExtname)
	m.Value["isAbsolute"] = NativeFunction(pathIsAbsolute)
	m.Value["relative"] = NativeFunction(pathRelative)
	m.Value["separator"] = &String{Value: string(filepath.Separator)}
	m.Value["listSeparator"] = &String{Value: string(filepath.ListSeparator)}
	return m
}

func pathJoin(_ *Context, args ...Value) (Value, error) {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if s, ok := a.(*String); ok {
			parts = append(parts, s.Value)
		}
	}
	return &String{Value: filepath.Join(parts...)}, nil
}

func pathResolve(_ *Context, args ...Value) (Value, error) {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if s, ok := a.(*String); ok {
			parts = append(parts, s.Value)
		}
	}
	abs, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: abs}, nil
}

func pathNormalize(_ *Context, args ...Value) (Value, error) {
	p, err := requireStringArg(args, 0, "normalize")
	if err != nil {
		return err, nil
	}
	return &String{Value: filepath.Clean(p)}, nil
}

func pathDirname(_ *Context, args ...Value) (Value, error) {
	p, err := requireStringArg(args, 0, "dirname")
	if err != nil {
		return err, nil
	}
	return &String{Value: filepath.Dir(p)}, nil
}

func pathBasename(_ *Context, args ...Value) (Value, error) {
	p, err := requireStringArg(args, 0, "basename")
	if err != nil {
		return err, nil
	}
	return &String{Value: filepath.Base(p)}, nil
}

func pathExtname(_ *Context, args ...Value) (Value, error) {
	p, err := requireStringArg(args, 0, "extname")
	if err != nil {
		return err, nil
	}
	return &String{Value: filepath.Ext(p)}, nil
}

func pathIsAbsolute(_ *Context, args ...Value) (Value, error) {
	p, err := requireStringArg(args, 0, "isAbsolute")
	if err != nil {
		return err, nil
	}
	return boolToValue(filepath.IsAbs(p)), nil
}

func pathRelative(_ *Context, args ...Value) (Value, error) {
	base, err := requireStringArg(args, 0, "relative")
	if err != nil {
		return err, nil
	}
	target, err := requireStringArg(args, 1, "relative")
	if err != nil {
		return err, nil
	}
	rel, e := filepath.Rel(base, target)
	if e != nil {
		return &VidaError{Message: &String{Value: e.Error()}}, nil
	}
	return &String{Value: rel}, nil
}
