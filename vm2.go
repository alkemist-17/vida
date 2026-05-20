package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/verror"
)

func (vm *VM) runv2() (Result, error) {
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
				vm.Frame.stack[B] = (*vm.Script.Store)[A]
			default:
				vm.Frame.stack[B] = vm.Frame.lambda.Free[A]
			}
		case store:
			switch P >> shift16 {
			case storeFromGlobal:
				switch P & clean16 {
				case storeFromLocal:
					(*vm.Script.Store)[B] = vm.Frame.stack[A]
				case storeFromKonst:
					(*vm.Script.Store)[B] = (*vm.Script.Konstants)[A]
				case storeFromGlobal:
					(*vm.Script.Store)[B] = (*vm.Script.Store)[A]
				default:
					(*vm.Script.Store)[B] = vm.Frame.lambda.Free[A]
				}
			default:
				switch P & clean16 {
				case storeFromLocal:
					vm.Frame.lambda.Free[B] = vm.Frame.stack[A]
				case storeFromKonst:
					vm.Frame.lambda.Free[B] = (*vm.Script.Konstants)[A]
				case storeFromGlobal:
					vm.Frame.lambda.Free[B] = (*vm.Script.Store)[A]
				default:
					vm.Frame.lambda.Free[B] = vm.Frame.stack[A]
				}
			}
		case check:
			if P == 0 && !vm.Frame.stack[A].Boolean() {
				ip = int(B)
			}
		case jump:
			ip = int(B)
		case binopG:
			val, err := (*vm.Script.Store)[A].Binop(P>>shift16, (*vm.Script.Store)[P&clean16])
			if err != nil {
				return Failure, err
			}
			vm.Frame.stack[B] = val
		case binop:
			val, err := vm.Frame.stack[A].Binop(P>>shift16, vm.Frame.stack[P&clean16])
			if err != nil {
				return Failure, err
			}
			vm.Frame.stack[B] = val
		case binopK:
			val, err := vm.Frame.stack[P&clean16].Binop(P>>shift16, (*vm.Script.Konstants)[A])
			if err != nil {
				return Failure, err
			}
			vm.Frame.stack[B] = val
		case binopQ:
			val, err := (*vm.Script.Konstants)[A].Binop(P>>shift16, vm.Frame.stack[P&clean16])
			if err != nil {
				return Failure, err
			}
			vm.Frame.stack[B] = val
		case eq:
			var val bool
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
					val = vm.Frame.stack[P&clean16].Equals((*vm.Script.Store)[A])
				default:
					val = vm.Frame.stack[P&clean16].Equals(vm.Frame.lambda.Free[A])
				}
			case storeFromKonst:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromGlobal:
					val = (*vm.Script.Konstants)[P&clean16].Equals((*vm.Script.Store)[A])
				default:
					val = (*vm.Script.Konstants)[P&clean16].Equals(vm.Frame.lambda.Free[A])
				}
			case storeFromGlobal:
				switch r {
				case storeFromLocal:
					val = (*vm.Script.Store)[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = (*vm.Script.Store)[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = (*vm.Script.Store)[P&clean16].Equals((*vm.Script.Store)[A])
				default:
					val = (*vm.Script.Store)[P&clean16].Equals(vm.Frame.lambda.Free[A])
				}
			default:
				switch r {
				case storeFromLocal:
					val = vm.Frame.lambda.Free[P&clean16].Equals(vm.Frame.stack[A])
				case storeFromKonst:
					val = vm.Frame.lambda.Free[P&clean16].Equals((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val = vm.Frame.lambda.Free[P&clean16].Equals((*vm.Script.Store)[A])
				default:
					val = vm.Frame.lambda.Free[P&clean16].Equals(vm.Frame.lambda.Free[A])
				}
			}
			if s>>4 == 1 {
				val = !val
			}
			vm.Frame.stack[B] = BoolVal(val)
		case prefix:
			val, err := vm.Frame.stack[A].Prefix(P)
			if err != nil {
				return Failure, err
			}
			vm.Frame.stack[B] = val
		case iGet:
			var val Value
			var err error
			scopeIndex := P >> shift20
			scopeIndexable := (P >> shift16) & clean8
			switch scopeIndex {
			case storeFromLocal:
				switch scopeIndexable {
				case storeFromLocal:
					val, err = vm.Frame.stack[P&clean16].IGet(vm.Frame.stack[A])
				case storeFromGlobal:
					val, err = (*vm.Script.Store)[P&clean16].IGet(vm.Frame.stack[A])
				default:
					val, err = vm.Frame.lambda.Free[P&clean16].IGet(vm.Frame.stack[A])
				}
			case storeFromKonst:
				switch scopeIndexable {
				case storeFromLocal:
					val, err = vm.Frame.stack[P&clean16].IGet((*vm.Script.Konstants)[A])
				case storeFromGlobal:
					val, err = (*vm.Script.Store)[P&clean16].IGet((*vm.Script.Konstants)[A])
				default:
					val, err = vm.Frame.lambda.Free[P&clean16].IGet((*vm.Script.Konstants)[A])
				}
			case storeFromGlobal:
				switch scopeIndexable {
				case storeFromLocal:
					val, err = vm.Frame.stack[P&clean16].IGet((*vm.Script.Store)[A])
				case storeFromGlobal:
					val, err = (*vm.Script.Store)[P&clean16].IGet((*vm.Script.Store)[A])
				default:
					val, err = vm.Frame.lambda.Free[P&clean16].IGet((*vm.Script.Store)[A])
				}
			default:
				switch scopeIndexable {
				case storeFromLocal:
					val, err = vm.Frame.stack[P&clean16].IGet(vm.Frame.lambda.Free[A])
				case storeFromGlobal:
					val, err = (*vm.Script.Store)[P&clean16].IGet(vm.Frame.lambda.Free[A])
				default:
					val, err = vm.Frame.lambda.Free[P&clean16].IGet(vm.Frame.lambda.Free[A])
				}
			}
			if err != nil {
				return Failure, err
			}
			vm.Frame.stack[B] = val
		case iSet:
			var err error
			scopeIdx := P >> shift20
			scopeExp := (P >> shift16) & clean8
			switch scopeIdx {
			case storeFromLocal:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.stack[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.stack[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.stack[A], (*vm.Script.Store)[B])
				default:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.stack[A], vm.Frame.lambda.Free[B])
				}
			case storeFromKonst:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Konstants)[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Konstants)[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Konstants)[A], (*vm.Script.Store)[B])
				default:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Konstants)[A], vm.Frame.lambda.Free[B])
				}
			case storeFromGlobal:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Store)[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Store)[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Store)[A], (*vm.Script.Store)[B])
				default:
					err = vm.Frame.stack[P&clean16].ISet((*vm.Script.Store)[A], vm.Frame.lambda.Free[B])
				}
			default:
				switch scopeExp {
				case storeFromLocal:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.lambda.Free[A], vm.Frame.stack[B])
				case storeFromKonst:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.lambda.Free[A], (*vm.Script.Konstants)[B])
				case storeFromGlobal:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.lambda.Free[A], (*vm.Script.Store)[B])
				default:
					err = vm.Frame.stack[P&clean16].ISet(vm.Frame.lambda.Free[A], vm.Frame.lambda.Free[B])
				}
			}
			if err != nil {
				return Failure, err
			}
		case slice:
			vm.Frame.stack[B] = NilVal()
		case array:
			xs := make([]Value, P)
			F := A
			for i := 0; i < int(P); i++ {
				xs[i] = vm.Frame.stack[F]
				F++
			}
			vm.Frame.stack[B] = ArrayVal(&Array{Value: xs})
		case object:
			vm.Frame.stack[B] = ObjectVal(&Object{Value: make(map[string]Value)})
		case forSet:
			ip = int(A)
		case iForSet:
			iterable := vm.Frame.stack[A]
			vm.Frame.stack[B] = iterable.Iterator()
			ip = int(P)
		case forLoop:
			i := vm.Frame.stack[B].ival
			e := vm.Frame.stack[B+1].ival
			s := vm.Frame.stack[B+2].ival
			if s > 0 {
				if i < e {
					vm.Frame.stack[B+3] = IntVal(i)
					i += s
					vm.Frame.stack[B] = IntVal(i)
					ip = int(A)
				}
			} else {
				if i > e {
					vm.Frame.stack[B+3] = IntVal(i)
					i += s
					vm.Frame.stack[B] = IntVal(i)
					ip = int(A)
				}
			}
		case iForLoop:
			continue
		case fun:
			fn := &Function{CoreFn: (*vm.Script.Konstants)[A].CoreFn()}
			if fn.CoreFn.Free > 0 {
				vm.Frame.stack[B] = FunctionVal(fn)
				var free []Value
				for i := 0; i < fn.CoreFn.Free; i++ {
					if fn.CoreFn.Info[i].IsLocal {
						free = append(free, vm.Frame.stack[fn.CoreFn.Info[i].Index])
					} else {
						free = append(free, vm.Frame.lambda.Free[fn.CoreFn.Info[i].Index])
					}
				}
				fn.Free = free
			}
			vm.Frame.stack[B] = FunctionVal(fn)
		case call:
			val := vm.Frame.stack[B]
			nargs := int(A)
			P = P & clean16
			if !val.IsCallable() {
				return Failure, verror.ErrValueNotCallable
			}
			if val.ttype == TFunction {
				fn := val.Fn()
				if vm.fp >= frameSize {
					return Failure, verror.ErrStackOverflow
				}
				if P != 0 {
					switch P {
					case ellipsisFirst:
						return Failure, verror.ErrValueNotCallable
					case ellipsisLast:
						return Failure, verror.ErrValueNotCallable
					}
				}
				if fn.CoreFn.IsVar {
					if fn.CoreFn.Arity > nargs {
						return Failure, verror.ErrValueNotCallable
					}
					init := int(B) + 1 + fn.CoreFn.Arity
					count := nargs - fn.CoreFn.Arity
					xs := make([]Value, count)
					for i := 0; i < count; i++ {
						xs[i] = vm.Frame.stack[init+i]
					}
				} else if nargs != fn.CoreFn.Arity {
					return Failure, verror.ErrValueNotCallable

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
				v, _ := val.GFunction()(varargs...)
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
				val = (*vm.Script.Store)[A]
			default:
				val = vm.Frame.lambda.Free[A]
			}
			vm.fp--
			vm.Frame = &vm.Frames[vm.fp]
			ip = vm.Frame.ip
			vm.Frame.stack = vm.Stack[vm.Frame.bp:]
			vm.Frame.stack[vm.Frame.ret] = val
		case end:
			return Success, nil
		default:
			message := fmt.Sprintf("unknown opcode %v", op)
			return Failure, verror.New(vm.Frame.lambda.CoreFn.ScriptName, message, verror.RunTimeErrType, 0)
		}
	}
}
