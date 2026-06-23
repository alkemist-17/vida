package vida

import "fmt"

const (
	v = 'v'
	i = 'i'
	d = 'd'
	a = 'a'
)

const (
	major     = 0
	minor     = 4
	patch     = 0
	inception = 25
)

const (
	header   uint64 = v<<56 | i<<48 | d<<40 | a<<32 | major<<24 | minor<<16 | patch<<8 | inception
	langName        = "	Vida 🌿🌻"
)

func Name() string {
	return langName
}

func Version() string {
	return fmt.Sprintf("	Version %v.%v.%v", major, minor, patch)
}

func About() string {
	return "\n\n\n" + langName + "\n" + Version() + `
	
	
	Welcome to Vida 🌿🌻!


	Vida is a simple, general purpose and bytecode interpreted programming language.
	It is extensible, clean and predictable.
	
	Happy Vida coding!


	`
}
