package vida

import (
	"fmt"
	"os"
	"strings"
)

type Script struct {
	GlobalStore  *[]Value
	Konstants    *[]Value
	MainFunction *Function
}

func newScript(scriptID string, extensionsLoader ExtensionsLoader) *Script {
	s := Script{
		Konstants:    nil,
		GlobalStore:  loadCoreLib(new([]Value), extensionsLoader),
		MainFunction: &Function{CoreFn: &CoreFunction{ScriptID: scriptID, MapScriptIPLine: createNewMapScriptIPLine(scriptID)}},
	}
	return &s
}

func newSubScript(scriptID string, store *[]Value, extensionsLoader ExtensionsLoader) *Script {
	s := Script{
		Konstants:    nil,
		GlobalStore:  loadCoreLib(store, extensionsLoader),
		MainFunction: &Function{CoreFn: &CoreFunction{ScriptID: scriptID, MapScriptIPLine: createNewMapScriptIPLine(scriptID)}},
	}
	return &s
}

func (s Script) String() string {
	return fmt.Sprintf("Script(%v)", s.MainFunction.CoreFn.ScriptID)
}

func LoadScriptFromFile(path string) ([]byte, error) {
	if strings.HasSuffix(path, VidaFileExtension) {
		if data, err := os.ReadFile(path); err == nil {
			return data, nil
		} else {
			return nil, NewRuntimeError(path, err.Error(), FileErrType, 0)
		}
	}
	return nil, NewRuntimeError(path, "It is not a vida script", FileErrType, 0)
}
