package vida

import "regexp"

func loadFoundationRegexp() Value {
	m := &Object{Value: make(map[string]Value, 10)}
	m.Value["match"] = NativeFunction(regexpMatch)
	m.Value["replaceAll"] = NativeFunction(regexpReplaceAll)
	m.Value["replaceAllLiteral"] = NativeFunction(regexpReplaceAllLit)
	m.Value["find"] = NativeFunction(regexpFindString)
	m.Value["findAll"] = NativeFunction(regexpFindAllString)
	m.Value["findFirstIndex"] = NativeFunction(regexpFindFirstIndex)
	m.Value["findAllIndex"] = NativeFunction(regexpFindAllIndex)
	m.Value["findSubMatch"] = NativeFunction(regexpFindSubmatch)
	m.Value["split"] = NativeFunction(regexpSplit)
	m.Value["escape"] = NativeFunction(regexpEscape)
	return m
}

func regexpMatch(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(re.MatchString(input.Value)), nil
		}
	}
	return Nil, nil
}

func regexpReplaceAll(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		source, okIn := args[1].(*String)
		replacement, okRepl := args[2].(*String)
		if okPatt && okIn && okRepl {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &String{Value: re.ReplaceAllString(source.Value, replacement.Value)}, nil
		}
	}
	return Nil, nil
}

func regexpReplaceAllLit(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		source, okIn := args[1].(*String)
		replacement, okRepl := args[2].(*String)
		if okPatt && okIn && okRepl {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &String{Value: re.ReplaceAllLiteralString(source.Value, replacement.Value)}, nil
		}
	}
	return Nil, nil
}

func regexpSplit(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		n, okn := args[2].(Integer)
		if okPatt && okIn && okn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			result := re.Split(input.Value, int(n))
			if result == nil {
				return Nil, nil
			}
			arr := &Array{Value: make([]Value, len(result))}
			for i, v := range result {
				arr.Value[i] = &String{Value: v}
			}
			return arr, nil
		}
	}
	return Nil, nil
}

func regexpFindFirstIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			res := re.FindStringIndex(input.Value)
			if res == nil {
				return Nil, nil
			}
			arr := &Array{Value: make([]Value, 2)}
			arr.Value[0] = Integer(res[0])
			arr.Value[1] = Integer(res[1])
			return arr, nil
		}
	}
	return Nil, nil
}

func regexpFindAllIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		n, okn := args[2].(Integer)
		if okPatt && okIn && okn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			result := re.FindAllStringIndex(input.Value, int(n))
			if result == nil {
				return Nil, nil
			}
			arr := &Array{Value: make([]Value, len(result))}
			for i, v := range result {
				a := make([]Value, len(v))
				for k, w := range v {
					a[k] = Integer(w)
				}
				arr.Value[i] = &Array{Value: a}
			}
			return arr, nil
		}
	}
	return Nil, nil
}

func regexpEscape(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if input, ok := args[0].(*String); ok {
			return &String{Value: regexp.QuoteMeta(input.Value)}, nil
		}
	}
	return Nil, nil
}

func regexpFindString(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			return &String{Value: re.FindString(input.Value)}, nil
		}
	}
	return Nil, nil
}

func regexpFindAllString(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		n, okn := args[2].(Integer)
		if okPatt && okIn && okn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			result := re.FindAllString(input.Value, int(n))
			if result == nil {
				return Nil, nil
			}
			arr := &Array{Value: make([]Value, len(result))}
			for i, v := range result {
				arr.Value[i] = &String{Value: v}
			}
			return arr, nil
		}
	}
	return Nil, nil
}

func regexpFindSubmatch(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &VidaError{Message: &String{Value: err.Error()}}, nil
			}
			result := re.FindStringSubmatch(input.Value)
			if result == nil {
				return Nil, nil
			}
			arr := &Array{Value: make([]Value, len(result))}
			for i, v := range result {
				arr.Value[i] = &String{Value: v}
			}
			return arr, nil
		}
	}
	return Nil, nil
}
