package vida

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/alkemist-17/vida/token"
)

func loadFoundationBuilders() Value {
	m := &Object{Value: make(map[string]Value, 1)}
	m.Value["new"] = NativeFunction(buildersNew)
	return m
}

func buildersNew(ctx *Context, args ...Value) (Value, error) {
	const fn = "builders::new"

	if len(args) == 0 {
		var sb strings.Builder
		return &VidaStringBuilder{Builder: &sb}, nil
	}

	t, err := asString(fn, args, 0)
	if err != nil {
		return Nil, err
	}

	switch strings.ToLower(t.Value) {
	case stringT, "s":
		var sb strings.Builder
		return &VidaStringBuilder{Builder: &sb}, nil
	case bytesT, "b":
		return &VidaBytesBuilder{Buffer: new(bytes.Buffer)}, nil
	default:
		return Nil, runtimeErrorf(fn, "argument 1: unknown builder kind %q (expected %q, \"s\", %q, or \"b\")", t.Value, stringT, bytesT)
	}
}

// String Builder Value
type VidaStringBuilder struct {
	ReferenceSemanticsImpl
	Builder *strings.Builder
}

func (sb *VidaStringBuilder) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, ErrPrefixOpNotDefined
	}
}

func (sb *VidaStringBuilder) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return Nil, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, sb, rhs)
	default:
		return Nil, ErrBinaryOpNotDefined
	}
}

func (sb *VidaStringBuilder) Equals(ctx *Context, other Value) Bool {
	if b, ok := other.(*VidaStringBuilder); ok {
		return sb == b
	}
	return false
}

func (sb *VidaStringBuilder) String() string {
	return fmt.Sprintf("StringBuilder[%p]", sb)
}

func (sb *VidaStringBuilder) Type() string {
	return stringBuilderT
}

func (sb *VidaStringBuilder) Clone() Value {
	var newBuilder strings.Builder
	newBuilder.WriteString(sb.Builder.String())
	return &VidaStringBuilder{Builder: &newBuilder}
}

func (sb *VidaStringBuilder) ObjectKey() string {
	return sb.String()
}

func (sb *VidaStringBuilder) GetVTable(ctx *Context) Value {
	if ctx.vtables[stringBuilderT] == nil {
		ctx.loadStringBuilderVT()
	}
	return ctx.vtables[stringBuilderT]
}

func (sb *VidaStringBuilder) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[stringBuilderT] == nil {
		ctx.loadStringBuilderVT()
	}
	if vtable, ok := ctx.vtables[stringBuilderT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func stringBuilderBuildString(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.build"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return &String{Value: sb.Builder.String()}, nil
}

func stringBuilderLen(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.len"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return Integer(sb.Builder.Len()), nil
}

func stringBuilderCap(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.cap"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return Integer(sb.Builder.Cap()), nil
}

func stringBuilderIsEmpty(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.isEmpty"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return Bool(sb.Builder.Len() == 0), nil
}

func stringBuilderGrow(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.grow"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	n, err := asInteger(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	if n < 0 {
		return Nil, runtimeErrorf(fn, "argument 2: expected a non-negative Integer, got %d", n)
	}
	sb.Builder.Grow(int(n))
	return sb, nil
}

func stringBuilderReset(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.reset"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	sb.Builder.Reset()
	return sb, nil
}

func stringBuilderWriteString(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.write"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	s, err := asString(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	sb.Builder.WriteString(s.Value)
	return sb, nil
}

func stringBuilderWriteLine(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.writeLine"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	s, err := asString(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	sb.Builder.WriteString(s.Value)
	sb.Builder.WriteByte('\n')
	return sb, nil
}

func stringBuilderWriteBytes(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.writeBytes"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	b, err := asBytesValue(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	sb.Builder.Write(b.Value)
	return sb, nil
}

func stringBuilderWriteByte(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.writeByte"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	n, err := asInteger(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	if n < 0 || n > 255 {
		return Nil, runtimeErrorf(fn, "argument 2: byte value out of range: %d (expected 0-255)", n)
	}
	sb.Builder.WriteByte(byte(n))
	return sb, nil
}

func stringBuilderWriteCodePoint(ctx *Context, args ...Value) (Value, error) {
	const fn = "StringBuilder.writeCodePoint"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	sb, err := asStringBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	cp, err := asInteger(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	if cp < 0 || cp > utf8.MaxRune {
		return Nil, runtimeErrorf(fn, "argument 2: invalid code point: %d", cp)
	}
	sb.Builder.WriteRune(rune(cp))
	return sb, nil
}

// Bytes Builder Value
type VidaBytesBuilder struct {
	ReferenceSemanticsImpl
	Buffer *bytes.Buffer
}

func (bb *VidaBytesBuilder) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, ErrPrefixOpNotDefined
	}
}

func (bb *VidaBytesBuilder) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return Nil, nil
	case uint64(token.OR):
		return rhs, nil
	case uint64(token.IN):
		return IsMemberOf(ctx, bb, rhs)
	default:
		return Nil, ErrBinaryOpNotDefined
	}
}

func (bb *VidaBytesBuilder) Equals(ctx *Context, other Value) Bool {
	if o, ok := other.(*VidaBytesBuilder); ok {
		return bb == o
	}
	return false
}

func (bb *VidaBytesBuilder) String() string {
	return fmt.Sprintf("BytesBuilder[%p]", bb)
}

func (bb *VidaBytesBuilder) Type() string {
	return bytesBuilderT
}

func (bb *VidaBytesBuilder) Clone() Value {
	cp := append([]byte(nil), bb.Buffer.Bytes()...)
	return &VidaBytesBuilder{Buffer: bytes.NewBuffer(cp)}
}

func (bb *VidaBytesBuilder) ObjectKey() string {
	return bb.String()
}

func (bb *VidaBytesBuilder) GetVTable(ctx *Context) Value {
	if ctx.vtables[bytesBuilderT] == nil {
		ctx.loadBytesBuilderVT()
	}
	return ctx.vtables[bytesBuilderT]
}

func (bb *VidaBytesBuilder) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[bytesBuilderT] == nil {
		ctx.loadBytesBuilderVT()
	}
	if vtable, ok := ctx.vtables[bytesBuilderT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func bytesBuilderBuild(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.build"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	// Copy out: callers must not be able to mutate the builder's internal
	// buffer through the returned Bytes value.
	out := append([]byte(nil), bb.Buffer.Bytes()...)
	return &Bytes{Value: out}, nil
}

func bytesBuilderLen(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.len"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return Integer(bb.Buffer.Len()), nil
}

func bytesBuilderCap(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.cap"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return Integer(bb.Buffer.Cap()), nil
}

func bytesBuilderIsEmpty(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.isEmpty"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	return Bool(bb.Buffer.Len() == 0), nil
}

func bytesBuilderGrow(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.grow"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	n, err := asInteger(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	if n < 0 {
		return Nil, runtimeErrorf(fn, "argument 2: expected a non-negative Integer, got %d", n)
	}
	bb.Buffer.Grow(int(n))
	return bb, nil
}

func bytesBuilderReset(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.reset"
	if err := requireArgCount(fn, args, 1); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	bb.Buffer.Reset()
	return bb, nil
}

func bytesBuilderWriteBytes(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.write"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	b, err := asBytesValue(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	bb.Buffer.Write(b.Value)
	return bb, nil
}

func bytesBuilderWriteString(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.writeString"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	s, err := asString(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	bb.Buffer.WriteString(s.Value)
	return bb, nil
}

func bytesBuilderWriteLine(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.writeLine"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	s, err := asString(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	bb.Buffer.WriteString(s.Value)
	bb.Buffer.WriteByte('\n')
	return bb, nil
}

func bytesBuilderWriteByte(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.writeByte"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	n, err := asInteger(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	if n < 0 || n > 255 {
		return Nil, runtimeErrorf(fn, "argument 2: byte value out of range: %d (expected 0-255)", n)
	}
	bb.Buffer.WriteByte(byte(n))
	return bb, nil
}

func bytesBuilderWriteCodePoint(ctx *Context, args ...Value) (Value, error) {
	const fn = "BytesBuilder.writeCodePoint"
	if err := requireArgCount(fn, args, 2); err != nil {
		return Nil, err
	}
	bb, err := asBytesBuilder(fn, args, 0)
	if err != nil {
		return Nil, err
	}
	cp, err := asInteger(fn, args, 1)
	if err != nil {
		return Nil, err
	}
	if cp < 0 || cp > utf8.MaxRune {
		return Nil, runtimeErrorf(fn, "argument 2: invalid code point: %d", cp)
	}
	bb.Buffer.WriteRune(rune(cp))
	return bb, nil
}
