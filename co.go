package vida

import (
	"errors"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationCoroutine() Value {
	m := &Object{Value: make(map[string]Value, 13)}
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
	return m
}

func coNewThread(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	switch l {
	case 1:
		if fn, ok := args[0].(*Function); ok {
			return coNewThreadWithSizeControl(fn, ctx.script, minFrameSize, minStackSize), nil
		}
	case 2:
		fn, okFn := args[0].(*Function)
		frameSize, ok := args[1].(Integer)
		if okFn && ok && minFrameSize <= frameSize && frameSize <= maxFrameSize {
			return coNewThreadWithSizeControl(fn, ctx.script, frameSize, minStackSize), nil
		}
		config, okConfig := args[1].(*Object)
		fSize, okFSize := config.Value["frame"].(Integer)
		sSize, okSSize := config.Value["stack"].(Integer)
		if okFn && okConfig && okFSize && okSSize && minFrameSize <= fSize && fSize <= maxFrameSize && minStackSize <= sSize && sSize <= maxStackSize {
			return coNewThreadWithSizeControl(fn, ctx.script, fSize, sSize), nil
		}
	case 3:
		fn, okFn := args[0].(*Function)
		frameSize, okFS := args[1].(Integer)
		stackSize, ok := args[2].(Integer)
		if okFn && okFS && ok && minFrameSize <= frameSize && frameSize <= maxFrameSize && minStackSize <= stackSize && stackSize <= maxStackSize {
			return coNewThreadWithSizeControl(fn, ctx.script, frameSize, stackSize), nil
		}
	}
	return Nil, errors.New("expected a function as first argument")
}

func coGetThreadState(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return &String{Value: th.State.String(), VTable: ctx.initialVTables[stringVT]}, nil
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coRunThread(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok && (th.State == Suspended || th.State == Ready) {
			var signal error
			if th.State == Ready {
				signal = verror.ErrStartThreadSignal
			} else {
				signal = verror.ErrResumeThreadSignal
			}
			th.Invoker = ctx.currentThread
			ctx.currentThread = th
			th.State = Running
			th.Invoker.State = Waiting
			ctx.vm.Thread = th
			return Nil, signal
		} else if !ok {
			return Nil, verror.ErrNotThread
		} else if th.State == Running || th.State == Done || th.State == Waiting {
			return Nil, verror.ErrResumingNotSuspendedThread
		}
	}
	return Nil, nil
}

func coSuspendThread(ctx *Context, args ...Value) (Value, error) {
	if ctx.IsMainThreadRunning() {
		return Nil, verror.ErrSuspendingMainThread
	}
	th := ctx.currentThread
	th.State = Suspended
	if len(args) > 0 {
		th.Channel = args[0]
	} else {
		th.Channel = Nil
	}
	return Nil, verror.ErrSuspendThreadSignal
}

func coGetCurrentRunningThread(ctx *Context, args ...Value) (Value, error) {
	return ctx.currentThread, nil
}

func coRecycleThread(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if th, ok := args[0].(*Thread); ok && th.State == Done {
			if fn, okfn := args[1].(*Function); okfn {
				th.Channel = Nil
				th.Script.MainFunction = fn
				th.State = Ready
				return th, nil
			}
		} else if !ok {
			return Nil, verror.ErrNotThread
		} else if th.State != Done {
			return Nil, verror.ErrRecyclingThread
		}
	}
	return Nil, nil
}

func coCompleteThread(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			if th.State == Ready || th.State == Suspended {
				th.State = Done
				th.Channel = Nil
				return th, nil
			} else {
				return Nil, verror.ErrClosingAThread
			}
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coIsActive(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State != Done), nil
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coIsDone(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State == Done), nil
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coIsMain(ctx *Context, args ...Value) (Value, error) {
	return Bool(ctx.IsMainThreadRunning()), nil
}

func coGetStackSize(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Integer(len(th.Stack)), nil
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coGetFrameSize(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Integer(len(th.Frames)), nil
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coValue(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return th.Channel, nil
		}
		return Nil, verror.ErrNotThread
	}
	return Nil, nil
}

func coNewThreadWithSizeControl(fn *Function, script *Script, frameSize, stackSize Integer) *Thread {
	return &Thread{
		Script: &Script{
			Konstants:    script.Konstants,
			GlobalStore:  script.GlobalStore,
			ErrorInfo:    script.ErrorInfo,
			MainFunction: fn,
		},
		Frames:  make([]frame, frameSize),
		Stack:   make([]Value, stackSize),
		Channel: Nil,
	}
}
