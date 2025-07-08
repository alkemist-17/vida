package extension

import "github.com/alkemist-17/vida"

func loadHelloExtension() vida.Value {
	m := &vida.Object{Value: make(map[string]vida.Value)}
	m.Value["sayHello"] = vida.GFn(greet)
	return m
}

func greet(args ...vida.Value) (vida.Value, error) {
	if len(args) > 0 {
		return &vida.String{Value: "Hello, " + args[0].String()}, nil
	} else {
		return &vida.String{Value: "Hello, World!"}, nil
	}
}
