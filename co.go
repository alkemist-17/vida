package vida

import (
	"fmt"
)

func loadFoundationCoroutine() Value {
	m := &Object{Value: make(map[string]Value, 18)}
	m.Value["new"] = NativeFunction(coNewThread)
	m.Value["run"] = NativeFunction(coRunThread)
	m.Value["suspend"] = NativeFunction(coSuspendThread)
	m.Value["complete"] = NativeFunction(coCompleteThread)
	m.Value["isActive"] = NativeFunction(coIsActive)
	m.Value["isDone"] = NativeFunction(coIsDone)
	m.Value["recycle"] = NativeFunction(coRecycleThread)
	m.Value["state"] = NativeFunction(coGetThreadState)
	m.Value["running"] = NativeFunction(coGetCurrentRunningThread)
	m.Value["isMain"] = NativeFunction(coIsMain)
	m.Value["getStackSize"] = NativeFunction(coGetStackSize)
	m.Value["getFrameSize"] = NativeFunction(coGetFrameSize)
	m.Value["value"] = NativeFunction(coValue)
	m.Value["READY"] = &String{Value: "ready"}
	m.Value["RUNNING"] = &String{Value: "running"}
	m.Value["SUSPENDED"] = &String{Value: "suspended"}
	m.Value["WAITING"] = &String{Value: "waiting"}
	m.Value["DONE"] = &String{Value: "done"}
	return m
}

func coNewThread(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return argError("co.new", "expected a function argument")
	}
	fn, okFn := args[0].(*Function)
	if !okFn {
		return argError("co.new", fmt.Sprintf("argument 1 must be a function, got %s", args[0].Type()))
	}
	switch l {
	case 1:
		return coNewThreadWithSizeControl(fn, ctx.script, minFrameSize, minStackSize), nil
	case 2:
		if frameSize, ok := args[1].(Integer); ok {
			if frameSize < minFrameSize || frameSize > maxFrameSize {
				return argError("co.new", fmt.Sprintf("frame size must be between %d and %d, got %d", minFrameSize, maxFrameSize, frameSize))
			}
			return coNewThreadWithSizeControl(fn, ctx.script, frameSize, minStackSize), nil
		}
		config, okConfig := args[1].(*Object)
		if !okConfig {
			return argError("co.new", fmt.Sprintf("argument 2 must be an integer (frame size) or an object ({frame, stack}), got %s", args[1].Type()))
		}
		fSize, okFSize := config.Value["frame"].(Integer)
		if !okFSize {
			return argError("co.new", "config object must have an integer 'frame' field")
		}
		sSize, okSSize := config.Value["stack"].(Integer)
		if !okSSize {
			return argError("co.new", "config object must have an integer 'stack' field")
		}
		if fSize < minFrameSize || fSize > maxFrameSize {
			return argError("co.new", fmt.Sprintf("frame size must be between %d and %d, got %d", minFrameSize, maxFrameSize, fSize))
		}
		if sSize < minStackSize || sSize > maxStackSize {
			return argError("co.new", fmt.Sprintf("stack size must be between %d and %d, got %d", minStackSize, maxStackSize, sSize))
		}
		return coNewThreadWithSizeControl(fn, ctx.script, fSize, sSize), nil
	case 3:
		frameSize, okFS := args[1].(Integer)
		if !okFS {
			return argError("co.new", fmt.Sprintf("argument 2 (frame size) must be an integer, got %s", args[1].Type()))
		}
		stackSize, ok := args[2].(Integer)
		if !ok {
			return argError("co.new", fmt.Sprintf("argument 3 (stack size) must be an integer, got %s", args[2].Type()))
		}
		if frameSize < minFrameSize || frameSize > maxFrameSize {
			return argError("co.new", fmt.Sprintf("frame size must be between %d and %d, got %d", minFrameSize, maxFrameSize, frameSize))
		}
		if stackSize < minStackSize || stackSize > maxStackSize {
			return argError("co.new", fmt.Sprintf("stack size must be between %d and %d, got %d", minStackSize, maxStackSize, stackSize))
		}
		return coNewThreadWithSizeControl(fn, ctx.script, frameSize, stackSize), nil
	default:
		return argError("co.new", fmt.Sprintf("expected 1 to 3 arguments, got %d", l))
	}
}

func coGetThreadState(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return &String{Value: th.State.String()}, nil
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.state", "expected a thread argument")
}

func coRunThread(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok && (th.State == Suspended || th.State == Ready) {
			var signal error
			if th.State == Ready {
				signal = ErrStartThreadSignal
			} else {
				signal = ErrResumeThreadSignal
			}
			th.Invoker = ctx.currentThread
			ctx.currentThread = th
			th.State = Running
			th.Invoker.State = Waiting
			ctx.vm.Thread = th
			return Nil, signal
		} else if !ok {
			return Nil, ErrNotThread
		} else if th.State == Running || th.State == Done || th.State == Waiting {
			return Nil, ErrResumingNotSuspendedThread
		}
	}
	return Nil, runtimeErrorf("co.run", "expected a thread argument")
}

func coSuspendThread(ctx *Context, args ...Value) (Value, error) {
	if ctx.IsMainThreadRunning() {
		return Nil, ErrSuspendingMainThread
	}
	th := ctx.currentThread
	th.State = Suspended
	if len(args) > 0 {
		th.Channel = args[0]
	} else {
		th.Channel = Nil
	}
	return Nil, ErrSuspendThreadSignal
}

func coGetCurrentRunningThread(ctx *Context, args ...Value) (Value, error) {
	return ctx.currentThread, nil
}

func coRecycleThread(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return Nil, runtimeErrorf("co.recycle", "expected 2 arguments (thread, function), got %d", len(args))
	}
	th, ok := args[0].(*Thread)
	if !ok {
		return Nil, ErrNotThread
	}
	if th.State != Done {
		return Nil, ErrRecyclingThread
	}
	fn, okfn := args[1].(*Function)
	if !okfn {
		return Nil, runtimeErrorf("co.recycle", "argument 2 must be a function, got %s", args[1].Type())
	}
	th.Channel = Nil
	th.Script.MainFunction = fn
	th.State = Ready
	return th, nil
}

func coCompleteThread(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			if th.State == Ready || th.State == Suspended {
				th.State = Done
				th.Channel = Nil
				return th, nil
			} else {
				return Nil, ErrClosingAThread
			}
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.complete", "expected a thread argument")
}

func coIsActive(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State != Done), nil
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.isActive", "expected a thread argument")
}

func coIsDone(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State == Done), nil
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.isDone", "expected a thread argument")
}

func coIsMain(ctx *Context, args ...Value) (Value, error) {
	return Bool(ctx.IsMainThreadRunning()), nil
}

func coGetStackSize(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Integer(len(th.Stack)), nil
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.getStackSize", "expected a thread argument")
}

func coGetFrameSize(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Integer(len(th.Frames)), nil
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.getFrameSize", "expected a thread argument")
}

func coValue(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return th.Channel, nil
		}
		return Nil, ErrNotThread
	}
	return Nil, runtimeErrorf("co.value", "expected a thread argument")
}

func coNewThreadWithSizeControl(fn *Function, script *Script, frameSize, stackSize Integer) *Thread {
	return &Thread{
		Script: &Script{
			Konstants:    script.Konstants,
			GlobalStore:  script.GlobalStore,
			MainFunction: fn,
		},
		Frames:  make([]frame, frameSize),
		Stack:   make([]Value, stackSize),
		Channel: Nil,
	}
}
