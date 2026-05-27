package vida

import (
	"github.com/alkemist-17/vida/verror"
)

func loadFoundationCoroutine() Value {
	m := &Object{Value: make(map[string]Value, 10)}
	m.Value["new"] = NativeFunction(coNewThread)
	m.Value["run"] = NativeFunction(coRunThread)
	m.Value["suspend"] = NativeFunction(coSuspendThread)
	m.Value["complete"] = NativeFunction(coCompleteThread)
	m.Value["isActive"] = NativeFunction(coIsActive)
	m.Value["isCompleted"] = NativeFunction(coIsCompleted)
	m.Value["recycle"] = NativeFunction(coRecycleThread)
	m.Value["state"] = NativeFunction(coGetThreadState)
	m.Value["running"] = NativeFunction(coGetCurrentRunningThread)
	m.Value["isMain"] = NativeFunction(coIsMain)
	return m
}

func coNewThread(args ...Value) (Value, error) {
	l := len(args)
	switch l {
	case 1:
		if fn, ok := args[0].(*Function); ok {
			script := ((*clbu)[globalStateIndex].(*GlobalState)).Script
			return coNewThreadWithSizeControl(fn, script, minFrameSize, minStackSize), nil
		}
	case 2:
		fn, okFn := args[0].(*Function)
		frameSize, ok := args[1].(Integer)
		if okFn && ok && minFrameSize <= frameSize && frameSize <= maxFrameSize {
			script := ((*clbu)[globalStateIndex].(*GlobalState)).Script
			return coNewThreadWithSizeControl(fn, script, frameSize, minStackSize), nil
		}
	case 3:
		fn, okFn := args[0].(*Function)
		frameSize, okFS := args[1].(Integer)
		stackSize, ok := args[2].(Integer)
		if okFn && okFS && ok && minFrameSize <= frameSize && frameSize <= maxFrameSize && minStackSize <= stackSize && stackSize <= maxStackSize {
			script := ((*clbu)[globalStateIndex].(*GlobalState)).Script
			return coNewThreadWithSizeControl(fn, script, frameSize, stackSize), nil
		}
	}
	return GlobalNil, nil
}

func coGetThreadState(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return &String{Value: th.State.String()}, nil
		}
		return GlobalNil, verror.ErrNotThread
	}
	return GlobalNil, nil
}

func coRunThread(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok && (th.State == Suspended || th.State == Ready) {
			var signal error
			if th.State == Ready {
				signal = verror.ErrStartThreadSignal
			} else {
				signal = verror.ErrResumeThreadSignal
			}
			vm := (*clbu)[globalStateIndex].(*GlobalState).VM
			th.Invoker = (*clbu)[globalStateIndex].(*GlobalState).Current
			(*clbu)[globalStateIndex].(*GlobalState).Current = th
			th.State = Running
			th.Invoker.State = Waiting
			vm.Thread = th
			return GlobalNil, signal
		} else if !ok {
			return GlobalNil, verror.ErrNotThread
		} else if th.State == Running || th.State == Completed || th.State == Waiting {
			return GlobalNil, verror.ErrResumingNotSuspendedThread
		}
	}
	return GlobalNil, nil
}

func coSuspendThread(args ...Value) (Value, error) {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Main == ((*clbu)[globalStateIndex].(*GlobalState)).Current {
		return GlobalNil, verror.ErrSuspendingMainThread
	}
	th := (*clbu)[globalStateIndex].(*GlobalState).Current
	th.State = Suspended
	if len(args) > 0 {
		th.Channel = args[0]
	} else {
		th.Channel = GlobalNil
	}
	return GlobalNil, verror.ErrSuspendThreadSignal
}

func coGetCurrentRunningThread(args ...Value) (Value, error) {
	return ((*clbu)[globalStateIndex].(*GlobalState)).Current, nil
}

func coRecycleThread(args ...Value) (Value, error) {
	if len(args) > 1 {
		if th, ok := args[0].(*Thread); ok && th.State == Completed {
			if fn, okfn := args[1].(*Function); okfn {
				th.Script.MainFunction = fn
				th.State = Ready
				return th, nil
			}
		} else if !ok {
			return GlobalNil, verror.ErrNotThread
		} else if th.State != Completed {
			return GlobalNil, verror.ErrRecyclingThread
		}
	}
	return GlobalNil, nil
}

func coCompleteThread(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			if th.State == Ready || th.State == Suspended {
				th.State = Completed
				return th, nil
			} else {
				return GlobalNil, verror.ErrClosingAThread
			}
		}
		return GlobalNil, verror.ErrNotThread
	}
	return GlobalNil, nil
}

func coIsActive(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State != Completed), nil
		}
		return GlobalNil, verror.ErrNotThread
	}
	return GlobalNil, nil
}

func coIsCompleted(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State == Completed), nil
		}
		return GlobalNil, verror.ErrNotThread
	}
	return GlobalNil, nil
}

func coIsMain(args ...Value) (Value, error) {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Main == ((*clbu)[globalStateIndex].(*GlobalState)).Current {
		return Bool(true), nil
	}
	return Bool(false), nil
}

func coNewThreadWithSizeControl(fn *Function, script *Script, frameSize, stackSize Integer) *Thread {
	return &Thread{
		Script: &Script{
			Konstants:    script.Konstants,
			Store:        script.Store,
			ErrorInfo:    script.ErrorInfo,
			MainFunction: fn,
		},
		Frames: make([]frame, frameSize),
		Stack:  make([]Value, stackSize),
	}
}
