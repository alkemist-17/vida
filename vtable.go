package vida

const (
	stringVT = "stringVT"
	arrayVT  = "arrayVT"
	objectVT = "objectVT"
)

func loadStringVTable() Value {
	m := &Object{Value: make(map[string]Value, 13)}
	m.Value["hasPrefix"] = NativeFunction(textHasPrefix)
	m.Value["hasSuffix"] = NativeFunction(textHasSuffix)
	m.Value["trim"] = NativeFunction(textTrim)
	m.Value["split"] = NativeFunction(textSplit)
	m.Value["fields"] = NativeFunction(textFields)
	m.Value["repeat"] = NativeFunction(textRepeat)
	m.Value["replaceAll"] = NativeFunction(textReplaceAll)
	m.Value["contains"] = NativeFunction(textContains)
	m.Value["index"] = NativeFunction(textIndex)
	m.Value["toLower"] = NativeFunction(textToLowerCase)
	m.Value["toUpper"] = NativeFunction(textToUpperCase)
	m.Value["count"] = NativeFunction(textCount)
	m.Value["compare"] = NativeFunction(textCompare)
	return m
}
