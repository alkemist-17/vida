package vida

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alkemist-17/vida/verror"
)

func loadFoundationText() Value {
	m := &Object{Value: make(map[string]Value, 42)}
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
	return m
}

func textHasPrefix(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		v, okV := args[0].(*String)
		p, okP := args[1].(*String)
		if okV && okP {
			return Bool(strings.HasPrefix(v.Value, p.Value)), nil
		}
	}
	return Nil, nil
}

func textHasSuffix(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		v, okV := args[0].(*String)
		p, okP := args[1].(*String)
		if okV && okP {
			return Bool(strings.HasSuffix(v.Value, p.Value)), nil
		}
	}
	return Nil, nil
}

func textFromCodepoints(ctx *Context, args ...Value) (Value, error) {
	runes := make([]rune, 0, len(args))
	for _, a := range args {
		if v, ok := a.(Integer); ok && v > 0 {
			runes = append(runes, int32(v))
		}
	}
	return &String{Value: string(runes), Runes: runes, VTable: ctx.vtables[stringVT]}, nil
}

func textTrim(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if p, ok := args[1].(*String); ok {
					return &String{Value: strings.Trim(v.Value, p.Value), VTable: ctx.vtables[stringVT]}, nil
				}
			}
			return &String{Value: strings.Trim(v.Value, " "), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textTrimLeft(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if p, ok := args[1].(*String); ok {
					return &String{Value: strings.TrimLeft(v.Value, p.Value), VTable: ctx.vtables[stringVT]}, nil
				}
			}
			return &String{Value: strings.TrimLeft(v.Value, " "), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textTrimRight(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if p, ok := args[1].(*String); ok {
					return &String{Value: strings.TrimRight(v.Value, p.Value), VTable: ctx.vtables[stringVT]}, nil
				}
			}
			return &String{Value: strings.TrimRight(v.Value, " "), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textSplit(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if sep, ok := args[1].(*String); ok {
					return textStringToArray(ctx, strings.Split(v.Value, sep.Value)), nil
				}
			}
			return textStringToArray(ctx, strings.Split(v.Value, " ")), nil
		}
	}
	return Nil, nil
}

func textFields(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return textStringToArray(ctx, strings.Fields(v.Value)), nil
		}
	}
	return Nil, nil
}

func textRepeat(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if v, ok := args[0].(*String); ok {
			if times, ok := args[1].(Integer); ok && times >= 0 {
				if StringLength(v)*times > verror.MaxMemSize {
					return Nil, nil
				}
				return &String{Value: strings.Repeat(v.Value, int(times)), VTable: ctx.vtables[stringVT]}, nil
			}
		}
	}
	return Nil, nil
}

func textReplaceN(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 3 {
		if s, ok := args[0].(*String); ok {
			if old, ok := args[1].(*String); ok {
				if nnew, ok := args[2].(*String); ok {
					if k, ok := args[3].(Integer); ok {
						return &String{Value: strings.Replace(s.Value, old.Value, nnew.Value, int(k)), VTable: ctx.vtables[stringVT]}, nil
					}
				}
			}
		}
	}
	return Nil, nil
}

func textReplaceAll(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 2 {
		if s, ok := args[0].(*String); ok {
			if old, ok := args[1].(*String); ok {
				if nnew, ok := args[2].(*String); ok {
					return &String{Value: strings.ReplaceAll(s.Value, old.Value, nnew.Value), VTable: ctx.vtables[stringVT]}, nil
				}
			}
		}
	}
	return Nil, nil
}

func textCenterString(ctx *Context, s *String, width int, sep string) *String {
	strLen := int(StringLength(s))
	if width <= strLen {
		return s
	}
	padding := width - strLen
	leftPad := padding / 2
	rightPad := padding - leftPad
	return &String{Value: strings.Repeat(sep, leftPad) + s.Value + strings.Repeat(sep, rightPad), VTable: ctx.vtables[stringVT]}
}

func textCenter(ctx *Context, args ...Value) (Value, error) {
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
				return textCenterString(ctx, str, int(width), sep), nil
			}
		}
	}
	return Nil, nil
}

func textContains(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		substr, okV := args[1].(*String)
		if okS && okV {
			return Bool(strings.Contains(s.Value, substr.Value)), nil
		}
	}
	return Nil, nil
}

func textContainsAny(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		substr, okV := args[1].(*String)
		if okS && okV {
			return Bool(strings.ContainsAny(s.Value, substr.Value)), nil
		}
	}
	return Nil, nil
}

func textIndex(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, okS := args[0].(*String)
		substr, okV := args[1].(*String)
		if okS && okV {
			return Integer(strings.Index(s.Value, substr.Value)), nil
		}
	}
	return Nil, nil
}

func textJoin(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		xs, ok := args[0].(*Array)
		sep, okSep := args[1].(*String)
		if ok && okSep {
			r := make([]string, 0, len(xs.Value))
			for _, v := range xs.Value {
				r = append(r, v.String())
			}
			return &String{Value: strings.Join(r, sep.Value), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textToLowerCase(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.ToLower(v.Value), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textToUpperCase(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if v, ok := args[0].(*String); ok {
			return &String{Value: strings.ToUpper(v.Value), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textCount(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if s, ok := args[0].(*String); ok {
			if substr, ok := args[1].(*String); ok {
				return Integer(strings.Count(s.Value, substr.Value)), nil
			}
		}
	}
	return Nil, nil
}

func textIsAscii(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textIsDecimal(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textIsDigit(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textIsHexDigit(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textIsLetter(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textIsNumber(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textIsSpace(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
	}
	return False, nil
}

func textCodepoints(ctx *Context, args ...Value) (Value, error) {
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
	return Nil, nil
}

func textIsSpaceChar(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok && StringLength(s) == 1 {
			return Bool(unicode.IsSpace(s.Runes[0])), nil
		}
	}
	return False, nil
}

func textBytesLen(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if val, ok := args[0].(*String); ok {
			return Integer(len(val.Value)), nil
		}
	}
	return Nil, nil
}

func textStringToArray(ctx *Context, slice []string) Value {
	l := len(slice)
	xs := make([]Value, l)
	for i := range l {
		xs[i] = &String{Value: slice[i], VTable: ctx.vtables[stringVT]}
	}
	return &Array{Value: xs}
}

func textEqualFold(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, oks := args[0].(*String)
		t, okt := args[1].(*String)
		if oks && okt {
			return Bool(strings.EqualFold(s.Value, t.Value)), nil
		}
	}
	return Nil, nil
}

func textCapitalize(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == EmptyString {
				return s, nil
			}
			first, size := utf8.DecodeRuneInString(s.Value)
			if size == 0 {
				return s, nil
			}
			return &String{Value: string(unicode.ToUpper(first)) + strings.ToLower(s.Value[size:]), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textPadLeft(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		s, ok1 := args[0].(*String)
		w, ok2 := args[1].(Integer)
		if ok1 && ok2 && w >= 0 {
			pad := " "
			if l > 2 {
				if p, ok := args[2].(*String); ok {
					pad = p.Value
				}
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			strLen := int(len(s.Runes))
			if int(w) <= strLen {
				return s, nil
			}
			return &String{Value: strings.Repeat(pad, int(w)-strLen) + s.Value, VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textPadRight(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		s, ok1 := args[0].(*String)
		w, ok2 := args[1].(Integer)
		if ok1 && ok2 && w >= 0 {
			pad := " "
			if l > 2 {
				if p, ok := args[2].(*String); ok {
					pad = p.Value
				}
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			strLen := int(len(s.Runes))
			if int(w) <= strLen {
				return s, nil
			}
			return &String{Value: s.Value + strings.Repeat(pad, int(w)-strLen), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textLines(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
			return textStringToArray(ctx, parts), nil
		}
	}
	return Nil, nil
}

func textTruncate(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if l > 1 {
		s, ok1 := args[0].(*String)
		maxx, ok2 := args[1].(Integer)
		if ok1 && ok2 && maxx >= 0 {
			suffix := EmptyString
			if l > 2 {
				if sf, ok := args[2].(*String); ok {
					suffix = sf.Value
				}
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			if len(s.Runes) <= int(maxx) {
				return s, nil
			}
			avail := max(int(maxx)-len([]rune(suffix)), 0)
			return &String{Value: string(s.Runes[:avail]) + suffix, VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textWrap(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, ok1 := args[0].(*String)
		w, ok2 := args[1].(Integer)
		if ok1 && ok2 && w > 0 {
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
			return &String{Value: b.String(), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textSlugify(ctx *Context, args ...Value) (Value, error) {
	l := len(args)
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			sep := "-"
			asciiOnly := false
			if l > 1 {
				if sepArg, ok := args[1].(*String); ok {
					sep = sepArg.Value
				}
				if sep == "" {
					sep = "-"
				}
			}
			if l > 2 {
				if asciiFlag, ok := args[2].(Bool); ok {
					asciiOnly = bool(asciiFlag)
				}
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
			return &String{Value: res, VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
}

func textStartsWithAny(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, ok1 := args[0].(*String)
		arr, ok2 := args[1].(*Array)
		if ok1 && ok2 {
			for _, v := range arr.Value {
				if p, ok := v.(*String); ok && strings.HasPrefix(s.Value, p.Value) {
					return True, nil
				}
			}
		}
	}
	return False, nil
}

func textEndsWithAny(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		s, ok1 := args[0].(*String)
		arr, ok2 := args[1].(*Array)
		if ok1 && ok2 {
			for _, v := range arr.Value {
				if p, ok := v.(*String); ok && strings.HasSuffix(s.Value, p.Value) {
					return True, nil
				}
			}
		}
	}
	return False, nil
}

func textCompare(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		a, ok1 := args[0].(*String)
		b, ok2 := args[1].(*String)
		if ok1 && ok2 {
			return Integer(strings.Compare(a.Value, b.Value)), nil
		}
	}
	return Nil, nil
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
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
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
			return &String{Value: b.String(), VTable: ctx.vtables[stringVT]}, nil
		}
	}
	return Nil, nil
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
		return nil, errors.New("text.urlDecode, expected 1 argument: (str: String)")
	}
	s, ok := args[0].(*String)
	if !ok {
		return nil, errors.New("text.urlDecode, argument must be a String")
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
				return nil, fmt.Errorf("text.urlDecode, incomplete percent-encoding at position %d: %q", i, input[i:])
			}

			h1, ok1 := hexDigitToByte(input[i+1])
			h2, ok2 := hexDigitToByte(input[i+2])
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("text.urlDecode, invalid hex digits at position %d: %q", i, input[i:i+3])
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
		return nil, errors.New("text.urlDecode, decoded result contains invalid UTF-8 sequence")
	}

	return &String{Value: result, VTable: ctx.vtables[stringVT]}, nil
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
