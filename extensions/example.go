package extensions

import "github.com/alkemist-17/vida"

func loadExampleExtension() vida.Value {
	m := &vida.Object{Value: make(map[string]vida.Value)}
	m.Value["sayHello"] = vida.GFnVal(greet)
	return vida.ObjectVal(m)
}

func greet(args ...vida.Value) (vida.Value, error) {
	if len(args) > 0 {
		return vida.StringVal("Hello, "+args[0].String(), nil), nil
	}
	return vida.StringVal("Hello, World!", nil), nil
}
