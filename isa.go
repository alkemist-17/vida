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
	send
)

var opcodes = [...]string{
	end:      "End",
	load:     "Load",
	store:    "Store",
	prefix:   "Prefix",
	binop:    "Binop",
	eq:       "Eq",
	binopG:   "BinopG",
	binopK:   "BinopK",
	binopQ:   "BinopQ",
	array:    "Array",
	object:   "Object",
	get:      "Get",
	set:      "Set",
	slice:    "Slice",
	forSet:   "For",
	forLoop:  "Loop",
	jump:     "Jump",
	iForSet:  "IFor",
	iForLoop: "ILoop",
	check:    "Check",
	fun:      "Fun",
	ret:      "Ret",
	call:     "Call",
	send:     "send",
}
