package extensions

import "github.com/alkemist-17/vida"

var Success = vida.Bool(true)

func GetLoader() vida.LibsLoader {
	l := make(map[string]func() vida.Value)
	l["hello"] = loadHelloExtension
	return l
}
