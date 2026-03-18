package vida

import (
	"strconv"
)

func loadFoundationCasting() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["toString"] = GFn(castToString)
	m.Value["toInt"] = GFn(castToInt)
	m.Value["toFloat"] = GFn(casttoFloat)
	m.Value["toBool"] = GFn(castToBool)
	m.Value["toArray"] = GFn(castToArray)
	m.Value["toObject"] = GFn(castToObject)
	return m
}

func castToString(args ...Value) (Value, error) {
	if len(args) > 0 {
		return &String{Value: args[0].String()}, nil
	}
	return NilValue, nil
}

func castToInt(args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		switch v := args[0].(type) {
		case *String:
			i, e := strconv.ParseInt(v.Value, 0, 64)
			if e == nil {
				return Integer(i), nil
			}
		case Integer:
			return v, nil
		case Bool:
			if v {
				return Integer(1), nil
			}
			return Integer(0), nil
		case Float:
			return Integer(v), nil
		case Nil:
			return Integer(0), nil
		}
	case 2:
		if v, ok := args[0].(*String); ok {
			if b, ok := args[1].(Integer); ok {
				i, e := strconv.ParseInt(v.Value, int(b), 64)
				if e == nil {
					return Integer(i), nil
				}
			}
		}
	}
	return NilValue, nil
}

func casttoFloat(args ...Value) (Value, error) {
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
		case Nil:
			return Float(0), nil
		case Bool:
			if v {
				return Float(1), nil
			}
			return Float(0), nil
		}
	}
	return NilValue, nil
}

func castToBool(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			switch v.Value {
			case "true":
				return Bool(true), nil
			case "false":
				return Bool(false), nil
			}
		}
		return args[0].Boolean(), nil
	}
	return NilValue, nil
}

func castToArray(args ...Value) (Value, error) {
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
			a := make([]Value, len(v.Value)*2)
			var idx int
			for k, v := range v.Value {
				a[idx] = &String{Value: k}
				idx++
				a[idx] = v
				idx++
			}
			return &Array{Value: a}, nil
		case *Enum:
			a := make([]Value, len(v.Pairs)*2)
			var idx int
			for k, v := range v.Pairs {
				a[idx] = &String{Value: k}
				idx++
				a[idx] = v
				idx++
			}
			return &Array{Value: a}, nil
		case Error:
			a := make([]Value, 2)
			a[0] = &String{Value: errorMessageFieldName}
			a[1] = v.Message
			return &Array{Value: a}, nil
		}
	}
	return NilValue, nil
}

func castToObject(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *Object:
			return v.Clone(), nil
		case *Array:
			o := &Object{Value: make(map[string]Value)}
			for i, v := range v.Value {
				o.Value[Integer(i).String()] = v
			}
			return o, nil
		case *Bytes:
			o := &Object{Value: make(map[string]Value)}
			for i, v := range v.Value {
				o.Value[Integer(i).String()] = Integer(v)
			}
			return o, nil
		case Error:
			o := &Object{Value: make(map[string]Value)}
			o.Value[errorMessageFieldName] = v.Message
			return o, nil
		case *Enum:
			o := &Object{Value: make(map[string]Value)}
			for k, v := range v.Pairs {
				o.Value[k] = v
			}
			return o, nil
		default:
			o := &Object{Value: make(map[string]Value)}
			o.Value["value"] = v
			return o, nil
		}
	}
	return NilValue, nil
}
