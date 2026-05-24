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
	m.Value["hasPrefix"] = GFn(textHasPrefix)
	m.Value["hasSuffix"] = GFn(textHasSuffix)
	m.Value["fromCodePoints"] = GFn(textFromCodepoints)
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
	m.Value["capitalize"] = GFn(textCapitalize)
	m.Value["padLeft"] = GFn(textPadLeft)
	m.Value["padRight"] = GFn(textPadRight)
	m.Value["lines"] = GFn(textLines)
	m.Value["truncate"] = GFn(textTruncate)
	m.Value["wrap"] = GFn(textWrap)
	m.Value["slugify"] = GFn(textSlugify)
	m.Value["startsWithAny"] = GFn(textStartsWithAny)
	m.Value["endsWithAny"] = GFn(textEndsWithAny)
	m.Value["compare"] = GFn(textCompare)
	m.Value["urlEncode"] = GFn(textUrlEncode)
	m.Value["urlDecode"] = GFn(textUrlDecode)
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

func textFromCodepoints(args ...Value) (Value, error) {
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
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if p, ok := args[1].(*String); ok {
					return &String{Value: strings.Trim(v.Value, p.Value)}, nil
				}
			}
			return &String{Value: strings.Trim(v.Value, " ")}, nil
		}
	}
	return NilValue, nil
}

func textTrimLeft(args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if p, ok := args[1].(*String); ok {
					return &String{Value: strings.TrimLeft(v.Value, p.Value)}, nil
				}
			}
			return &String{Value: strings.TrimLeft(v.Value, " ")}, nil
		}
	}
	return NilValue, nil
}

func textTrimRight(args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if p, ok := args[1].(*String); ok {
					return &String{Value: strings.TrimRight(v.Value, p.Value)}, nil
				}
			}
			return &String{Value: strings.TrimRight(v.Value, " ")}, nil
		}
	}
	return NilValue, nil
}

func textSplit(args ...Value) (Value, error) {
	l := len(args)
	if l > 0 {
		if v, ok := args[0].(*String); ok {
			if l > 1 {
				if sep, ok := args[1].(*String); ok {
					return textStringToArray(strings.Split(v.Value, sep.Value)), nil
				}
			}
			return textStringToArray(strings.Split(v.Value, " ")), nil
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
			if s.Value == EmptyString {
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
			if s.Value == EmptyString {
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
			if s.Value == EmptyString {
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
			if s.Value == EmptyString {
				return Bool(false), nil
			}
			if s.Runes == nil {
				s.Runes = []rune(s.Value)
			}
			for _, r := range s.Runes {
				lower := r | 32
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
			if s.Value == EmptyString {
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
			if s.Value == EmptyString {
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
			if s.Value == EmptyString {
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

func textCapitalize(args ...Value) (Value, error) {
	if len(args) > 0 {
		if s, ok := args[0].(*String); ok {
			if s.Value == EmptyString {
				return s, nil
			}
			first, size := utf8.DecodeRuneInString(s.Value)
			if size == 0 {
				return s, nil
			}
			return &String{Value: string(unicode.ToUpper(first)) + strings.ToLower(s.Value[size:])}, nil
		}
	}
	return NilValue, nil
}

func textPadLeft(args ...Value) (Value, error) {
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
			return &String{Value: strings.Repeat(pad, int(w)-strLen) + s.Value}, nil
		}
	}
	return NilValue, nil
}

func textPadRight(args ...Value) (Value, error) {
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
			return &String{Value: s.Value + strings.Repeat(pad, int(w)-strLen)}, nil
		}
	}
	return NilValue, nil
}

func textLines(args ...Value) (Value, error) {
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
			return textStringToArray(parts), nil
		}
	}
	return NilValue, nil
}

func textTruncate(args ...Value) (Value, error) {
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
			return &String{Value: string(s.Runes[:avail]) + suffix}, nil
		}
	}
	return NilValue, nil
}

func textWrap(args ...Value) (Value, error) {
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
			return &String{Value: b.String()}, nil
		}
	}
	return NilValue, nil
}

func textSlugify(args ...Value) (Value, error) {
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
			return &String{Value: res}, nil
		}
	}
	return NilValue, nil
}

func textStartsWithAny(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, ok1 := args[0].(*String)
		arr, ok2 := args[1].(*Array)
		if ok1 && ok2 {
			for _, v := range arr.Value {
				if p, ok := v.(*String); ok && strings.HasPrefix(s.Value, p.Value) {
					return Bool(true), nil
				}
			}
		}
	}
	return Bool(false), nil
}

func textEndsWithAny(args ...Value) (Value, error) {
	if len(args) > 1 {
		s, ok1 := args[0].(*String)
		arr, ok2 := args[1].(*Array)
		if ok1 && ok2 {
			for _, v := range arr.Value {
				if p, ok := v.(*String); ok && strings.HasSuffix(s.Value, p.Value) {
					return Bool(true), nil
				}
			}
		}
	}
	return Bool(false), nil
}

func textCompare(args ...Value) (Value, error) {
	if len(args) > 1 {
		a, ok1 := args[0].(*String)
		b, ok2 := args[1].(*String)
		if ok1 && ok2 {
			return Integer(strings.Compare(a.Value, b.Value)), nil
		}
	}
	return NilValue, nil
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
func textUrlEncode(args ...Value) (Value, error) {
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
			return &String{Value: b.String()}, nil
		}
	}
	return NilValue, nil
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
func textUrlDecode(args ...Value) (Value, error) {
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
