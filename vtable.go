package vida

const (
	stringVT      = "StrVT"
	arrayVT       = "ArrVT"
	objectVT      = "ObjVT"
	bytesVT       = "BtsVT"
	threadVT      = "ThrVT"
	colorVT       = "ColVT"
	fileHandlerVT = "FhdVT"
	httpClientVT  = "htpVT"
)

func loadStringVT() Value {
	vt := &Object{Value: make(map[string]Value, 16)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["toLower"] = NativeFunction(textToLowerCase)
	vt.Value["toUpper"] = NativeFunction(textToUpperCase)
	vt.Value["contains"] = NativeFunction(textContains)
	vt.Value["index"] = NativeFunction(textIndex)
	vt.Value["repeat"] = NativeFunction(textRepeat)
	vt.Value["count"] = NativeFunction(textCount)
	vt.Value["hasPrefix"] = NativeFunction(textHasPrefix)
	vt.Value["hasSuffix"] = NativeFunction(textHasSuffix)
	vt.Value["join"] = NativeFunction(textJoin)
	vt.Value["replace"] = NativeFunction(textReplaceN)
	vt.Value["match"] = NativeFunction(textMatch)
	vt.Value["findFirstIndex"] = NativeFunction(textFindFirstIndex)
	vt.Value["trim"] = NativeFunction(textTrim)
	vt.Value["isEmpty"] = NativeFunction(textIsEmpty)
	vt.Value["format"] = NativeFunction(coreFormat)
	return vt
}

func loadArrayVT() Value {
	vt := &Object{Value: make(map[string]Value, 16)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["isEmpty"] = NativeFunction(arrayIsEmpty)
	vt.Value["view"] = NativeFunction(arrayView)
	vt.Value["clear"] = NativeFunction(arrayClear)
	vt.Value["index"] = NativeFunction(arrayIndex)
	vt.Value["insert"] = NativeFunction(arrayInsert)
	vt.Value["pop"] = NativeFunction(arrayPop)
	vt.Value["append"] = NativeFunction(coreAppend)
	vt.Value["replace"] = NativeFunction(arrayReplace)
	vt.Value["clone"] = NativeFunction(coreClone)
	vt.Value["sortBy"] = NativeFunction(arraySortWithCompareVidaFunction)
	vt.Value["sort"] = NativeFunction(arraySort)
	vt.Value["reverse"] = NativeFunction(arrayReverse)
	vt.Value["delete"] = NativeFunction(arrayDelete)
	vt.Value["contains"] = NativeFunction(arrayContains)
	vt.Value["concat"] = NativeFunction(arrayConcat)
	return vt
}

func loadObjectVT() Value {
	vt := &Object{Value: make(map[string]Value, 12)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["inject"] = NativeFunction(objectInjectProperties)
	vt.Value["override"] = NativeFunction(objectInjectAndOverrideProperties)
	vt.Value["extract"] = NativeFunction(objectExtractProperties)
	vt.Value["implements"] = NativeFunction(objectCheckProperties)
	vt.Value["has"] = NativeFunction(objectCircumventHasValue)
	vt.Value["del"] = NativeFunction(objectCircumventDeleteProperty)
	vt.Value["keys"] = NativeFunction(objectGetKeys)
	vt.Value["values"] = NativeFunction(objectGetValues)
	vt.Value["isEmpty"] = NativeFunction(objectIsEmpty)
	vt.Value["clear"] = NativeFunction(objectClear)
	vt.Value["getset"] = NativeFunction(objectGetOrSet)
	return vt
}

func loadBytesVT() Value {
	vt := &Object{Value: make(map[string]Value, 10)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["toFile"] = NativeFunction(bytesToFile)
	vt.Value["xor"] = NativeFunction(bytesXOR)
	vt.Value["toString"] = NativeFunction(bytesToString)
	vt.Value["view"] = NativeFunction(bytesView)
	vt.Value["reverse"] = NativeFunction(bytesReverse)
	vt.Value["timingSafeEqual"] = NativeFunction(bytesTimingSafeEqual)
	vt.Value["bitLen"] = NativeFunction(bytesBitLen)
	vt.Value["fill"] = NativeFunction(bytesFill)
	vt.Value["concat"] = NativeFunction(bytesConcat)
	return vt
}

func loadThreadVT() Value {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["run"] = NativeFunction(coRunThread)
	vt.Value["complete"] = NativeFunction(coCompleteThread)
	vt.Value["isActive"] = NativeFunction(coIsActive)
	vt.Value["isDone"] = NativeFunction(coIsDone)
	vt.Value["state"] = NativeFunction(coGetThreadState)
	vt.Value["value"] = NativeFunction(coValue)
	return vt
}

func loadColorVT() Value {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["string"] = NativeFunction(colorString)
	vt.Value["format"] = NativeFunction(colorFormat)
	vt.Value["bg"] = NativeFunction(colorSetBG)
	vt.Value["fg"] = NativeFunction(colorSetFG)
	vt.Value["reset"] = NativeFunction(colorSetReset)
	vt.Value["resets"] = NativeFunction(colorGetReset)
	return vt
}

func loadFileHandlerVT() Value {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["close"] = NativeFunction(fileClose)
	vt.Value["isClosed"] = NativeFunction(fileIsClosed)
	vt.Value["name"] = NativeFunction(fileName)
	vt.Value["write"] = NativeFunction(fileWrite)
	vt.Value["lines"] = NativeFunction(fileReadLines)
	vt.Value["read"] = NativeFunction(fileRead)
	return vt
}

func loadHttpClientVT() Value {
	vt := &Object{Value: make(map[string]Value, 7)}
	vt.Value["get"] = NativeFunction(makeRequestFn(httpGET))
	vt.Value["post"] = NativeFunction(makeRequestFn(httpPOST))
	vt.Value["put"] = NativeFunction(makeRequestFn(httpPUT))
	vt.Value["delete"] = NativeFunction(makeRequestFn(httpDELETE))
	vt.Value["patch"] = NativeFunction(makeRequestFn(httpPATCH))
	vt.Value["head"] = NativeFunction(makeRequestFn(httpHEAD))
	vt.Value["options"] = NativeFunction(makeRequestFn(httpOPTIONS))
	return vt
}
