package vida

import (
	"os"
	"os/exec"
	"runtime"
)

func loadFoundationOS() Value {
	m := &Object{Value: make(map[string]Value, 17)}
	m.Value["args"] = NativeFunction(osArgs)
	m.Value["env"] = NativeFunction(osEnviron)
	m.Value["exit"] = NativeFunction(osExit)
	m.Value["getFromEnv"] = NativeFunction(osGetEnv)
	m.Value["pwd"] = NativeFunction(osGetWD)
	m.Value["hostname"] = NativeFunction(osHostname)
	m.Value["pathSeparator"] = NativeFunction(osGetPathSeparator)
	m.Value["mkdir"] = NativeFunction(osMkdir)
	m.Value["mkdirAll"] = NativeFunction(osMkdirAll)
	m.Value["rm"] = NativeFunction(osRemove)
	m.Value["rmAll"] = NativeFunction(osRemoveAll)
	m.Value["name"] = NativeFunction(osName)
	m.Value["arch"] = NativeFunction(osArch)
	m.Value["run"] = NativeFunction(osRunCMD)
	m.Value["stdin"] = &FileHandler{Handler: os.Stdin}
	m.Value["stdout"] = &FileHandler{Handler: os.Stdout}
	m.Value["stderr"] = &FileHandler{Handler: os.Stderr}
	return m
}

func osArgs(ctx *Context, args ...Value) (Value, error) {
	xs := &Array{}
	for _, v := range os.Args {
		xs.Value = append(xs.Value, &String{Value: v, VTable: ctx.initialVTables[stringVT]})
	}
	return xs, nil
}

func osEnviron(ctx *Context, args ...Value) (Value, error) {
	xs := &Array{}
	for _, v := range os.Environ() {
		xs.Value = append(xs.Value, &String{Value: v, VTable: ctx.initialVTables[stringVT]})
	}
	return xs, nil
}

func osExit(ctx *Context, args ...Value) (Value, error) {
	os.Exit(0)
	return Nil, nil
}

func osGetEnv(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if val, ok := args[0].(*String); ok {
			xs := make([]Value, 0, 2)
			if r, ok := os.LookupEnv(val.Value); ok {
				xs = append(xs, &String{Value: r, VTable: ctx.initialVTables[stringVT]})
				xs = append(xs, Bool(ok))
			} else {
				xs = append(xs, &String{Value: EmptyString, VTable: ctx.initialVTables[stringVT]})
				xs = append(xs, Bool(ok))
			}
			return &Array{Value: xs}, nil
		}
	}
	return Nil, nil
}

func osGetWD(ctx *Context, args ...Value) (Value, error) {
	if d, e := os.Getwd(); e == nil {
		return &String{Value: d, VTable: ctx.initialVTables[stringVT]}, nil
	} else {
		return &VidaError{Message: &String{Value: e.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
	}
}

func osHostname(ctx *Context, args ...Value) (Value, error) {
	if h, e := os.Hostname(); e == nil {
		return &String{Value: h, VTable: ctx.initialVTables[stringVT]}, nil
	} else {
		return &VidaError{Message: &String{Value: e.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
	}
}

func osGetPathSeparator(ctx *Context, args ...Value) (Value, error) {
	return &String{Value: string(os.PathSeparator), VTable: ctx.initialVTables[stringVT]}, nil
}

func osMkdir(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.Mkdir(d.Value, 0660)
			if err != nil && !os.IsExist(err) {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return True, nil
		}
	}
	return Nil, nil
}

func osMkdirAll(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.MkdirAll(d.Value, 0660)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return True, nil
		}
	}
	return Nil, nil
}

func osRemove(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.Remove(d.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return True, nil
		}
	}
	return Nil, nil
}

func osRemoveAll(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.RemoveAll(d.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return True, nil
		}
	}
	return Nil, nil
}

func osName(ctx *Context, args ...Value) (Value, error) {
	return &String{Value: runtime.GOOS}, nil
}

func osArch(ctx *Context, args ...Value) (Value, error) {
	return &String{Value: runtime.GOARCH, VTable: ctx.initialVTables[stringVT]}, nil
}

func osRunCMD(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if val, ok := args[0].(*String); ok {
			var arr []string
			for i := 1; i < l; i++ {
				if v, ok := args[i].(*String); ok {
					arr = append(arr, v.Value)
				}
			}
			cmd := exec.Command(val.Value, arr...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error(), VTable: ctx.initialVTables[stringVT]}}, nil
			}
			return True, nil
		}
	}
	return Nil, nil
}
