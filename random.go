package vida

import (
	cryptoRand "crypto/rand"
	"math"
	"math/rand/v2"
)

const nanoIDDefaultAlphabet = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const nanoIDDefaultSize = 21
const nanoIDMaxSize = 36

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
	m.Value["nanoid"] = GFn(randNanoID)
	m.Value["customNanoid"] = GFn(randNanoIDCustomAlphabeth)
	m.Value["nanoidDefaultSize"] = Integer(nanoIDDefaultSize)
	m.Value["alphanumeric"] = &String{Value: "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	m.Value["numeric"] = &String{Value: "0123456789"}
	m.Value["alpha"] = &String{Value: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}
	m.Value["password"] = &String{Value: " !#$%&'()*+,-./:;<=>?@[\\]^_`{|}~0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"}
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

func randNanoID(args ...Value) (Value, error) {
	switch len(args) {
	case 0:
		b := make([]byte, nanoIDDefaultSize)
		r := []rune(nanoIDDefaultAlphabet)
		cryptoRand.Read(b)
		nanoid := make([]rune, nanoIDDefaultSize)
		for i := range nanoIDDefaultSize {
			nanoid[i] = r[b[i]&63]
		}
		return &String{Value: string(nanoid), Runes: nanoid}, nil
	case 1:
		if size, ok := args[0].(Integer); ok && 0 < size && size <= nanoIDMaxSize {
			b := make([]byte, size)
			r := []rune(nanoIDDefaultAlphabet)
			cryptoRand.Read(b)
			nanoid := make([]rune, size)
			for i := range size {
				nanoid[i] = r[b[i]&63]
			}
			return &String{Value: string(nanoid), Runes: nanoid}, nil
		}
	}
	return NilValue, nil
}

func randNanoIDCustomAlphabeth(args ...Value) (Value, error) {
	if len(args) > 1 {
		alpha, okAlpha := args[0].(*String)
		size, oksize := args[1].(Integer)
		if okAlpha && oksize && 0 < len(alpha.Value) && len(alpha.Value) < 256 && size > 0 {
			mask := generateMask(len(alpha.Value))
			steps := int(math.Ceil(1.6 * float64(mask*size) / float64(len(alpha.Value))))
			nanoid := make([]rune, size)
			b := make([]byte, steps)
			r := []rune(alpha.Value)
			lenr := len(r)
			return GFn(func(args ...Value) (Value, error) {
				for {
					cryptoRand.Read(b)
					var j int
					for i := range steps {
						idx := int(b[i] & byte(mask))
						if idx < lenr {
							nanoid[j] = r[idx]
							j++
							if j == int(size) {
								return &String{Value: string(nanoid), Runes: nanoid}, nil
							}
						}
					}
				}
			}), nil
		}
	}
	return NilValue, nil
}

func generateMask(length int) Integer {
	for i := range 9 {
		mask := (2 << uint(i)) - 1
		if mask >= length-1 {
			return Integer(mask)
		}
	}
	return 0
}
