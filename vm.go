package vida

import (
	"fmt"
)

const frameSize = 1024
const stacksize = 1024

type frame struct {
	code   []uint64
	stack  []Value
	lambda *Function
	ip     int
	bp     int
	ret    int
}

type VM struct {
	*Thread
	ctx *Context
}

func (vm *VM) run() error {
	vm.Frame = &vm.Frames[vm.fp]
	vm.Frame.code = vm.Script.MainFunction.CoreFn.Code
	vm.Frame.lambda = vm.Script.MainFunction
	vm.Frame.stack = vm.Stack[:]
	ip := 1
	var i, op, A, B, P uint64
	for {
		i = vm.Frame.code[ip]
		op = i >> shift56
		A = i >> shift16 & clean16
		B = i & clean16
		P = i >> shift32 & clean24
		ip++
		switch op {
		case load:
			switch P {
			case loadFromLocal:
				vm.Frame.stack[B] = vm.Frame.stack[A]
			case loadFromKonst:
				vm.Frame.stack[B] = (*vm.Script.Konstants)[A]
			case loadFromGlobal:
				vm.Frame.stack[B] = (*vm.Script.GlobalStore)[A]
			default:
				vm.Frame.stack[B] = vm.Frame.lambda.FreeVarStore[A]
			}
		case store:
			switch P >> shift16 {
			case storeFromGlobal:
				switch P & clean16 {
				case storeFromLocal:
					(*vm.Script.GlobalStore)[B] = vm.Frame.stack[A]
				case storeFromKonst:
					(*vm.Script.GlobalStore)[B] = (*vm.Script.Konstants)[A]
				case storeFromGlobal:
					(*vm.Script.GlobalStore)[B] = (*vm.Script.GlobalStore)[A]
				default:
					(*vm.Script.GlobalStore)[B] = vm.Frame.lambda.FreeVarStore[A]
				}
			default:
				switch P & clean16 {
				case storeFromLocal:
					vm.Frame.lambda.FreeVarStore[B] = vm.Frame.stack[A]
				case storeFromKonst:
					vm.Frame.lambda.FreeVarStore[B] = (*vm.Script.Konstants)[A]
				case storeFromGlobal:
					vm.Frame.lambda.FreeVarStore[B] = (*vm.Script.GlobalStore)[A]
				default:
					vm.Frame.lambda.FreeVarStore[B] = vm.Frame.stack[A]
				}
			}
		case check:
			if vm.Frame.stack[A].Boolean() == (P == 1) {
				ip = int(B)
			}
		case jump:
			ip = int(B)
		case binopG:
			val, err := (*vm.Script.GlobalStore)[A].Binop(vm.ctx, P>>shift16, (*vm.Script.GlobalStore)[P&clean16])
			if err != nil {
				return vm.createError(ip, err)
			}
			vm.Frame.stack[B] = val
		case binop:
			val, err := vm.Frame.stack[A].Binop(vm.ctx, P>>shift16, vm.Frame.stack[P&clean16])
			if err != nil {
				return vm.createError(ip, err)
			}
			vm.Frame.stack[B] = val
		case binopK:
			val, err := vm.Frame.stack[P&clean16].Binop(vm.ctx, P>>shift16, (*vm.Script.Konstants)[A])
			if err != nil {
				return vm.createError(ip, err)
			}
			vm.Frame.stack[B] = val
		case binopQ:
			val, err := (*vm.Script.Konstants)[A].Binop(vm.ctx, P>>shift16, vm.Frame.stack[P&clean16])
			if err != nil {
				return vm.createError(ip, err)
			}
			vm.Frame.stack[B] = val
		case eq:
			var val Bool
			var s byte = byte(P >> shift16)
			l := s >> shift2 & clean2bits
			r := s & clean2bits
			switch l {
			case storeFromLocal:
				switch r {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Equals(vm.ctx, vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.stack[P&clean16].Equals(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.stack[P&clean16].Equals(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.stack[P&clean16].Equals(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				}
			case storeFromKonst:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.ctx, vm.Frame.stack[A])
				case storeFromGlobal:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				}
			case storeFromGlobal:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.ctx, vm.Frame.stack[A])
				case storeFromKonst:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				}
			default:
				switch r {
				case storeFromLocal:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.ctx, vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				}
			}
			if s>>4 == 1 {
				val = !val
			}
			vm.Frame.stack[B] = val
		case prefix:
			val, err := vm.Frame.stack[A].Prefix(vm.ctx, P)
			if err != nil {
				return vm.createError(ip, err)
			}
			vm.Frame.stack[B] = val
		case get:
			var val Value = Nil
			scopeIndex := P >> shift20
			scopeIndexable := (P >> shift16) & clean8
			switch scopeIndex {
			case storeFromLocal:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, vm.Frame.stack[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, vm.Frame.stack[A])
				case storeFromKonst:
					val = (*vm.Script.Konstants)[P&clean16].Get(vm.ctx, vm.Frame.stack[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, vm.Frame.stack[A])
				}
			case storeFromKonst:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromKonst:
					val = (*vm.Script.Konstants)[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				}
			case storeFromGlobal:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				case storeFromKonst:
					val = (*vm.Script.Konstants)[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				}
			default:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				case storeFromKonst:
					val = (*vm.Script.Konstants)[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				}
			}
			vm.Frame.stack[B] = val
		case set:
			var err error
			scopeIdx := P >> shift20
			scopeExp := (P >> shift16) & clean8
			switch scopeIdx {
			case storeFromLocal:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.stack[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.stack[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.stack[A], (*vm.Script.GlobalStore)[B])
				default:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.stack[A], vm.Frame.lambda.FreeVarStore[B])
				}
			case storeFromKonst:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.Konstants)[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.Konstants)[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.Konstants)[A], (*vm.Script.GlobalStore)[B])
				default:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.Konstants)[A], vm.Frame.lambda.FreeVarStore[B])
				}
			case storeFromGlobal:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.GlobalStore)[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.GlobalStore)[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.GlobalStore)[A], (*vm.Script.GlobalStore)[B])
				default:
					err = vm.Frame.stack[P&clean16].Set((*vm.Script.GlobalStore)[A], vm.Frame.lambda.FreeVarStore[B])
				}
			default:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.lambda.FreeVarStore[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.lambda.FreeVarStore[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.lambda.FreeVarStore[A], (*vm.Script.GlobalStore)[B])
				default:
					err = vm.Frame.stack[P&clean16].Set(vm.Frame.lambda.FreeVarStore[A], vm.Frame.lambda.FreeVarStore[B])
				}
			}
			if err != nil {
				return vm.createError(ip, err)
			}
		case lookup:
			vm.Frame.stack[B] = vm.Frame.stack[P&clean16].LookUp(vm.ctx, (*vm.Script.Konstants)[A])
		case slice:
			val, err := vm.processSlice(P, A)
			if err != nil {
				return vm.createError(ip, err)
			}
			vm.Frame.stack[B] = val
		case array:
			xs := make([]Value, P)
			F := A
			for i := 0; i < int(P); i++ {
				xs[i] = vm.Frame.stack[F]
				F++
			}
			vm.Frame.stack[B] = &Array{Value: xs}
		case object:
			vm.Frame.stack[B] = &Object{Value: make(map[string]Value)}
		case forSet:
			if _, isInteger := vm.Frame.stack[B].(Integer); !isInteger {
				return vm.createError(ip, ErrExpectedInteger)
			}
			if _, isInteger := vm.Frame.stack[B+1].(Integer); !isInteger {
				return vm.createError(ip, ErrExpectedInteger)
			}
			if v, isInteger := vm.Frame.stack[B+2].(Integer); !isInteger {
				return vm.createError(ip, ErrExpectedInteger)
			} else if v == 0 {
				return vm.createError(ip, ErrExpectedIntegerDifferentFromZero)
			}
			ip = int(A)
		case iForSet:
			iterable := vm.Frame.stack[A]
			if !iterable.IsIterable() {
				return vm.createError(ip, ErrValueNotIterable)
			}
			vm.Frame.stack[B] = iterable.Iterator()
			ip = int(P)
		case forLoop:
			i := vm.Frame.stack[B].(Integer)
			e := vm.Frame.stack[B+1].(Integer)
			s := vm.Frame.stack[B+2].(Integer)
			if s > 0 {
				if i < e {
					vm.Frame.stack[B+3] = i
					i += s
					vm.Frame.stack[B] = i
					ip = int(A)
				}
			} else {
				if i > e {
					vm.Frame.stack[B+3] = i
					i += s
					vm.Frame.stack[B] = i
					ip = int(A)
				}
			}
		case iForLoop:
			i, _ := vm.Frame.stack[B].(Iterator)
			if i.Next() {
				vm.Frame.stack[B+1] = i.Key(vm.ctx)
				vm.Frame.stack[B+2] = i.Value(vm.ctx)
				ip = int(A)
				continue
			}
		case fun:
			fn := &Function{CoreFn: (*vm.Script.Konstants)[A].(*CoreFunction)}
			if fn.CoreFn.FreeVarsCount > 0 {
				vm.Frame.stack[B] = fn
				freeVarStore := make([]Value, 0, fn.CoreFn.FreeVarsCount)
				for i := 0; i < fn.CoreFn.FreeVarsCount; i++ {
					if fn.CoreFn.FreeVarsInfo[i].IsLocal {
						freeVarStore = append(freeVarStore, vm.Frame.stack[fn.CoreFn.FreeVarsInfo[i].Index])
					} else {
						freeVarStore = append(freeVarStore, vm.Frame.lambda.FreeVarStore[fn.CoreFn.FreeVarsInfo[i].Index])
					}
				}
				fn.FreeVarStore = freeVarStore
			}
			vm.Frame.stack[B] = fn
		case call:
			val := vm.Frame.stack[B]
			nargs := int(A)
			F := P >> shift16
			P = P & clean16
			if !val.IsCallable() {
				return vm.createError(ip, ErrValueNotCallable)
			}
			if fn, ok := val.(*Function); ok {
				if vm.fp >= frameSize {
					return vm.createError(ip, ErrStackOverflow)
				}
				if P != 0 {
					switch P {
					case spreadFirst:
						if xs, ok := vm.Frame.stack[B+F].(*Array); ok && len(xs.Value) < len(vm.Frame.stack) {
							nargs = len(xs.Value) + int(F) - 1
							for i, v := range xs.Value {
								vm.Frame.stack[int(B)+int(F)+i] = v
							}
						} else {
							return vm.createError(ip, ErrVariadicArgs)
						}
					case spreadLast:
						if xs, ok := vm.Frame.stack[int(B)+nargs].(*Array); ok && len(xs.Value) < len(vm.Frame.stack) {
							nargs += len(xs.Value) - 1
							for i, v := range xs.Value {
								vm.Frame.stack[int(B)+int(A)+i] = v
							}
						} else {
							return vm.createError(ip, ErrVariadicArgs)
						}
					}
				}
				if fn.CoreFn.IsVarArg {
					if fn.CoreFn.Arity > nargs {
						return vm.createError(ip, ErrNotEnoughArgs)
					}
					init := int(B) + 1 + fn.CoreFn.Arity
					count := nargs - fn.CoreFn.Arity
					xs := make([]Value, count)
					for i := range count {
						xs[i] = vm.Frame.stack[init+i]
					}
					vm.Frame.stack[init] = &Array{Value: xs}
				} else if nargs != fn.CoreFn.Arity {
					return vm.createError(ip, ErrArity)
				}
				if fn == vm.Frame.lambda && vm.Frame.code[ip]>>shift56 == ret {
					for i := 0; i < nargs; i++ {
						vm.Frame.stack[i] = vm.Frame.stack[int(B)+1+i]
					}
					ip = 0
					continue
				}
				vm.Frame.ip = ip
				vm.Frame.ret = int(B)
				bs := vm.Frame.bp
				vm.fp++
				vm.Frame = &vm.Frames[vm.fp]
				vm.Frame.lambda = fn
				vm.Frame.bp = bs + int(B) + 1
				vm.Frame.code = fn.CoreFn.Code
				vm.Frame.stack = vm.Stack[vm.Frame.bp:]
				ip = 0
			} else {
				varargs := vm.Frame.stack[B+1 : B+A+1]
				if P != 0 {
				nonnecessary:
					switch P {
					case spreadFirst:
						if arr, ok := varargs[0].(*Array); ok {
							for i := 0; i < len(arr.Value); i++ {
								vm.Frame.stack[int(B)+1+i] = arr.Value[i]
							}
							varargs = vm.Frame.stack[B+1 : int(B)+int(A)+len(arr.Value)]
						} else {
							return vm.createError(ip, ErrVariadicArgs)
						}
					case spreadLast:
						if arr, ok := varargs[len(varargs)-1].(*Array); ok {
							if len(arr.Value) == 0 {
								break nonnecessary
							}
							for i, v := range arr.Value {
								vm.Frame.stack[int(B)+len(varargs)+i] = v
							}
							varargs = vm.Frame.stack[B+1 : B+A+uint64(len(arr.Value))]
						} else {
							return vm.createError(ip, ErrVariadicArgs)
						}
					}
				}
				v, err := val.Call(vm.ctx, varargs...)
				if err != nil {
					switch err {
					case ErrResumeThreadSignal:
						threadError := vm.runThread(vm.fp, vm.Frame.ip, false, varargs[1:]...)
						if threadError != nil {
							return vm.createError(ip, threadError)
						}
						switch vm.State {
						case Done, Suspended:
							v = vm.Channel
							invoker := vm.Thread.Invoker
							invoker.State = Running
							vm.Thread.Invoker = nil
							vm.ctx.currentThread = invoker
							vm.Thread = invoker
						}
					case ErrStartThreadSignal:
						threadError := vm.runThread(vm.fp, 0, true, varargs[1:]...)
						if threadError != nil {
							return vm.createError(ip, threadError)
						}
						switch vm.State {
						case Done, Suspended:
							v = vm.Channel
							invoker := vm.Thread.Invoker
							invoker.State = Running
							vm.Thread.Invoker = nil
							vm.ctx.currentThread = invoker
							vm.Thread = invoker
						}
					default:
						return vm.createError(ip, err)
					}
				}
				vm.Frame.stack[B] = v
			}
		case ret:
			var val Value
			switch B {
			case storeFromLocal:
				val = vm.Frame.stack[A]
			case storeFromKonst:
				val = (*vm.Script.Konstants)[A]
			case storeFromGlobal:
				val = (*vm.Script.GlobalStore)[A]
			default:
				val = vm.Frame.lambda.FreeVarStore[A]
			}
			vm.fp--
			vm.Frame = &vm.Frames[vm.fp]
			ip = vm.Frame.ip
			vm.Frame.stack = vm.Stack[vm.Frame.bp:]
			vm.Frame.stack[vm.Frame.ret] = val
		case end:
			return nil
		default:
			message := fmt.Sprintf("unknown opcode %v", op)
			return NewRuntimeError(vm.Frame.lambda.CoreFn.ScriptID, message, RunTimeErrType, 0)
		}
	}
}

func sliceBounds(start, end, length Integer) (Integer, Integer, bool) {
	if start < 0 {
		start += length
	}
	if end < 0 {
		end += length
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		return 0, 0, true
	}
	if start > length {
		return 0, 0, true
	}
	return start, end, false
}

func (vm *VM) processSlice(mode, sliceable uint64) (Value, error) {
	val := vm.Frame.stack[sliceable]
	var startIdx, endIdx Integer
	var hasStart, hasEnd bool

	if mode == ecv || mode == ece {
		e := vm.Frame.stack[sliceable+1]
		i, ok := e.(Integer)
		if !ok {
			return Nil, ErrSlice
		}
		startIdx, hasStart = i, true
	}
	if mode == vce {
		e := vm.Frame.stack[sliceable+1]
		i, ok := e.(Integer)
		if !ok {
			return Nil, ErrSlice
		}
		endIdx, hasEnd = i, true
	}
	if mode == ece {
		e := vm.Frame.stack[sliceable+2]
		i, ok := e.(Integer)
		if !ok {
			return Nil, ErrSlice
		}
		endIdx, hasEnd = i, true
	}

	resolve := func(length Integer) (Integer, Integer, bool) {
		s := Integer(0)
		if hasStart {
			s = startIdx
		}
		e := length
		if hasEnd {
			e = endIdx
		}
		return sliceBounds(s, e, length)
	}

	switch v := val.(type) {
	case *Array:
		length := Integer(len(v.Value))
		if mode == vcv {
			data := make([]Value, length)
			copy(data, v.Value)
			return &Array{Value: data}, nil
		}
		s, e, empty := resolve(length)
		if empty {
			return &Array{}, nil
		}
		slc := v.Value[s:e]
		data := make([]Value, len(slc))
		copy(data, slc)
		return &Array{Value: data}, nil

	case *String:
		if v.Runes == nil {
			v.Runes = []rune(v.Value)
		}
		length := Integer(len(v.Runes))
		if mode == vcv {
			return v, nil
		}
		s, e, empty := resolve(length)
		if empty {
			return &String{}, nil
		}
		return &String{Value: string(v.Runes[s:e])}, nil

	case *Bytes:
		length := Integer(len(v.Value))
		if mode == vcv {
			data := make([]byte, length)
			copy(data, v.Value)
			return &Bytes{Value: data}, nil
		}
		s, e, empty := resolve(length)
		if empty {
			return &Bytes{}, nil
		}
		slc := v.Value[s:e]
		data := make([]byte, len(slc))
		copy(data, slc)
		return &Bytes{Value: data}, nil
	}

	return Nil, ErrSlice
}

func (vm *VM) printCallStack() {
	fmt.Printf("\t[Call Stack]\n")
	fmt.Printf("\t(Most recent call first)\n\n\n\n")
	for i := vm.fp; i >= 0; i-- {
		modName := vm.Frames[i].lambda.CoreFn.ScriptID
		ip := vm.Frames[i].ip
		fn := vm.Frames[i].lambda
		var nearLine uint
		if fn.CoreFn.MapScriptIPLine[modName][ip] == 0 {
			nearLine = getNonZeroLine(fn, modName, ip)
		} else {
			nearLine = fn.CoreFn.MapScriptIPLine[modName][ip]
		}
		err := NewStackFrameInfo(modName, nearLine, uint(i))
		fmt.Printf("%v\n\n\n", err)
	}
}

func (vm *VM) createError(ip int, err error) error {
	vm.Thread.State = Done
	modName := vm.Frame.lambda.CoreFn.ScriptID
	fn := vm.Frame.lambda
	vm.Frame.ip = ip
	var nearLine uint
	if fn.CoreFn.MapScriptIPLine[modName][ip] == 0 {
		nearLine = getNonZeroLine(fn, modName, ip)
	} else {
		nearLine = fn.CoreFn.MapScriptIPLine[modName][ip]
	}
	return NewRuntimeError(modName, err.Error(), RunTimeErrType, nearLine)
}

func getNonZeroLine(fn *Function, modName string, ip int) uint {
	var nearLine uint
	for i := ip; i >= 0; i-- {
		nearLine = fn.CoreFn.MapScriptIPLine[modName][i]
		if nearLine != 0 {
			break
		}
	}
	return nearLine
}

func checkISACompatibility(script *Script) error {
	majorFromCode := (script.MainFunction.CoreFn.Code[0] >> 24) & 255
	if majorFromCode == major {
		return nil
	}
	return NewRuntimeError(script.MainFunction.CoreFn.ScriptID, "script compiled with an uncompatible interpreter version", FileErrType, 0)
}
