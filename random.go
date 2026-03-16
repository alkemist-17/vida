package vida

import (
	cryptoRand "crypto/rand"
	"math"
	"math/rand/v2"
)

func loadFoundationRandom() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["N"] = GFn(randN)
	m.Value["I"] = GFn(randNextI)
	m.Value["U32"] = GFn(randNextU32)
	m.Value["F"] = randNextF(rand.Float64)
	m.Value["norm"] = randNextF(rand.NormFloat64)
	m.Value["exp"] = randNextF(rand.ExpFloat64)
	m.Value["perm"] = GFn(randPerm)
	m.Value["shuffled"] = GFn(randShuffled)
	m.Value["zipf"] = randNextZipf(rand.NewZipf(rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())), rand.Float64()+1.0, rand.ExpFloat64()+1.0, math.MaxUint64))
	m.Value["pcg"] = reandNewPermutedCongruentialGenerator(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	m.Value["CC8"] = randNewChaCha8()
	m.Value["bytes"] = GFn(randBytes)
	m.Value["text"] = GFn(randText)
	return m
}

func randNextI(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(Integer); ok {
			if v > 0 {
				return Integer(rand.Int64N(int64(v))), nil
			}
			if v < 0 {
				return Integer(rand.Int64N(int64(-v))), nil
			}
		}
	}
	return Integer(rand.Int64()), nil
}

func randNextF(fn func() float64) GFn {
	return func(args ...Value) (Value, error) {
		return Float(fn()), nil
	}
}

func randPerm(args ...Value) (Value, error) {
	if len(args) > 0 {
		if inputVal, ok := args[0].(Integer); ok {
			size := int(inputVal)
			if 0 < size && size < math.MaxInt32 {
				xs := make([]Value, size)
				for i := range xs {
					xs[i] = Integer(i)
				}
				rand.Shuffle(size, func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
				return &Array{Value: xs}, nil
			}
		}
	}
	return NilValue, nil
}

func randShuffled(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *Array:
			c := v.Clone().(*Array)
			rand.Shuffle(len(v.Value), func(i, j int) { c.Value[i], c.Value[j] = c.Value[j], c.Value[i] })
			return c, nil
		case *String:
			if v.Runes == nil {
				v.Runes = []rune(v.Value)
			}
			l := len(v.Runes)
			r := make([]rune, l)
			copy(r, v.Runes)
			rand.Shuffle(l, func(i, j int) { r[i], r[j] = r[j], r[i] })
			return &String{Value: string(r), Runes: r}, nil
		case *Bytes:
			c := v.Clone().(*Bytes)
			rand.Shuffle(len(v.Value), func(i, j int) { c.Value[i], c.Value[j] = c.Value[j], c.Value[i] })
			return c, nil
		}
	}
	return NilValue, nil
}

func randNextU32(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(Integer); ok {
			if v > 0 {
				return Integer(uint32(rand.Int64N(int64(v)))), nil
			}
			if v < 0 {
				return Integer(uint32(rand.Int64N(int64(-v)))), nil
			}
		}
	}
	return Integer(rand.Int32()), nil
}

func randNextZipf(zipf *rand.Zipf) GFn {
	return func(args ...Value) (Value, error) {
		i64 := Integer(zipf.Uint64())
		if i64 < 0 {
			return -i64, nil
		}
		return i64, nil
	}
}

func reandNewPermutedCongruentialGenerator(pcg *rand.PCG) GFn {
	return func(args ...Value) (Value, error) {
		i64 := Integer(pcg.Uint64())
		if i64 < 0 {
			return -i64, nil
		}
		return i64, nil
	}
}

func randN(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(Integer); ok {
			if v > 0 {
				return Integer(rand.N(int(v))), nil
			}
			if v < 0 {
				return Integer(rand.N(int(-v))), nil
			}
		}
	}
	return Integer(rand.N(math.MaxInt)), nil
}

func randNewChaCha8() GFn {
	b := make([]byte, 32)
	cryptoRand.Read(b)
	chacha8 := rand.NewChaCha8([32]byte(b))
	return func(args ...Value) (Value, error) {
		i64 := Integer(chacha8.Uint64())
		if i64 < 0 {
			return -i64, nil
		}
		return i64, nil
	}
}

func randBytes(args ...Value) (Value, error) {
	if len(args) > 0 {
		if inputValue, ok := args[0].(Integer); ok {
			size := int(inputValue)
			if 0 < size && size < math.MaxInt32 {
				b := make([]byte, size)
				cryptoRand.Read(b)
				return &Bytes{Value: b}, nil
			}
		}
	}
	return NilValue, nil
}

func randText(args ...Value) (Value, error) {
	return &String{Value: cryptoRand.Text()}, nil
}
