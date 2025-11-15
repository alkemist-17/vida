package vida

import (
	"encoding/json"
	"errors"
)

func loadFoundationJSON() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["stringify"] = GFn(jsonStringEncoding)
	m.Value["parse"] = GFn(jsonParse)
	m.Value["isValid"] = GFn(jsonIsValid)
	m.Value["pretty"] = GFn(jsonPretty)
	return m
}

func jsonStringEncoding(args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, err := json.Marshal(args[0]); err == nil {
			return &String{Value: string(b)}, nil
		} else {
			return NilValue, err
		}
	}
	b, _ := json.Marshal(NilValue)
	return &String{Value: string(b)}, nil
}

func jsonParse(args ...Value) (Value, error) {
	if len(args) > 0 {
		valid, _ := jsonIsValid(args[0])
		switch t := valid.(type) {
		case Nil:
			return NilValue, errors.New("invalid json")
		case Bool:
			if !t {
				return NilValue, errors.New("invalid json")
			}
		}
		switch t := args[0].(type) {
		case *String:
			var value any
			if err := json.Unmarshal([]byte(t.Value), &value); err == nil {
				switch v := value.(type) {
				case nil:
					return NilValue, nil
				case bool:
					return Bool(v), nil
				case string:
					return &String{Value: v}, nil
				case float64:
					return Float(v), nil
				case map[string]any:
					return parseObject(v), nil
				case []any:
					return parseArray(v), nil
				}
			} else {
				return NilValue, err
			}
		}
	}
	return NilValue, nil
}

func jsonIsValid(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch t := args[0].(type) {
		case *String:
			return Bool(json.Valid([]byte(t.Value))), nil
		case *Bytes:
			return Bool(json.Valid(t.Value)), nil
		}
	}
	return NilValue, nil
}

func jsonPretty(args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, err := json.MarshalIndent(args[0], "", "  "); err == nil {
			return &String{Value: string(b)}, nil
		} else {
			return NilValue, err
		}
	}
	return NilValue, nil
}

func parseObject(input map[string]any) *Object {
	o := &Object{Value: make(map[string]Value)}
	for kk, vv := range input {
		switch tt := vv.(type) {
		case nil:
			o.Value[kk] = NilValue
		case bool:
			o.Value[kk] = Bool(tt)
		case string:
			o.Value[kk] = &String{Value: tt}
		case float64:
			o.Value[kk] = Float(tt)
		case map[string]any:
			o.Value[kk] = parseObject(tt)
		case []any:
			o.Value[kk] = parseArray(tt)
		}
	}
	return o
}

func parseArray(input []any) *Array {
	A := &Array{Value: make([]Value, len(input))}
	for ii, vv := range input {
		switch tt := vv.(type) {
		case nil:
			A.Value[ii] = NilValue
		case bool:
			A.Value[ii] = Bool(tt)
		case string:
			A.Value[ii] = &String{Value: tt}
		case float64:
			A.Value[ii] = Float(tt)
		case map[string]any:
			A.Value[ii] = parseObject(tt)
		case []any:
			A.Value[ii] = parseArray(tt)
		}
	}
	return A
}
