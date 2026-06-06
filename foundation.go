package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationException() Value {
	if ((*clbu)[globalStateIndex].(*GlobalState)).Pool == nil {
		((*clbu)[globalStateIndex].(*GlobalState)).Pool = newThreadPool()
	}
	m := &Object{Value: make(map[string]Value, 2)}
	m.Value["raise"] = NativeFunction(exceptionRaise)
	m.Value["protected"] = NativeFunction(exceptionProtectedCall)
	return m
}

func exceptionRaise(args ...Value) (Value, error) {
	if len(args) > 0 {
		err := fmt.Errorf("\n\n\t[%v]\n\tMessage : %v\n\n", verror.ExceptionErrType, args[0].String())
		return Nil, err
	}
	err := fmt.Errorf("\n\n\t[%v]\n\n", verror.ExceptionErrType)
	return Nil, err
}

func exceptionProtectedCall(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch fn := args[0].(type) {
		case *Function:
			th := ((*clbu)[globalStateIndex].(*GlobalState)).Pool.getThread()
			th.State = Ready
			th.Script.MainFunction = fn
			v, err := coRunThread(th)
			vm := (*clbu)[globalStateIndex].(*GlobalState).VM
			if err != nil {
				switch err {
				case verror.ErrResumeThreadSignal:
					_, threadError := vm.runThread(vm.fp, vm.Frame.ip, false, args[1:]...)
					((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
					if threadError != nil {
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
						switch e := threadError.(type) {
						case *verror.VidaError:
							return &VidaError{Message: &String{Value: e.Message}}, nil
						default:
							return &VidaError{Message: &String{Value: threadError.Error()}}, nil
						}
					}
					switch vm.State {
					case Done, Suspended:
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
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
						switch e := threadError.(type) {
						case *verror.VidaError:
							return &VidaError{Message: &String{Value: e.Message}}, nil
						default:
							return &VidaError{Message: &String{Value: threadError.Error()}}, nil
						}
					}
					switch vm.State {
					case Done, Suspended:
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
					}
				default:
					switch e := err.(type) {
					case *verror.VidaError:
						return &VidaError{Message: &String{Value: e.Message}}, nil
					default:
						return &VidaError{Message: &String{Value: e.Error()}}, nil
					}
				}
			}
			return v, nil
		case NativeFunction:
			v, err := fn.Call(args[1:]...)
			vm := (*clbu)[globalStateIndex].(*GlobalState).VM
			if err != nil {
				switch err {
				case verror.ErrResumeThreadSignal:
					_, threadError := vm.runThread(vm.fp, vm.Frame.ip, false, args[2:]...)
					((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
					if threadError != nil {
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
						switch e := threadError.(type) {
						case *verror.VidaError:
							return &VidaError{Message: &String{Value: e.Message}}, nil
						default:
							return &VidaError{Message: &String{Value: threadError.Error()}}, nil
						}
					}
					switch vm.State {
					case Done, Suspended:
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
					}
				case verror.ErrStartThreadSignal:
					_, threadError := vm.runThread(vm.fp, 0, true, args[2:]...)
					((*clbu)[globalStateIndex].(*GlobalState)).Pool.releaseThread()
					if threadError != nil {
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
						switch e := threadError.(type) {
						case *verror.VidaError:
							return &VidaError{Message: &String{Value: e.Message}}, nil
						default:
							return &VidaError{Message: &String{Value: threadError.Error()}}, nil
						}
					}
					switch vm.State {
					case Done, Suspended:
						v = vm.Channel
						invoker := vm.Thread.Invoker
						invoker.State = Running
						vm.Thread.Invoker = nil
						(*clbu)[globalStateIndex].(*GlobalState).Current = invoker
						vm.Thread = invoker
					}
				default:
					switch e := err.(type) {
					case *verror.VidaError:
						return &VidaError{Message: &String{Value: e.Message}}, nil
					default:
						return &VidaError{Message: &String{Value: e.Error()}}, nil
					}
				}
			}
			return v, nil
		}
	}
	return Nil, nil
}
