package vida

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/alkemist-17/vida/token"
)

func loadFoundationStyle() Value {
	m := &Object{Value: make(map[string]Value, 6)}
	m.Value["new"] = NativeFunction(styleNewFn)
	m.Value["hex"] = NativeFunction(hexToRGBFn)
	m.Value["rgb"] = NativeFunction(rgbToHexFn)
	m.Value["name"] = NativeFunction(nameToRGBFn)
	m.Value["lerp"] = NativeFunction(lerpFn)
	m.Value["reset"] = NativeFunction(styleResetFn)
	m.Value["showcase"] = NativeFunction(RunColorShowcase)
	return m
}

type RGB struct {
	R, G, B uint8
}

func (c RGB) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

type Style struct {
	ReferenceSemanticsImpl
	fg, bg                                               *RGB
	fgIdx                                                *uint8
	bgIdx                                                *uint8
	fg16, bg16                                           *uint8
	width                                                int
	bold, dim, italic, underline, blink, reverse, strike bool
	noReset                                              bool
	enabled                                              bool
}

func NewStyle() *Style {
	return &Style{enabled: true}
}

func (s *Style) clone() *Style {
	cp := *s
	return &cp
}

func (s *Style) FgRGB(r, g, b uint8) *Style { c := s.clone(); c.fg = &RGB{r, g, b}; return c }
func (s *Style) BgRGB(r, g, b uint8) *Style { c := s.clone(); c.bg = &RGB{r, g, b}; return c }

func (s *Style) FgHex(hex string) (*Style, error) {
	rgb, err := ParseHex(hex)
	if err != nil {
		return s, err
	}
	c := s.clone()
	c.fg = &rgb
	return c, nil
}

func (s *Style) BgHex(hex string) (*Style, error) {
	rgb, err := ParseHex(hex)
	if err != nil {
		return s, err
	}
	c := s.clone()
	c.bg = &rgb
	return c, nil
}

func (s *Style) FgName(name string) (*Style, error) {
	rgb, ok := namedColors[strings.ToLower(name)]
	if !ok {
		return s, fmt.Errorf("std.style: unknown color name %q", name)
	}
	c := s.clone()
	c.fg = &rgb
	return c, nil
}

func (s *Style) BgName(name string) (*Style, error) {
	rgb, ok := namedColors[strings.ToLower(name)]
	if !ok {
		return s, fmt.Errorf("std.style: unknown color name %q", name)
	}
	c := s.clone()
	c.bg = &rgb
	return c, nil
}

func (s *Style) Fg256(idx uint8) *Style { c := s.clone(); c.fgIdx = &idx; return c }
func (s *Style) Bg256(idx uint8) *Style { c := s.clone(); c.bgIdx = &idx; return c }

func (s *Style) Fg16(idx uint8) *Style { c := s.clone(); idx %= 16; c.fg16 = &idx; return c }
func (s *Style) Bg16(idx uint8) *Style { c := s.clone(); idx %= 16; c.bg16 = &idx; return c }

func (s *Style) Bold() *Style      { c := s.clone(); c.bold = true; return c }
func (s *Style) Dim() *Style       { c := s.clone(); c.dim = true; return c }
func (s *Style) Italic() *Style    { c := s.clone(); c.italic = true; return c }
func (s *Style) Underline() *Style { c := s.clone(); c.underline = true; return c }
func (s *Style) Blink() *Style     { c := s.clone(); c.blink = true; return c }
func (s *Style) Reverse() *Style   { c := s.clone(); c.reverse = true; return c }
func (s *Style) Strike() *Style    { c := s.clone(); c.strike = true; return c }
func (s *Style) NoReset() *Style   { c := s.clone(); c.noReset = true; return c }

func (s *Style) Width(w int) *Style     { c := s.clone(); c.width = w; return c }
func (s *Style) Enabled(on bool) *Style { c := s.clone(); c.enabled = on; return c }

func (s *Style) prefix() string {
	var codes []string
	if s.bold {
		codes = append(codes, "1")
	}
	if s.dim {
		codes = append(codes, "2")
	}
	if s.italic {
		codes = append(codes, "3")
	}
	if s.underline {
		codes = append(codes, "4")
	}
	if s.blink {
		codes = append(codes, "5")
	}
	if s.reverse {
		codes = append(codes, "7")
	}
	if s.strike {
		codes = append(codes, "9")
	}

	switch {
	case s.fg != nil:
		codes = append(codes, "38;2", strconv.Itoa(int(s.fg.R)), strconv.Itoa(int(s.fg.G)), strconv.Itoa(int(s.fg.B)))
	case s.fgIdx != nil:
		codes = append(codes, "38;5", strconv.Itoa(int(*s.fgIdx)))
	case s.fg16 != nil:
		codes = append(codes, strconv.Itoa(ansi16Code(*s.fg16, false)))
	}
	switch {
	case s.bg != nil:
		codes = append(codes, "48;2", strconv.Itoa(int(s.bg.R)), strconv.Itoa(int(s.bg.G)), strconv.Itoa(int(s.bg.B)))
	case s.bgIdx != nil:
		codes = append(codes, "48;5", strconv.Itoa(int(*s.bgIdx)))
	case s.bg16 != nil:
		codes = append(codes, strconv.Itoa(ansi16Code(*s.bg16, true)))
	}

	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// FgHexMust is FgHex without the error return, for call sites using a
// compile-time-known-valid literal. Panics on malformed input — do not
// use with user- or script-supplied strings.
func (s *Style) FgHexMust(hex string) *Style {
	c, err := s.FgHex(hex)
	if err != nil {
		panic(err)
	}
	return c
}

// BgHexMust is BgHex without the error return, for call sites using a
// compile-time-known-valid literal. Panics on malformed input — do not
// use with user- or script-supplied strings.
func (s *Style) BgHexMust(hex string) *Style {
	c, err := s.BgHex(hex)
	if err != nil {
		panic(err)
	}
	return c
}

func ansi16Code(idx uint8, background bool) int {
	base := 30
	if idx >= 8 {
		base = 90
		idx -= 8
	}
	if background {
		base += 10
	}
	return base + int(idx)
}

func (s *Style) Sprint(a ...any) string {
	if !s.enabled {
		return fmt.Sprint(a...)
	}
	var sb strings.Builder
	sb.WriteString(s.prefix())
	fmt.Fprint(&sb, a...)
	if !s.noReset {
		sb.WriteString(resetCode)
	}
	return sb.String()
}

func (s *Style) Sprintf(format string, a ...any) string {
	return s.Sprint(fmt.Sprintf(format, a...))
}

func (s *Style) Fill(a ...any) string {
	text := fmt.Sprint(a...)
	if s.width <= 0 {
		return s.Sprint(text)
	}
	runes := []rune(text)
	if len(runes) > s.width {
		runes = runes[:s.width]
	}
	padded := string(runes) + strings.Repeat(" ", s.width-len(runes))
	return s.Sprint(padded)
}

func ParseHex(hex string) (RGB, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return RGB{}, fmt.Errorf("std.style: invalid hex %q", hex)
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return RGB{}, fmt.Errorf("std.style: invalid hex %q: %w", hex, err)
	}
	return RGB{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v)}, nil
}

func Lerp(a, b RGB, t float64) RGB {
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}
	return RGB{lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B)}
}

const resetCode = "\x1b[0m"

// ---- Value interface impl for *Style -----------------------------------

func (s *Style) Boolean() Bool { return true }

func (s *Style) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, ErrPrefixOpNotDefined
	}
}

func (s *Style) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.OR):
		return s, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, s, rhs)
	default:
		return Nil, ErrBinaryOpNotDefined
	}
}

func (s *Style) Equals(ctx *Context, other Value) Bool {
	if val, ok := other.(*Style); ok {
		return s == val
	}
	return false
}

func (s *Style) String() string    { return fmt.Sprintf("style[%p]", s) }
func (s *Style) Type() string      { return styleT }
func (s *Style) Clone() Value      { return s.clone() }
func (s *Style) ObjectKey() string { return s.String() }

func (s *Style) GetVTable(ctx *Context) Value {
	if ctx.vtables[styleT] == nil {
		ctx.loadStyleVT()
	}
	return ctx.vtables[styleT]
}

func (s *Style) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[styleT] == nil {
		ctx.loadStyleVT()
	}
	if vtable, ok := ctx.vtables[styleT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func styleNewFn(ctx *Context, args ...Value) (Value, error) {
	return NewStyle(), nil
}

func styleResetFn(ctx *Context, args ...Value) (Value, error) {
	fmt.Print(resetCode)
	return Nil, nil
}

// ---- RGB <-> hex module helpers -----------------------------------------

// hex(str) -> {r, g, b} object, or an error if malformed
func hexToRGBFn(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return Nil, ErrInvalidNumberOfArguments
	}
	str, ok := args[0].(*String)
	if !ok {
		return Nil, ErrExpectedString
	}
	rgb, err := ParseHex(str.Value)
	if err != nil {
		return Nil, err
	}
	return rgbToObject(rgb), nil
}

// rgb(r, g, b) -> "#rrggbb" string
func rgbToHexFn(ctx *Context, args ...Value) (Value, error) {
	rgb, ok := objectFromArgs(args)
	if !ok {
		return Nil, ErrInvalidNumberOfArguments
	}
	return &String{Value: rgb.Hex()}, nil
}

// name(str) -> {r, g, b} object, or an error if the name isn't in the
// namedColors table. Lets scripts resolve a named color to raw RGB
func nameToRGBFn(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 1 {
		return Nil, ErrInvalidNumberOfArguments
	}
	str, ok := args[0].(*String)
	if !ok {
		return Nil, ErrExpectedString
	}
	rgb, ok := namedColors[strings.ToLower(str.Value)]
	if !ok {
		return Nil, fmt.Errorf("vida.style: unknown color name %q", str.Value)
	}
	return rgbToObject(rgb), nil
}

// lerp(rgbA, rgbB, t) -> {r, g, b} object interpolated
func lerpFn(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 3 {
		return Nil, ErrInvalidNumberOfArguments
	}
	a, aok := objectToRGB(args[0])
	b, bok := objectToRGB(args[1])
	t, tok := toFloat(args[2])
	if !aok || !bok || !tok {
		return Nil, ErrInvalidTypeOfArgument
	}
	return rgbToObject(Lerp(a, b, t)), nil
}

func rgbToObject(c RGB) *Object {
	return &Object{Value: map[string]Value{
		"r": Integer(c.R),
		"g": Integer(c.G),
		"b": Integer(c.B),
	}}
}

func objectToRGB(v Value) (RGB, bool) {
	obj, ok := v.(*Object)
	if !ok {
		return RGB{}, false
	}
	r, rok := obj.Value["r"].(Integer)
	g, gok := obj.Value["g"].(Integer)
	b, bok := obj.Value["b"].(Integer)
	if !rok || !gok || !bok {
		return RGB{}, false
	}
	return RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, true
}

// objectFromArgs supports both rgb({r,g,b}) and rgb(r, g, b) call shapes.
func objectFromArgs(args []Value) (RGB, bool) {
	if len(args) == 1 {
		return objectToRGB(args[0])
	}
	if len(args) >= 3 {
		r, rok := args[0].(Integer)
		g, gok := args[1].(Integer)
		b, bok := args[2].(Integer)
		if rok && gok && bok {
			return RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, true
		}
	}
	return RGB{}, false
}

func toFloat(v Value) (float64, bool) {
	switch v := v.(type) {
	case Integer:
		return float64(v), true
	case Float:
		return float64(v), true
	default:
		return 0, false
	}
}

// self returns args[0] as *Style, or ok=false.
func styleSelf(args []Value) (*Style, bool) {
	if len(args) < 1 {
		return nil, false
	}
	s, ok := args[0].(*Style)
	return s, ok
}

func styleFgRGB(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	rgb, rok := objectFromArgs(args[1:])
	if !ok || !rok {
		return Nil, ErrInvalidTypeOfArgument
	}
	return s.FgRGB(rgb.R, rgb.G, rgb.B), nil
}

func styleBgRGB(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	rgb, rok := objectFromArgs(args[1:])
	if !ok || !rok {
		return Nil, ErrInvalidTypeOfArgument
	}
	return s.BgRGB(rgb.R, rgb.G, rgb.B), nil
}

func styleFgHex(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	str, sok := args[1].(*String)
	if !sok {
		return Nil, ErrExpectedString
	}
	return s.FgHex(str.Value)
}

func styleBgHex(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	str, sok := args[1].(*String)
	if !sok {
		return Nil, ErrExpectedString
	}
	return s.BgHex(str.Value)
}

func styleFgName(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	str, sok := args[1].(*String)
	if !sok {
		return Nil, ErrExpectedString
	}
	return s.FgName(str.Value)
}

func styleBgName(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	str, sok := args[1].(*String)
	if !sok {
		return Nil, ErrExpectedString
	}
	return s.BgName(str.Value)
}

func styleFg256(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	idx, iok := args[1].(Integer)
	if !iok {
		return Nil, ErrExpectedInteger
	}
	return s.Fg256(uint8(idx)), nil
}

func styleBg256(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	idx, iok := args[1].(Integer)
	if !iok {
		return Nil, ErrExpectedInteger
	}
	return s.Bg256(uint8(idx)), nil
}

func styleFg16(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	idx, iok := args[1].(Integer)
	if !iok {
		return Nil, ErrExpectedInteger
	}
	return s.Fg16(uint8(idx)), nil
}

func styleBg16(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	idx, iok := args[1].(Integer)
	if !iok {
		return Nil, ErrExpectedInteger
	}
	return s.Bg16(uint8(idx)), nil
}

// styleFlag wraps a no-arg Style->Style transform (Bold, Dim, ...) into a
// NativeFunction so the 7 boolean toggles don't each need boilerplate.
func styleFlag(f func(*Style) *Style) func(*Context, ...Value) (Value, error) {
	return func(ctx *Context, args ...Value) (Value, error) {
		s, ok := styleSelf(args)
		if !ok {
			return Nil, ErrInvalidNumberOfArguments
		}
		return f(s), nil
	}
}

func styleWidth(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	w, wok := args[1].(Integer)
	if !wok {
		return Nil, ErrExpectedInteger
	}
	return s.Width(int(w)), nil
}

func styleEnabled(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	on, bok := args[1].(Bool)
	if !bok {
		return Nil, ErrExpectedBool
	}
	return s.Enabled(bool(on)), nil
}

func styleSprint(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	msg, mok := args[1].(*String)
	if !mok {
		return Nil, ErrExpectedString
	}
	return &String{Value: s.Sprint(msg.Value)}, nil
}

func styleSprintf(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	format, fok := args[1].(*String)
	if !fok {
		return Nil, ErrExpectedString
	}
	msg, err := VSprintf(format.Value, args[2:]...)
	if err != nil {
		return Nil, err
	}
	if !s.enabled {
		return &String{Value: msg}, nil
	}
	return &String{Value: s.Sprint(msg)}, nil
}

func styleFill(ctx *Context, args ...Value) (Value, error) {
	s, ok := styleSelf(args)
	if !ok || len(args) < 2 {
		return Nil, ErrInvalidNumberOfArguments
	}
	msg, mok := args[1].(*String)
	if !mok {
		return Nil, ErrExpectedString
	}
	return &String{Value: s.Fill(msg.Value)}, nil
}

// namedColors is the full CSS3/X11 standard color name table (147 names,
// gray/grey spelling variants included as aliases to the same RGB).
var namedColors = map[string]RGB{
	"aliceblue":            {240, 248, 255},
	"antiquewhite":         {250, 235, 215},
	"aqua":                 {0, 255, 255},
	"aquamarine":           {127, 255, 212},
	"azure":                {240, 255, 255},
	"beige":                {245, 245, 220},
	"bisque":               {255, 228, 196},
	"black":                {0, 0, 0},
	"blanchedalmond":       {255, 235, 205},
	"blue":                 {0, 0, 255},
	"blueviolet":           {138, 43, 226},
	"brown":                {165, 42, 42},
	"burlywood":            {222, 184, 135},
	"cadetblue":            {95, 158, 160},
	"chartreuse":           {127, 255, 0},
	"chocolate":            {210, 105, 30},
	"coral":                {255, 127, 80},
	"cornflowerblue":       {100, 149, 237},
	"cornsilk":             {255, 248, 220},
	"crimson":              {220, 20, 60},
	"cyan":                 {0, 255, 255},
	"darkblue":             {0, 0, 139},
	"darkcyan":             {0, 139, 139},
	"darkgoldenrod":        {184, 134, 11},
	"darkgray":             {169, 169, 169},
	"darkgrey":             {169, 169, 169},
	"darkgreen":            {0, 100, 0},
	"darkkhaki":            {189, 183, 107},
	"darkmagenta":          {139, 0, 139},
	"darkolivegreen":       {85, 107, 47},
	"darkorange":           {255, 140, 0},
	"darkorchid":           {153, 50, 204},
	"darkred":              {139, 0, 0},
	"darksalmon":           {233, 150, 122},
	"darkseagreen":         {143, 188, 143},
	"darkslateblue":        {72, 61, 139},
	"darkslategray":        {47, 79, 79},
	"darkslategrey":        {47, 79, 79},
	"darkturquoise":        {0, 206, 209},
	"darkviolet":           {148, 0, 211},
	"deeppink":             {255, 20, 147},
	"deepskyblue":          {0, 191, 255},
	"dimgray":              {105, 105, 105},
	"dimgrey":              {105, 105, 105},
	"dodgerblue":           {30, 144, 255},
	"firebrick":            {178, 34, 34},
	"floralwhite":          {255, 250, 240},
	"forestgreen":          {34, 139, 34},
	"fuchsia":              {255, 0, 255},
	"gainsboro":            {220, 220, 220},
	"ghostwhite":           {248, 248, 255},
	"gold":                 {255, 215, 0},
	"goldenrod":            {218, 165, 32},
	"gray":                 {128, 128, 128},
	"grey":                 {128, 128, 128},
	"green":                {0, 128, 0},
	"greenyellow":          {173, 255, 47},
	"honeydew":             {240, 255, 240},
	"hotpink":              {255, 105, 180},
	"indianred":            {205, 92, 92},
	"indigo":               {75, 0, 130},
	"ivory":                {255, 255, 240},
	"khaki":                {240, 230, 140},
	"lavender":             {230, 230, 250},
	"lavenderblush":        {255, 240, 245},
	"lawngreen":            {124, 252, 0},
	"lemonchiffon":         {255, 250, 205},
	"lightblue":            {173, 216, 230},
	"lightcoral":           {240, 128, 128},
	"lightcyan":            {224, 255, 255},
	"lightgoldenrodyellow": {250, 250, 210},
	"lightgray":            {211, 211, 211},
	"lightgrey":            {211, 211, 211},
	"lightgreen":           {144, 238, 144},
	"lightpink":            {255, 182, 193},
	"lightsalmon":          {255, 160, 122},
	"lightseagreen":        {32, 178, 170},
	"lightskyblue":         {135, 206, 250},
	"lightslategray":       {119, 136, 153},
	"lightslategrey":       {119, 136, 153},
	"lightsteelblue":       {176, 196, 222},
	"lightyellow":          {255, 255, 224},
	"lime":                 {0, 255, 0},
	"limegreen":            {50, 205, 50},
	"linen":                {250, 240, 230},
	"magenta":              {255, 0, 255},
	"maroon":               {128, 0, 0},
	"mediumaquamarine":     {102, 205, 170},
	"mediumblue":           {0, 0, 205},
	"mediumorchid":         {186, 85, 211},
	"mediumpurple":         {147, 112, 219},
	"mediumseagreen":       {60, 179, 113},
	"mediumslateblue":      {123, 104, 238},
	"mediumspringgreen":    {0, 250, 154},
	"mediumturquoise":      {72, 209, 204},
	"mediumvioletred":      {199, 21, 133},
	"midnightblue":         {25, 25, 112},
	"mintcream":            {245, 255, 250},
	"mistyrose":            {255, 228, 225},
	"moccasin":             {255, 228, 181},
	"navajowhite":          {255, 222, 173},
	"navy":                 {0, 0, 128},
	"oldlace":              {253, 245, 230},
	"olive":                {128, 128, 0},
	"olivedrab":            {107, 142, 35},
	"orange":               {255, 165, 0},
	"orangered":            {255, 69, 0},
	"orchid":               {218, 112, 214},
	"palegoldenrod":        {238, 232, 170},
	"palegreen":            {152, 251, 152},
	"paleturquoise":        {175, 238, 238},
	"palevioletred":        {219, 112, 147},
	"papayawhip":           {255, 239, 213},
	"peachpuff":            {255, 218, 185},
	"peru":                 {205, 133, 63},
	"pink":                 {255, 192, 203},
	"plum":                 {221, 160, 221},
	"powderblue":           {176, 224, 230},
	"purple":               {128, 0, 128},
	"rebeccapurple":        {102, 51, 153},
	"red":                  {255, 0, 0},
	"rosybrown":            {188, 143, 143},
	"royalblue":            {65, 105, 225},
	"saddlebrown":          {139, 69, 19},
	"salmon":               {250, 128, 114},
	"sandybrown":           {244, 164, 96},
	"seagreen":             {46, 139, 87},
	"seashell":             {255, 245, 238},
	"sienna":               {160, 82, 45},
	"silver":               {192, 192, 192},
	"skyblue":              {135, 206, 235},
	"slateblue":            {106, 90, 205},
	"slategray":            {112, 128, 144},
	"slategrey":            {112, 128, 144},
	"snow":                 {255, 250, 250},
	"springgreen":          {0, 255, 127},
	"steelblue":            {70, 130, 180},
	"tan":                  {210, 180, 140},
	"teal":                 {0, 128, 128},
	"thistle":              {216, 191, 216},
	"tomato":               {255, 99, 71},
	"turquoise":            {64, 224, 208},
	"violet":               {238, 130, 238},
	"wheat":                {245, 222, 179},
	"white":                {255, 255, 255},
	"whitesmoke":           {245, 245, 245},
	"yellow":               {255, 255, 0},
	"yellowgreen":          {154, 205, 50},
	"mint":                 {189, 252, 201},
}

// RGB truecolor + Lerp.
func RunColorShowcase(ctx *Context, args ...Value) (Value, error) {
	fmt.Print("\n")
	bannerGradient()
	fmt.Print("\n")
	namedColorShowcase()
	fmt.Print("\n")
	hexRoundTrip()
	fmt.Print("\n")
	styleCombinations()
	fmt.Print("\n")
	rainbowSweep()
	fmt.Print("\n")
	fmt.Println(NewStyle().Dim().Sprint("\n\nShowcase complete.\n\n"))
	fmt.Print("\n")
	return Nil, nil
}

// ---- 1. Gradient banner: interpolate fg color letter-by-letter --------

func bannerGradient() {
	text := "   VIDA COLOR STYLE — truecolor styling, done right   "
	from := RGB{255, 0, 128} // hot pink
	to := RGB{0, 200, 255}   // cyan-blue

	runes := []rune(text)
	n := len(runes)
	var sb strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(max(n-1, 1))
		c := Lerp(from, to, t)
		sb.WriteString(NewStyle().FgRGB(c.R, c.G, c.B).Bold().Sprint(string(r)))
	}
	fmt.Printf("\n\n%s\n\n\n", sb.String())
}

// ---- 2. Named colors, swatch + label -----------------------------------

func namedColorShowcase() {
	fmt.Println(NewStyle().Underline().Sprint("Named colors\n\n"))
	names := []string{
		"tomato", "gold", "mint", "skyblue", "orchid",
		"crimson", "forestgreen", "royalblue", "hotpink", "chocolate",
	}
	var row strings.Builder
	for _, name := range names {
		sw := NewStyle()
		sw, _ = sw.BgName(name)
		sw = sw.Width(3)
		sw, _ = sw.FgName("black")
		row.WriteString(sw.Fill(""))
		row.WriteString(" ")
		row.WriteString(NewStyle().Sprint(name))
		row.WriteString("\n")
		fmt.Println(row.String())
		row.Reset()
	}
	if row.Len() > 0 {
		fmt.Println(row.String())
	}
}

// ---- 3. Hex <-> RGB round trip ------------------------------------------

func hexRoundTrip() {
	fmt.Println(NewStyle().Underline().Sprint("Hex <-> RGB\n\n"))
	hexes := []string{"#FF6B35", "#4ECDC4", "#A5D8FF", "#1A535C"}
	for _, h := range hexes {
		c, err := ParseHex(h)
		if err != nil {
			continue
		}
		swatch := NewStyle().BgRGB(c.R, c.G, c.B).Width(4).Sprint()
		fmt.Printf("%s  %s  -> rgb(%3d, %3d, %3d) -> %s\n",
			swatch, h, c.R, c.G, c.B, c.Hex())
	}
}

// ---- 4. Style combinations table ----------------------------------------

func styleCombinations() {
	fmt.Println(NewStyle().Underline().Sprint("\n\nStyle combinations\n\n"))
	base := NewStyle().FgHexMust("#EAEAEA")
	combos := []struct {
		label string
		style *Style
	}{
		{"bold", base.Bold()},
		{"italic", base.Italic()},
		{"underline", base.Underline()},
		{"dim", base.Dim()},
		{"reverse", base.Reverse()},
		{"strike", base.Strike()},
		{"bold+underline", base.Bold().Underline()},
		{"bold+bg navy", func() *Style { s, _ := base.Bold().BgName("navy"); return s }()},
	}
	for _, c := range combos {
		fmt.Printf("  %-18s %s\n", c.label, c.style.Sprint("The quick brown fox"))
	}
}

// ---- 5. Rainbow sweep: full HSV wheel via RGB interpolation ------------

func rainbowSweep() {
	fmt.Println(NewStyle().Underline().Sprint("\n\nRainbow sweep (HSV -> RGB)\n\n"))
	width := 60
	for row := range 3 {
		var sb strings.Builder
		for i := range width {
			hue := float64(i) / float64(width) * 360.0
			c := hsvToRGB(hue, 0.85, 0.95-float64(row)*0.2)
			sb.WriteString(NewStyle().BgRGB(c.R, c.G, c.B).Sprint(" "))
		}
		fmt.Println(sb.String())
	}
}

// hsvToRGB is a small local helper purely for the demo's rainbow sweep;
// not part of the public Style API surface.
func hsvToRGB(h, s, v float64) RGB {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return RGB{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
	}
}
