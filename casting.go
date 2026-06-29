package vida

import (
	"strconv"
)

func loadFoundationCasting() Value {
	m := &Object{Value: make(map[string]Value, 6)}
	m.Value["toString"] = NativeFunction(castToString)
	m.Value["toInt"] = NativeFunction(castToInt)
	m.Value["toFloat"] = NativeFunction(castToFloat)
	m.Value["toBool"] = NativeFunction(castToBool)
	m.Value["toArray"] = NativeFunction(castToArray)
	m.Value["toObject"] = NativeFunction(castToObject)
	return m
}

func castToNumber(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 2:
		inputStr, okInput := args[0].(*String)
		typeOf, okType := args[1].(*String)
		if okInput && okType {
			switch typeOf.Value {
			case integerT:
				return castToInt(ctx, inputStr)
			case floatT:
				return castToFloat(ctx, inputStr)
			}
		}
	case 3:
		inputStr, okInput := args[0].(*String)
		typeOf, okType := args[1].(*String)
		base, ok := args[2].(Integer)
		if ok && okType && okInput && base == 0 || (2 <= base && base <= 36) {
			switch typeOf.Value {
			case integerT:
				return castToInt(ctx, inputStr, base)
			case floatT:
				return castToFloat(ctx, inputStr)
			}
		}
	}
	return &VidaError{Message: &String{Value: "error is toNum: expected type, str or type, str, base"}}, nil
}

func castToString(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		return &String{Value: args[0].String()}, nil
	}
	return &String{Value: EmptyString}, nil
}

func castToInt(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		switch v := args[0].(type) {
		case *String:
			i, e := strconv.ParseInt(v.Value, 0, 64)
			if e == nil {
				return Integer(i), nil
			}
			f, e := strconv.ParseFloat(v.Value, 64)
			if e == nil {
				return Integer(f), nil
			}
			return Nil, nil
		case Integer:
			return v, nil
		case Bool:
			if v {
				return Integer(1), nil
			}
			return Integer(0), nil
		case Float:
			return Integer(int64(v)), nil
		case NilValue:
			return Integer(0), nil
		}
	case 2:
		if v, ok := args[0].(*String); ok {
			if base, ok := args[1].(Integer); ok && base == 0 || (2 <= base && base <= 36) {
				i, e := strconv.ParseInt(v.Value, int(base), 64)
				if e == nil {
					return Integer(i), nil
				}
			}
		}
	}
	return Integer(0), nil
}

func castToFloat(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *String:
			r, e := strconv.ParseFloat(v.Value, 64)
			if e == nil {
				return Float(r), nil
			}
		case Integer:
			return Float(v), nil
		case Float:
			return v, nil
		case NilValue:
			return Float(0), nil
		case Bool:
			if v {
				return Float(1), nil
			}
			return Float(0), nil
		}
	}
	return Float(0), nil
}

func castToBool(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			switch v.Value {
			case "true":
				return True, nil
			case "false":
				return False, nil
			}
		}
		return args[0].Boolean(), nil
	}
	return False, nil
}

func castToArray(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *Array:
			return v.Clone(), nil
		case *Bytes:
			a := make([]Value, len(v.Value))
			for i, b := range v.Value {
				a[i] = Integer(b)
			}
			return &Array{Value: a}, nil
		case *Object:
			pairs := make([]Value, len(v.Value))
			i := 0
			for k, val := range v.Value {
				pairs[i] = &Array{Value: []Value{&String{Value: k}, val}}
				i++
			}
			return &Array{Value: pairs}, nil
		case *Enum:
			pairs := make([]Value, len(v.Pairs))
			i := 0
			for k, val := range v.Pairs {
				pairs[i] = &Array{Value: []Value{&String{Value: k}, val}}
				i++
			}
			return &Array{Value: pairs}, nil
		case *VidaError:
			a := make([]Value, 1)
			a[0] = v.Message
			return &Array{Value: a}, nil
		}
	}
	return &Array{Value: make([]Value, 0)}, nil
}

func castToObject(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		switch val := args[0].(type) {
		case *Object:
			return val.Clone(), nil
		case *Array:
			o := &Object{Value: make(map[string]Value, len(val.Value))}
			for i, v := range val.Value {
				o.Value[Integer(i).ObjectKey()] = v
			}
			return o, nil
		case *Bytes:
			o := &Object{Value: make(map[string]Value, len(val.Value))}
			for i, v := range val.Value {
				o.Value[Integer(i).ObjectKey()] = Integer(v)
			}
			return o, nil
		case *VidaError:
			o := &Object{Value: make(map[string]Value, 1)}
			o.Value[errorMessageFieldName] = val.Message
			return o, nil
		case *Enum:
			o := &Object{Value: make(map[string]Value, len(val.Pairs))}
			for k, v := range val.Pairs {
				o.Value[k] = v
			}
			return o, nil
		default:
			o := &Object{Value: make(map[string]Value, 4)}
			o.Value[DefaultValField] = val
			return o, nil
		}
	}
	return &Object{Value: make(map[string]Value)}, nil
}
