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
	iGet
	iSet
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
	iGet:     "IGet",
	iSet:     "ISet",
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
}
