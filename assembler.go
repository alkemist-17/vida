package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/token"
)

const (
	opLoad     = uint64(load) << shift56
	opStore    = uint64(store) << shift56
	opPrefix   = uint64(prefix) << shift56
	opBinop    = uint64(binop) << shift56
	opBinopG   = uint64(binopG) << shift56
	opBinopK   = uint64(binopK) << shift56
	opBinopQ   = uint64(binopQ) << shift56
	opEq       = uint64(eq) << shift56
	opArray    = uint64(array) << shift56
	opObject   = uint64(object) << shift56
	opIGet     = uint64(iGet) << shift56
	opISet     = uint64(iSet) << shift56
	opSlice    = uint64(slice) << shift56
	opForSet   = uint64(forSet) << shift56
	opForLoop  = uint64(forLoop) << shift56
	opIForSet  = uint64(iForSet) << shift56
	opIForLoop = uint64(iForLoop) << shift56
	opJump     = uint64(jump) << shift56
	opCheck    = uint64(check) << shift56
	opFun      = uint64(fun) << shift56
	opCall     = uint64(call) << shift56
	opRet      = uint64(ret) << shift56
	opHeader   = uint64(header)
	opEnd      = uint64(end)
)

const (
	rKonst = iota
	rLoc
	rGlob
	rFree
	rNotDefined
)

const (
	storeFromLocal = iota
	storeFromKonst
	storeFromGlobal
	storeFromFree
)

const (
	loadFromLocal = iota
	loadFromKonst
	loadFromGlobal
	loadFromFree
)

const (
	vcv = 2
	vce = 3
	ecv = 6
	ece = 7
)

const (
	shift2     = 2
	shift4     = 4
	shift16    = 16
	shift20    = 20
	shift32    = 32
	shift48    = 48
	shift56    = 56
	clean2bits = 0b00000011
	clean8     = 0x000000000000000F
	clean16    = 0x000000000000FFFF
	clean24    = 0x0000000000FFFFFF
)

func (c *compiler) appendHeader() {
	c.currentFn.Code = append(c.currentFn.Code, opHeader)
}

func (c *compiler) appendEnd() {
	c.currentFn.Code = append(c.currentFn.Code, opEnd)
}

func (c *compiler) emitLoad(from, to, kind int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(from)<<shift16)|
			(uint64(kind)<<shift32)|
			opLoad)
}

func (c *compiler) emitStore(from, to, src, dst int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(from)<<shift16)|
			(uint64(src)<<shift32)|
			(uint64(dst)<<shift48)|
			opStore)
}

func (c *compiler) emitPrefix(from, to int, operator token.Token) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(from)<<shift16)|
			(uint64(operator)<<shift32)|
			opPrefix)
}

func (c *compiler) emitBinop(lidx, ridx, to int, operator token.Token) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(lidx)<<shift16)|
			(uint64(ridx)<<shift32)|
			(uint64(operator)<<shift48)|
			opBinop)
}

func (c *compiler) emitBinopG(lidx, ridx, to int, operator token.Token) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(lidx)<<shift16)|
			(uint64(ridx)<<shift32)|
			(uint64(operator)<<shift48)|
			opBinopG)
}

func (c *compiler) emitBinopK(kidx, regAddr, to int, operator token.Token) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(kidx)<<shift16)|
			(uint64(regAddr)<<shift32)|
			(uint64(operator)<<shift48)|
			opBinopK)
}

func (c *compiler) emitBinopQ(kidx, regAddr, to int, operator token.Token) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(kidx)<<shift16)|
			(uint64(regAddr)<<shift32)|
			(uint64(operator)<<shift48)|
			opBinopQ)
}

func (c *compiler) emitSuperEq(lidx, ridx, to, scopeLeft, scopeRight int, operator token.Token) {
	var flags byte = byte(scopeRight) | (byte(scopeLeft) << shift2)
	if operator == token.NEQ {
		flags |= 1 << shift4
	}
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(ridx)<<shift16)|
			(uint64(lidx)<<shift32)|
			(uint64(flags)<<shift48)|
			opEq)
}

func (c *compiler) emitArray(length, root, to int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(root)<<shift16)|
			(uint64(length)<<shift32)|
			opArray)
}

func (c *compiler) emitObject(to int) {
	c.currentFn.Code = append(c.currentFn.Code, uint64(to)|opObject)
}

func (c *compiler) emitIGet(indexable, index, to, scopeIndex, scopeIndexable int) {
	var flags byte = byte(scopeIndexable) | (byte(scopeIndex) << shift4)
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(index)<<shift16)|
			(uint64(indexable)<<shift32)|
			(uint64(flags)<<shift48)|
			opIGet)
}

func (c *compiler) emitISet(indexable, index, expr, scopeIdx, scopeExpr int) {
	var flags byte = byte(scopeExpr) | (byte(scopeIdx) << shift4)
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(expr)|
			(uint64(index)<<shift16)|
			(uint64(indexable)<<shift32)|
			(uint64(flags)<<shift48)|
			opISet)
}

func (c *compiler) emitSlice(mode, sliceable, to int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(sliceable)<<shift16)|
			(uint64(mode)<<shift32)|
			opSlice)
}

func (c *compiler) emitForSet(iReg, loop int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(iReg)|
			(uint64(loop)<<shift16)|
			opForSet)
}

func (c *compiler) emitForLoop(iReg, loop int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(iReg)|
			(uint64(loop)<<shift16)|
			opForLoop)
}

func (c *compiler) emitIForSet(loop, iterable, ireg int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(ireg)|
			(uint64(iterable)<<shift16)|
			(uint64(loop)<<shift32)|
			opIForSet)
}

func (c *compiler) emitIForLoop(iReg, loop int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(iReg)|
			(uint64(loop)<<shift16)|
			opIForLoop)
}

func (c *compiler) emitJump(to int) {
	c.currentFn.Code = append(c.currentFn.Code, uint64(to)|opJump)
}

func (c *compiler) emitCheck(against, reg, jump int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(jump)|
			(uint64(reg)<<shift16)|
			(uint64(against)<<shift32)|
			opCheck)
}

func (c *compiler) emitFun(from, to int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(to)|
			(uint64(from)<<shift16)|
			opFun)
}

func (c *compiler) emitCall(callable, argCount, ellipsis, firstArg int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(callable)|
			(uint64(argCount)<<shift16)|
			(uint64(ellipsis)<<shift32)|
			(uint64(firstArg)<<shift48)|
			opCall)
}

func (c *compiler) emitRet(source, index int) {
	c.currentFn.Code = append(c.currentFn.Code,
		uint64(source)|
			(uint64(index)<<shift16)|
			opRet)
}

func (c *compiler) refScope(id string) (int, int) {
	if to, isLocal, key := c.sb.isLocal(id); isLocal {
		if key.level != c.level {
			fn := c.fn[c.level]
			for i := 0; i < len(fn.Info); i++ {
				if fn.Info[i].Id == id {
					return i, rFree
				}
			}
			fn.Free++
			if key.level+1 == c.level {
				fn.Info = append(fn.Info, freeInfo{Index: int(to), IsLocal: true, Id: key.id})
			} else {
				for i := key.level + 1; i < c.level; i++ {
					if i == key.level+1 {
						c.fn[i].Free++
						c.fn[i].Info = append(c.fn[i].Info, freeInfo{Index: int(to), IsLocal: true, Id: key.id})
					} else {
						idx := len(c.fn[i-1].Info) - 1
						c.fn[i].Info = append(c.fn[i].Info, freeInfo{Index: idx, IsLocal: false, Id: key.id})
						c.fn[i].Free++
					}
				}
				fn.Info = append(fn.Info, freeInfo{Index: len(c.fn[c.level-1].Info) - 1, IsLocal: false, Id: key.id})
			}
			return len(fn.Info) - 1, rFree
		}
		return to, rLoc
	}
	if idx, isGlobal := c.sb.isGlobal(id); isGlobal {
		return idx, rGlob
	}
	c.hadError = true
	return 0, rNotDefined
}

func (c *compiler) generateReferenceError(ref string, line uint) {
	c.lineErr = line
	c.errMsg = fmt.Sprintf("reference '%v' not found", ref)
}
