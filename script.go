package vida

import (
	"fmt"
	"os"

	"github.com/alkemist-17/vida/verror"
)

const vidaFileExtension = ".vida"

type Script struct {
	Store        *[]Value
	Konstants    *[]Value
	MainFunction *Function
	ErrorInfo
}

func newMainScript(name string) *Script {
	s := Script{
		Konstants:    nil,
		Store:        loadCoreLib(new([]Value)),
		MainFunction: &Function{CoreFn: &CoreFunction{ScriptName: name}},
	}
	return &s
}

func newScript(name string, store *[]Value) *Script {
	s := Script{
		Konstants:    nil,
		Store:        loadCoreLib(store),
		MainFunction: &Function{CoreFn: &CoreFunction{ScriptName: name}},
	}
	return &s
}

func (s Script) String() string {
	return fmt.Sprintf("Script(%v)", s.MainFunction.CoreFn.ScriptName)
}

func readScript(scriptName string) ([]byte, error) {
	if data, err := os.ReadFile(scriptName); err == nil {
		return data, nil
	} else {
		return nil, verror.New(scriptName, err.Error(), verror.FileErrType, 0)
	}
}
