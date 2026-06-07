package extensions

import "github.com/alkemist-17/vida"

const Success = vida.True

func GetLoader() vida.ExtensionsLoader {
	l := make(map[string]func() vida.Value)
	l["example"] = loadExampleExtension
	return l
}
