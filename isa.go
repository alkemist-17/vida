package vida

const (
	end = iota
	load
	store
	prefix
	binop
	eq
	binopG
	binopK
	binopQ
	array
	object
	get
	set
	slice
	forSet
	forLoop
	jump
	iForSet
	iForLoop
	check
	fun
	ret
	call
	lookup
)

var opcodes = [...]string{
	end:      "end",
	load:     "load",
	store:    "store",
	prefix:   "prefix",
	binop:    "binop",
	eq:       "eq",
	binopG:   "binopG",
	binopK:   "binopK",
	binopQ:   "binopQ",
	array:    "array",
	object:   "object",
	get:      "get",
	set:      "set",
	slice:    "slice",
	forSet:   "for",
	forLoop:  "loop",
	jump:     "jump",
	iForSet:  "iFor",
	iForLoop: "iLoop",
	check:    "check",
	fun:      "fun",
	ret:      "ret",
	call:     "call",
	lookup:   "lookup",
}
