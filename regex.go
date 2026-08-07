package vida

import (
	"fmt"
	"regexp"
)

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
	if len(args) < 2 {
		return argError("re.match", fmt.Sprintf("expected 2 arguments (pattern, string), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.match", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.match", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Bool(re.MatchString(input.Value)), nil
}

func regexpReplaceAll(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("re.replaceAll", fmt.Sprintf("expected 3 arguments (pattern, string, replacement), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.replaceAll", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	source, okIn := args[1].(*String)
	if !okIn {
		return argError("re.replaceAll", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	replacement, okRepl := args[2].(*String)
	if !okRepl {
		return argError("re.replaceAll", fmt.Sprintf("argument 3 (replacement) must be a string, got %s", args[2].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: re.ReplaceAllString(source.Value, replacement.Value)}, nil
}

func regexpReplaceAllLit(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("re.replaceAllLiteral", fmt.Sprintf("expected 3 arguments (pattern, string, replacement), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.replaceAllLiteral", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	source, okIn := args[1].(*String)
	if !okIn {
		return argError("re.replaceAllLiteral", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	replacement, okRepl := args[2].(*String)
	if !okRepl {
		return argError("re.replaceAllLiteral", fmt.Sprintf("argument 3 (replacement) must be a string, got %s", args[2].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: re.ReplaceAllLiteralString(source.Value, replacement.Value)}, nil
}

func regexpSplit(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("re.split", fmt.Sprintf("expected 3 arguments (pattern, string, n), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.split", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.split", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	n, okn := args[2].(Integer)
	if !okn {
		return argError("re.split", fmt.Sprintf("argument 3 (n) must be an integer, got %s", args[2].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	result := re.Split(input.Value, int(n))
	arr := &Array{Value: make([]Value, len(result))}
	for i, v := range result {
		arr.Value[i] = &String{Value: v}
	}
	return arr, nil
}

func regexpFindFirstIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("re.findFirstIndex", fmt.Sprintf("expected 2 arguments (pattern, string), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.findFirstIndex", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.findFirstIndex", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	res := re.FindStringIndex(input.Value)
	if res == nil {
		return Nil, nil // no match is a valid result, not an error
	}
	arr := &Array{Value: make([]Value, 2)}
	arr.Value[0] = Integer(res[0])
	arr.Value[1] = Integer(res[1])
	return arr, nil
}

func regexpFindAllIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("re.findAllIndex", fmt.Sprintf("expected 3 arguments (pattern, string, n), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.findAllIndex", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.findAllIndex", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	n, okn := args[2].(Integer)
	if !okn {
		return argError("re.findAllIndex", fmt.Sprintf("argument 3 (n) must be an integer, got %s", args[2].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	result := re.FindAllStringIndex(input.Value, int(n))
	if result == nil {
		return Nil, nil // no match is a valid result, not an error
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

func regexpEscape(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("re.escape", "expected a string argument")
	}
	input, ok := args[0].(*String)
	if !ok {
		return argError("re.escape", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	return &String{Value: regexp.QuoteMeta(input.Value)}, nil
}

func regexpFindString(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("re.find", fmt.Sprintf("expected 2 arguments (pattern, string), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.find", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.find", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: re.FindString(input.Value)}, nil
}

func regexpFindAllString(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("re.findAll", fmt.Sprintf("expected 3 arguments (pattern, string, n), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.findAll", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.findAll", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	n, okn := args[2].(Integer)
	if !okn {
		return argError("re.findAll", fmt.Sprintf("argument 3 (n) must be an integer, got %s", args[2].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	result := re.FindAllString(input.Value, int(n))
	if result == nil {
		return Nil, nil // no match is a valid result, not an error
	}
	arr := &Array{Value: make([]Value, len(result))}
	for i, v := range result {
		arr.Value[i] = &String{Value: v}
	}
	return arr, nil
}

func regexpFindSubmatch(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("re.findSubMatch", fmt.Sprintf("expected 2 arguments (pattern, string), got %d", len(args)))
	}
	pattern, okPatt := args[0].(*String)
	if !okPatt {
		return argError("re.findSubMatch", fmt.Sprintf("argument 1 (pattern) must be a string, got %s", args[0].Type()))
	}
	input, okIn := args[1].(*String)
	if !okIn {
		return argError("re.findSubMatch", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	result := re.FindStringSubmatch(input.Value)
	if result == nil {
		return Nil, nil // no match is a valid result, not an error
	}
	arr := &Array{Value: make([]Value, len(result))}
	for i, v := range result {
		arr.Value[i] = &String{Value: v}
	}
	return arr, nil
}
