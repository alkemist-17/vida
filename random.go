package vida

import (
	"math/rand/v2"
)

func loadFoundationRandom() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["nextI"] = GFn(randNextI)
	m.Value["nextU32"] = GFn(randNextU32)
	m.Value["nextF"] = randNextF(rand.Float64)
	m.Value["norm"] = randNextF(rand.NormFloat64)
	m.Value["exp"] = randNextF(rand.ExpFloat64)
	m.Value["perm"] = GFn(randPerm)
	m.Value["shuffled"] = GFn(randShuffled)
	return m
}

func randNextI(args ...Value) (Value, error) {
	if len(args) == 0 {
		return Integer(rand.Int64()), nil
	}
	if len(args) > 0 {
		if v, ok := args[0].(Integer); ok {
			if v > 0 {
				return Integer(rand.Int64N(int64(v))), nil
			}
			if v < 0 {
				return Integer(rand.Int64N(int64(-v))), nil
			}
			return Integer(0), nil
		} else {
			return Integer(rand.Int64()), nil
		}
	}
	return NilValue, nil
}

func randNextF(fn func() float64) GFn {
	return func(args ...Value) (Value, error) {
		return Float(fn()), nil
	}
}

func randPerm(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(Integer); ok {
			if 0 <= v && v <= 0x7FFF_FFFF {
				xs := make([]Value, v)
				for i := range xs {
					xs[i] = Integer(i)
				}
				rand.Shuffle(int(v), func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
				return &Array{Value: xs}, nil
			}
		}
	}
	return NilValue, nil
}

func randShuffled(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*Array); ok {
			c := v.Clone().(*Array)
			rand.Shuffle(len(v.Value), func(i, j int) { c.Value[i], c.Value[j] = c.Value[j], c.Value[i] })
			return c, nil
		}
		if v, ok := args[0].(*String); ok {
			if v.Runes == nil {
				v.Runes = []rune(v.Value)
			}
			l := len(v.Runes)
			r := make([]rune, l)
			copy(r, v.Runes)
			rand.Shuffle(l, func(i, j int) { r[i], r[j] = r[j], r[i] })
			return &String{Value: string(r), Runes: r}, nil
		}
	}
	return NilValue, nil
}

func randNextU32(args ...Value) (Value, error) {
	if len(args) == 0 {
		return Integer(rand.Int32()), nil
	}
	if len(args) > 0 {
		if v, ok := args[0].(Integer); ok {
			if v > 0 {
				return Integer(uint32(rand.Int64N(int64(v)))), nil
			}
			if v < 0 {
				return Integer(uint32(rand.Int64N(int64(-v)))), nil
			}
			return Integer(0), nil
		} else {
			return Integer(rand.Int32()), nil
		}
	}
	return NilValue, nil
}
