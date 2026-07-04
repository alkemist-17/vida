package vida

import (
	"math"
)

func loadFoundationMath() Value {
	m := &Object{Value: make(map[string]Value, 35)}
	m.Value["pi"] = Float(math.Pi)
	m.Value["tau"] = Float(math.Pi * 2)
	m.Value["phi"] = Float(math.Phi)
	m.Value["e"] = Float(math.E)
	m.Value["inf"] = mathInf(math.Inf)
	m.Value["isNan"] = mathIsNan(math.IsNaN)
	m.Value["isInf"] = mathIsInf(math.IsInf)
	m.Value["nan"] = mathNan(math.NaN)
	m.Value["ceil"] = mathFromFloatToFloat(math.Ceil)
	m.Value["floor"] = mathFromFloatToFloat(math.Floor)
	m.Value["round"] = mathFromFloatToFloat(math.Round)
	m.Value["roundToEven"] = mathFromFloatToFloat(math.RoundToEven)
	m.Value["abs"] = mathFromFloatToFloat(math.Abs)
	m.Value["sqrt"] = mathFromFloatToFloat(math.Sqrt)
	m.Value["cbrt"] = mathFromFloatToFloat(math.Cbrt)
	m.Value["sin"] = mathFromFloatToFloat(math.Sin)
	m.Value["cos"] = mathFromFloatToFloat(math.Cos)
	m.Value["tan"] = mathFromFloatToFloat(math.Tan)
	m.Value["asin"] = mathFromFloatToFloat(math.Asin)
	m.Value["acos"] = mathFromFloatToFloat(math.Acos)
	m.Value["atan"] = mathFromFloatToFloat(math.Atan)
	m.Value["atan2"] = mathFrom2FloatsToFloat(math.Atan2)
	m.Value["sinh"] = mathFromFloatToFloat(math.Sinh)
	m.Value["cosh"] = mathFromFloatToFloat(math.Cosh)
	m.Value["tanh"] = mathFromFloatToFloat(math.Tanh)
	m.Value["asinh"] = mathFromFloatToFloat(math.Asinh)
	m.Value["acosh"] = mathFromFloatToFloat(math.Acosh)
	m.Value["atanh"] = mathFromFloatToFloat(math.Atanh)
	m.Value["pow"] = mathPow(math.Pow)
	m.Value["exp"] = mathFromFloatToFloat(math.Exp)
	m.Value["exp2"] = mathFromFloatToFloat(math.Exp2)
	m.Value["gamma"] = mathFromFloatToFloat(math.Gamma)
	m.Value["hypot"] = mathFrom2FloatsToFloat(math.Hypot)
	m.Value["max"] = mathFrom2FloatsToFloat(math.Max)
	m.Value["min"] = mathFrom2FloatsToFloat(math.Min)
	return m
}

func mathInf(fn func(int) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if v, ok := args[0].(Integer); ok {
				return Float(fn(int(v))), nil
			}
		}
		return Nil, nil
	}
}

func mathIsNan(fn func(float64) bool) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if v, ok := args[0].(Float); ok {
				return Bool(fn(float64(v))), nil
			}
			if v, ok := args[0].(Integer); ok {
				return Bool(fn(float64(v))), nil
			}
		}
		return Nil, nil
	}
}

func mathIsInf(fn func(float64, int) bool) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 1 {
			if v, ok := args[0].(Float); ok {
				if i, oki := args[1].(Integer); oki {
					return Bool(fn(float64(v), int(i))), nil
				}
			}
			if v, ok := args[0].(Integer); ok {
				if i, oki := args[1].(Integer); oki {
					return Bool(fn(float64(v), int(i))), nil
				}
			}
		}
		return Nil, nil
	}
}

func mathNan(fn func() float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		return Float(fn()), nil
	}
}

func mathFromFloatToFloat(fn func(float64) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 0 {
			if v, ok := args[0].(Float); ok {
				return Float(fn(float64(v))), nil
			}
			if v, ok := args[0].(Integer); ok {
				return Float(fn(float64(v))), nil
			}
		}
		return Nil, nil
	}
}

func mathFrom2FloatsToFloat(fn func(float64, float64) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 1 {
			l, okl := getFloat(args[0])
			r, okr := getFloat(args[1])
			if okl && okr {
				return Float(fn(float64(l), float64(r))), nil
			}
		}
		return Nil, nil
	}
}

func getFloat(value Value) (Float, bool) {
	if v, ok := value.(Float); ok {
		return v, true
	}
	if v, ok := value.(Integer); ok {
		return Float(v), true
	}
	return 0.0, false
}

func mathPow(fn func(float64, float64) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) > 1 {
			switch l := args[0].(type) {
			case Integer:
				switch r := args[1].(type) {
				case Integer:
					return Integer(fn(float64(l), float64(r))), nil
				case Float:
					return Float(fn(float64(l), float64(r))), nil
				}
			case Float:
				switch r := args[1].(type) {
				case Integer:
					return Float(fn(float64(l), float64(r))), nil
				case Float:
					return Float(fn(float64(l), float64(r))), nil
				}
			}
		}
		return Nil, nil
	}
}
