package vida

import "regexp"

func loadFoundationRegexp() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["match"] = GFn(regexpMatch)
	m.Value["replaceAll"] = GFn(regexpReplaceAll)
	m.Value["findString"] = GFn(regexpFindString)
	m.Value["findAllStrings"] = GFn(regexpFindAllStrings)
	m.Value["findFirstIndex"] = GFn(regexpFindFirstIndex)
	m.Value["escape"] = GFn(regexpEscape)
	return m
}

func regexpMatch(args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &Error{Message: &String{Value: err.Error()}}, nil
			}
			return Bool(re.MatchString(input.Value)), nil
		}
	}
	return NilValue, nil
}

func regexpReplaceAll(args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		source, okIn := args[1].(*String)
		replacement, okRepl := args[2].(*String)
		if okPatt && okIn && okRepl {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &Error{Message: &String{Value: err.Error()}}, nil
			}
			return &String{Value: re.ReplaceAllString(source.Value, replacement.Value)}, nil
		}
	}
	return NilValue, nil
}

func regexpFindFirstIndex(args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &Error{Message: &String{Value: err.Error()}}, nil
			}
			res := re.FindStringIndex(input.Value)
			if res == nil {
				return NilValue, nil
			}
			arr := &Array{Value: make([]Value, 2)}
			arr.Value[0] = Integer(res[0])
			arr.Value[1] = Integer(res[1])
			return arr, nil
		}
	}
	return NilValue, nil
}

func regexpEscape(args ...Value) (Value, error) {
	if len(args) > 0 {
		if input, ok := args[0].(*String); ok {
			return &String{Value: regexp.QuoteMeta(input.Value)}, nil
		}
	}
	return NilValue, nil
}

func regexpFindString(args ...Value) (Value, error) {
	if len(args) > 1 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		if okPatt && okIn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &Error{Message: &String{Value: err.Error()}}, nil
			}
			return &String{Value: re.FindString(input.Value)}, nil
		}
	}
	return NilValue, nil
}

func regexpFindAllStrings(args ...Value) (Value, error) {
	if len(args) > 2 {
		pattern, okPatt := args[0].(*String)
		input, okIn := args[1].(*String)
		n, okn := args[2].(Integer)
		if okPatt && okIn && okn {
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return &Error{Message: &String{Value: err.Error()}}, nil
			}
			result := re.FindAllString(input.Value, int(n))
			arr := &Array{Value: make([]Value, len(result))}
			for i, v := range result {
				arr.Value[i] = &String{Value: v}
			}
			return arr, nil
		}
	}
	return NilValue, nil
}
