package vida

// Vtables' names are data type ones.
const (
	stringT      = "string"
	arrayT       = "array"
	objectT      = "object"
	bytesT       = "bytes"
	threadT      = "thread"
	colorT       = "color"
	fileHandlerT = "file"
	httpClientT  = "httpClient"
	booleanT     = "bool"
	enumT        = "enum"
	errorT       = "error"
	nativeFuncT  = "nativeFunction"
	functionT    = "function"
	nilT         = "nil"
	integerT     = "int"
	floatT       = "float"
	timeT        = "time"
	universalT   = "universal"
)

func (ctx *Context) loadStringVT() {
	vt := &Object{Value: make(map[string]Value, 19)}
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
	vt.Value["find"] = NativeFunction(textFindFirstIndex)
	vt.Value["trim"] = NativeFunction(textExtendedTrim)
	vt.Value["isEmpty"] = NativeFunction(textIsEmpty)
	vt.Value["format"] = NativeFunction(coreFormat)
	vt.Value["toNum"] = NativeFunction(castToNumber)
	vt.Value["bytes"] = NativeFunction(textGetBytes)
	vt.Value["codePoints"] = NativeFunction(textCodepoints)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[stringT] = vt
}

func (ctx *Context) loadArrayVT() {
	vt := &Object{Value: make(map[string]Value, 15)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["isEmpty"] = NativeFunction(arrayIsEmpty)
	vt.Value["view"] = NativeFunction(collectionProcessView)
	vt.Value["clear"] = NativeFunction(arrayClear)
	vt.Value["index"] = NativeFunction(arrayIndex)
	vt.Value["insert"] = NativeFunction(arrayInsert)
	vt.Value["pop"] = NativeFunction(arrayPop)
	vt.Value["append"] = NativeFunction(coreAppend)
	vt.Value["replace"] = NativeFunction(arrayReplace)
	vt.Value["sortBy"] = NativeFunction(arraySortWithCompareVidaFunction)
	vt.Value["sort"] = NativeFunction(arraySort)
	vt.Value["reverse"] = NativeFunction(arrayReverse)
	vt.Value["remove"] = NativeFunction(arrayRemove)
	vt.Value["contains"] = NativeFunction(arrayContains)
	vt.Value["concat"] = NativeFunction(arrayConcat)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[arrayT] = vt
}

func (ctx *Context) loadObjectVT() {
	vt := &Object{Value: make(map[string]Value, 14)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["inject"] = NativeFunction(objectInjectProperties)
	vt.Value["override"] = NativeFunction(objectInjectAndOverrideProperties)
	vt.Value["extract"] = NativeFunction(objectExtractProperties)
	vt.Value["implements"] = NativeFunction(objectCheckProperties)
	vt.Value["get"] = NativeFunction(objectCircumventGetValue)
	vt.Value["set"] = NativeFunction(objectCircumventSetValue)
	vt.Value["has"] = NativeFunction(objectCircumventHasValue)
	vt.Value["del"] = NativeFunction(objectCircumventDeleteProperty)
	vt.Value["keys"] = NativeFunction(objectGetKeys)
	vt.Value["values"] = NativeFunction(objectGetValues)
	vt.Value["isEmpty"] = NativeFunction(objectIsEmpty)
	vt.Value["clear"] = NativeFunction(objectClear)
	vt.Value["getset"] = NativeFunction(objectGetOrSet)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[objectT] = vt
}

func (ctx *Context) loadBytesVT() {
	vt := &Object{Value: make(map[string]Value, 10)}
	vt.Value["len"] = NativeFunction(coreLen)
	vt.Value["toFile"] = NativeFunction(bytesToFile)
	vt.Value["xor"] = NativeFunction(bytesXOR)
	vt.Value["toString"] = NativeFunction(bytesToString)
	vt.Value["view"] = NativeFunction(collectionProcessView)
	vt.Value["reverse"] = NativeFunction(bytesReverse)
	vt.Value["timingSafeEqual"] = NativeFunction(bytesTimingSafeEqual)
	vt.Value["bitLen"] = NativeFunction(bytesBitLen)
	vt.Value["fill"] = NativeFunction(bytesFill)
	vt.Value["concat"] = NativeFunction(bytesConcat)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[bytesT] = vt
}

func (ctx *Context) loadThreadVT() {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["run"] = NativeFunction(coRunThread)
	vt.Value["complete"] = NativeFunction(coCompleteThread)
	vt.Value["isActive"] = NativeFunction(coIsActive)
	vt.Value["isDone"] = NativeFunction(coIsDone)
	vt.Value["state"] = NativeFunction(coGetThreadState)
	vt.Value["value"] = NativeFunction(coValue)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[threadT] = vt
}

func (ctx *Context) loadColorVT() {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["string"] = NativeFunction(colorString)
	vt.Value["format"] = NativeFunction(colorFormat)
	vt.Value["bg"] = NativeFunction(colorSetBG)
	vt.Value["fg"] = NativeFunction(colorSetFG)
	vt.Value["reset"] = NativeFunction(colorSetReset)
	vt.Value["resets"] = NativeFunction(colorGetReset)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[colorT] = vt
}

func (ctx *Context) loadFileHandlerVT() {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["close"] = NativeFunction(fileClose)
	vt.Value["isClosed"] = NativeFunction(fileIsClosed)
	vt.Value["name"] = NativeFunction(fileName)
	vt.Value["write"] = NativeFunction(fileWrite)
	vt.Value["lines"] = NativeFunction(fileReadLines)
	vt.Value["read"] = NativeFunction(fileRead)
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[fileHandlerT] = vt
}

func (ctx *Context) loadHttpClientVT() {
	vt := &Object{Value: make(map[string]Value, 7)}
	vt.Value["get"] = NativeFunction(makeRequestFn(httpGET))
	vt.Value["post"] = NativeFunction(makeRequestFn(httpPOST))
	vt.Value["put"] = NativeFunction(makeRequestFn(httpPUT))
	vt.Value["delete"] = NativeFunction(makeRequestFn(httpDELETE))
	vt.Value["patch"] = NativeFunction(makeRequestFn(httpPATCH))
	vt.Value["head"] = NativeFunction(makeRequestFn(httpHEAD))
	vt.Value["options"] = NativeFunction(makeRequestFn(httpOPTIONS))
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[httpClientT] = vt
}

func (ctx *Context) loadUniversalVT() {
	vt := &Object{Value: make(map[string]Value, 6)}
	vt.Value["type"] = NativeFunction(coreType)
	vt.Value["clone"] = NativeFunction(coreClone)
	vt.Value["isError"] = NativeFunction(coreIsError)
	vt.Value["isNil"] = NativeFunction(coreIsNil)
	vt.Value["getvt"] = NativeFunction(coreGetVTable)
	vt.Value["extendvt"] = NativeFunction(coreExtendVTable)
	ctx.vtables[universalT] = vt
}

func (ctx *Context) loadBooleanVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[booleanT] = vt
}

func (ctx *Context) loadEnumVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[enumT] = vt
}

func (ctx *Context) loadErrorVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[errorT] = vt
}

func (ctx *Context) loadFunctionVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[functionT] = vt
}

func (ctx *Context) loadNativeFunctionVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[nativeFuncT] = vt
}

func (ctx *Context) loadNilVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[nilT] = vt
}

func (ctx *Context) loadIntegerVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[integerT] = vt
}

func (ctx *Context) loadFloatVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[floatT] = vt
}

func (ctx *Context) loadTimeVT() {
	vt := &Object{Value: make(map[string]Value)}
	vt.VTable = ctx.vtables[universalT].(*Object)
	ctx.vtables[timeT] = vt
}
