package vida

import (
	"fmt"

	"github.com/alkemist-17/vida/token"
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
	shift24    = 24
	shift32    = 32
	shift48    = 48
	shift56    = 56
	clean2bits = 0b00000011
	clean8     = 0x000000000000000F
	clean16    = 0x000000000000FFFF
	clean24    = 0x0000000000FFFFFF
)

func (c *compiler) appendHeader() {
	c.currentFn.Code = append(c.currentFn.Code, header)
}

func (c *compiler) appendEnd() {
	c.currentFn.Code = append(c.currentFn.Code, end)
}

func (c *compiler) emitLoad(from, to, kind int) {
	var i uint64 = uint64(to)
	i |= uint64(from) << shift16
	i |= uint64(kind) << shift32
	i |= load << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitStore(from, to, src, dst int) {
	var i uint64 = uint64(to)
	i |= uint64(from) << shift16
	i |= uint64(src) << shift32
	i |= uint64(dst) << shift48
	i |= store << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitPrefix(from, to int, operator token.Token) {
	var i uint64 = uint64(to)
	i |= uint64(from) << shift16
	i |= uint64(operator) << shift32
	i |= prefix << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitBinop(lidx, ridx, to int, operator token.Token) {
	var i uint64 = uint64(to)
	i |= uint64(lidx) << shift16
	i |= uint64(ridx) << shift32
	i |= uint64(operator) << shift48
	i |= binop << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitBinopG(lidx, ridx, to int, operator token.Token) {
	var i uint64 = uint64(to)
	i |= uint64(lidx) << shift16
	i |= uint64(ridx) << shift32
	i |= uint64(operator) << shift48
	i |= binopG << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitBinopK(kidx, regAddr, to int, operator token.Token) {
	var i uint64 = uint64(to)
	i |= uint64(kidx) << shift16
	i |= uint64(regAddr) << shift32
	i |= uint64(operator) << shift48
	i |= binopK << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitBinopQ(kidx, regAddr, to int, operator token.Token) {
	var i uint64 = uint64(to)
	i |= uint64(kidx) << shift16
	i |= uint64(regAddr) << shift32
	i |= uint64(operator) << shift48
	i |= binopQ << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitSuperEq(lidx, ridx, to, scopeLeft, scopeRight int, operator token.Token) {
	var s byte = byte(scopeRight)
	s |= byte(scopeLeft) << shift2
	if operator == token.NEQ {
		s |= 1 << shift4
	}
	var i uint64 = uint64(to)
	i |= uint64(ridx) << shift16
	i |= uint64(lidx) << shift32
	i |= uint64(s) << shift48
	i |= eq << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitArray(length, root, to int) {
	var i uint64 = uint64(to)
	i |= uint64(root) << shift16
	i |= uint64(length) << shift32
	i |= array << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitObject(to int) {
	var i uint64 = uint64(to)
	i |= object << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitIGet(indexable, index, to, scopeIndex, scopeIndexable int) {
	var s byte = byte(scopeIndexable)
	s |= byte(scopeIndex) << shift4
	var i uint64 = uint64(to)
	i |= uint64(index) << shift16
	i |= uint64(indexable) << shift32
	i |= uint64(s) << shift48
	i |= iGet << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitISet(indexable, index, expr, scopeIdx, scopeExpr int) {
	var s byte = byte(scopeExpr)
	s |= byte(scopeIdx) << shift4
	var i uint64 = uint64(expr)
	i |= uint64(index) << shift16
	i |= uint64(indexable) << shift32
	i |= uint64(s) << shift48
	i |= iSet << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitSlice(mode, sliceable, to int) {
	var i uint64 = uint64(to)
	i |= uint64(sliceable) << shift16
	i |= uint64(mode) << shift32
	i |= slice << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitForSet(iReg, loop int) {
	var i uint64 = uint64(iReg)
	i |= uint64(loop) << shift16
	i |= forSet << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitForLoop(iReg, loop int) {
	var i uint64 = uint64(iReg)
	i |= uint64(loop) << shift16
	i |= forLoop << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitIForSet(loop, iterable, ireg int) {
	var i uint64 = uint64(ireg)
	i |= uint64(iterable) << shift16
	i |= uint64(loop) << shift32
	i |= iForSet << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitIForLoop(iReg, loop int) {
	var i uint64 = uint64(iReg)
	i |= uint64(loop) << shift16
	i |= iForLoop << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitJump(to int) {
	var i uint64 = uint64(to)
	i |= jump << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitCheck(against, reg, jump int) {
	var i uint64 = uint64(jump)
	i |= uint64(reg) << shift16
	i |= uint64(against) << shift32
	i |= check << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitFun(from, to int) {
	var i uint64 = uint64(to)
	i |= uint64(from) << shift16
	i |= fun << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitCall(callable, argCount, ellipsis, firstArg int) {
	var i uint64 = uint64(callable)
	i |= uint64(argCount) << shift16
	i |= uint64(ellipsis) << shift32
	i |= uint64(firstArg) << shift48
	i |= call << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
}

func (c *compiler) emitRet(source, index int) {
	var i uint64 = uint64(source)
	i |= uint64(index) << shift16
	i |= ret << shift56
	c.currentFn.Code = append(c.currentFn.Code, i)
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
