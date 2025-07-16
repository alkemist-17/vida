package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationException() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["raise"] = GFn(raiseException)
	m.Value["protected"] = GFn(catchException)
	return m
}

func raiseException(args ...Value) (Value, error) {
	if len(args) > 0 {
		err := fmt.Errorf("%s", fmt.Sprintf("\n\n  [%v]\n   Message : %v\n\n", verror.ExceptionErrType, args[0].String()))
		return NilValue, err
	}
	err := fmt.Errorf("%s", fmt.Sprintf("\n\n  [%v]\n\n", verror.ExceptionErrType))
	return NilValue, err
}

func catchException(args ...Value) (Value, error) {
	if len(args) > 0 {
		if fn, ok := args[0].(*Function); ok {
			vm := &VM{newThread(fn, ((*clbu)[globalStateIndex].(*GlobalState)).Script, picoStack)}
			if _, err := vm.runThread(0, 0, true, args[1:]...); err == nil {
				return vm.Channel, nil
			} else {
				switch e := err.(type) {
				case verror.VidaError:
					return Error{Message: &String{Value: e.Message}}, nil
				}
			}
		}
	}
	return NilValue, nil
}
