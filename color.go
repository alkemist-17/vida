package vida

import (
	"fmt"
	"math"
	"strings"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

func loadFoundationColor() Value {
	m := &Object{Value: make(map[string]Value)}
	m.Value["string"] = GFn(colorQuickSprint)
	m.Value["format"] = GFn(colorFormatQuickSprint)
	m.Value["reset"] = GFn(colorReset)
	m.Value["new"] = GFn(colorNew)
	m.Value["printPaletteChart"] = GFn(colorPaletteChart)
	return m
}

// Vida Interface
func colorQuickSprint(args ...Value) (Value, error) {
	if len(args) > 1 {
		if obj, ok := args[0].(*Object); ok {
			fgval, okfg := obj.Value[fg]
			bgval, okbg := obj.Value[bg]
			message, okMsg := args[1].(*String)
			if okfg && okbg && okMsg {
				fgColor, fgok := colorValue(fgval)
				bgColor, bgok := colorValue(bgval)
				if fgok && bgok {
					return &String{Value: Sprint256(fgColor, bgColor, message)}, nil
				}
			}
		}
	}
	return NilValue, nil
}

func colorFormatQuickSprint(args ...Value) (Value, error) {
	if len(args) > 2 {
		if obj, ok := args[0].(*Object); ok {
			fgval, okfg := obj.Value[fg]
			bgval, okbg := obj.Value[bg]
			format, okFmt := args[1].(*String)
			if okfg && okbg && okFmt {
				fgColor, fgok := colorValue(fgval)
				bgColor, bgok := colorValue(bgval)
				if fgok && bgok {
					msg, e := FormatValue(format.Value, args[2:]...)
					return &String{Value: Sprint256(fgColor, bgColor, msg)}, e
				}
			}
		}
	}
	return NilValue, nil
}

func colorNew(args ...Value) (Value, error) {
	if len(args) > 0 {
		if obj, ok := args[0].(*Object); ok {
			fgval, okfg := obj.Value[fg]
			bgval, okbg := obj.Value[bg]
			if okfg && okbg {
				fgColor, fgok := colorValue(fgval)
				bgColor, bgok := colorValue(bgval)
				if fgok && bgok {
					return generateColorInterface(NewColor().Bg(bgColor).Fg(fgColor)), nil
				}
			}
		}
	}
	return generateColorInterface(NewColor()), nil
}

func colorReset(args ...Value) (Value, error) {
	fmt.Print(Sprint256(-1, -1, ""))
	return NilValue, nil
}

func colorPaletteChart(args ...Value) (Value, error) {
	fmt.Println("\n\n\n256 Color Palette Chart")
	// Standard Colors (0-15)
	printSection("Standard Colors (0-15)", 0, 15)
	// 216 Color Cube (16-231)
	printSection("216 Color Cube (16-231)", 16, 231)
	// Grayscale (232-255)
	printSection("Grayscale (232-255)", 232, 255)
	fmt.Printf("\nPalette Chart complete.\n\n\n")
	return NilValue, nil
}

// Library
// Value Color represents a 256-color style configuration.
type Color struct {
	ReferenceSemanticsImpl      // Reference semantics default impl
	fg                     *int // Foreground color (0-255)
	bg                     *int // Background color (0-255)
	reset                  bool
	width                  int
}

func (c *Color) Boolean() Bool {
	return true
}

func (c *Color) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return Bool(true), nil
	default:
		return NilValue, verror.ErrPrefixOpNotDefined
	}
}

func (c *Color) Binop(op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return NilValue, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(c, rhs)
	default:
		return NilValue, verror.ErrBinaryOpNotDefined
	}
}

func (c *Color) Equals(other Value) Bool {
	if val, ok := other.(*Color); ok {
		return c == val
	}
	return false
}

func (c *Color) String() string {
	return fmt.Sprintf("Color(%p)", c)
}

func (c *Color) Type() string {
	return "color"
}

func (c *Color) Clone() Value {
	return &Color{fg: c.fg, bg: c.bg, reset: c.reset}
}

func (c *Color) ObjectKey() string {
	return fmt.Sprintf("Color(%p)", c)
}

// ANSI Codes
const (
	resetCode = "\x1b[0m"
	fgPrefix  = "\x1b[38;5;"
	bgPrefix  = "\x1b[48;5;"
	suffix    = "m"
	colorMin  = 0
	colorMax  = 255
	defaultFG = 15
	defaultBG = 0
	colorName = "color"
	fg        = "fg"
	bg        = "bg"
)

// NewColor creates a new Color instance.
func NewColor() *Color {
	return &Color{reset: true}
}

// Fg sets the foreground color (0-255).
func (c *Color) Fg(color int) *Color {
	if color < colorMin || color > colorMax {
		color = defaultFG // Default to white if color is out of bounds
	}
	c.fg = &color
	return c
}

// Bg sets the background color (0-255).
func (c *Color) Bg(color int) *Color {
	if color < colorMin || color > colorMax {
		color = defaultBG // Default to black if color is out of bounds
	}
	c.bg = &color
	return c
}

// NoReset disables the automatic reset code at the end of the string.
// Use this if you want to chain colors manually.
func (c *Color) NoReset() *Color {
	c.reset = false
	return c
}

func (c *Color) Reset() *Color {
	c.reset = true
	return c
}

// Sprint formats the text with the configured colors.
func (c *Color) Sprint(a ...any) string {
	var sb strings.Builder

	// Build Foreground
	if c.fg != nil {
		sb.WriteString(fgPrefix)
		fmt.Fprintf(&sb, "%d", *c.fg)
		sb.WriteString(suffix)
	}

	// Build Background
	if c.bg != nil {
		sb.WriteString(bgPrefix)
		fmt.Fprintf(&sb, "%d", *c.bg)
		sb.WriteString(suffix)
	}

	// Add Text
	fmt.Fprint(&sb, a...)

	// Add Reset
	if c.reset {
		sb.WriteString(resetCode)
	}

	return sb.String()
}

func (c *Color) Fill(a ...any) string {
	text := fmt.Sprint(a...)

	// If no width is set, just return normal Sprint
	if c.width <= 0 {
		return c.Sprint(text)
	}

	// Calculate padding needed
	// Note: ANSI codes don't count toward visual width
	textLen := min(len(text), c.width) // Truncate if text is too long
	padding := c.width - textLen

	// Pad with spaces
	paddedText := text + strings.Repeat(" ", padding)

	// Apply colors to the entire padded string
	return c.Sprint(paddedText)
}

// Sprint256 is a quick helper: Sprint256(fg, bg, text)
// Pass -1 for fg or bg if you don't want to set that attribute.
func Sprint256(fg, bg int, a ...any) string {
	c := NewColor()
	if fg >= 0 {
		c.Fg(fg)
	}
	if bg >= 0 {
		c.Bg(bg)
	}
	return c.Sprint(a...)
}

// Color Value Interface
func generateColorInterface(color *Color) Value {
	o := &Object{Value: make(map[string]Value)}
	o.Value[colorName] = color
	o.Value["string"] = colorString()
	o.Value["format"] = colorFormat()
	o.Value["bg"] = GFn(colorSetBG)
	o.Value["fg"] = GFn(colorSetFG)
	o.Value["reset"] = GFn(colorSetReset)
	o.Value["resets"] = GFn(colorGetReset)
	return o
}

func colorString() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 1 {
			if obj, ok := args[0].(*Object); ok {
				if c, ok := obj.Value[colorName].(*Color); ok {
					return &String{Value: c.Sprint(args[1])}, nil
				}
			}
		}
		return NilValue, nil
	}
}

func colorFormat() GFn {
	return func(args ...Value) (Value, error) {
		if len(args) > 2 {
			if obj, ok := args[0].(*Object); ok {
				if c, ok := obj.Value[colorName].(*Color); ok {
					if format, ok := args[1].(*String); ok {
						message, e := FormatValue(format.Value, args[2:]...)
						return &String{Value: c.Sprint(message)}, e
					}
				}
			}
		}
		return NilValue, nil
	}
}

func colorSetBG(args ...Value) (Value, error) {
	if len(args) > 1 {
		if obj, ok := args[0].(*Object); ok {
			if c, ok := obj.Value[colorName].(*Color); ok {
				if val, ok := args[1].(Integer); ok {
					c.Bg(int((Integer(math.Abs(float64(val))) % 256)))
					return obj, nil
				}
			}
		}
	}
	return NilValue, nil
}

func colorSetFG(args ...Value) (Value, error) {
	if len(args) > 1 {
		if obj, ok := args[0].(*Object); ok {
			if c, ok := obj.Value[colorName].(*Color); ok {
				if val, ok := args[1].(Integer); ok {
					c.Fg(int((Integer(math.Abs(float64(val))) % 256)))
					return obj, nil
				}
			}
		}
	}
	return NilValue, nil
}

func colorSetReset(args ...Value) (Value, error) {
	if len(args) > 1 {
		if obj, ok := args[0].(*Object); ok {
			if c, ok := obj.Value[colorName].(*Color); ok {
				if val, ok := args[1].(Bool); ok {
					c.reset = bool(val)
					return obj, nil
				}
			}
		}
	}
	return NilValue, nil
}

func colorGetReset(args ...Value) (Value, error) {
	if len(args) > 0 {
		if obj, ok := args[0].(*Object); ok {
			if c, ok := obj.Value[colorName].(*Color); ok {
				return Bool(c.reset), nil
			}
		}
	}
	return NilValue, nil
}

func printSection(title string, start, end int) {
	fmt.Printf("\n%s:\n", title)

	// Print rowCount colors per row
	rowCount := 10
	count := 0

	for i := start; i <= end; i++ {
		// Create a colored block with the number inside
		// We use Sprint256 to color the text, but we need a background to see the color well
		// For the chart, let's color the Background with the color code, and text white/black

		// Determine text color based on brightness (simple heuristic)
		textFg := 15 // White
		if i > 230 { // Grayscale gets dark text on light bg
			textFg = 0
		}

		// Format the number to be 4 digits for alignment
		codeStr := fmt.Sprintf(" %3d ", i)

		// Print the block: BG = i, FG = contrast
		fmt.Print(Sprint256(textFg, i, codeStr))

		count++
		if count%rowCount == 0 {
			fmt.Println() // New line every rowCount colors
		}
	}

	// Ensure we end with a newline if the last row wasn't full
	if count%rowCount != 0 {
		fmt.Println()
	}

	// Reset after section
	fmt.Print(Sprint256(-1, -1, ""))
}

func colorValue(val Value) (int, bool) {
	switch val := val.(type) {
	case Integer:
		return int(val), true
	case Nil:
		return -1, true
	default:
		return 0, false
	}
}
