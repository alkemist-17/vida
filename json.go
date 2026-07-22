package vida

import (
	"encoding/json"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationJSON() Value {
	m := &Object{Value: make(map[string]Value, 6)}
	m.Value["stringify"] = NativeFunction(jsonValueToJsonString)
	m.Value["parse"] = NativeFunction(jsonParse)
	m.Value["encode"] = NativeFunction(jsonValueToJsonBytes)
	m.Value["decode"] = NativeFunction(jsonDecode)
	m.Value["isValid"] = NativeFunction(jsonIsValid)
	m.Value["pretty"] = NativeFunction(jsonPretty)
	return m
}

func jsonValueToJsonString(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, err := json.Marshal(args[0]); err == nil {
			return &String{Value: string(b)}, nil
		} else {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	}
	b, _ := json.Marshal(Nil)
	return &String{Value: string(b)}, nil
}

func jsonValueToJsonBytes(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, err := json.Marshal(args[0]); err == nil {
			return &Bytes{Value: b}, nil
		} else {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	}
	b, _ := json.Marshal(Nil)
	return &Bytes{Value: b}, nil
}

func jsonParse(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		valid, _ := jsonIsValid(ctx, args[0])
		switch t := valid.(type) {
		case NilValue:
			return &VidaError{Message: &String{Value: verror.ErrInvalidJSON.Error()}}, nil
		case Bool:
			if !t {
				return &VidaError{Message: &String{Value: verror.ErrInvalidJSON.Error()}}, nil
			}
		}
		switch t := args[0].(type) {
		case *String:
			var value any
			if err := json.Unmarshal([]byte(t.Value), &value); err == nil {
				switch v := value.(type) {
				case nil:
					return Nil, nil
				case bool:
					return Bool(v), nil
				case string:
					return &String{Value: v}, nil
				case float64:
					return Float(v), nil
				case map[string]any:
					return parseObject(ctx, v), nil
				case []any:
					return parseArray(ctx, v), nil
				}
			} else {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
		case *Bytes:
			var value any
			if err := json.Unmarshal(t.Value, &value); err == nil {
				switch v := value.(type) {
				case nil:
					return Nil, nil
				case bool:
					return Bool(v), nil
				case string:
					return &String{Value: v}, nil
				case float64:
					return Float(v), nil
				case map[string]any:
					return parseObject(ctx, v), nil
				case []any:
					return parseArray(ctx, v), nil
				}
			} else {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
		}
	}
	return Nil, nil
}

func jsonDecode(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		valid, _ := jsonIsValid(ctx, args[0])
		switch t := valid.(type) {
		case NilValue:
			return &VidaError{Message: &String{Value: verror.ErrInvalidJSON.Error()}}, nil
		case Bool:
			if !t {
				return &VidaError{Message: &String{Value: verror.ErrInvalidJSON.Error()}}, nil
			}
		}
		switch t := args[0].(type) {
		case *Bytes:
			var value any
			if err := json.Unmarshal(t.Value, &value); err == nil {
				switch v := value.(type) {
				case nil:
					return Nil, nil
				case bool:
					return Bool(v), nil
				case string:
					return &String{Value: v}, nil
				case float64:
					return Float(v), nil
				case map[string]any:
					return parseObject(ctx, v), nil
				case []any:
					return parseArray(ctx, v), nil
				}
			} else {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
		}
	}
	return Nil, nil
}

func jsonIsValid(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		switch t := args[0].(type) {
		case *String:
			return Bool(json.Valid([]byte(t.Value))), nil
		case *Bytes:
			return Bool(json.Valid(t.Value)), nil
		default:
			return False, nil
		}
	}
	return Nil, nil
}

func jsonPretty(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if b, err := json.MarshalIndent(args[0], EmptyString, "  "); err == nil {
			return &String{Value: string(b)}, nil
		} else {
			return &VidaError{Message: &String{Value: err.Error()}}, nil
		}
	}
	return Nil, nil
}

func parseObject(ctx *Context, input map[string]any) *Object {
	o := &Object{Value: make(map[string]Value, len(input))}
	for kk, vv := range input {
		switch tt := vv.(type) {
		case nil:
			o.Value[kk] = Nil
		case bool:
			o.Value[kk] = Bool(tt)
		case string:
			o.Value[kk] = &String{Value: tt}
		case float64:
			o.Value[kk] = Float(tt)
		case map[string]any:
			o.Value[kk] = parseObject(ctx, tt)
		case []any:
			o.Value[kk] = parseArray(ctx, tt)
		}
	}
	return o
}

func parseArray(ctx *Context, input []any) *Array {
	A := &Array{Value: make([]Value, len(input))}
	for ii, vv := range input {
		switch tt := vv.(type) {
		case nil:
			A.Value[ii] = Nil
		case bool:
			A.Value[ii] = Bool(tt)
		case string:
			A.Value[ii] = &String{Value: tt}
		case float64:
			A.Value[ii] = Float(tt)
		case map[string]any:
			A.Value[ii] = parseObject(ctx, tt)
		case []any:
			A.Value[ii] = parseArray(ctx, tt)
		}
	}
	return A
}
