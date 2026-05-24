package vida

import (
	"strings"
	"unicode"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationText() Value {
	m := &Object{Value: make(map[string]Value, 30)}
	m.Value["hasPrefix"] = GFn(textHasPrefix)
	m.Value["hasSuffix"] = GFn(textHasSuffix)
	m.Value["fromCodePoint"] = GFn(textFromCodepoint)
	m.Value["trim"] = GFn(textTrim)
	m.Value["trimLeft"] = GFn(textTrimLeft)
	m.Value["trimRight"] = GFn(textTrimRight)
	m.Value["split"] = GFn(textSplit)
	m.Value["fields"] = GFn(textFields)
	m.Value["repeat"] = GFn(textRepeat)
	m.Value["replaceN"] = GFn(textReplaceN)
	m.Value["replaceAll"] = GFn(textReplaceAll)
	m.Value["center"] = GFn(textCenter)
	m.Value["contains"] = GFn(textContains)
	m.Value["containsAny"] = GFn(textContainsAny)
	m.Value["index"] = GFn(textIndex)
	m.Value["join"] = GFn(textJoin)
	m.Value["toLower"] = GFn(textToLowerCase)
	m.Value["toUpper"] = GFn(textToUpperCase)
	m.Value["count"] = GFn(textCount)
	m.Value["isAscii"] = GFn(textIsAscii)
	m.Value["isDecimal"] = GFn(textIsDecimal)
	m.Value["isDigit"] = GFn(textIsDigit)
	m.Value["isHexDigit"] = GFn(textIsHexDigit)
	m.Value["isLetter"] = GFn(textIsLetter)
	m.Value["isNumber"] = GFn(textIsNumber)
	m.Value["isSpace"] = GFn(textIsSpace)
	m.Value["isSpaceChar"] = GFn(textIsSpaceChar)
	m.Value["codePoints"] = GFn(textCodepoints)
	m.Value["bytesLen"] = GFn(textBytesLen)
	m.Value["equalFold"] = GFn(textEqualFold)
	return m
}

func textHasPrefix(args ...Value) (Value, error) {
	if len(args) > 1 {
		v, okV := args[0].(*String)
		p, okP := args[1].(*String)
		if okV && okP {
			return Bool(strings.HasPrefix(v.Value, p.Value)), nil
		}
	}
	return NilValue, nil
}

func textHasSuffix(args ...Value) (Value, error) {
	if len(args) > 1 {
		v, okV := args[0].(*String)
		p, okP := args[1].(*String)
		if okV && okP {
			return Bool(strings.HasSuffix(v.Value, p.Value)), nil
		}
	}
	return NilValue, nil
}

func textFromCodepoint(args ...Value) (Value, error) {
	runes := make([]rune, 0, len(args))
	for _, a := range args {
		if v, ok := a.(Integer); ok && v > 0 {
			runes = append(runes, int32(v))
		}
	}
	return &String{Value: string(runes), Runes: runes}, nil
}

func textTrim(args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if v, ok := args[0].(*String); ok {
			if p, ok := args[1].(*String); ok {
				return &String{Value: strings.Trim(v.Value, p.Value)}, nil
			}
		}
	} else if l == 1 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.Trim(v.Value, " ")}, nil
		}
	}
	return NilValue, nil
}

func textTrimLeft(args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if v, ok := args[0].(*String); ok {
			if p, ok := args[1].(*String); ok {
				return &String{Value: strings.TrimLeft(v.Value, p.Value)}, nil
			}
		}
	} else if l == 1 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.TrimLeft(v.Value, " ")}, nil
		}
	}
	return NilValue, nil
}

func textTrimRight(args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if v, ok := args[0].(*String); ok {
			if p, ok := args[1].(*String); ok {
				return &String{Value: strings.TrimRight(v.Value, p.Value)}, nil
			}
		}
	} else if l == 1 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.TrimRight(v.Value, " ")}, nil
		}
	}
	return NilValue, nil
}

func textSplit(args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if v, ok := args[0].(*String); ok {
			if sep, ok := args[1].(*String); ok {
				return textStringToArray(strings.Split(v.Value, sep.Value)), nil
			}
		}
	} else if l == 1 {
		if v, ok := args[0].(*String); ok {
			return textStringToArray(strings.Split(v.Value, "")), nil
		}
	}
	return NilValue, nil
}

func textFields(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return textStringToArray(strings.Fields(v.Value)), nil
		}
	}
	return NilValue, nil
}

func textRepeat(args ...Value) (Value, error) {
	if len(args) > 1 {
		if v, ok := args[0].(*String); ok {
			if times, ok := args[1].(Integer); ok && times >= 0 {
				if StringLength(v)*times > verror.MaxMemSize {
					return NilValue, nil
				}
				return &String{Value: strings.Repeat(v.Value, int(times))}, nil
			}
		}
	}
	return NilValue, nil
}

func textReplaceN(args ...Value) (Value, error) {
	if len(args) > 3 {
		if s, ok := args[0].(*String); ok {
			if old, ok := args[1].(*String); ok {
				if nnew, ok := args[2].(*String); ok {
					if k, ok := args[3].(Integer); ok {
						return &String{Value: strings.Replace(s.Value, old.Value, nnew.Value, int(k))}, nil
					}
				}
			}
		}
	}
	return NilValue, nil
}

func textReplaceAll(args ...Value) (Value, error) {
	if len(args) > 2 {
		if s, ok := args[0].(*String); ok {
			if old, ok := args[1].(*String); ok {
				if nnew, ok := args[2].(*String); ok {
					return &String{Value: strings.ReplaceAll(s.Value, old.Value, nnew.Value)}, nil
				}
			}
		}
	}
	return NilValue, nil
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

func textCenter(args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if str, ok := args[0].(*String); ok {
			if width, ok := args[1].(Integer); ok {
				sep := " "
				if l > 2 {
					if s, ok := args[2].(*String); ok {
						sep = s.Value
					}
				}
				return textCenterString(str, int(width), sep), nil
			}
		}
	}
	return NilValue, nil
}

func textContains(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		substr, okV := args[1].(*String)
		if okS && okV {
			return Bool(strings.Contains(s.Value, substr.Value)), nil
		}
	}
	return NilValue, nil
}

func textContainsAny(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		substr, okV := args[1].(*String)
		if okS && okV {
			return Bool(strings.ContainsAny(s.Value, substr.Value)), nil
		}
	}
	return NilValue, nil
}

func textIndex(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		substr, okV := args[1].(*String)
		if okS && okV {
			return Integer(strings.Index(s.Value, substr.Value)), nil
		}
	}
	return NilValue, nil
}

func textJoin(args ...Value) (Value, error) {
	if len(args) > 1 {
		xs, ok := args[0].(*Array)
		sep, okSep := args[1].(*String)
		if ok && okSep {
			r := make([]string, 0, len(xs.Value))
			for _, v := range xs.Value {
				r = append(r, v.String())
			}
			return &String{Value: strings.Join(r, sep.Value)}, nil
		}
	}
	return NilValue, nil
}

func textToLowerCase(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.ToLower(v.Value)}, nil
		}
	}
	return NilValue, nil
}

func textToUpperCase(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.ToUpper(v.Value)}, nil
		}
	}
	return NilValue, nil
}

func textCount(args ...Value) (Value, error) {
	if len(args) > 1 {
		if s, ok := args[0].(*String); ok {
			if substr, ok := args[1].(*String); ok {
				return Integer(strings.Count(s.Value, substr.Value)), nil
			}
		}
	}
	return NilValue, nil
}

func textIsAscii(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			for i := 0; i < len(s.Value); i++ {
				if s.Value[i] > unicode.MaxASCII {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textIsDecimal(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				if r < '0' || r > '9' {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textIsDigit(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				if !unicode.IsDigit(r) {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textIsHexDigit(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				lower := r | 32 // ASCII lowercase trick; safe for non-ASCII (they won't match anyway)
				if !((r >= '0' && r <= '9') || (lower >= 'a' && lower <= 'f')) {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textIsLetter(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				if !unicode.IsLetter(r) {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textIsNumber(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				if !unicode.IsNumber(r) {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textIsSpace(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == "" {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				if !unicode.IsSpace(r) {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
	}
	return Bool(false), nil
}

func textCodepoints(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			result := make([]Value, len(s.Runes))
			for i, r := range s.Runes {
				result[i] = Integer(r)
			}
			return &Array{Value: result}, nil
		}
	}
	return NilValue, nil
}

func textIsSpaceChar(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsSpace(s.Runes[0])), nil
		}
	}
	return Bool(false), nil
}

func textBytesLen(args ...Value) (Value, error) {
	if len(args) > 0 {
		if val, ok := args[0].(*String); ok {
			return Integer(len(val.Value)), nil
		}
	}
	return NilValue, nil
}

func textStringToArray(slice []string) Value {
	l := len(slice)
	xs := make([]Value, l)
	for i := range l {
		xs[i] = &String{Value: slice[i]}
	}
	return &Array{Value: xs}
}

func textEqualFold(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, oks := args[0].(*String)
		t, okt := args[1].(*String)
		if oks && okt {
			return Bool(strings.EqualFold(s.Value, t.Value)), nil
		}
	}
	return NilValue, nil
}
