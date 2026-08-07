package vida

import (
	"fmt"
	"math"
)

func loadFoundationMath() Value {
	m := &Object{Value: make(map[string]Value, 42)}
	m.Value["pi"] = Float(math.Pi)
	m.Value["tau"] = Float(math.Pi * 2)
	m.Value["phi"] = Float(math.Phi)
	m.Value["e"] = Float(math.E)
	m.Value["inf"] = mathIntFloat("inf", math.Inf)
	m.Value["isNan"] = mathIsNan("isNan", math.IsNaN)
	m.Value["isInf"] = mathIsInf("isInf", math.IsInf)
	m.Value["nan"] = mathNan(math.NaN)
	m.Value["ceil"] = mathFromFloatToFloat("ceil", math.Ceil)
	m.Value["floor"] = mathFromFloatToFloat("floor", math.Floor)
	m.Value["round"] = mathFromFloatToFloat("round", math.Round)
	m.Value["roundToEven"] = mathFromFloatToFloat("roundToEven", math.RoundToEven)
	m.Value["abs"] = mathFromFloatToFloat("abs", math.Abs)
	m.Value["sqrt"] = mathFromFloatToFloat("sqrt", math.Sqrt)
	m.Value["cbrt"] = mathFromFloatToFloat("cbrt", math.Cbrt)
	m.Value["sin"] = mathFromFloatToFloat("sin", math.Sin)
	m.Value["cos"] = mathFromFloatToFloat("cos", math.Cos)
	m.Value["tan"] = mathFromFloatToFloat("tan", math.Tan)
	m.Value["asin"] = mathFromFloatToFloat("asin", math.Asin)
	m.Value["acos"] = mathFromFloatToFloat("acos", math.Acos)
	m.Value["atan"] = mathFromFloatToFloat("atan", math.Atan)
	m.Value["atan2"] = mathFrom2FloatsToFloat("atan2", math.Atan2)
	m.Value["sinh"] = mathFromFloatToFloat("sinh", math.Sinh)
	m.Value["cosh"] = mathFromFloatToFloat("cosh", math.Cosh)
	m.Value["tanh"] = mathFromFloatToFloat("tanh", math.Tanh)
	m.Value["asinh"] = mathFromFloatToFloat("asinh", math.Asinh)
	m.Value["acosh"] = mathFromFloatToFloat("acosh", math.Acosh)
	m.Value["atanh"] = mathFromFloatToFloat("atanh", math.Atanh)
	m.Value["pow"] = mathPow(math.Pow)
	m.Value["pow10"] = mathIntFloat("pow10", math.Pow10)
	m.Value["mod"] = mathFrom2FloatsToFloat("mod", math.Mod)
	m.Value["rem"] = mathFrom2FloatsToFloat("rem", math.Remainder)
	m.Value["exp"] = mathFromFloatToFloat("exp", math.Exp)
	m.Value["exp2"] = mathFromFloatToFloat("exp2", math.Exp2)
	m.Value["gamma"] = mathFromFloatToFloat("gamma", math.Gamma)
	m.Value["hypot"] = mathFrom2FloatsToFloat("hypot", math.Hypot)
	m.Value["max"] = mathFrom2FloatsToFloat("max", math.Max)
	m.Value["min"] = mathFrom2FloatsToFloat("min", math.Min)
	m.Value["log"] = mathFromFloatToFloat("log", math.Log)
	m.Value["log2"] = mathFromFloatToFloat("log2", math.Log2)
	m.Value["log10"] = mathFromFloatToFloat("log10", math.Log10)
	m.Value["trunc"] = mathFromFloatToFloat("trunc", math.Trunc)
	return m
}

func mathIntFloat(name string, fn func(int) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) == 0 {
			return argError(name, "expected an integer argument")
		}
		v, ok := args[0].(Integer)
		if !ok {
			return argError(name, fmt.Sprintf("expected an integer argument, got %s", args[0].Type()))
		}
		return Float(fn(int(v))), nil
	}
}

func mathIsNan(name string, fn func(float64) bool) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) == 0 {
			return argError(name, "expected a numeric argument")
		}
		if v, ok := args[0].(Float); ok {
			return Bool(fn(float64(v))), nil
		}
		if v, ok := args[0].(Integer); ok {
			return Bool(fn(float64(v))), nil
		}
		return argError(name, fmt.Sprintf("expected a numeric argument, got %s", args[0].Type()))
	}
}

func mathIsInf(name string, fn func(float64, int) bool) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) < 2 {
			return argError(name, fmt.Sprintf("expected 2 arguments (value, sign), got %d", len(args)))
		}
		i, oki := args[1].(Integer)
		if !oki {
			return argError(name, fmt.Sprintf("argument 2 (sign) must be an integer, got %s", args[1].Type()))
		}
		if v, ok := args[0].(Float); ok {
			return Bool(fn(float64(v), int(i))), nil
		}
		if v, ok := args[0].(Integer); ok {
			return Bool(fn(float64(v), int(i))), nil
		}
		return argError(name, fmt.Sprintf("argument 1 must be numeric, got %s", args[0].Type()))
	}
}

func mathNan(fn func() float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		return Float(fn()), nil
	}
}

func mathFromFloatToFloat(name string, fn func(float64) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) == 0 {
			return argError(name, "expected a numeric argument")
		}
		if v, ok := args[0].(Float); ok {
			return Float(fn(float64(v))), nil
		}
		if v, ok := args[0].(Integer); ok {
			return Float(fn(float64(v))), nil
		}
		return argError(name, fmt.Sprintf("expected a numeric argument, got %s", args[0].Type()))
	}
}

func mathFrom2FloatsToFloat(name string, fn func(float64, float64) float64) NativeFunction {
	return func(ctx *Context, args ...Value) (Value, error) {
		if len(args) < 2 {
			return argError(name, fmt.Sprintf("expected 2 numeric arguments, got %d", len(args)))
		}
		l, okl := getFloat(args[0])
		if !okl {
			return argError(name, fmt.Sprintf("argument 1 must be numeric, got %s", args[0].Type()))
		}
		r, okr := getFloat(args[1])
		if !okr {
			return argError(name, fmt.Sprintf("argument 2 must be numeric, got %s", args[1].Type()))
		}
		return Float(fn(float64(l), float64(r))), nil
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
		if len(args) < 2 {
			return argError("pow", fmt.Sprintf("expected 2 numeric arguments, got %d", len(args)))
		}
		switch l := args[0].(type) {
		case Integer:
			switch r := args[1].(type) {
			case Integer:
				return Integer(fn(float64(l), float64(r))), nil
			case Float:
				return Float(fn(float64(l), float64(r))), nil
			default:
				return argError("pow", fmt.Sprintf("argument 2 must be numeric, got %s", args[1].Type()))
			}
		case Float:
			switch r := args[1].(type) {
			case Integer:
				return Float(fn(float64(l), float64(r))), nil
			case Float:
				return Float(fn(float64(l), float64(r))), nil
			default:
				return argError("pow", fmt.Sprintf("argument 2 must be numeric, got %s", args[1].Type()))
			}
		default:
			return argError("pow", fmt.Sprintf("argument 1 must be numeric, got %s", args[0].Type()))
		}
	}
}
