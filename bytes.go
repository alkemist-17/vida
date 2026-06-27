package vida

import (
	"crypto/hmac"
	"crypto/md5"
	cryptoRand "crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type encodingType int

const (
	bytesEncodingHex encodingType = iota
	bytesEncodingHEX
	bytesEncodingBase64
	bytesEncodingBase64URL
	bytesEncodingBinary
)

type checksumType int

const (
	csCRC32 checksumType = iota
	csCRC32C
	csAdler32
	csMD5
	csSHA1
	csSHA256
	csSHA3
	csSHA512
)

type hmacAlgoType int

const (
	hmacMD5 hmacAlgoType = iota
	hmacSHA1
	hmacSHA256
	hmacSHA512
)

const (
	littleEndian = "little"
	bigEndian    = "big"
)

const bytesUUIDLen = 16

type Bytes struct {
	ReferenceSemanticsImpl
	Value []byte
}

func (b *Bytes) Boolean() Bool {
	return True
}

func (b *Bytes) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (b *Bytes) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch r := rhs.(type) {
	case *Bytes:
		switch op {
		case uint64(token.ADD):
			rLen := len(r.Value)
			if rLen == 0 {
				return b, nil
			}
			lLen := len(b.Value)
			if rLen+lLen >= verror.MaxMemSize {
				return Nil, verror.ErrMaxMemSize
			}
			values := make([]byte, lLen+rLen)
			copy(values[:lLen], b.Value)
			copy(values[lLen:], r.Value)
			return &Bytes{Value: values}, nil
		}
	}
	switch op {
	case uint64(token.OR):
		return b, nil
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(b, rhs)
	}
	return Nil, verror.ErrBinaryOpNotDefined
}

func (b *Bytes) Get(ctx *Context, index Value) Value {
	switch r := index.(type) {
	case Integer:
		l := Integer(len(b.Value))
		if r < 0 {
			r += l
		}
		if 0 <= r && r < l {
			return Integer(b.Value[r])
		}
	}
	return Nil
}

func (b *Bytes) Set(index, val Value) error {
	return verror.ErrValueIsConstant
}

func (b *Bytes) Equals(other Value) Bool {
	if val, ok := other.(*Bytes); ok {
		return b == val
	}
	if val, ok := other.(*String); ok {
		return string(b.Value) == val.Value
	}
	return false
}

func (b *Bytes) IsIterable() Bool {
	return true
}

func (b *Bytes) IsCallable() Bool {
	return false
}

func (b *Bytes) Iterator() Value {
	return &BytesIterator{Bytes: b.Value, Init: -1, End: len(b.Value)}
}

func (b Bytes) String() string {
	return fmt.Sprintf("bytes[% x]", b.Value)
}

func (b *Bytes) ObjectKey() string {
	return fmt.Sprintf("Bytes(%p)", b)
}

func (b *Bytes) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[bytesVT] == nil {
		ctx.vtables[bytesVT] = loadBytesVT()
	}
	if vtable, ok := ctx.vtables[bytesVT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (b *Bytes) Type() string {
	return "bytes"
}

func (b *Bytes) Clone() Value {
	return &Bytes{Value: b.Value}
}

func loadFoundationBytes() Value {
	m := &Object{Value: make(map[string]Value, 32)}
	m.Value["new"] = NativeFunction(bytesCreateNewBytesValue)
	m.Value["from"] = NativeFunction(bytesFromValue)
	m.Value["genCryptoRand"] = NativeFunction(bytesGenerateCryptoRand)
	m.Value["timingSafeEqual"] = NativeFunction(bytesTimingSafeEqual)
	m.Value["encode"] = NativeFunction(bytesEncode)
	m.Value["decode"] = NativeFunction(bytesDecode)
	m.Value["encoding"] = bytesEncodings()
	m.Value["endian"] = bytesEndian()
	m.Value["fromFile"] = NativeFunction(bytesFromFile)
	m.Value["toFile"] = NativeFunction(bytesToFile)
	m.Value["xor"] = NativeFunction(bytesXOR)
	m.Value["uuid"] = NativeFunction(bytesUUID)
	m.Value["parseUUID"] = NativeFunction(bytesParseUUID)
	m.Value["toString"] = NativeFunction(bytesToString)
	m.Value["bitLen"] = NativeFunction(bytesBitLen)
	m.Value["hexdump"] = NativeFunction(bytesDump)
	m.Value["getSystemEndianess"] = NativeFunction(bytesEndianess)
	m.Value["view"] = NativeFunction(bytesView)
	m.Value["copyTo"] = NativeFunction(bytesCopyTo)
	m.Value["fill"] = NativeFunction(bytesFill)
	m.Value["reverse"] = NativeFunction(bytesReverse)
	m.Value["reversed"] = NativeFunction(bytesReversed)
	m.Value["bitView"] = NativeFunction(bytesBitView)
	m.Value["getBit"] = NativeFunction(bytesGetBit)
	m.Value["setBit"] = NativeFunction(bytesSetBit)
	m.Value["readUInt"] = NativeFunction(bytesReadUInt)
	m.Value["fromUInt"] = NativeFunction(bytesFromUInt)
	m.Value["concat"] = NativeFunction(bytesConcat)
	m.Value["checksums"] = bytesChecksums()
	m.Value["checksum"] = NativeFunction(bytesChecksum)
	m.Value["hmac"] = NativeFunction(bytesHMAC)
	m.Value["hmacAlgorithms"] = bytesHMACAlgorithms()
	return m
}

func bytesCreateNewBytesValue(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		switch v := args[0].(type) {
		case Integer:
			var init byte = 0
			if l > 1 {
				if val, ok := args[1].(Integer); ok {
					init = byte(val)
				}
			}
			if v > 0 && v < verror.MaxMemSize {
				b := make([]byte, v)
				for i := range b {
					b[i] = init
				}
				return &Bytes{Value: b}, nil
			}
		case *String:
			return &Bytes{Value: []byte(v.Value)}, nil
		case *Bytes:
			return v.Clone(), nil
		case *Array:
			var bts []byte
			for _, val := range v.Value {
				if i, ok := val.(Integer); ok {
					bts = append(bts, byte(i))
				}
			}
			return &Bytes{Value: bts}, nil
		}
	}
	return &Bytes{}, nil
}

func bytesFromValue(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *String:
			return &Bytes{Value: []byte(v.Value)}, nil
		case *Bytes:
			return v.Clone(), nil
		case *Array:
			var bts []byte
			for _, val := range v.Value {
				if i, ok := val.(Integer); ok {
					bts = append(bts, byte(i))
				}
			}
			return &Bytes{Value: bts}, nil
		}
	}
	return &Bytes{}, nil
}

func bytesGenerateCryptoRand(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		if inputValue, ok := args[0].(Integer); ok {
			size := int(inputValue)
			if 0 < size && size < math.MaxInt32 {
				b := make([]byte, size)
				cryptoRand.Read(b)
				return &Bytes{Value: b}, nil
			}
		}
	case 2:
		s, okS := args[0].(Integer)
		e, okE := args[1].(Integer)
		if okS && okE {
			size := int(s)
			if 0 < size && size < math.MaxInt32 {
				b := make([]byte, size)
				cryptoRand.Read(b)
				switch e {
				case Integer(bytesEncodingBase64):
					return &String{Value: base64.StdEncoding.EncodeToString(b)}, nil
				case Integer(bytesEncodingHex):
					return &String{Value: hex.EncodeToString(b)}, nil
				case Integer(bytesEncodingHEX):
					return &String{Value: strings.ToUpper(hex.EncodeToString(b))}, nil
				case Integer(bytesEncodingBinary):
					var sb strings.Builder
					sb.Grow(len(b) * 8)
					for _, v := range b {
						fmt.Fprintf(&sb, "%08b", v)
					}
					return &String{Value: sb.String()}, nil
				default:
					return &Bytes{Value: b}, nil
				}
			}
		}
	}
	return Nil, nil
}

func bytesTimingSafeEqual(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		lhs, okl := args[0].(*Bytes)
		rhs, okr := args[1].(*Bytes)
		if okl && okr {
			return Bool(subtle.ConstantTimeCompare(lhs.Value, rhs.Value) == 1), nil
		}
		return False, nil
	}
	return Nil, nil
}

func bytesEncode(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		b, okI := args[0].(*Bytes)
		e, okE := args[1].(Integer)
		if okI && okE {
			switch e {
			case Integer(bytesEncodingBase64):
				return &String{Value: base64.StdEncoding.EncodeToString(b.Value)}, nil
			case Integer(bytesEncodingHex):
				return &String{Value: hex.EncodeToString(b.Value)}, nil
			case Integer(bytesEncodingHEX):
				return &String{Value: strings.ToUpper(hex.EncodeToString(b.Value))}, nil
			case Integer(bytesEncodingBase64URL):
				return &String{Value: base64.URLEncoding.EncodeToString(b.Value)}, nil
			case Integer(bytesEncodingBinary):
				var sb strings.Builder
				sb.Grow(len(b.Value) * 8)
				for _, v := range b.Value {
					fmt.Fprintf(&sb, "%08b", v)
				}
				return &String{Value: sb.String()}, nil
			default:
				return b, nil
			}
		}
	}
	return Nil, nil
}

func bytesDecode(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		e, okE := args[1].(Integer)
		if okS && okE {
			var r []byte
			var err error
			switch e {
			case Integer(bytesEncodingBase64):
				r, err = base64.StdEncoding.DecodeString(s.Value)
				goto resolve
			case Integer(bytesEncodingHex), Integer(bytesEncodingHEX):
				r, err = hex.DecodeString(s.Value)
				goto resolve
			case Integer(bytesEncodingBase64URL):
				r, err = base64.URLEncoding.DecodeString(s.Value)
				goto resolve
			case Integer(bytesEncodingBinary):
				l := len(s.Value)
				if l%8 == 0 {
					r = make([]byte, l/8)
					for i := 0; i < l; i += 8 {
						var b byte
						for j := range 8 {
							switch s.Value[i+j] {
							case '1':
								b = (b << 1) | 1
							case '0':
								b <<= 1
							default:
								return Nil, nil
							}
						}
						r[i/8] = b
					}
					goto resolve
				}
				goto nilvalue
			default:
				return &Bytes{Value: []byte(s.Value)}, nil
			}
		resolve:
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &Bytes{Value: r}, nil
		}
	}
nilvalue:
	return Nil, nil
}

func bytesEncodings() *Object {
	e := make(map[string]Value, 5)
	e["base64"] = Integer(bytesEncodingBase64)
	e["hex"] = Integer(bytesEncodingHex)
	e["HEX"] = Integer(bytesEncodingHEX)
	e["base64url"] = Integer(bytesEncodingBase64URL)
	e["binary"] = Integer(bytesEncodingBinary)
	return &Object{Value: e}
}

func bytesEndian() *Object {
	e := make(map[string]Value, 2)
	e["bigEndian"] = &String{Value: littleEndian}
	e["littleEndian"] = &String{Value: bigEndian}
	return &Object{Value: e}
}

func bytesChecksums() *Object {
	m := make(map[string]Value, 8)
	m["crc32"] = Integer(csCRC32)
	m["crc32c"] = Integer(csCRC32C)
	m["adler32"] = Integer(csAdler32)
	m["md5"] = Integer(csMD5)
	m["sha1"] = Integer(csSHA1)
	m["sha256"] = Integer(csSHA256)
	m["sha512"] = Integer(csSHA512)
	m["sha3_384"] = Integer(csSHA3)
	return &Object{Value: m}
}

func bytesHMACAlgorithms() *Object {
	m := make(map[string]Value, 4)
	m["md5"] = Integer(hmacMD5)
	m["sha1"] = Integer(hmacSHA1)
	m["sha256"] = Integer(hmacSHA256)
	m["sha512"] = Integer(hmacSHA512)
	return &Object{Value: m}
}

func bytesToFile(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		b, okB := args[0].(*Bytes)
		p, okP := args[1].(*String)
		if okB && okP {
			f, err := os.Create(p.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			defer f.Close()
			n, err := f.Write(b.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
		s, okS := args[0].(*String)
		if okS && okP {
			f, err := os.Create(p.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			defer f.Close()
			n, err := f.Write([]byte(s.Value))
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
	}
	return Nil, nil
}

func bytesFromFile(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if path, ok := args[0].(*String); ok {
			data, err := os.ReadFile(path.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &Bytes{Value: data}, nil
		}
	}
	return Nil, nil
}

func bytesXOR(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		lhs, okl := args[0].(*Bytes)
		rhs, okr := args[1].(*Bytes)
		if okl && okr {
			l := min(len(lhs.Value), len(rhs.Value))
			dst := make([]byte, l)
			subtle.XORBytes(dst, lhs.Value, rhs.Value)
			return &Bytes{Value: dst}, nil
		}
	}
	return Nil, nil
}

func bytesUUID(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 1 {
		if _, ok := args[0].(NilValue); ok {
			return &String{Value: "00000000-0000-0000-0000-000000000000"}, nil
		}
		if b, ok := args[0].(*Bytes); ok && len(b.Value) == bytesUUIDLen {
			return &String{Value: fmt.Sprintf("%x-%x-%x-%x-%x", b.Value[0:4], b.Value[4:6], b.Value[6:8], b.Value[8:10], b.Value[10:])}, nil
		}
		return &String{Value: "FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF"}, nil
	}
	b := make([]byte, bytesUUIDLen)
	cryptoRand.Read(b)
	return &String{Value: fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])}, nil
}

func bytesParseUUID(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && len(s.Value) == 2*bytesUUIDLen+4 && s.Value[8] == '-' && s.Value[13] == '-' && s.Value[18] == '-' && s.Value[23] == '-' {
			decoded, err := hex.DecodeString(fmt.Sprintf("%v%v%v%v%v", s.Value[0:8], s.Value[9:13], s.Value[14:18], s.Value[19:23], s.Value[24:36]))
			if err == nil && len(decoded) == bytesUUIDLen {
				return &Bytes{Value: decoded}, nil
			}
		}
	}
	return Nil, nil
}

func bytesToString(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, ok := args[0].(*Bytes); ok {
			return &String{Value: string(b.Value)}, nil
		}
	}
	return Nil, nil
}

func bytesBitLen(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, ok := args[0].(*Bytes); ok {
			return Integer(len(b.Value) * 8), nil
		}
	}
	return Nil, nil
}

func bytesDump(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, ok := args[0].(*Bytes); ok {
			const rowSize = 16
			var sb strings.Builder
			l := len(b.Value)
			for i := 0; i < l; i += rowSize {
				end := min(i+rowSize, l)
				row := b.Value[i:end]
				fmt.Fprintf(&sb, "%08x ", i)
				var space int
				for j := range rowSize {
					if j < len(row) {
						space++
						fmt.Fprintf(&sb, "%02x", row[j])
						if space > 1 {
							space = 0
							sb.WriteByte(' ')
						}
					}
				}
				sb.WriteByte('\n')
			}
			return &String{Value: sb.String()}, nil
		}
	}
	return Nil, nil
}

func bytesEndianess(ctx *Context, args ...Value) (Value, error) {
	b := uint16(0xFF)
	if *(*byte)(unsafe.Pointer(&b)) == 0 {
		return &String{Value: bigEndian}, nil
	}
	return &String{Value: littleEndian}, nil
}

func bytesView(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		b, okB := args[0].(*Bytes)
		offset, okO := args[1].(Integer)
		length, okL := args[2].(Integer)
		if okB && okO && okL {
			srclen := Integer(len(b.Value))
			start, end, _ := sliceBounds(offset, offset+length, srclen)
			return &Bytes{Value: b.Value[start:end]}, nil
		}
	}
	return Nil, nil
}

func bytesCopyTo(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if len(args) > 2 {
		dst, okD := args[0].(*Bytes)
		src, okS := args[1].(*Bytes)
		srcOffset, okO := args[2].(Integer)

		if okD && okS && okO {
			dstLen := len(dst.Value)
			srcLen := len(src.Value)
			offset := int(srcOffset)
			length := dstLen
			if l > 3 {
				if l, ok := args[3].(Integer); ok && int(l) <= dstLen {
					length = int(l)
				}
			}

			if offset < 0 || offset+length > srcLen {
				return &VidaError{Message: &String{Value: "source range out of bounds"}}, nil
			}

			copy(dst.Value, src.Value[offset:offset+length])
			return Integer(length), nil
		}
	}
	return Nil, nil

}

func bytesFill(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		b, okB := args[0].(*Bytes)
		val, okV := args[1].(Integer)
		if okB && okV {
			for i := range b.Value {
				b.Value[i] = byte(val % 256)
			}
			return b, nil
		}
	}
	return Nil, nil
}

func bytesReverse(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, okB := args[0].(*Bytes); okB {
			for i, j := 0, len(b.Value)-1; i < j; i, j = i+1, j-1 {
				b.Value[i], b.Value[j] = b.Value[j], b.Value[i]
			}
			return b, nil
		}
	}
	return Nil, nil
}

func bytesReversed(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, okB := args[0].(*Bytes); okB {
			n := len(b.Value)
			nb := make([]byte, n)
			for i := range n {
				nb[i] = b.Value[n-1-i]
			}
			return &Bytes{Value: nb}, nil
		}
	}
	return Nil, nil
}

func bytesConcat(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return &Bytes{}, nil
	}

	totalLen := 0
	inputs := make([][]byte, len(args))

	for i, arg := range args {
		switch v := arg.(type) {
		case *Bytes:
			inputs[i] = v.Value
			totalLen += len(v.Value)
		case *String:
			b := []byte(v.Value)
			inputs[i] = b
			totalLen += len(b)
		default:
			return &VidaError{Message: &String{Value: "bytes concat only accepts Bytes or String arguments"}}, nil
		}
	}

	dst := make([]byte, totalLen)
	offset := 0
	for _, src := range inputs {
		copy(dst[offset:], src)
		offset += len(src)
	}

	return &Bytes{Value: dst}, nil
}

func bytesChecksum(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return &VidaError{Message: &String{Value: "bytes checksum requires: data, algorithm"}}, nil
	}

	data, okData := args[0].(*Bytes)
	algo, okAlgo := args[1].(Integer)
	if !okData || !okAlgo {
		return &VidaError{Message: &String{Value: "invalid arguments: expected Bytes and checksum algorithm in bytes lib"}}, nil
	}

	var hashBytes []byte

	switch algo {
	case Integer(csCRC32):
		h := crc32.NewIEEE()
		h.Write(data.Value)
		sum := h.Sum32()
		hashBytes = []byte{byte(sum >> 24), byte(sum >> 16), byte(sum >> 8), byte(sum)}

	case Integer(csCRC32C):
		h := crc32.New(crc32.MakeTable(crc32.Castagnoli))
		h.Write(data.Value)
		sum := h.Sum32()
		hashBytes = []byte{byte(sum >> 24), byte(sum >> 16), byte(sum >> 8), byte(sum)}

	case Integer(csAdler32):
		h := adler32.New()
		h.Write(data.Value)
		sum := h.Sum32()
		hashBytes = []byte{byte(sum >> 24), byte(sum >> 16), byte(sum >> 8), byte(sum)}

	case Integer(csMD5):
		h := md5.New()
		h.Write(data.Value)
		hashBytes = h.Sum(nil)

	case Integer(csSHA1):
		h := sha1.New()
		h.Write(data.Value)
		hashBytes = h.Sum(nil)

	case Integer(csSHA3):
		h := sha3.New384()
		h.Write(data.Value)
		hashBytes = h.Sum(nil)

	case Integer(csSHA256):
		h := sha256.New()
		h.Write(data.Value)
		hashBytes = h.Sum(nil)

	case Integer(csSHA512):
		h := sha512.New()
		h.Write(data.Value)
		hashBytes = h.Sum(nil)

	default:
		return &VidaError{Message: &String{Value: "unsupported checksum algorithm"}}, nil
	}

	return &Bytes{Value: hashBytes}, nil
}

func bytesHMAC(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 3 {
		return &VidaError{Message: &String{Value: "bytes hmac requires: data, key, algorithm"}}, nil
	}

	data, okD := args[0].(*Bytes)
	key, okK := args[1].(*Bytes)
	algo, okA := args[2].(Integer)

	if !okD || !okK || !okA {
		return &VidaError{Message: &String{Value: "invalid arguments: expected Bytes, Bytes, algorithm in bytes hmac"}}, nil
	}

	var h func() hash.Hash
	switch algo {
	case Integer(hmacMD5):
		h = md5.New
	case Integer(hmacSHA1):
		h = sha1.New
	case Integer(hmacSHA256):
		h = sha256.New
	case Integer(hmacSHA512):
		h = sha512.New
	default:
		return &VidaError{Message: &String{Value: "unsupported HMAC algorithm"}}, nil
	}

	mac := hmac.New(h, key.Value)

	_, err := mac.Write(data.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}

	return &Bytes{Value: mac.Sum(nil)}, nil
}

// =============================================================================
// Bit-manipulation functions for the Vida bytes library.
//
// BIT INDEXING CONVENTION (MSB-first, left-to-right):
//
//   Byte index:    [    0    ] [    1    ] [    2    ] …
//   Bit index:      0  1 … 7   8  9 … 15  16 …
//   Bit weight:    2⁷ 2⁶… 2⁰  2⁷ 2⁶… 2⁰  2⁷ …
//
//   getBit(buf, 0) returns the MSB of buf[0]  (the 0x80 bit).
//   getBit(buf, 7) returns the LSB of buf[0]  (the 0x01 bit).
//   getBit(buf, 8) returns the MSB of buf[1]  (the 0x80 bit).
//
//   This matches RFC/protocol diagrams and natural human binary reading.
//
// ENDIANNESS (byte order only, never bit reversal):
//
//   readUInt / fromUInt interpret multi-byte values extracted by bitView.
//   After bitView produces a packed []byte with MSB-first bits:
//
//   big-endian:    result[0] is the most-significant byte
//                  value = result[0]<<((n-1)*8) | … | result[n-1]
//
//   little-endian: result[0] is the least-significant byte
//                  value = result[n-1]<<((n-1)*8) | … | result[0]
//
//   Endianness NEVER reverses individual bits. It only changes which byte
//   is treated as most- vs least-significant.
// =============================================================================

// bitByteIdx returns the byte index and the bit mask for global bit index i
// using MSB-first convention.
//
//	byteIndex = i / 8
//	mask      = 0x80 >> (i % 8)
func bitByteIdx(i int) (byteIndex int, mask byte) {
	return i / 8, 0x80 >> uint(i%8)
}

// bytesGetBit returns the value (0 or 1) of the bit at global index bitIdx
// using MSB-first numbering.
//
// Signature: getBit(buf Bytes, bitIdx Int) → Int | &VidaError
func bytesGetBit(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 2 {
		return &VidaError{Message: &String{Value: "bytes.getBit requires: bytes, bitIndex"}}, nil
	}
	b, ok := args[0].(*Bytes)
	bitIdx, okIdx := args[1].(Integer)
	if !ok || !okIdx || bitIdx < 0 {
		return &VidaError{Message: &String{Value: "bytes.getBit: invalid arguments (expected Bytes, non-negative Int)"}}, nil
	}

	totalBits := len(b.Value) * 8
	if int(bitIdx) >= totalBits {
		return &VidaError{Message: &String{Value: fmt.Sprintf("bytes.getBit: bit index %d out of range (buffer has %d bits)", bitIdx, totalBits)}}, nil
	}

	byteIndex, mask := bitByteIdx(int(bitIdx))
	if b.Value[byteIndex]&mask != 0 {
		return Integer(1), nil
	}
	return Integer(0), nil
}

// bytesSetBit returns a new Bytes value with the bit at global index bitIdx
// set to val (0 or 1). The original buffer is not modified.
//
// Signature: setBit(buf Bytes, bitIdx Int, val Int) → Bytes | &VidaError
func bytesSetBit(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 3 {
		return &VidaError{Message: &String{Value: "bytes.setBit requires: bytes, bitIndex, value (0 or 1)"}}, nil
	}
	b, ok := args[0].(*Bytes)
	bitIdx, okIdx := args[1].(Integer)
	val, okVal := args[2].(Integer)

	if !ok || !okIdx || !okVal || bitIdx < 0 || (val != 0 && val != 1) {
		return &VidaError{Message: &String{Value: "bytes.setBit: invalid arguments (expected Bytes, non-negative Int bitIndex, 0 or 1 value)"}}, nil
	}

	totalBits := len(b.Value) * 8
	if int(bitIdx) >= totalBits {
		return &VidaError{Message: &String{Value: "bytes.setBit: bit index out of range"}}, nil
	}

	clone := make([]byte, len(b.Value))
	copy(clone, b.Value)

	byteIndex, mask := bitByteIdx(int(bitIdx))
	if val == 1 {
		clone[byteIndex] |= mask
	} else {
		clone[byteIndex] &^= mask
	}
	return &Bytes{Value: clone}, nil
}

// bytesBitView extracts a contiguous run of bits [start, start+length) from buf
// and packs them into a new Bytes value, MSB-first.
//
// The returned buffer has ceil(length/8) bytes. The first extracted bit becomes
// the MSB of result[0]; any unused low bits in the last byte are zero-padded.
//
// Example: buf = [0b10110001], bitView(buf, 2, 4)
//
//	extracts bits at global indices 2,3,4,5 → values 1,1,0,0
//	packs as 0b11000000 → result = [0xC0]
//
// Signature: bitView(buf Bytes, start Int, length Int) → Bytes | &VidaError
func bytesBitView(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return &VidaError{Message: &String{Value: "bytes.bitView requires: bytes, startBit, bitLength"}}, nil
	}
	b, ok := args[0].(*Bytes)
	start, okS := args[1].(Integer)
	length, okL := args[2].(Integer)

	if !ok || !okS || !okL || start < 0 || length < 0 {
		return &VidaError{Message: &String{Value: "bytes.bitView: invalid arguments (expected Bytes, non-negative Int start and length)"}}, nil
	}
	if length == 0 {
		return &Bytes{}, nil
	}
	if int(start)+int(length) > len(b.Value)*8 {
		return &VidaError{Message: &String{Value: "bytes.bitView: bit range exceeds buffer size"}}, nil
	}

	dstLen := (int(length) + 7) / 8
	dst := make([]byte, dstLen)

	for i := 0; i < int(length); i++ {
		srcGlobal := int(start) + i
		srcByteIdx := srcGlobal / 8
		srcMask := byte(0x80 >> uint(srcGlobal%8))
		bitVal := (b.Value[srcByteIdx] & srcMask) != 0

		dstByteIdx := i / 8
		dstMask := byte(0x80 >> uint(i%8))
		if bitVal {
			dst[dstByteIdx] |= dstMask
		}
	}

	return &Bytes{Value: dst}, nil
}

// bytesReadUInt reads a little-endian or big-endian unsigned integer from a
// run of bits [startBit, startBit+bitLength) in buf.
//
// The bit field is first extracted MSB-first by bitView, producing a packed
// []byte. Endianness then controls how those bytes are assembled into a uint64:
//
//	big-endian:    result[0] is most-significant  (network byte order)
//	little-endian: result[0] is least-significant (x86 byte order)
//
// bitLength must be in [1, 64].
//
// Signature: readUInt(buf Bytes, startBit Int, bitLength Int, endian String) → Int | &VidaError
func bytesReadUInt(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 4 {
		return &VidaError{Message: &String{Value: "bytes.readUInt requires: bytes, startBit, bitLength, endian"}}, nil
	}
	b, ok := args[0].(*Bytes)
	start, okS := args[1].(Integer)
	bitLen, okL := args[2].(Integer)
	endian, okE := args[3].(*String)

	if !ok || !okS || !okL || !okE || start < 0 || bitLen < 1 || bitLen > 64 {
		return &VidaError{Message: &String{Value: "bytes.readUInt: invalid arguments (bitLength must be 1-64)"}}, nil
	}
	if int(start)+int(bitLen) > len(b.Value)*8 {
		return &VidaError{Message: &String{Value: "bytes.readUInt: bit range exceeds buffer"}}, nil
	}

	extracted, err := bytesBitView(ctx, b, start, bitLen)
	if err != nil {
		return extracted, err
	}
	extBytes := extracted.(*Bytes).Value

	var val uint64
	n := len(extBytes)

	switch endian.Value {
	case bigEndian:
		for i := range n {
			val = (val << 8) | uint64(extBytes[i])
		}
	case littleEndian:
		for i := n - 1; i >= 0; i-- {
			val = (val << 8) | uint64(extBytes[i])
		}
	default:
		return &VidaError{Message: &String{Value: fmt.Sprintf("bytes.readUInt: unknown endian %q (use bytes.endian.big or bytes.endian.little)", endian)}}, nil
	}

	unusedBits := (n * 8) - int(bitLen)
	val >>= uint(unusedBits)

	return Integer(val), nil
}

// bytesFromUInt packs an unsigned integer value into a minimal []byte using
// MSB-first bit layout.
//
// bitLength specifies how many bits of value to use (1–64).  The result has
// ceil(bitLength/8) bytes.  Bits are placed MSB-first: the highest-order
// requested bit goes into the MSB position of result[0].
//
// Endianness controls byte order of the result:
//
//	big-endian:    result[0] is most-significant (standard / network order)
//	little-endian: result[0] is least-significant (x86 order)
//
// Signature: fromUInt(value Int, bitLength Int, endian String) → Bytes | &VidaError
func bytesFromUInt(ctx *Context, args ...Value) (Value, error) {
	if len(args) != 3 {
		return &VidaError{Message: &String{Value: "bytes.fromUInt requires: value, bitLength, endian"}}, nil
	}
	val, okV := args[0].(Integer)
	bitLen, okL := args[1].(Integer)
	endian, okE := args[2].(*String)

	if !okV || !okL || !okE || bitLen < 1 || bitLen > 64 || val < 0 {
		return &VidaError{Message: &String{Value: "bytes.fromUInt: invalid arguments (value must be ≥ 0, bitLength must be 1-64)"}}, nil
	}

	u := uint64(val)
	maxVal := uint64(0)
	if int(bitLen) == 64 {
		maxVal = ^uint64(0)
	} else {
		maxVal = (uint64(1) << uint(bitLen)) - 1
	}
	if u > maxVal {
		return &VidaError{Message: &String{Value: fmt.Sprintf("bytes.fromUInt: value %d exceeds %d-bit capacity (%d max)", val, bitLen, maxVal)}}, nil
	}

	dstLen := (int(bitLen) + 7) / 8
	dst := make([]byte, dstLen)

	for i := 0; i < int(bitLen); i++ {
		bitPos := uint(int(bitLen) - 1 - i)
		bitVal := (u >> bitPos) & 1

		dstByteIdx := i / 8
		dstMask := byte(0x80 >> uint(i%8))
		if bitVal == 1 {
			dst[dstByteIdx] |= dstMask
		}
	}

	switch endian.Value {
	case bigEndian:
	case littleEndian:
		for i, j := 0, len(dst)-1; i < j; i, j = i+1, j-1 {
			dst[i], dst[j] = dst[j], dst[i]
		}
	default:
		return &VidaError{Message: &String{Value: fmt.Sprintf("bytes.fromUInt: unknown endian %q (use bytes.endian.big or bytes.endian.little)", endian)}}, nil
	}

	return &Bytes{Value: dst}, nil
}
