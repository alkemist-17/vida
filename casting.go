package vida

import "strconv"

func loadFoundationCasting() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["toString"] = GFn(toString)
	m.Value["toInt"] = GFn(toInt)
	m.Value["toFloat"] = GFn(toFloat)
	m.Value["toBool"] = GFn(toBool)
	return m
}

func toString(args ...Value) (Value, error) {
	if len(args) > 0 {
		return &String{Value: args[0].String()}, nil
	}
	return NilValue, nil
}

func toInt(args ...Value) (Value, error) {
	l := len(args)
	switch l {
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

func toFloat(args ...Value) (Value, error) {
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

func toBool(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			if v.Value == "true" {
				return Bool(true), nil
			}
			if v.Value == "false" {
				return Bool(false), nil
			}
		}
	}
	return NilValue, nil
}
