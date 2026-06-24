package extensions

import "github.com/alkemist-17/vida"

func GetLoader() vida.ExtensionsLoader {
	l := make(map[string]func() vida.Value, 1)
	l["extension.example"] = loadExampleExtension
	return l
}
