package vida

import "fmt"

const v uint64 = 'v'
const i uint64 = 'i'
const d uint64 = 'd'
const a uint64 = 'a'

const major uint64 = 0
const minor uint64 = 3
const patch uint64 = 19
const inception uint64 = 25
const header uint64 = v<<56 | i<<48 | d<<40 | a<<32 | major<<24 | minor<<16 | patch<<8 | inception
const name = "Vida 🌱🐝🌻"

func Name() string {
	return name
}

func Version() string {
	return fmt.Sprintf("Version %v.%v.%v", major, minor, patch)
}

func About() string {
	return "\n\n\t" + name + "\n\t" + Version() + `
	
	
	What is Vida?

	Vida is a simple, elegant, interpreted computer language.
	Vida has a minimal set of constructs that makes it 
	easy to learn and suitable for most common programming tasks.
	Vida is a high level language, and it can be extended 
	by implementing the Value interface.
	The core and foundation libs are implemented in such way, so anyone can add to the
	language all what could be needed for your projects.
	Vida can be embedded in a host environment or can be used as self-sufficient language.
	Vida is a living language!
	Contributions are always welcome.
	
	Happy Vida coding!


	`
}
