package vida

import "github.com/alkemist-17/vida/verror"

func loadFoundationBytes() Value {
	m := &Object{Value: make(map[string]Value, 2)}
	m.Value["new"] = GFn(bytesCreateNewBytesBuffer)
	m.Value["from"] = GFn(bytesFromValue)
	return m
}

func bytesCreateNewBytesBuffer(args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		switch v := args[0].(type) {
		case Integer:
			var init byte = 0
			if l > 1 {
				if val, ok := args[1].(Integer); ok {
					init = byte(val)
				}
			}
			if v > 0 && v < verror.MaxMemSize {
				b := make([]byte, v)
				for i := range v {
					b[i] = init
				}
				return &Bytes{Value: b}, nil
			}
		case *String:
			return &Bytes{Value: []byte(v.Value)}, nil
		case *Bytes:
			return v.Clone(), nil
		case *Array:
			var bts []byte
			for _, val := range v.Value {
				if i, ok := val.(Integer); ok {
					bts = append(bts, byte(i))
				}
			}
			return &Bytes{Value: bts}, nil
		}
	}
	return &Bytes{}, nil
}

func bytesFromValue(args ...Value) (Value, error) {
	if len(args) > 0 {
		switch v := args[0].(type) {
		case *String:
			return &Bytes{Value: []byte(v.Value)}, nil
		case *Bytes:
			return v.Clone(), nil
		case *Array:
			var bts []byte
			for _, val := range v.Value {
				if i, ok := val.(Integer); ok {
					bts = append(bts, byte(i))
				}
			}
			return &Bytes{Value: bts}, nil
		}
	}
	return &Bytes{}, nil
}
