package vida

import (
	cryptoRand "crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"github.com/alkemist-17/vida/verror"
)

const (
	bytesBase64URL = "base64url"
	bytesBase64    = "base64"
	bytesHex       = "hex"
	bytesUUIDLen   = 16
)

func loadFoundationBytes() Value {
	m := &Object{Value: make(map[string]Value, 12)}
	m.Value["new"] = GFn(bytesCreateNewBytesValue)
	m.Value["from"] = GFn(bytesFromValue)
	m.Value["cryptoRandom"] = GFn(bytesCryptoRandom)
	m.Value["timingSafeEqual"] = GFn(bytesTimingSafeEqual)
	m.Value["encode"] = GFn(bytesEncode)
	m.Value["decode"] = GFn(bytesDecode)
	m.Value["encoding"] = bytesEncodings()
	m.Value["toFile"] = GFn(bytesToFile)
	m.Value["xor"] = GFn(bytesXOR)
	m.Value["uuid"] = GFn(bytesUUID)
	m.Value["parseUUID"] = GFn(bytesParseUUID)
	m.Value["toString"] = GFn(bytesToString)
	return m
}

func bytesCreateNewBytesValue(args ...Value) (Value, error) {
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
				for i := range v {
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

func bytesFromValue(args ...Value) (Value, error) {
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

func bytesCryptoRandom(args ...Value) (Value, error) {
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
		e, okE := args[1].(*String)
		if okS && okE {
			size := int(s)
			if 0 < size && size < math.MaxInt32 {
				b := make([]byte, size)
				cryptoRand.Read(b)
				switch e.Value {
				case bytesBase64:
					return &String{Value: base64.StdEncoding.EncodeToString(b)}, nil
				case bytesHex:
					return &String{Value: hex.EncodeToString(b)}, nil
				default:
					return &Bytes{Value: b}, nil
				}
			}
		}
	}
	return NilValue, nil
}

func bytesTimingSafeEqual(args ...Value) (Value, error) {
	if len(args) > 1 {
		lhs, okl := args[0].(*Bytes)
		rhs, okr := args[1].(*Bytes)
		if okl && okr {
			return Bool(subtle.ConstantTimeCompare(lhs.Value, rhs.Value) == 1), nil
		}
		return Bool(false), nil
	}
	return NilValue, nil
}

func bytesEncode(args ...Value) (Value, error) {
	if len(args) > 1 {
		b, okI := args[0].(*Bytes)
		e, okE := args[1].(*String)
		if okI && okE {
			switch e.Value {
			case bytesBase64:
				return &String{Value: base64.StdEncoding.EncodeToString(b.Value)}, nil
			case bytesHex:
				return &String{Value: hex.EncodeToString(b.Value)}, nil
			case bytesBase64URL:
				return &String{Value: base64.URLEncoding.EncodeToString(b.Value)}, nil
			default:
				return b, nil
			}
		}
	}
	return NilValue, nil
}

func bytesDecode(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		e, okE := args[1].(*String)
		if okS && okE {
			var r []byte
			var err error
			switch e.Value {
			case bytesBase64:
				r, err = base64.StdEncoding.DecodeString(s.Value)
				goto resolve
			case bytesHex:
				r, err = hex.DecodeString(s.Value)
				goto resolve
			case bytesBase64URL:
				r, err = base64.URLEncoding.DecodeString(s.Value)
				goto resolve
			default:
				return &Bytes{Value: []byte(s.Value)}, nil
			}
		resolve:
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &Bytes{Value: r}, nil
		}
	}
	return NilValue, nil
}

func bytesEncodings() *Object {
	e := make(map[string]Value, 3)
	e[bytesBase64] = &String{Value: bytesBase64}
	e[bytesHex] = &String{Value: bytesHex}
	e[bytesBase64URL] = &String{Value: bytesBase64URL}
	return &Object{Value: e}
}

func bytesToFile(args ...Value) (Value, error) {
	if len(args) > 1 {
		b, okB := args[0].(*Bytes)
		p, okP := args[1].(*String)
		if okB && okP {
			f, err := os.Create(p.Value)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			defer f.Close()
			n, err := f.Write(b.Value)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
		s, okS := args[0].(*String)
		if okS && okP {
			f, err := os.Create(p.Value)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			defer f.Close()
			n, err := f.WriteString(s.Value)
			if err != nil {
				return VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return Integer(n), nil
		}
	}
	return NilValue, nil
}

func bytesXOR(args ...Value) (Value, error) {
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
	return NilValue, nil
}

func bytesUUID(args ...Value) (Value, error) {
	if len(args) == 1 {
		if _, ok := args[0].(Nil); ok {
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

func bytesParseUUID(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && len(s.Value) == 2*bytesUUIDLen+4 && s.Value[8] == '-' && s.Value[13] == '-' && s.Value[18] == '-' && s.Value[23] == '-' {
			decoded, err := hex.DecodeString(fmt.Sprintf("%v%v%v%v%v", s.Value[0:8], s.Value[9:13], s.Value[14:18], s.Value[19:23], s.Value[24:36]))
			if err == nil && len(decoded) == bytesUUIDLen {
				return &Bytes{Value: decoded}, nil
			}
		}
	}
	return NilValue, nil
}

func bytesToString(args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, ok := args[0].(*Bytes); ok {
			return &String{Value: string(b.Value)}, nil
		}
	}
	return NilValue, nil
}
