package vida

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationText() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["hasPrefix"] = GFn(textHasPrefix)
	m.Value["hasSuffix"] = GFn(textHasSuffix)
	m.Value["fromCodePoint"] = GFn(textFromCodepoint)
	m.Value["trim"] = GFn(textTrim)
	m.Value["trimLeft"] = GFn(textTrimLeft)
	m.Value["trimRight"] = GFn(textTrimRight)
	m.Value["split"] = GFn(textSplit)
	m.Value["fields"] = GFn(textFields)
	m.Value["repeat"] = GFn(textRepeat)
	m.Value["replace"] = GFn(textReplace)
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
	m.Value["codePoint"] = GFn(textCodepoint)
	m.Value["byteslen"] = GFn(textBytesLen)
	return m
}

func textHasPrefix(args ...Value) (Value, error) {
	if len(args) > 1 {
		if v, ok := args[0].(*String); ok {
			if p, ok := args[1].(*String); ok {
				return Bool(strings.HasPrefix(v.Value, p.Value)), nil
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textHasSuffix(args ...Value) (Value, error) {
	if len(args) > 1 {
		if v, ok := args[0].(*String); ok {
			if p, ok := args[1].(*String); ok {
				return Bool(strings.HasSuffix(v.Value, p.Value)), nil
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textFromCodepoint(args ...Value) (Value, error) {
	runes := make([]rune, 0)
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
			return &String{Value: strings.Trim(v.Value, " ")}, nil
		}
		return NilValue, nil
	}
	if l == 1 {
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
			return &String{Value: strings.TrimLeft(v.Value, " ")}, nil
		}
		return NilValue, nil
	}
	if l == 1 {
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
			return &String{Value: strings.TrimRight(v.Value, " ")}, nil
		}
		return NilValue, nil
	}
	if l == 1 {
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
			if p, ok := args[1].(*String); ok {
				return textStringSliceToArray(strings.Split(v.Value, p.Value)), nil
			}
			return textStringSliceToArray(strings.Split(v.Value, "")), nil
		}
		return NilValue, nil
	}
	if l == 1 {
		if v, ok := args[0].(*String); ok {
			return textStringSliceToArray(strings.Split(v.Value, "")), nil
		}
	}
	return NilValue, nil
}

func textFields(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return textStringSliceToArray(strings.Fields(v.Value)), nil
		}
	}
	return NilValue, nil
}

func textRepeat(args ...Value) (Value, error) {
	if len(args) >= 2 {
		if v, ok := args[0].(*String); ok {
			if times, ok := args[1].(Integer); ok && times >= 0 {
				if StringLength(v)*times > verror.MaxMemSize {
					return NilValue, nil
				}
				return &String{Value: strings.Repeat(v.Value, int(times))}, nil
			}
			return NilValue, nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textReplace(args ...Value) (Value, error) {
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
		return NilValue, nil
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
		return NilValue, nil
	}
	return NilValue, nil
}

func textCenter(args ...Value) (Value, error) {
	l := len(args)
	if l == 2 {
		if str, ok := args[0].(*String); ok {
			if width, ok := args[1].(Integer); ok {
				strlen := StringLength(str)
				if width <= strlen {
					return str, nil
				}
				padding := width - strlen
				newString := str.Value
				sep := " "
				for i := Integer(0); i < padding; i++ {
					if i%2 == 0 {
						newString = newString + sep
					} else {
						newString = sep + newString
					}
				}
				return &String{Value: newString}, nil
			}
		}
		return NilValue, nil
	}
	if l > 2 {
		if str, ok := args[0].(*String); ok {
			if width, ok := args[1].(Integer); ok {
				if sep, ok := args[2].(*String); ok {
					strlen := StringLength(str)
					if width <= strlen {
						return str, nil
					}
					padding := width - strlen
					newString := str.Value
					for i := Integer(0); i < padding; i++ {
						if i%2 == 0 {
							newString = newString + sep.Value
						} else {
							newString = sep.Value + newString
						}
					}
					return &String{Value: newString}, nil
				}
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textContains(args ...Value) (Value, error) {
	if len(args) > 1 {
		if s, ok := args[0].(*String); ok {
			if substr, ok := args[1].(*String); ok {
				return Bool(strings.Contains(s.Value, substr.Value)), nil
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textContainsAny(args ...Value) (Value, error) {
	if len(args) > 1 {
		if s, ok := args[0].(*String); ok {
			if substr, ok := args[1].(*String); ok {
				return Bool(strings.ContainsAny(s.Value, substr.Value)), nil
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIndex(args ...Value) (Value, error) {
	if len(args) > 1 {
		if s, ok := args[0].(*String); ok {
			if substr, ok := args[1].(*String); ok {
				return Integer(strings.Index(s.Value, substr.Value)), nil
			}
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textJoin(args ...Value) (Value, error) {
	if len(args) > 1 {
		if xs, ok := args[0].(*Array); ok {
			if sep, ok := args[1].(*String); ok {
				var r []string
				for _, v := range xs.Value {
					r = append(r, v.String())
				}
				return &String{Value: strings.Join(r, sep.Value)}, nil
			}
		}
		return NilValue, nil
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
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsAscii(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool(0 <= c && c <= unicode.MaxASCII), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsDecimal(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool('0' <= c && c <= '9'), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsDigit(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsDigit(s.Runes[0])), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsHexDigit(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool('0' <= c && c <= '9' || 'a' <= (32|c) && (32|c) <= 'f'), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsLetter(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool('a' <= (32|c) && (32|c) <= 'z' || c == '_' || c >= utf8.RuneSelf && unicode.IsLetter(c)), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsNumber(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsNumber(s.Runes[0])), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textIsSpace(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsSpace(s.Runes[0])), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textCodepoint(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Integer(s.Runes[0]), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func textBytesLen(args ...Value) (Value, error) {
	if len(args) > 0 {
		if val, ok := args[0].(*String); ok {
			return Integer(len(val.Value)), nil
		}
	}
	return NilValue, nil
}

func textStringSliceToArray(slice []string) Value {
	l := len(slice)
	xs := make([]Value, l)
	for i := 0; i < l; i++ {
		xs[i] = &String{Value: slice[i]}
	}
	return &Array{Value: xs}
}
