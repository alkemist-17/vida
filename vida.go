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
	minor     = 3
	patch     = 63
	inception = 25
)

const (
	header uint64 = v<<56 | i<<48 | d<<40 | a<<32 | major<<24 | minor<<16 | patch<<8 | inception
	name          = "	Vida 🌿 🐝 🌻 🦖"
)

func Name() string {
	return name
}

func Version() string {
	return fmt.Sprintf("	Version %v.%v.%v", major, minor, patch)
}

func About() string {
	return "\n\n\n" + name + "\n" + Version() + `
	
	
	Welcome to Vida!


	Vida is a simple and elegant computer language.
	It features a minimal set of constructs
	that makes it easy to learn and 
	suitable for most common programming tasks.
	Vida can be seamlessly extended and
	embedded in host environments or
	used as a standalone language.
	
	
	Happy Vida coding!


	`
}
