package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationException() Value {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
		((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
	}
	m := &Object{Value: make(map[string]Value)}
	m.Value["raise"] = GFn(exceptionRaise)
	m.Value["protected"] = GFn(exceptionCatch)
	return m
}

func exceptionRaise(args ...Value) (Value, error) {
	if len(args) > 0 {
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n  [%v]\n   Message : %v\n\n", verror.ExceptionErrType, args[0].String()))
		return NilValue, err
	}
	err := fmt.Errorf("%s", fmt.Sprintf("\n\n  [%v]\n\n", verror.ExceptionErrType))
	return NilValue, err
}

func exceptionCatch(args ...Value) (Value, error) {
	if len(args) > 0 {
		if fn, ok := args[0].(*Function); ok {
			th := ((*clbu)[globalStateIndex].(*GlobalState)).Pool.getThread()
			th.State = Ready
			th.Script.MainFunction = fn
			v, err := gfnRunThread(th)
			vm := (*clbu)[globalStateIndex].(*GlobalState).VM
			if err != nil {
				switch err {
				case verror.ErrResumeThreadSignal:
					_, threadError := vm.runThread(vm.fp, vm.Frame.ip, false, args[1:]...)
					((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
					if threadError != nil {
						switch e := threadError.(type) {
						case verror.VidaError:
							return Error{Message: &String{Value: e.Message}}, nil
						}
					}
					switch vm.State {
					case Completed, Suspended:
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
					}
				case verror.ErrStartThreadSignal:
					_, threadError := vm.runThread(vm.fp, 0, true, args[1:]...)
					((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
					if threadError != nil {
						switch e := threadError.(type) {
						case verror.VidaError:
							return Error{Message: &String{Value: e.Message}}, nil
						}
					}
					switch vm.State {
					case Completed, Suspended:
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
					}
				default:
					switch e := err.(type) {
					case verror.VidaError:
						return Error{Message: &String{Value: e.Message}}, nil
					}
				}
			}
			return v, nil
		}
	}
	return NilValue, nil
}
