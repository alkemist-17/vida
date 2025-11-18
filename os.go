package vida

import (
	"os"
	"os/exec"
	"runtime"
)

func loadFoundationOS() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["args"] = GFn(osArgs)
	m.Value["env"] = GFn(osEnviron)
	m.Value["exit"] = GFn(osExit)
	m.Value["getFromEnv"] = GFn(osGetEnv)
	m.Value["pwd"] = GFn(osGetWD)
	m.Value["hostname"] = GFn(osHostname)
	m.Value["pathSeparator"] = &String{Value: string(os.PathSeparator)}
	m.Value["mkdir"] = GFn(osMkdir)
	m.Value["mkdirAll"] = GFn(osMkdirAll)
	m.Value["rm"] = GFn(osRemove)
	m.Value["rmAll"] = GFn(osRemoveAll)
	m.Value["name"] = GFn(osName)
	m.Value["arch"] = GFn(osArch)
	m.Value["run"] = GFn(osRunCMD)
	m.Value["stdin"] = &FileHandler{Handler: os.Stdin}
	m.Value["stdout"] = &FileHandler{Handler: os.Stdout}
	m.Value["stderr"] = &FileHandler{Handler: os.Stderr}
	return m
}

func osArgs(args ...Value) (Value, error) {
	xs := &Array{}
	for _, v := range os.Args {
		xs.Value = append(xs.Value, &String{Value: v})
	}
	return xs, nil
}

func osEnviron(args ...Value) (Value, error) {
	xs := &Array{}
	for _, v := range os.Environ() {
		xs.Value = append(xs.Value, &String{Value: v})
	}
	return xs, nil
}

func osExit(args ...Value) (Value, error) {
	os.Exit(0)
	return NilValue, nil
}

func osGetEnv(args ...Value) (Value, error) {
	if len(args) > 0 {
		if val, ok := args[0].(*String); ok {
			xs := make([]Value, 0, 2)
			if r, ok := os.LookupEnv(val.Value); ok {
				xs = append(xs, &String{Value: r})
				xs = append(xs, Bool(ok))
			} else {
				xs = append(xs, &String{Value: ""})
				xs = append(xs, Bool(ok))
			}
			return &Array{Value: xs}, nil
		}
	}
	return NilValue, nil
}

func osGetWD(args ...Value) (Value, error) {
	if d, e := os.Getwd(); e == nil {
		return &String{Value: d}, nil
	} else {
		return Error{Message: &String{Value: e.Error()}}, nil
	}
}

func osHostname(args ...Value) (Value, error) {
	if h, e := os.Hostname(); e == nil {
		return &String{Value: h}, nil
	} else {
		return Error{Message: &String{Value: e.Error()}}, nil
	}
}

func osMkdir(args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.Mkdir(d.Value, 0660)
			if err != nil && !os.IsExist(err) {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func osMkdirAll(args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.MkdirAll(d.Value, 0660)
			if err != nil {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func osRemove(args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.Remove(d.Value)
			if err != nil {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func osRemoveAll(args ...Value) (Value, error) {
	if len(args) > 0 {
		if d, ok := args[0].(*String); ok {
			err := os.RemoveAll(d.Value)
			if err != nil {
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}

func osName(args ...Value) (Value, error) {
	return &String{Value: runtime.GOOS}, nil
}

func osArch(args ...Value) (Value, error) {
	return &String{Value: runtime.GOARCH}, nil
}

func osRunCMD(args ...Value) (Value, error) {
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
				return Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(true), nil
		}
	}
	return NilValue, nil
}
