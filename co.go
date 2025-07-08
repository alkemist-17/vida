package vida

import (
	"github.com/alkemist-17/vida/verror"
)

func loadFoundationCoroutine() Value {
	m := &Object{Value: make(map[string]Value)}
	// Co API
	m.Value["new"] = GFn(gfnNewThread)
	m.Value["run"] = GFn(gfnRunThread)
	m.Value["suspend"] = GFn(gfnSuspendThread)
	m.Value["close"] = GFn(gfnCloseThread)
	m.Value["isAlive"] = GFn(gfnIsAlive)
	m.Value["isClosed"] = GFn(gfnIsAlive)
	//m.Value["recycle"] = GFn(gfnRecycleThread)
	m.Value["state"] = GFn(gfnGetThreadState)
	m.Value["running"] = GFn(gfnRunningThread)
	// Thread stack sizes
	m.Value["stack"] = getStackSizes()
	return m
}

func gfnNewThread(args ...Value) (Value, error) {
	l := len(args)
	if l == 1 {
		if fn, ok := args[0].(*Function); ok {
			return newThread(fn, ((*clbu)[globalStateIndex].(*GlobalState)).Script, primeStack), nil
		} else {
			return NilValue, verror.ErrNotAFunction
		}
	} else if l > 1 {
		if fn, ok := args[0].(*Function); ok {
			if s, ok := args[1].(Integer); ok && femtoStack <= s && s <= fullStack {
				return newThread(fn, ((*clbu)[globalStateIndex].(*GlobalState)).Script, int(s)), nil
			} else {
				return NilValue, verror.ErrStackSize
			}
		} else {
			return NilValue, verror.ErrNotAFunction
		}
	}
	return NilValue, nil
}

func gfnGetThreadState(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return &String{Value: th.State.String()}, nil
		} else {
			return NilValue, verror.ErrNotThread
		}
	}
	return NilValue, nil
}

func gfnRunThread(args ...Value) (Value, error) {
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
			return NilValue, signal
		} else if !ok {
			return NilValue, verror.ErrNotThread
		} else if th.State == Running || th.State == Closed || th.State == Waiting {
			return NilValue, verror.ErrResumingNotSuspendedThread
		}
	}
	return NilValue, nil
}

func gfnSuspendThread(args ...Value) (Value, error) {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Main == ((*clbu)[globalStateIndex].(*GlobalState)).Current {
		return NilValue, verror.ErrSuspendingMainThread
	}
	th := (*clbu)[globalStateIndex].(*GlobalState).Current
	th.State = Suspended
	if len(args) > 0 {
		th.Channel = args[0]
	} else {
		th.Channel = NilValue
	}
	return NilValue, verror.ErrSuspendThreadSignal
}

func gfnRunningThread(args ...Value) (Value, error) {
	return ((*clbu)[globalStateIndex].(*GlobalState)).Current, nil
}

func gfnRecycleThread(args ...Value) (Value, error) {
	if len(args) > 1 {
		if th, ok := args[0].(*Thread); ok && th.State == Closed {
			if fn, okfn := args[1].(*Function); okfn {
				th.Script.MainFunction = fn
				th.State = Ready
				return th, nil
			}
		}
	}
	return NilValue, nil
}

func gfnCloseThread(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			if th.State == Ready || th.State == Suspended {
				th.State = Closed
			} else {
				return NilValue, verror.ErrClosingAThread
			}
		} else {
			return NilValue, verror.ErrNotThread
		}
	}
	return NilValue, nil
}

func gfnIsAlive(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State != Closed), nil
		} else {
			return NilValue, verror.ErrNotThread
		}
	}
	return NilValue, nil
}

func gfnIsClosed(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State == Closed), nil
		} else {
			return NilValue, verror.ErrNotThread
		}
	}
	return NilValue, nil
}

func getStackSizes() *Object {
	m := &Object{Value: make(map[string]Value)}
	m.Value["of1024"] = Integer(fullStack)
	m.Value["of512"] = Integer(halfStack)
	m.Value["of256"] = Integer(quarterStack)
	m.Value["of128"] = Integer(microStack)
	m.Value["of64"] = Integer(milliStack)
	m.Value["of32"] = Integer(nanoStack)
	m.Value["of23"] = Integer(primeStack)
	m.Value["of16"] = Integer(picoStack)
	m.Value["of8"] = Integer(femtoStack)
	return m
}
