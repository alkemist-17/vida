package vida

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func loadFoundationText() Value {
	m := &Object{Value: make(map[string]Value, 43)}
	m.Value["randomElement"] = NativeFunction(arrayRandomElement)
	m.Value["hasPrefix"] = NativeFunction(textHasPrefix)
	m.Value["hasSuffix"] = NativeFunction(textHasSuffix)
	m.Value["fromCodePoints"] = NativeFunction(textFromCodepoints)
	m.Value["trim"] = NativeFunction(textTrim)
	m.Value["trimLeft"] = NativeFunction(textTrimLeft)
	m.Value["trimRight"] = NativeFunction(textTrimRight)
	m.Value["split"] = NativeFunction(textSplit)
	m.Value["fields"] = NativeFunction(textFields)
	m.Value["repeat"] = NativeFunction(textRepeat)
	m.Value["replaceN"] = NativeFunction(textReplaceN)
	m.Value["replaceAll"] = NativeFunction(textReplaceAll)
	m.Value["center"] = NativeFunction(textCenter)
	m.Value["contains"] = NativeFunction(textContains)
	m.Value["containsAny"] = NativeFunction(textContainsAny)
	m.Value["index"] = NativeFunction(textIndex)
	m.Value["join"] = NativeFunction(textJoin)
	m.Value["toLower"] = NativeFunction(textToLowerCase)
	m.Value["toUpper"] = NativeFunction(textToUpperCase)
	m.Value["count"] = NativeFunction(textCount)
	m.Value["isAscii"] = NativeFunction(textIsAscii)
	m.Value["isDecimal"] = NativeFunction(textIsDecimal)
	m.Value["isDigit"] = NativeFunction(textIsDigit)
	m.Value["isHexDigit"] = NativeFunction(textIsHexDigit)
	m.Value["isLetter"] = NativeFunction(textIsLetter)
	m.Value["isNumber"] = NativeFunction(textIsNumber)
	m.Value["isSpace"] = NativeFunction(textIsSpace)
	m.Value["isSpaceChar"] = NativeFunction(textIsSpaceChar)
	m.Value["codePoints"] = NativeFunction(textCodepoints)
	m.Value["bytesLen"] = NativeFunction(textBytesLen)
	m.Value["equalFold"] = NativeFunction(textEqualFold)
	m.Value["capitalize"] = NativeFunction(textCapitalize)
	m.Value["padLeft"] = NativeFunction(textPadLeft)
	m.Value["padRight"] = NativeFunction(textPadRight)
	m.Value["lines"] = NativeFunction(textLines)
	m.Value["truncate"] = NativeFunction(textTruncate)
	m.Value["wrap"] = NativeFunction(textWrap)
	m.Value["slugify"] = NativeFunction(textSlugify)
	m.Value["startsWithAny"] = NativeFunction(textStartsWithAny)
	m.Value["endsWithAny"] = NativeFunction(textEndsWithAny)
	m.Value["compare"] = NativeFunction(textCompare)
	m.Value["urlEncode"] = NativeFunction(textUrlEncode)
	m.Value["urlDecode"] = NativeFunction(textUrlDecode)
	m.Value["toInterpreted"] = NativeFunction(textToInterpreted)
	return m
}

func textMatch(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("match", fmt.Sprintf("expected 2 arguments (string, pattern), got %d", len(args)))
	}
	input, okIn := args[0].(*String)
	if !okIn {
		return argError("match", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	pattern, okPatt := args[1].(*String)
	if !okPatt {
		return argError("match", fmt.Sprintf("argument 2 (pattern) must be a string, got %s", args[1].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Bool(re.MatchString(input.Value)), nil
}

func textFindFirstIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("findFirstIndex", fmt.Sprintf("expected 2 arguments (string, pattern), got %d", len(args)))
	}
	input, okIn := args[0].(*String)
	if !okIn {
		return argError("findFirstIndex", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	pattern, okPatt := args[1].(*String)
	if !okPatt {
		return argError("findFirstIndex", fmt.Sprintf("argument 2 (pattern) must be a string, got %s", args[1].Type()))
	}
	re, err := regexp.Compile(pattern.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	res := re.FindAllStringIndex(input.Value, -1)
	if res == nil {
		return &Array{}, nil
	}
	arr := &Array{Value: make([]Value, 0, len(res))}
	for _, v := range res {
		idx := &Array{Value: make([]Value, 2)}
		idx.Value[0] = Integer(v[0])
		idx.Value[1] = Integer(v[1])
		arr.Value = append(arr.Value, idx)
	}
	return arr, nil
}

func textIsEmpty(ctx *Context, args ...Value) (Value, error) {
	l, err := coreLen(ctx, args...)
	if err != nil {
		return argError("isEmpty", err.Error())
	}
	v, ok := l.(Integer)
	if !ok {
		return argError("isEmpty", "expected an argument with a length (string, array, object, or bytes)")
	}
	return Bool(v == 0), nil
}

func textEscapeUnescapedQuotes(s string) string {
	var sb strings.Builder
	backslashes := 0
	for _, r := range s {
		if r == '"' {
			if backslashes%2 == 0 {
				sb.WriteByte('\\')
			}
			sb.WriteRune(r)
			backslashes = 0
			continue
		}
		if r == '\\' {
			backslashes++
		} else {
			backslashes = 0
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func textToInterpreted(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("toInterpreted", "expected a string argument")
	}
	val, ok := args[0].(*String)
	if !ok {
		return argError("toInterpreted", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	s := val.Value
	res, err := strconv.Unquote(`"` + textEscapeUnescapedQuotes(s) + `"`)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return &String{Value: res}, nil
}

func textHasPrefix(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("hasPrefix", fmt.Sprintf("expected 2 arguments (string, prefix), got %d", len(args)))
	}
	v, okV := args[0].(*String)
	if !okV {
		return argError("hasPrefix", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	p, okP := args[1].(*String)
	if !okP {
		return argError("hasPrefix", fmt.Sprintf("argument 2 (prefix) must be a string, got %s", args[1].Type()))
	}
	return Bool(strings.HasPrefix(v.Value, p.Value)), nil
}

func textHasSuffix(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("hasSuffix", fmt.Sprintf("expected 2 arguments (string, suffix), got %d", len(args)))
	}
	v, okV := args[0].(*String)
	if !okV {
		return argError("hasSuffix", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	p, okP := args[1].(*String)
	if !okP {
		return argError("hasSuffix", fmt.Sprintf("argument 2 (suffix) must be a string, got %s", args[1].Type()))
	}
	return Bool(strings.HasSuffix(v.Value, p.Value)), nil
}

func textFromCodepoints(ctx *Context, args ...Value) (Value, error) {
	runes := make([]rune, 0, len(args))
	for i, a := range args {
		v, ok := a.(Integer)
		if !ok {
			return argError("fromCodePoints", fmt.Sprintf("argument %d must be an integer code point, got %s", i+1, a.Type()))
		}
		if v < 0 || v > utf8.MaxRune {
			return argError("fromCodePoints", fmt.Sprintf("argument %d (%d) is not a valid Unicode code point", i+1, v))
		}
		runes = append(runes, rune(v))
	}
	// NFKC - Compatibility Decomposition, followed by Canonical Composition
	// for compatibility with Vida approach to codepoints.
	normalized := norm.NFKC.String(string(runes))
	return &String{Value: normalized, Runes: []rune(normalized)}, nil
}

func textTrim(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return argError("trim", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("trim", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	if l > 1 {
		p, ok := args[1].(*String)
		if !ok {
			return argError("trim", fmt.Sprintf("argument 2 (cutset) must be a string, got %s", args[1].Type()))
		}
		return &String{Value: strings.Trim(v.Value, p.Value)}, nil
	}
	return &String{Value: strings.Trim(v.Value, " ")}, nil
}

func textExtendedTrim(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 1:
		v, ok := args[0].(*String)
		if !ok {
			return argError("trim", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
		}
		return &String{Value: strings.Trim(v.Value, " ")}, nil
	case 2:
		val, ok := args[0].(*String)
		if !ok {
			return argError("trim", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
		}
		switch p := args[1].(type) {
		case *String:
			return &String{Value: strings.Trim(val.Value, p.Value)}, nil
		case *Object:
			target, ok := p.Value["target"].(*String)
			if !ok {
				return &VidaError{Message: &String{Value: "'target' is a required property of type string for string.trim config object"}}, nil
			}
			switch target.Value {
			case "left":
				if cutset, ok := p.Value["cutset"].(*String); ok {
					return &String{Value: strings.TrimLeft(val.Value, cutset.Value)}, nil
				}
				return &String{Value: strings.TrimLeft(val.Value, " ")}, nil
			case "right":
				if cutset, ok := p.Value["cutset"].(*String); ok {
					return &String{Value: strings.TrimRight(val.Value, cutset.Value)}, nil
				}
				return &String{Value: strings.TrimRight(val.Value, " ")}, nil
			default:
				return &VidaError{Message: &String{Value: "'target' should have 'left' or 'right' values for string.trim config object"}}, nil
			}
		default:
			return argError("trim", fmt.Sprintf("argument 2 must be a string or a config object, got %s", args[1].Type()))
		}
	default:
		return argError("trim", fmt.Sprintf("expected 1 or 2 arguments, got %d", len(args)))
	}
}

func textTrimLeft(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return argError("trimLeft", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("trimLeft", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	if l > 1 {
		p, ok := args[1].(*String)
		if !ok {
			return argError("trimLeft", fmt.Sprintf("argument 2 (cutset) must be a string, got %s", args[1].Type()))
		}
		return &String{Value: strings.TrimLeft(v.Value, p.Value)}, nil
	}
	return &String{Value: strings.TrimLeft(v.Value, " ")}, nil
}

func textTrimRight(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return argError("trimRight", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("trimRight", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	if l > 1 {
		p, ok := args[1].(*String)
		if !ok {
			return argError("trimRight", fmt.Sprintf("argument 2 (cutset) must be a string, got %s", args[1].Type()))
		}
		return &String{Value: strings.TrimRight(v.Value, p.Value)}, nil
	}
	return &String{Value: strings.TrimRight(v.Value, " ")}, nil
}

func textSplit(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return argError("split", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("split", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	if l > 1 {
		sep, ok := args[1].(*String)
		if !ok {
			return argError("split", fmt.Sprintf("argument 2 (separator) must be a string, got %s", args[1].Type()))
		}
		return textStringToArray(strings.Split(v.Value, sep.Value)), nil
	}
	return textStringToArray(strings.Split(v.Value, " ")), nil
}

func textFields(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("fields", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("fields", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	return textStringToArray(strings.Fields(v.Value)), nil
}

func textRepeat(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("repeat", fmt.Sprintf("expected 2 arguments (string, count), got %d", len(args)))
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("repeat", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	times, ok := args[1].(Integer)
	if !ok {
		return argError("repeat", fmt.Sprintf("argument 2 (count) must be an integer, got %s", args[1].Type()))
	}
	if times < 0 {
		return argError("repeat", fmt.Sprintf("count must be non-negative, got %d", times))
	}
	if StringLength(v)*times > MaxMemSize {
		return argError("repeat", "result would exceed the maximum allowed string size")
	}
	return &String{Value: strings.Repeat(v.Value, int(times))}, nil
}

func textReplaceN(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 3 {
		return textReplaceAll(ctx, args...)
	}
	if len(args) < 4 {
		return argError("replaceN", fmt.Sprintf("expected 4 arguments (string, old, new, n), got %d", len(args)))
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("replaceN", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	old, ok := args[1].(*String)
	if !ok {
		return argError("replaceN", fmt.Sprintf("argument 2 (old) must be a string, got %s", args[1].Type()))
	}
	nnew, ok := args[2].(*String)
	if !ok {
		return argError("replaceN", fmt.Sprintf("argument 3 (new) must be a string, got %s", args[2].Type()))
	}
	k, ok := args[3].(Integer)
	if !ok {
		return argError("replaceN", fmt.Sprintf("argument 4 (n) must be an integer, got %s", args[3].Type()))
	}
	return &String{Value: strings.Replace(s.Value, old.Value, nnew.Value, int(k))}, nil
}

func textReplaceAll(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return argError("replaceAll", fmt.Sprintf("expected 3 arguments (string, old, new), got %d", len(args)))
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("replaceAll", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	old, ok := args[1].(*String)
	if !ok {
		return argError("replaceAll", fmt.Sprintf("argument 2 (old) must be a string, got %s", args[1].Type()))
	}
	nnew, ok := args[2].(*String)
	if !ok {
		return argError("replaceAll", fmt.Sprintf("argument 3 (new) must be a string, got %s", args[2].Type()))
	}
	return &String{Value: strings.ReplaceAll(s.Value, old.Value, nnew.Value)}, nil
}

func textCenterString(s *String, width int, sep string) *String {
	strLen := int(StringLength(s))
	if width <= strLen {
		return s
	}
	padding := width - strLen
	leftPad := padding / 2
	rightPad := padding - leftPad
	return &String{Value: strings.Repeat(sep, leftPad) + s.Value + strings.Repeat(sep, rightPad)}
}

func textCenter(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l < 2 {
		return argError("center", fmt.Sprintf("expected at least 2 arguments (string, width), got %d", l))
	}
	str, ok := args[0].(*String)
	if !ok {
		return argError("center", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	width, ok := args[1].(Integer)
	if !ok {
		return argError("center", fmt.Sprintf("argument 2 (width) must be an integer, got %s", args[1].Type()))
	}
	sep := " "
	if l > 2 {
		s, ok := args[2].(*String)
		if !ok {
			return argError("center", fmt.Sprintf("argument 3 (separator) must be a string, got %s", args[2].Type()))
		}
		sep = s.Value
	}
	return textCenterString(str, int(width), sep), nil
}

func textContains(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("contains", fmt.Sprintf("expected 2 arguments (string, substring), got %d", len(args)))
	}
	s, okS := args[0].(*String)
	if !okS {
		return argError("contains", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	substr, okV := args[1].(*String)
	if !okV {
		return argError("contains", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	return Bool(strings.Contains(s.Value, substr.Value)), nil
}

func textContainsAny(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("containsAny", fmt.Sprintf("expected 2 arguments (string, chars), got %d", len(args)))
	}
	s, okS := args[0].(*String)
	if !okS {
		return argError("containsAny", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	substr, okV := args[1].(*String)
	if !okV {
		return argError("containsAny", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	return Bool(strings.ContainsAny(s.Value, substr.Value)), nil
}

func textIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("index", fmt.Sprintf("expected 2 arguments (string, substring), got %d", len(args)))
	}
	s, okS := args[0].(*String)
	if !okS {
		return argError("index", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	substr, okV := args[1].(*String)
	if !okV {
		return argError("index", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	return Integer(strings.Index(s.Value, substr.Value)), nil
}

func textJoin(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("join", fmt.Sprintf("expected 2 arguments (separator, array), got %d", len(args)))
	}
	sep, okSep := args[0].(*String)
	if !okSep {
		return argError("join", fmt.Sprintf("argument 1 (separator) must be a string, got %s", args[0].Type()))
	}
	xs, ok := args[1].(*Array)
	if !ok {
		return argError("join", fmt.Sprintf("argument 2 must be an array, got %s", args[1].Type()))
	}
	r := make([]string, 0, len(xs.Value))
	for _, v := range xs.Value {
		r = append(r, v.String())
	}
	return &String{Value: strings.Join(r, sep.Value)}, nil
}

func textToLowerCase(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("toLower", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("toLower", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	return &String{Value: strings.ToLower(v.Value)}, nil
}

func textToUpperCase(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("toUpper", "expected a string argument")
	}
	v, ok := args[0].(*String)
	if !ok {
		return argError("toUpper", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	return &String{Value: strings.ToUpper(v.Value)}, nil
}

func textCount(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("count", fmt.Sprintf("expected 2 arguments (string, substring), got %d", len(args)))
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("count", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	substr, ok := args[1].(*String)
	if !ok {
		return argError("count", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	return Integer(strings.Count(s.Value, substr.Value)), nil
}

func textIsAscii(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isAscii", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isAscii", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	for i := 0; i < len(s.Value); i++ {
		if s.Value[i] > unicode.MaxASCII {
			return False, nil
		}
	}
	return True, nil
}

func textIsDecimal(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isDecimal", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isDecimal", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	for _, r := range s.Runes {
		if r < '0' || r > '9' {
			return False, nil
		}
	}
	return True, nil
}

func textIsDigit(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isDigit", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isDigit", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	for _, r := range s.Runes {
		if !unicode.IsDigit(r) {
			return False, nil
		}
	}
	return True, nil
}

func textIsHexDigit(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isHexDigit", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isHexDigit", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	for _, r := range s.Runes {
		lower := r | 32
		if !((r >= '0' && r <= '9') || (lower >= 'a' && lower <= 'f')) {
			return False, nil
		}
	}
	return True, nil
}

func textIsLetter(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isLetter", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isLetter", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	for _, r := range s.Runes {
		if !unicode.IsLetter(r) {
			return False, nil
		}
	}
	return True, nil
}

func textIsNumber(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isNumber", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isNumber", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	for _, r := range s.Runes {
		if !unicode.IsNumber(r) {
			return False, nil
		}
	}
	return True, nil
}

func textIsSpace(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isSpace", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isSpace", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return False, nil
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	for _, r := range s.Runes {
		if !unicode.IsSpace(r) {
			return False, nil
		}
	}
	return True, nil
}

func textCodepoints(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("codePoints", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("codePoints", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	result := make([]Value, len(s.Runes))
	for i, r := range s.Runes {
		result[i] = Integer(r)
	}
	return &Array{Value: result}, nil
}

func textIsSpaceChar(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("isSpaceChar", "expected a single-character string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("isSpaceChar", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if StringLength(s) != 1 {
		return argError("isSpaceChar", fmt.Sprintf("expected a single-character string, got a string of length %d", StringLength(s)))
	}
	return Bool(unicode.IsSpace(s.Runes[0])), nil
}

func textBytesLen(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("bytesLen", "expected a string argument")
	}
	val, ok := args[0].(*String)
	if !ok {
		return argError("bytesLen", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	return Integer(len(val.Value)), nil
}

func textStringToArray(slice []string) Value {
	l := len(slice)
	xs := make([]Value, l)
	for i := range l {
		xs[i] = &String{Value: slice[i]}
	}
	return &Array{Value: xs}
}

func textEqualFold(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("equalFold", fmt.Sprintf("expected 2 string arguments, got %d", len(args)))
	}
	s, oks := args[0].(*String)
	if !oks {
		return argError("equalFold", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	t, okt := args[1].(*String)
	if !okt {
		return argError("equalFold", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	return Bool(strings.EqualFold(s.Value, t.Value)), nil
}

func textCapitalize(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("capitalize", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("capitalize", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return s, nil
	}
	first, size := utf8.DecodeRuneInString(s.Value)
	if size == 0 {
		return s, nil
	}
	return &String{Value: string(unicode.ToUpper(first)) + strings.ToLower(s.Value[size:])}, nil
}

func textPadLeft(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l < 2 {
		return argError("padLeft", fmt.Sprintf("expected at least 2 arguments (string, width), got %d", l))
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("padLeft", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	w, ok := args[1].(Integer)
	if !ok {
		return argError("padLeft", fmt.Sprintf("argument 2 (width) must be an integer, got %s", args[1].Type()))
	}
	if w < 0 {
		return argError("padLeft", fmt.Sprintf("width must be non-negative, got %d", w))
	}
	pad := " "
	if l > 2 {
		p, ok := args[2].(*String)
		if !ok {
			return argError("padLeft", fmt.Sprintf("argument 3 (pad) must be a string, got %s", args[2].Type()))
		}
		pad = p.Value
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	strLen := int(len(s.Runes))
	if int(w) <= strLen {
		return s, nil
	}
	return &String{Value: strings.Repeat(pad, int(w)-strLen) + s.Value}, nil
}

func textPadRight(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l < 2 {
		return argError("padRight", fmt.Sprintf("expected at least 2 arguments (string, width), got %d", l))
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("padRight", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	w, ok := args[1].(Integer)
	if !ok {
		return argError("padRight", fmt.Sprintf("argument 2 (width) must be an integer, got %s", args[1].Type()))
	}
	if w < 0 {
		return argError("padRight", fmt.Sprintf("width must be non-negative, got %d", w))
	}
	pad := " "
	if l > 2 {
		p, ok := args[2].(*String)
		if !ok {
			return argError("padRight", fmt.Sprintf("argument 3 (pad) must be a string, got %s", args[2].Type()))
		}
		pad = p.Value
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	strLen := int(len(s.Runes))
	if int(w) <= strLen {
		return s, nil
	}
	return &String{Value: s.Value + strings.Repeat(pad, int(w)-strLen)}, nil
}

func textLines(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("lines", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("lines", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	if s.Value == EmptyString {
		return &Array{}, nil
	}
	norm := strings.ReplaceAll(s.Value, "\r\n", "\n")
	norm = strings.ReplaceAll(norm, "\r", "\n")
	parts := strings.Split(norm, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == EmptyString {
		parts = parts[:len(parts)-1]
	}
	if len(parts) > 0 && parts[0] == EmptyString {
		parts = parts[1:]
	}
	return textStringToArray(parts), nil
}

func textTruncate(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l < 2 {
		return argError("truncate", fmt.Sprintf("expected at least 2 arguments (string, max), got %d", l))
	}
	s, ok1 := args[0].(*String)
	if !ok1 {
		return argError("truncate", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	maxx, ok2 := args[1].(Integer)
	if !ok2 {
		return argError("truncate", fmt.Sprintf("argument 2 (max) must be an integer, got %s", args[1].Type()))
	}
	if maxx < 0 {
		return argError("truncate", fmt.Sprintf("max must be non-negative, got %d", maxx))
	}
	suffix := EmptyString
	if l > 2 {
		sf, ok := args[2].(*String)
		if !ok {
			return argError("truncate", fmt.Sprintf("argument 3 (suffix) must be a string, got %s", args[2].Type()))
		}
		suffix = sf.Value
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	if len(s.Runes) <= int(maxx) {
		return s, nil
	}
	avail := max(int(maxx)-len([]rune(suffix)), 0)
	return &String{Value: string(s.Runes[:avail]) + suffix}, nil
}

func textWrap(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("wrap", fmt.Sprintf("expected 2 arguments (string, width), got %d", len(args)))
	}
	s, ok1 := args[0].(*String)
	if !ok1 {
		return argError("wrap", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	w, ok2 := args[1].(Integer)
	if !ok2 {
		return argError("wrap", fmt.Sprintf("argument 2 (width) must be an integer, got %s", args[1].Type()))
	}
	if w <= 0 {
		return argError("wrap", fmt.Sprintf("width must be greater than 0, got %d", w))
	}
	if s.Runes == nil {
		s.Runes = []rune(s.Value)
	}
	width := int(w)
	var b strings.Builder
	lineLen := 0
	for _, r := range s.Runes {
		if r == '\n' || r == '\r' {
			b.WriteRune(r)
			lineLen = 0
			continue
		}
		if lineLen == 0 && lineLen+1 > width {
			b.WriteRune(r)
			lineLen = 1
			continue
		}
		if lineLen+1 > width {
			b.WriteByte('\n')
			lineLen = 0
		}
		b.WriteRune(r)
		lineLen++
	}
	return &String{Value: b.String()}, nil
}

func textSlugify(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l == 0 {
		return argError("slugify", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("slugify", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	sep := "-"
	asciiOnly := false
	if l > 1 {
		sepArg, ok := args[1].(*String)
		if !ok {
			return argError("slugify", fmt.Sprintf("argument 2 (separator) must be a string, got %s", args[1].Type()))
		}
		sep = sepArg.Value
		if sep == EmptyString {
			sep = "-"
		}
	}
	if l > 2 {
		asciiFlag, ok := args[2].(Bool)
		if !ok {
			return argError("slugify", fmt.Sprintf("argument 3 (asciiOnly) must be a bool, got %s", args[2].Type()))
		}
		asciiOnly = bool(asciiFlag)
	}
	var b strings.Builder
	prevSep := false
	for _, r := range s.Value {
		var keep bool
		if asciiOnly {
			lr := unicode.ToLower(r)
			keep = (lr >= 'a' && lr <= 'z') || (r >= '0' && r <= '9')
		} else {
			keep = unicode.IsLetter(r) || unicode.IsDigit(r)
		}
		if keep {
			b.WriteRune(unicode.ToLower(r))
			prevSep = false
		} else if !prevSep {
			b.WriteString(sep)
			prevSep = true
		}
	}
	res := strings.Trim(b.String(), sep)
	return &String{Value: res}, nil
}

func textStartsWithAny(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("startsWithAny", fmt.Sprintf("expected 2 arguments (string, prefixes), got %d", len(args)))
	}
	s, ok1 := args[0].(*String)
	if !ok1 {
		return argError("startsWithAny", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	arr, ok2 := args[1].(*Array)
	if !ok2 {
		return argError("startsWithAny", fmt.Sprintf("argument 2 must be an array of strings, got %s", args[1].Type()))
	}
	for _, v := range arr.Value {
		if p, ok := v.(*String); ok && strings.HasPrefix(s.Value, p.Value) {
			return True, nil
		}
	}
	return False, nil
}

func textEndsWithAny(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("endsWithAny", fmt.Sprintf("expected 2 arguments (string, suffixes), got %d", len(args)))
	}
	s, ok1 := args[0].(*String)
	if !ok1 {
		return argError("endsWithAny", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	arr, ok2 := args[1].(*Array)
	if !ok2 {
		return argError("endsWithAny", fmt.Sprintf("argument 2 must be an array of strings, got %s", args[1].Type()))
	}
	for _, v := range arr.Value {
		if p, ok := v.(*String); ok && strings.HasSuffix(s.Value, p.Value) {
			return True, nil
		}
	}
	return False, nil
}

func textCompare(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("compare", fmt.Sprintf("expected 2 string arguments, got %d", len(args)))
	}
	a, ok1 := args[0].(*String)
	if !ok1 {
		return argError("compare", fmt.Sprintf("argument 1 must be a string, got %s", args[0].Type()))
	}
	b, ok2 := args[1].(*String)
	if !ok2 {
		return argError("compare", fmt.Sprintf("argument 2 must be a string, got %s", args[1].Type()))
	}
	return Integer(strings.Compare(a.Value, b.Value)), nil
}

func textGetBytes(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("bytes", "expected a string argument")
	}
	src, ok := args[0].(*String)
	if !ok {
		return argError("bytes", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	return &Bytes{Value: []byte(src.Value)}, nil
}

// textUrlEncode percent-encodes a string for safe inclusion in URLs.
//
// Follows RFC 3986:
//   - Unreserved chars (A-Z a-z 0-9 - _ . ~) are NOT encoded
//   - All other characters (including UTF-8 multi-byte sequences) are percent-encoded
//   - Space is encoded as %20 (not +, which is form-specific)
//
// Examples:
//
//	"hello"           → "hello"
//	"hello world"     → "hello%20world"
//	"αβγ"             → "%CE%B1%CE%B2%CE%B3"  (UTF-8 bytes encoded)
//	"café_123"        → "caf%C3%A9_123"
//
// Use case: Combine with text.slugify for URL-safe slugs:
//
//	text.slugify("Hello αβγ!") → "hello-αβγ"
//	text.urlEncode(...)        → "hello-%CE%B1%CE%B2%CE%B3"
func textUrlEncode(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("urlEncode", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("urlEncode", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}
	// RFC 3986 unreserved characters: do NOT encode these
	// Using a lookup table for O(1) checks
	const unreserved = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_.~"
	var safe [256]bool
	for i := 0; i < len(unreserved); i++ {
		safe[unreserved[i]] = true
	}
	var b strings.Builder
	b.Grow(len(s.Value) * 3)
	for i := 0; i < len(s.Value); i++ {
		c := s.Value[i]
		if safe[c] {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperHex[c>>4])
			b.WriteByte(upperHex[c&0x0F])
		}
	}
	return &String{Value: b.String()}, nil
}

var upperHex = [16]byte{
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
}

// textUrlDecode decodes a percent-encoded string back to its original form.
//
// Follows RFC 3986:
//   - %XX sequences (XX = hex digits, case-insensitive) are decoded to bytes
//   - + is NOT converted to space (that's form-specific; use urlDecodeForm if needed)
//   - Result is validated as valid UTF-8; invalid sequences return an error
//
// Examples:
//
//	"hello"              → "hello"
//	"hello%20world"      → "hello world"
//	"%CE%B1%CE%B2%CE%B3" → "αβγ"
//	"caf%C3%A9_123"      → "café_123"
//
// Error cases:
//   - Incomplete % sequence: "hello%" → error
//   - Invalid hex: "hello%GG" → error
//   - Invalid UTF-8: "%FF%FE" → error (unless you want lenient mode)
//
// Use case: Reverse text.urlEncode for round-trip safety:
//
//	original = "Hello αβγ!"
//	encoded  = text.urlEncode(original)
//	decoded  = text.urlDecode(encoded)  // → original
func textUrlDecode(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return argError("urlDecode", "expected a string argument")
	}
	s, ok := args[0].(*String)
	if !ok {
		return argError("urlDecode", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
	}

	input := s.Value
	var b strings.Builder
	// Pre-allocate: decoded output is always ≤ encoded input length
	b.Grow(len(input))

	i := 0
	for i < len(input) {
		c := input[i]

		if c == '%' {
			// Percent-encoded sequence: need 2 hex digits
			if i+2 >= len(input) {
				return argError("urlDecode", fmt.Sprintf("incomplete percent-encoding at position %d: %q", i, input[i:]))
			}

			h1, ok1 := hexDigitToByte(input[i+1])
			h2, ok2 := hexDigitToByte(input[i+2])
			if !ok1 || !ok2 {
				return argError("urlDecode", fmt.Sprintf("invalid hex digits at position %d: %q", i, input[i:i+3]))
			}

			decodedByte := (h1 << 4) | h2
			b.WriteByte(decodedByte)
			i += 3
		} else {
			// Literal character (unreserved or already-decoded)
			b.WriteByte(c)
			i++
		}
	}

	// Validate that the result is valid UTF-8
	result := b.String()
	if !utf8.ValidString(result) {
		return argError("urlDecode", "decoded result contains an invalid UTF-8 sequence")
	}

	return &String{Value: result}, nil
}

func hexDigitToByte(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	default:
		return 0, false
	}
}
