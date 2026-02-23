package vida

import (
	"github.com/alkemist-17/vida/verror"
)

func loadFoundationCoroutine() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["new"] = GFn(coNewThread)
	m.Value["run"] = GFn(coRunThread)
	m.Value["suspend"] = GFn(coSuspendThread)
	m.Value["complete"] = GFn(coCompleteThread)
	m.Value["isActive"] = GFn(coIsActive)
	m.Value["isCompleted"] = GFn(coIsCompleted)
	m.Value["recycle"] = GFn(coRecycleThread)
	m.Value["state"] = GFn(coGetThreadState)
	m.Value["running"] = GFn(coGetCurrentRunningThread)
	m.Value["isMain"] = GFn(coIsMain)
	m.Value["min"] = Integer(femtoStack)
	m.Value["max"] = Integer(fullStack)
	return m
}

func coNewThread(args ...Value) (Value, error) {
	l := len(args)
	if l == 1 {
		if fn, ok := args[0].(*Function); ok {
			return newThread(fn, ((*clbu)[globalStateIndex].(*GlobalState)).Script, defaultThreadStackSize), nil
		}
		return NilValue, verror.ErrNotAFunction
	} else if l > 1 {
		if fn, ok := args[0].(*Function); ok {
			if s, ok := args[1].(Integer); ok && femtoStack <= s && s <= fullStack {
				return newThread(fn, ((*clbu)[globalStateIndex].(*GlobalState)).Script, int(s)), nil
			}
			return NilValue, verror.ErrStackSize
		}
		return NilValue, verror.ErrNotAFunction
	}
	return NilValue, nil
}

func coGetThreadState(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return &String{Value: th.State.String()}, nil
		}
		return NilValue, verror.ErrNotThread
	}
	return NilValue, nil
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
			return NilValue, signal
		} else if !ok {
			return NilValue, verror.ErrNotThread
		} else if th.State == Running || th.State == Completed || th.State == Waiting {
			return NilValue, verror.ErrResumingNotSuspendedThread
		}
	}
	return NilValue, nil
}

func coSuspendThread(args ...Value) (Value, error) {
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
			return NilValue, verror.ErrNotThread
		} else if th.State != Completed {
			return NilValue, verror.ErrRecyclingThread
		}
	}
	return NilValue, nil
}

func coCompleteThread(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			if th.State == Ready || th.State == Suspended {
				th.State = Completed
			} else {
				return NilValue, verror.ErrClosingAThread
			}
		}
		return NilValue, verror.ErrNotThread
	}
	return NilValue, nil
}

func coIsActive(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State != Completed), nil
		}
		return NilValue, verror.ErrNotThread
	}
	return NilValue, nil
}

func coIsCompleted(args ...Value) (Value, error) {
	if len(args) > 0 {
		if th, ok := args[0].(*Thread); ok {
			return Bool(th.State == Completed), nil
		}
		return NilValue, verror.ErrNotThread
	}
	return NilValue, nil
}

func coIsMain(args ...Value) (Value, error) {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Main == ((*clbu)[globalStateIndex].(*GlobalState)).Current {
		return Bool(true), nil
	}
	return Bool(false), nil
}
