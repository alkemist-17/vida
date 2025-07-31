package vida

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationText() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["hasPrefix"] = GFn(hasPrefix)
	m.Value["hasSuffix"] = GFn(hasSuffix)
	m.Value["fromCodePoint"] = GFn(fromCodepoint)
	m.Value["trim"] = GFn(trim)
	m.Value["trimLeft"] = GFn(trimLeft)
	m.Value["trimRight"] = GFn(trimRight)
	m.Value["split"] = GFn(split)
	m.Value["fields"] = GFn(fields)
	m.Value["repeat"] = GFn(repeat)
	m.Value["replace"] = GFn(replace)
	m.Value["replaceAll"] = GFn(replaceAll)
	m.Value["center"] = GFn(center)
	m.Value["contains"] = GFn(contains)
	m.Value["containsAny"] = GFn(containsAny)
	m.Value["index"] = GFn(index)
	m.Value["join"] = GFn(join)
	m.Value["toLower"] = GFn(lower)
	m.Value["toUpper"] = GFn(upper)
	m.Value["count"] = GFn(count)
	m.Value["isAscii"] = GFn(isAscii)
	m.Value["isDecimal"] = GFn(isDecimal)
	m.Value["isDigit"] = GFn(isDigit)
	m.Value["isHexDigit"] = GFn(isHexDigit)
	m.Value["isLetter"] = GFn(isLetter)
	m.Value["isNumber"] = GFn(isNumber)
	m.Value["isSpace"] = GFn(isSpace)
	m.Value["codePoint"] = GFn(codepoint)
	m.Value["byteslen"] = GFn(byteslen)
	return m
}

func hasPrefix(args ...Value) (Value, error) {
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

func hasSuffix(args ...Value) (Value, error) {
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

func fromCodepoint(args ...Value) (Value, error) {
	runes := make([]rune, 0)
	for _, a := range args {
		if v, ok := a.(Integer); ok && v > 0 {
			runes = append(runes, int32(v))
		}
	}
	return &String{Value: string(runes), Runes: runes}, nil
}

func trim(args ...Value) (Value, error) {
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

func trimLeft(args ...Value) (Value, error) {
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

func trimRight(args ...Value) (Value, error) {
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

func split(args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		if v, ok := args[0].(*String); ok {
			if p, ok := args[1].(*String); ok {
				return stringSliceToArray(strings.Split(v.Value, p.Value)), nil
			}
			return stringSliceToArray(strings.Split(v.Value, "")), nil
		}
		return NilValue, nil
	}
	if l == 1 {
		if v, ok := args[0].(*String); ok {
			return stringSliceToArray(strings.Split(v.Value, "")), nil
		}
	}
	return NilValue, nil
}

func fields(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return stringSliceToArray(strings.Fields(v.Value)), nil
		}
	}
	return NilValue, nil
}

func repeat(args ...Value) (Value, error) {
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

func replace(args ...Value) (Value, error) {
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

func replaceAll(args ...Value) (Value, error) {
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

func center(args ...Value) (Value, error) {
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

func contains(args ...Value) (Value, error) {
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

func containsAny(args ...Value) (Value, error) {
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

func index(args ...Value) (Value, error) {
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

func join(args ...Value) (Value, error) {
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

func lower(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.ToLower(v.Value)}, nil
		}
	}
	return NilValue, nil
}

func upper(args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.ToUpper(v.Value)}, nil
		}
	}
	return NilValue, nil
}

func count(args ...Value) (Value, error) {
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

func isAscii(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool(0 <= c && c <= unicode.MaxASCII), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func isDecimal(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool('0' <= c && c <= '9'), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func isDigit(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsDigit(s.Runes[0])), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func isHexDigit(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool('0' <= c && c <= '9' || 'a' <= (32|c) && (32|c) <= 'f'), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func isLetter(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			c := s.Runes[0]
			return Bool('a' <= (32|c) && (32|c) <= 'z' || c == '_' || c >= utf8.RuneSelf && unicode.IsLetter(c)), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func isNumber(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsNumber(s.Runes[0])), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func isSpace(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsSpace(s.Runes[0])), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func codepoint(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Integer(s.Runes[0]), nil
		}
		return NilValue, nil
	}
	return NilValue, nil
}

func byteslen(args ...Value) (Value, error) {
	if len(args) > 0 {
		if val, ok := args[0].(*String); ok {
			return Integer(len(val.Value)), nil
		}
	}
	return NilValue, nil
}

func stringSliceToArray(slice []string) Value {
	l := len(slice)
	xs := make([]Value, l)
	for i := 0; i < l; i++ {
		xs[i] = &String{Value: slice[i]}
	}
	return &Array{Value: xs}
}
