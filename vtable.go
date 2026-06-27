package vida

const (
	stringVT = "StringVT"
	arrayVT  = "ArrayVT"
	objectVT = "ObjectVT"
)

func loadStringVT() Value {
	m := &Object{Value: make(map[string]Value, 15)}
	m.Value["len"] = NativeFunction(coreLen)
	m.Value["toLower"] = NativeFunction(textToLowerCase)
	m.Value["toUpper"] = NativeFunction(textToUpperCase)
	m.Value["contains"] = NativeFunction(textContains)
	m.Value["index"] = NativeFunction(textIndex)
	m.Value["repeat"] = NativeFunction(textRepeat)
	m.Value["count"] = NativeFunction(textCount)
	m.Value["hasPrefix"] = NativeFunction(textHasPrefix)
	m.Value["hasSuffix"] = NativeFunction(textHasSuffix)
	m.Value["join"] = NativeFunction(textJoin)
	m.Value["replace"] = NativeFunction(textReplaceN)
	m.Value["match"] = NativeFunction(textMatch)
	m.Value["findFirstIndex"] = NativeFunction(textFindFirstIndex)
	m.Value["trim"] = NativeFunction(textTrim)
	m.Value["isEmpty"] = NativeFunction(textIsEmpty)
	return m
}
