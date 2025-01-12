package libs

import (
	"math"

	"github.com/ever-eduardo/vida"
)

func generateMath() vida.Value {
	m := &vida.Object{Value: make(map[string]vida.Value)}
	m.Value["pi"] = vida.Float(math.Pi)
	m.Value["tau"] = vida.Float(math.Pi * 2)
	m.Value["phi"] = vida.Float(math.Phi)
	m.Value["e"] = vida.Float(math.E)
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
	m.Value["sinh"] = mathFromFloatToFloat(math.Sinh)
	m.Value["cosh"] = mathFromFloatToFloat(math.Cosh)
	m.Value["tanh"] = mathFromFloatToFloat(math.Tanh)
	m.Value["asinh"] = mathFromFloatToFloat(math.Asinh)
	m.Value["acosh"] = mathFromFloatToFloat(math.Acosh)
	m.Value["atanh"] = mathFromFloatToFloat(math.Atanh)
	m.Value["pow"] = mathPow(math.Pow)
	m.UpdateKeys()
	return m
}

func mathInf(fn func(int) float64) vida.GFn {
	return func(args ...vida.Value) (vida.Value, error) {
		if len(args) > 0 {
			if v, ok := args[0].(vida.Integer); ok {
				return vida.Float(fn(int(v))), nil
			}
		}
		return vida.NilValue, nil
	}
}

func mathIsNan(fn func(float64) bool) vida.GFn {
	return func(args ...vida.Value) (vida.Value, error) {
		if len(args) > 0 {
			if v, ok := args[0].(vida.Float); ok {
				return vida.Bool(fn(float64(v))), nil
			}
			if v, ok := args[0].(vida.Integer); ok {
				return vida.Bool(fn(float64(v))), nil
			}
		}
		return vida.NilValue, nil
	}
}

func mathIsInf(fn func(float64, int) bool) vida.GFn {
	return func(args ...vida.Value) (vida.Value, error) {
		if len(args) > 1 {
			if v, ok := args[0].(vida.Float); ok {
				if i, oki := args[1].(vida.Integer); oki {
					return vida.Bool(fn(float64(v), int(i))), nil
				}
			}
			if v, ok := args[0].(vida.Integer); ok {
				if i, oki := args[1].(vida.Integer); oki {
					return vida.Bool(fn(float64(v), int(i))), nil
				}
			}
		}
		return vida.NilValue, nil
	}
}

func mathNan(fn func() float64) vida.GFn {
	return func(args ...vida.Value) (vida.Value, error) {
		return vida.Float(fn()), nil
	}
}

func mathFromFloatToFloat(fn func(float64) float64) vida.GFn {
	return func(args ...vida.Value) (vida.Value, error) {
		if len(args) > 0 {
			if v, ok := args[0].(vida.Float); ok {
				return vida.Float(fn(float64(v))), nil
			}
			if v, ok := args[0].(vida.Integer); ok {
				return vida.Float(fn(float64(v))), nil
			}
		}
		return vida.NilValue, nil
	}
}

func mathPow(fn func(float64, float64) float64) vida.GFn {
	return func(args ...vida.Value) (vida.Value, error) {
		if len(args) > 1 {
			switch l := args[0].(type) {
			case vida.Integer:
				switch r := args[1].(type) {
				case vida.Integer:
					return vida.Integer(fn(float64(l), float64(r))), nil
				case vida.Float:
					return vida.Float(fn(float64(l), float64(r))), nil
				}
			case vida.Float:
				switch r := args[1].(type) {
				case vida.Integer:
					return vida.Float(fn(float64(l), float64(r))), nil
				case vida.Float:
					return vida.Float(fn(float64(l), float64(r))), nil
				}
			}
		}
		return vida.NilValue, nil
	}
}
