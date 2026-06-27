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

func loadArrayVT() Value {
	m := &Object{Value: make(map[string]Value, 16)}
	m.Value["len"] = NativeFunction(coreLen)
	m.Value["isEmpty"] = NativeFunction(arrayIsEmpty)
	m.Value["view"] = NativeFunction(arrayView)
	m.Value["clear"] = NativeFunction(arrayClear)
	m.Value["index"] = NativeFunction(arrayIndex)
	m.Value["insert"] = NativeFunction(arrayInsert)
	m.Value["pop"] = NativeFunction(arrayPop)
	m.Value["append"] = NativeFunction(coreAppend)
	m.Value["replace"] = NativeFunction(arrayReplace)
	m.Value["clone"] = NativeFunction(coreClone)
	m.Value["sortBy"] = NativeFunction(arraySortWithCompareVidaFunction)
	m.Value["sort"] = NativeFunction(arraySort)
	m.Value["reverse"] = NativeFunction(arrayReverse)
	m.Value["delete"] = NativeFunction(arrayDelete)
	m.Value["contains"] = NativeFunction(arrayContains)
	m.Value["concat"] = NativeFunction(arrayConcat)
	return m
}
