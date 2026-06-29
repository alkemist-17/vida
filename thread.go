package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type ThreadState int

const (
	Ready ThreadState = iota
	Running
	Suspended
	Waiting
	Done
)

const (
	minFrameSize = 8
	minStackSize = 64
	maxFrameSize = 1024
	maxStackSize = 1024
)

func (state ThreadState) String() string {
	switch state {
	case Ready:
		return "ready"
	case Running:
		return "running"
	case Suspended:
		return "suspended"
	case Waiting:
		return "waiting"
	case Done:
		return "done"
	default:
		return "unknown"
	}
}

type Thread struct {
	ReferenceSemanticsImpl
	Frames  []frame
	Stack   []Value
	Script  *Script
	Frame   *frame
	Invoker *Thread
	State   ThreadState
	Channel Value
	Reg     uint64
	fp      int
}

func newThread(script *Script) *Thread {
	return &Thread{
		Frames:  make([]frame, frameSize),
		Stack:   make([]Value, stacksize),
		Script:  script,
		State:   Running,
		Channel: Nil,
	}
}

func newInternalThread(fn *Function, script *Script) *Thread {
	return &Thread{
		Script: &Script{
			Konstants:    script.Konstants,
			GlobalStore:  script.GlobalStore,
			ErrorInfo:    script.ErrorInfo,
			MainFunction: fn,
		},
		Frames:  make([]frame, frameSize),
		Stack:   make([]Value, stacksize),
		Channel: Nil,
	}
}

func (th *Thread) Boolean() Bool {
	return True
}

func (th *Thread) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (th *Thread) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.OR):
		return th, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return False, nil
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (th *Thread) Equals(other Value) Bool {
	if val, ok := other.(*Thread); ok {
		return th == val
	}
	return false
}

func (th *Thread) String() string {
	return fmt.Sprintf("thread[%p, %v]", th, th.State.String())
}

func (th *Thread) ObjectKey() string {
	return fmt.Sprintf("thread[%p]", th)
}

func (th *Thread) Type() string {
	return threadT
}

func (th *Thread) Clone() Value {
	return coNewThreadWithSizeControl(th.Script.MainFunction, th.Script, Integer(len(th.Frames)), Integer(len(th.Stack)))
}

func (th *Thread) GetVTable(ctx *Context) Value {
	if ctx.vtables[threadT] == nil {
		ctx.loadThreadVT()
	}
	return ctx.vtables[threadT]
}

func (th *Thread) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[threadT] == nil {
		ctx.loadThreadVT()
	}
	if vtable, ok := ctx.vtables[threadT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (th *Thread) Reset(fn *Function) *Thread {
	th.State = Ready
	th.Script.MainFunction = fn
	th.Channel = Nil
	return th
}

type internalThreadPool struct {
	pool map[int]*Thread
	key  int
}

func newInternalThreadPool() *internalThreadPool {
	return &internalThreadPool{pool: make(map[int]*Thread, 10)}
}

func (p *internalThreadPool) get(fn *Function, script *Script) *Thread {
	if th, isFree := p.pool[p.key]; isFree {
		p.key++
		return th.Reset(fn)
	}
	th := newInternalThread(fn, script)
	p.pool[p.key] = th
	p.key++
	return th
}

func (p *internalThreadPool) release() {
	p.key--
}

func (vm *VM) runThread(fp, givenIP int, start bool, args ...Value) error {
	ip := givenIP
	var i, op, A, B, P uint64
	largs := len(args)
	if start {
		vm.fp = fp
		vm.Frame = &vm.Frames[vm.fp]
		vm.Frame.code = vm.Script.MainFunction.CoreFn.Code
		vm.Frame.lambda = vm.Script.MainFunction
		vm.Frame.stack = vm.Stack[:]
		copy(vm.Frame.stack[0:], args)
		switch vm.Frame.lambda.CoreFn.getConfigType() {
		case cZT:
			A := make([]Value, len(args))
			copy(A, args)
			vm.Frame.stack[0] = &Array{Value: A}
		case cNT:
			if largs == 0 {
				var k int
				for i := range vm.Script.MainFunction.CoreFn.Arity {
					k++
					vm.Frame.stack[i] = Nil
				}
				vm.Frame.stack[k] = &Array{}
			} else {
				var k int
				l := min(largs, vm.Script.MainFunction.CoreFn.Arity)
				for i := range l {
					k++
					vm.Frame.stack[i] = args[i]
				}
				if l < vm.Script.MainFunction.CoreFn.Arity {
					for i := l; i < vm.Script.MainFunction.CoreFn.Arity; i++ {
						k++
						vm.Frame.stack[i] = Nil
					}
				}
				if vm.Script.MainFunction.CoreFn.Arity < largs {
					r := largs - vm.Script.MainFunction.CoreFn.Arity
					var A = make([]Value, r)
					for i := range r {
						A[i] = vm.Frame.stack[k+i]
					}
					vm.Frame.stack[k] = &Array{Value: A}
				} else {
					vm.Frame.stack[k] = &Array{}
				}
			}
		case cNF:
			if largs == 0 {
				for i := range vm.Script.MainFunction.CoreFn.Arity {
					vm.Frame.stack[i] = Nil
				}
			} else if largs < vm.Script.MainFunction.CoreFn.Arity {
				var k int
				for i := range largs {
					k++
					vm.Frame.stack[i] = args[i]
				}
				for i := k; i < vm.Script.MainFunction.CoreFn.Arity; i++ {
					vm.Frame.stack[i] = Nil
				}
			}
		}
	} else {
		// Setting up the data to return to the suspended thread after suspension.
		if largs > 0 {
			vm.Frame.stack[vm.Reg] = args[0]
		} else {
			vm.Frame.stack[vm.Reg] = Nil
		}
	}
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
			if P == 0 && !vm.Frame.stack[A].Boolean() {
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
					val = vm.Frame.stack[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.stack[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.stack[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.stack[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			case storeFromKonst:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromGlobal:
					val = (*vm.Script.Konstants)[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			case storeFromGlobal:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			default:
				switch r {
				case storeFromLocal:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			}
			if s>>4 == 1 {
				val = !val
			}
			vm.Frame.stack[B] = val
		case prefix:
			val, err := vm.Frame.stack[A].Prefix(P)
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
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, vm.Frame.stack[A])
				}
			case storeFromKonst:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				}
			case storeFromGlobal:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				}
			default:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
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
				return vm.createError(ip, verror.ErrExpectedInteger)
			}
			if _, isInteger := vm.Frame.stack[B+1].(Integer); !isInteger {
				return vm.createError(ip, verror.ErrExpectedInteger)
			}
			if v, isInteger := vm.Frame.stack[B+2].(Integer); !isInteger {
				return vm.createError(ip, verror.ErrExpectedInteger)
			} else if v == 0 {
				return vm.createError(ip, verror.ErrExpectedIntegerDifferentFromZero)
			}
			ip = int(A)
		case iForSet:
			iterable := vm.Frame.stack[A]
			if !iterable.IsIterable() {
				return vm.createError(ip, verror.ErrValueNotIterable)
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
				return vm.createError(ip, verror.ErrValueNotCallable)
			}
			if fn, ok := val.(*Function); ok {
				if vm.fp >= len(vm.Frames) {
					return vm.createError(ip, verror.ErrStackOverflow)
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
							return vm.createError(ip, verror.ErrVariadicArgs)
						}
					case spreadLast:
						if xs, ok := vm.Frame.stack[int(B)+nargs].(*Array); ok && len(xs.Value) < len(vm.Frame.stack) {
							nargs += len(xs.Value) - 1
							for i, v := range xs.Value {
								vm.Frame.stack[int(B)+int(A)+i] = v
							}
						} else {
							return vm.createError(ip, verror.ErrVariadicArgs)
						}
					}
				}
				if fn.CoreFn.IsVarArg {
					if fn.CoreFn.Arity > nargs {
						return vm.createError(ip, verror.ErrNotEnoughArgs)
					}
					init := int(B) + 1 + fn.CoreFn.Arity
					count := nargs - fn.CoreFn.Arity
					xs := make([]Value, count)
					for i := range count {
						xs[i] = vm.Frame.stack[init+i]
					}
					vm.Frame.stack[init] = &Array{Value: xs}
				} else if nargs != fn.CoreFn.Arity {
					return vm.createError(ip, verror.ErrArity)
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
							return vm.createError(ip, verror.ErrVariadicArgs)
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
							return vm.createError(ip, verror.ErrVariadicArgs)
						}
					}
				}
				v, err := val.Call(vm.ctx, varargs...)
				if err != nil {
					switch err {
					case verror.ErrResumeThreadSignal:
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
					case verror.ErrStartThreadSignal:
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
					case verror.ErrSuspendThreadSignal:
						vm.Frame.ip = ip
						vm.Reg = B
						return nil
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
			if vm.fp == 0 {
				vm.Channel = val
				vm.State = Done
				return nil
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
			return verror.New(vm.Frame.lambda.CoreFn.ScriptID, message, verror.RunTimeErrType, 0)
		}
	}
}

func (vm *VM) debugThread(fp, givenIP int, start bool, args ...Value) error {
	ip := givenIP
	var i, op, A, B, P uint64
	largs := len(args)
	if start {
		vm.fp = fp
		vm.Frame = &vm.Frames[vm.fp]
		vm.Frame.code = vm.Script.MainFunction.CoreFn.Code
		vm.Frame.lambda = vm.Script.MainFunction
		vm.Frame.stack = vm.Stack[:]
		copy(vm.Frame.stack[0:], args)
		if vm.Script.MainFunction.CoreFn.Arity > largs {
			for i := largs; i < vm.Script.MainFunction.CoreFn.Arity; i++ {
				vm.Frame.stack[i] = Nil
			}
		}
	} else {
		if largs > 0 {
			vm.Frame.stack[vm.Reg] = args[0]
		} else if largs == 0 && vm.Script.MainFunction.CoreFn.Arity > largs {
			for i := 1; i <= vm.Script.MainFunction.CoreFn.Arity; i++ {
				if _, ok := vm.Frame.stack[i].(Iterator); !ok {
					vm.Frame.stack[i] = Nil
				}
			}
		} else if largs == 0 && vm.Script.MainFunction.CoreFn.Arity == 0 {
			vm.Frame.stack[0] = Nil
		}
	}
	for {
		vm.Inspect(ip)
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
			if P == 0 && !vm.Frame.stack[A].Boolean() {
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
					val = vm.Frame.stack[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.stack[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.stack[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.stack[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			case storeFromKonst:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromGlobal:
					val = (*vm.Script.Konstants)[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			case storeFromGlobal:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = (*vm.Script.GlobalStore)[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			default:
				switch r {
				case storeFromLocal:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals((*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Equals(vm.Frame.lambda.FreeVarStore[A])
				}
			}
			if s>>4 == 1 {
				val = !val
			}
			vm.Frame.stack[B] = val
		case prefix:
			val, err := vm.Frame.stack[A].Prefix(P)
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
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, vm.Frame.stack[A])
				}
			case storeFromKonst:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, (*vm.Script.Konstants)[A])
				}
			case storeFromGlobal:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				default:
					val = vm.Frame.lambda.FreeVarStore[P&clean16].Get(vm.ctx, (*vm.Script.GlobalStore)[A])
				}
			default:
				switch scopeIndexable {
				case storeFromLocal:
					val = vm.Frame.stack[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
				case storeFromGlobal:
					val = (*vm.Script.GlobalStore)[P&clean16].Get(vm.ctx, vm.Frame.lambda.FreeVarStore[A])
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
				return vm.createError(ip, verror.ErrExpectedInteger)
			}
			if _, isInteger := vm.Frame.stack[B+1].(Integer); !isInteger {
				return vm.createError(ip, verror.ErrExpectedInteger)
			}
			if v, isInteger := vm.Frame.stack[B+2].(Integer); !isInteger {
				return vm.createError(ip, verror.ErrExpectedInteger)
			} else if v == 0 {
				return vm.createError(ip, verror.ErrExpectedIntegerDifferentFromZero)
			}
			ip = int(A)
		case iForSet:
			iterable := vm.Frame.stack[A]
			if !iterable.IsIterable() {
				return vm.createError(ip, verror.ErrValueNotIterable)
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
				return vm.createError(ip, verror.ErrValueNotCallable)
			}
			if fn, ok := val.(*Function); ok {
				if vm.fp >= len(vm.Frames) {
					return vm.createError(ip, verror.ErrStackOverflow)
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
							return vm.createError(ip, verror.ErrVariadicArgs)
						}
					case spreadLast:
						if xs, ok := vm.Frame.stack[int(B)+nargs].(*Array); ok && len(xs.Value) < len(vm.Frame.stack) {
							nargs += len(xs.Value) - 1
							for i, v := range xs.Value {
								vm.Frame.stack[int(B)+int(A)+i] = v
							}
						} else {
							return vm.createError(ip, verror.ErrVariadicArgs)
						}
					}
				}
				if fn.CoreFn.IsVarArg {
					if fn.CoreFn.Arity > nargs {
						return vm.createError(ip, verror.ErrNotEnoughArgs)
					}
					init := int(B) + 1 + fn.CoreFn.Arity
					count := nargs - fn.CoreFn.Arity
					xs := make([]Value, count)
					for i := range count {
						xs[i] = vm.Frame.stack[init+i]
					}
					vm.Frame.stack[init] = &Array{Value: xs}
				} else if nargs != fn.CoreFn.Arity {
					return vm.createError(ip, verror.ErrArity)
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
							return vm.createError(ip, verror.ErrVariadicArgs)
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
							return vm.createError(ip, verror.ErrVariadicArgs)
						}
					}
				}
				v, err := val.Call(vm.ctx, varargs...)
				if err != nil {
					switch err {
					case verror.ErrResumeThreadSignal:
						threadError := vm.debugThread(vm.fp, vm.Frame.ip, false, varargs[1:]...)
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
					case verror.ErrStartThreadSignal:
						threadError := vm.debugThread(vm.fp, 0, true, varargs[1:]...)
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
					case verror.ErrSuspendThreadSignal:
						vm.Frame.ip = ip
						vm.Reg = B
						return nil
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
			if vm.fp == 0 {
				vm.Channel = val
				vm.State = Done
				return nil
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
			return verror.New(vm.Frame.lambda.CoreFn.ScriptID, message, verror.RunTimeErrType, 0)
		}
	}
}
