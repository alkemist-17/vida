package vida

const (
	stringVT = "StringVT"
	arrayVT  = "ArrayVT"
	objectVT = "ObjectVT"
	bytesVT  = "BytesVT"
)

func loadStringVT() Value {
	m := &Object{Value: make(map[string]Value, 16)}
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
	m.Value["format"] = NativeFunction(coreFormat)
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

func loadObjectVT() Value {
	m := &Object{Value: make(map[string]Value, 14)}
	m.Value["len"] = NativeFunction(coreLen)
	m.Value["inject"] = NativeFunction(objectInjectProperties)
	m.Value["override"] = NativeFunction(objectInjectAndOverrideProperties)
	m.Value["extract"] = NativeFunction(objectExtractProperties)
	m.Value["implements"] = NativeFunction(objectCheckProperties)
	m.Value["set"] = NativeFunction(objectSetValue)
	m.Value["get"] = NativeFunction(objectGetValue)
	m.Value["has"] = NativeFunction(objectHasValue)
	m.Value["del"] = NativeFunction(objectDeleteProperty)
	m.Value["keys"] = NativeFunction(objectGetKeys)
	m.Value["values"] = NativeFunction(objectGetValues)
	m.Value["isEmpty"] = NativeFunction(objectIsEmpty)
	m.Value["clear"] = NativeFunction(objectClear)
	m.Value["getset"] = NativeFunction(objectGetOrSet)
	return m
}

func loadBytesVT() Value {
	m := &Object{Value: make(map[string]Value, 10)}
	m.Value["len"] = NativeFunction(coreLen)
	m.Value["toFile"] = NativeFunction(bytesToFile)
	m.Value["xor"] = NativeFunction(bytesXOR)
	m.Value["toString"] = NativeFunction(bytesToString)
	m.Value["view"] = NativeFunction(bytesView)
	m.Value["reverse"] = NativeFunction(bytesReverse)
	m.Value["timingSafeEqual"] = NativeFunction(bytesTimingSafeEqual)
	m.Value["bitLen"] = NativeFunction(bytesBitLen)
	m.Value["fill"] = NativeFunction(bytesFill)
	m.Value["concat"] = NativeFunction(bytesConcat)
	return m
}
