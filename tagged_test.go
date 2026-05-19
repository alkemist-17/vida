package vida

import (
	"math"
	"testing"
	"time"
	"unsafe"
)

func TestIntValZeroAlloc(t *testing.T) {
	var sink int64
	allocs := testing.AllocsPerRun(1000000, func() {
		v := IntVal(42)
		sink = v.Int()
	})
	_ = sink
	if allocs > 0 {
		t.Fatalf("got %v allocs, want 0", allocs)
	}
}

func TestFloatValZeroAlloc(t *testing.T) {
	var sink float64
	allocs := testing.AllocsPerRun(1000000, func() {
		v := FloatVal(1e+6)
		sink = v.Float()
	})
	_ = sink
	if allocs > 0 {
		t.Fatalf("got %v allocs, want 0", allocs)
	}
}

func TestBoolTrueValZeroAlloc(t *testing.T) {
	var sink bool
	allocs := testing.AllocsPerRun(1000000, func() {
		v := BoolVal(true)
		sink = v.Bool()
	})
	_ = sink
	if allocs > 0 {
		t.Fatalf("got %v allocs, want 0", allocs)
	}
}

func TestBoolFalseValZeroAlloc(t *testing.T) {
	var sink bool
	allocs := testing.AllocsPerRun(1000000, func() {
		v := BoolVal(false)
		sink = v.Bool()
	})
	_ = sink
	if allocs > 0 {
		t.Fatalf("got %v allocs, want 0", allocs)
	}
}

func TestNilValZeroAlloc(t *testing.T) {
	var nil TValue
	allocs := testing.AllocsPerRun(1000000, func() {
		nil = NilVal()
	})
	_ = nil
	if allocs > 0 {
		t.Fatalf("got %v allocs, want 0", allocs)
	}
}

func checkTag(t *testing.T, v TValue, want uint8) {
	t.Helper()
	if v.TType() != want {
		t.Fatalf("tag: got %d, want %d", v.TType(), want)
	}
}

func TestNilValTag(t *testing.T) {
	checkTag(t, NilVal(), TNil)
}

func TestNilValIsZeroTValue(t *testing.T) {
	var zero TValue
	if zero != NilVal() {
		t.Fatal("zero TValue is not equal to NilVal()")
	}
}

func TestBoolValTrueTag(t *testing.T) {
	checkTag(t, BoolVal(true), TBool)
}

func TestBoolValFalseTag(t *testing.T) {
	checkTag(t, BoolVal(false), TBool)
}

func TestBoolValRoundTrip(t *testing.T) {
	if !BoolVal(true).Bool() {
		t.Fatal("BoolVal(true).Bool() returned false")
	}
	if BoolVal(false).Bool() {
		t.Fatal("BoolVal(false).Bool() returned true")
	}
}

func TestBoolValDistinct(t *testing.T) {
	if BoolVal(true) == BoolVal(false) {
		t.Fatal("BoolVal(true) == BoolVal(false)")
	}
}

func TestIntValTag(t *testing.T) {
	checkTag(t, IntVal(0), TInt)
}

func TestIntValRoundTrip(t *testing.T) {
	cases := []int64{0, 1, -1, math.MaxInt64, math.MinInt64, 42, -99999}
	for _, n := range cases {
		if got := IntVal(n).Int(); got != n {
			t.Fatalf("IntVal(%d).Int() = %d", n, got)
		}
	}
}

func TestIntValZeroAllocAccessor(t *testing.T) {
	v := IntVal(7)
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Int() })
	if allocs > 0 {
		t.Fatalf("Int() accessor: got %v allocs, want 0", allocs)
	}
}

func TestFloatValTag(t *testing.T) {
	checkTag(t, FloatVal(0), TFloat)
}

func TestFloatValRoundTrip(t *testing.T) {
	cases := []float64{0, 1.5, -1.5, math.MaxFloat64, math.SmallestNonzeroFloat64, math.Pi, math.E}
	for _, f := range cases {
		if got := FloatVal(f).Float(); got != f {
			t.Fatalf("FloatVal(%v).Float() = %v", f, got)
		}
	}
}

func TestFloatValNaN(t *testing.T) {
	v := FloatVal(math.NaN())
	got := v.Float()
	if !math.IsNaN(got) {
		t.Fatalf("FloatVal(NaN).Float() = %v, want NaN", got)
	}
}

func TestFloatValInfinity(t *testing.T) {
	for _, f := range []float64{math.Inf(1), math.Inf(-1)} {
		if got := FloatVal(f).Float(); got != f {
			t.Fatalf("FloatVal(%v).Float() = %v", f, got)
		}
	}
}

func TestFloatValZeroAllocAccessor(t *testing.T) {
	v := FloatVal(3.141692)
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Float() })
	if allocs > 0 {
		t.Fatalf("Float() accessor: got %v allocs, want 0", allocs)
	}
}

func TestStringValTag(t *testing.T) {
	checkTag(t, StringVal("hello"), TString)
}

func TestStringValRoundTrip(t *testing.T) {
	cases := []string{"", "hello", "日本語", "a\x00b", "newline\n"}
	for _, s := range cases {
		if got := StringVal(s).Str().Value; got != s {
			t.Fatalf("StringVal(%q).Str().TValue = %q", s, got)
		}
	}
}

func TestStringValPtrNotNil(t *testing.T) {
	v := StringVal("x")
	if v.ptr == nil {
		t.Fatal("StringVal produced a nil ptr")
	}
}

func TestStringValAllocsOnConstruction(t *testing.T) {
	// Exactly one allocation: the stringData heap object.
	allocs := testing.AllocsPerRun(100_000, func() {
		v := StringVal("hello")
		if v.Str().Value != "hello" {
			t.Fatalf("stringVal: got %v, want %v", v.Str().Value, v)
		}
	})
	if allocs != 1 {
		t.Fatalf("StringVal: got %v allocs, want 1", allocs)
	}
}

func TestStringValAccessorZeroAlloc(t *testing.T) {
	v := StringVal("hello")
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Str() })
	if allocs > 0 {
		t.Fatalf("Str() accessor: got %v allocs, want 0", allocs)
	}
}

func TestArrayValTag(t *testing.T) {
	checkTag(t, ArrayVal(&TypeArray{}), TArray)
}

func TestArrayValRoundTrip(t *testing.T) {
	a := &TypeArray{Value: []TValue{IntVal(1), IntVal(2), IntVal(3)}}
	v := ArrayVal(a)
	if v.Arr() != a {
		t.Fatal("ArrayVal round-trip pointer mismatch")
	}
}

func TestArrayValNilElements(t *testing.T) {
	a := &TypeArray{Value: []TValue{NilVal(), NilVal()}}
	v := ArrayVal(a)
	if len(v.Arr().Value) != 2 {
		t.Fatal("array element count mismatch")
	}
}

func TestArrayValAccessorZeroAlloc(t *testing.T) {
	v := ArrayVal(&TypeArray{})
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Arr() })
	if allocs > 0 {
		t.Fatalf("Arr() accessor: got %v allocs, want 0", allocs)
	}
}

func TestObjectValTag(t *testing.T) {
	checkTag(t, ObjectVal(&TypeObject{}), TObject)
}

func TestObjectValRoundTrip(t *testing.T) {
	o := &TypeObject{Value: map[string]TValue{"key": IntVal(99)}}
	v := ObjectVal(o)
	if v.Obj() != o {
		t.Fatal("ObjectVal round-trip pointer mismatch")
	}
}

func TestObjectValEmptyMap(t *testing.T) {
	o := &TypeObject{Value: make(map[string]TValue)}
	v := ObjectVal(o)
	if len(v.Obj().Value) != 0 {
		t.Fatal("expected empty object")
	}
}

func TestObjectValAccessorZeroAlloc(t *testing.T) {
	v := ObjectVal(&TypeObject{Value: make(map[string]TValue)})
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Obj() })
	if allocs > 0 {
		t.Fatalf("Obj() accessor: got %v allocs, want 0", allocs)
	}
}

func TestFunctionValTag(t *testing.T) {
	fn := &Function{CoreFn: &CoreFunction{}}
	checkTag(t, FunctionVal(fn), TFunction)
}

func TestFunctionValRoundTrip(t *testing.T) {
	fn := &Function{CoreFn: &CoreFunction{Arity: 2}}
	v := FunctionVal(fn)
	if v.Fn() != fn {
		t.Fatal("FunctionVal round-trip pointer mismatch")
	}
}

func TestFunctionValCoreFnPreserved(t *testing.T) {
	core := &CoreFunction{Arity: 3, IsVar: true}
	fn := &Function{CoreFn: core}
	v := FunctionVal(fn)
	if v.Fn().CoreFn != core {
		t.Fatal("CoreFn pointer not preserved through FunctionVal")
	}
}

func TestFunctionValAccessorZeroAlloc(t *testing.T) {
	v := FunctionVal(&Function{CoreFn: &CoreFunction{}})
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Fn() })
	if allocs > 0 {
		t.Fatalf("Fn() accessor: got %v allocs, want 0", allocs)
	}
}

func TestGFnValTag(t *testing.T) {
	fn := func(args ...TValue) (TValue, error) { return NilVal(), nil }
	checkTag(t, GFnVal(fn), TGFn)
}

func TestGFnValRoundTrip(t *testing.T) {
	called := false
	fn := func(args ...TValue) (TValue, error) {
		called = true
		return IntVal(7), nil
	}
	v := GFnVal(fn)
	result, err := v.GFunction()(IntVal(1))
	if err != nil {
		t.Fatalf("GFn call returned error: %v", err)
	}
	if !called {
		t.Fatal("wrapped GFn was not called")
	}
	if result.Int() != 7 {
		t.Fatalf("GFn result: got %v, want 7", result.Int())
	}
}

func TestGFnValPtrNotNil(t *testing.T) {
	v := GFnVal(func(args ...TValue) (TValue, error) { return NilVal(), nil })
	if v.ptr == nil {
		t.Fatal("GFnVal produced a nil ptr")
	}
}

func TestGFnValAllocsOnConstruction(t *testing.T) {
	// Exactly one allocation: the GFnWrapper heap object.
	fn := func(args ...TValue) (TValue, error) { return NilVal(), nil }
	var sink TValue
	allocs := testing.AllocsPerRun(100_000, func() {
		sink = GFnVal(fn)
	})
	_ = sink
	if allocs != 1 {
		t.Fatalf("GFnVal: got %v allocs, want 1", allocs)
	}
}

func TestGFnValAccessorZeroAlloc(t *testing.T) {
	v := GFnVal(func(args ...TValue) (TValue, error) { return NilVal(), nil })
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.GFunction() })
	if allocs > 0 {
		t.Fatalf("GFn() accessor: got %v allocs, want 0", allocs)
	}
}

func TestBytesValTag(t *testing.T) {
	checkTag(t, BytesVal(&Bytes{}), TBytes)
}

func TestBytesValRoundTrip(t *testing.T) {
	b := &Bytes{Value: []byte{1, 2, 3, 4}}
	v := BytesVal(b)
	if v.BBytes() != b {
		t.Fatal("BytesVal round-trip pointer mismatch")
	}
}

func TestBytesValEmptySlice(t *testing.T) {
	b := &Bytes{Value: []byte{}}
	v := BytesVal(b)
	if len(v.BBytes().Value) != 0 {
		t.Fatal("expected empty bytes")
	}
}

func TestBytesValDataIntegrity(t *testing.T) {
	data := []byte("hello world")
	b := &Bytes{Value: data}
	v := BytesVal(b)
	for i, want := range data {
		if got := v.BBytes().Value[i]; got != want {
			t.Fatalf("bytes[%d]: got %v, want %v", i, got, want)
		}
	}
}

func TestBytesValAccessorZeroAlloc(t *testing.T) {
	v := BytesVal(&Bytes{Value: []byte{1, 2}})
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.BBytes() })
	if allocs > 0 {
		t.Fatalf("Byt() accessor: got %v allocs, want 0", allocs)
	}
}

func TestEnumValTag(t *testing.T) {
	e := &Enum{Pairs: map[string]Integer{"A": 0}}
	checkTag(t, EnumVal(e), TEnum)
}

func TestEnumValRoundTrip(t *testing.T) {
	e := &Enum{Pairs: map[string]Integer{"Red": 0, "Green": 1, "Blue": 2}}
	v := EnumVal(e)
	if v.Enm() != e {
		t.Fatal("EnumVal round-trip pointer mismatch")
	}
}

func TestEnumValPairsPreserved(t *testing.T) {
	e := &Enum{Pairs: map[string]Integer{"X": 10, "Y": 20}}
	v := EnumVal(e)
	if v.Enm().Pairs["X"] != 10 || v.Enm().Pairs["Y"] != 20 {
		t.Fatal("enum pairs not preserved through EnumVal")
	}
}

func TestEnumValAccessorZeroAlloc(t *testing.T) {
	v := EnumVal(&Enum{Pairs: map[string]Integer{"A": 0}})
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Enm() })
	if allocs > 0 {
		t.Fatalf("Enm() accessor: got %v allocs, want 0", allocs)
	}
}

func TestErrorValTag(t *testing.T) {
	checkTag(t, ErrorVal(StringVal("oops")), TError)
}

func TestErrorValRoundTripWithString(t *testing.T) {
	msg := StringVal("something went wrong")
	v := ErrorVal(msg)
	got := v.Err().Message
	if got.TType() != TString {
		t.Fatalf("message tag: got %d, want %d", got.TType(), TString)
	}
	if got.Str().Value != "something went wrong" {
		t.Fatalf("message text: got %q", got.Str().Value)
	}
}

func TestErrorValRoundTripWithInt(t *testing.T) {
	// Message can be any TValue, including an integer error code.
	msg := IntVal(404)
	v := ErrorVal(msg)
	if v.Err().Message.Int() != 404 {
		t.Fatal("integer error message not preserved")
	}
}

func TestErrorValRoundTripWithNil(t *testing.T) {
	v := ErrorVal(NilVal())
	if v.Err().Message.TType() != TNil {
		t.Fatal("nil error message tag mismatch")
	}
}

func TestErrorValPtrNotNil(t *testing.T) {
	v := ErrorVal(StringVal("err"))
	if v.ptr == nil {
		t.Fatal("ErrorVal produced a nil ptr")
	}
}

func TestErrorValAllocsOnConstruction(t *testing.T) {
	// Exactly one allocation: the vidaErrorData heap struct.
	// The message is a pre-built TValue so it does not allocate here.
	msg := IntVal(1)
	var sink TValue
	allocs := testing.AllocsPerRun(100_000, func() {
		v := ErrorVal(msg)
		sink = v
	})
	_ = sink
	if allocs != 1 {
		t.Fatalf("ErrorVal: got %v allocs, want 1", allocs)
	}
}

func TestErrorValAccessorZeroAlloc(t *testing.T) {
	v := ErrorVal(IntVal(0))
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Err() })
	if allocs > 0 {
		t.Fatalf("Err() accessor: got %v allocs, want 0", allocs)
	}
}

func TestTimeValTag(t *testing.T) {
	checkTag(t, TimeVal(time.Now()), TTime)
}

func TestTimeValRoundTrip(t *testing.T) {
	now := time.Now()
	v := TimeVal(now)
	got := v.Time()
	if !got.Equal(now) {
		t.Fatalf("TimeVal round-trip: got %v, want %v", got, now)
	}
}

func TestTimeValZeroTime(t *testing.T) {
	zero := time.Time{}
	v := TimeVal(zero)
	if !v.Time().IsZero() {
		t.Fatal("zero time not preserved through TimeVal")
	}
}

func TestTimeValUTC(t *testing.T) {
	utc := time.Now().UTC()
	v := TimeVal(utc)
	got := v.Time()
	if got.Location() != time.UTC {
		t.Fatalf("location: got %v, want UTC", got.Location())
	}
}

func TestTimeValLocationPreserved(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone not available:", err)
	}
	ts := time.Now().In(loc)
	v := TimeVal(ts)
	if v.Time().Location().String() != "America/New_York" {
		t.Fatalf("location: got %v, want America/New_York", v.Time().Location())
	}
}

func TestTimeValMonotonicStripped(t *testing.T) {
	// time.Time TValues from time.Now() carry a monotonic reading.
	// After storing and retrieving through TimeVal the wall time must match
	// even if the monotonic component is affected by the copy.
	now := time.Now()
	v := TimeVal(now)
	// Use Equal which compares wall clock, not the monotonic component.
	if !v.Time().Equal(now) {
		t.Fatal("wall clock not preserved through TimeVal")
	}
}

func TestTimeValAllocsOnConstruction(t *testing.T) {
	// Exactly one allocation: the vidaTime heap wrapper.
	ts := time.Now()
	var sink TValue
	allocs := testing.AllocsPerRun(100_000, func() {
		v := TimeVal(ts)
		sink = v
	})
	_ = sink
	if allocs != 1 {
		t.Fatalf("TimeVal: got %v allocs, want 1", allocs)
	}
}

func TestTimeValAccessorZeroAlloc(t *testing.T) {
	v := TimeVal(time.Now())
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Time() })
	if allocs > 0 {
		t.Fatalf("AsTime() accessor: got %v allocs, want 0", allocs)
	}
}

type stubExt struct {
	ExternalDefaults
	name string
	n    int
}

func (s *stubExt) TypeName() string  { return s.name }
func (s *stubExt) String() string    { return s.name }
func (s *stubExt) Clone() TValue     { return ExtVal(&stubExt{name: s.name, n: s.n}) }
func (s *stubExt) ObjectKey() string { return s.name }

func TestExtValTag(t *testing.T) {
	checkTag(t, ExtVal(&stubExt{name: "stub"}), TExtern)
}

func TestExtValRoundTrip(t *testing.T) {
	s := &stubExt{name: "color", n: 42}
	v := ExtVal(s)
	got, ok := v.Ext().(*stubExt)
	if !ok {
		t.Fatal("Ext() did not return the original *stubExt")
	}
	if got != s {
		t.Fatal("ExtVal round-trip pointer mismatch")
	}
	if got.n != 42 {
		t.Fatalf("ExtVal: n: got %d, want 42", got.n)
	}
}

func TestExtValPtrFieldIsNil(t *testing.T) {
	// For TagExt the ptr field is unused; only ext carries the TValue.
	v := ExtVal(&stubExt{name: "x"})
	if v.ptr != nil {
		t.Fatal("expected ptr == nil for TagExt TValue")
	}
}

func TestExtValExtFieldNotNil(t *testing.T) {
	v := ExtVal(&stubExt{name: "x"})
	if v.extern == nil {
		t.Fatal("expected ext != nil for TagExt TValue")
	}
}

func TestExtValAccessorZeroAlloc(t *testing.T) {
	v := ExtVal(&stubExt{name: "x"})
	allocs := testing.AllocsPerRun(1_000_000, func() { _ = v.Ext() })
	if allocs > 0 {
		t.Fatalf("Ext() accessor: got %v allocs, want 0", allocs)
	}
}

// ── Cross-type invariants ─────────────────────────────────────────────────────

func TestAllTagsDistinct(t *testing.T) {
	// Every tag constant must be unique.
	tags := map[uint8]string{
		TNil:      "TagNil",
		TBool:     "TagBool",
		TInt:      "TagInt",
		TFloat:    "TagFloat",
		TString:   "TagString",
		TArray:    "TagArray",
		TObject:   "TagObject",
		TFunction: "TagFunction",
		TCoreFn:   "TagCoreFn",
		TGFn:      "TagGFn",
		TBytes:    "TagBytes",
		TEnum:     "TagEnum",
		TError:    "TagError",
		TThread:   "TagThread",
		TTime:     "TagTime",
		TExtern:   "TagExt",
	}
	// If any two constants share a TValue the map will be smaller than 16.
	if len(tags) != 16 {
		t.Fatalf("tag collision detected: only %d distinct tags, want 15", len(tags))
	}
}

func TestTValueStructSize(t *testing.T) {
	// 32 bytes: 1 tag + 7 pad + 8 ival + 8 ptr + 16 ext (interface).
	const want = 40
	if got := unsafe.Sizeof(TValue{}); got != want {
		t.Fatalf("TValue size: got %d bytes, want %d", got, want)
	}
}

func TestNilTagIsZero(t *testing.T) {
	// TNil must be 0 so the zero TValue is automatically nil.
	if TNil != 0 {
		t.Fatalf("TagNil = %d, want 0", TNil)
	}
}

func TestDifferentTypesNotEqual(t *testing.T) {
	// TValues of different types must never compare equal,
	// even when their payload bits happen to overlap.
	vals := []TValue{
		NilVal(),
		BoolVal(false), // ival == 0, same as NilVal
		IntVal(0),      // ival == 0
		FloatVal(0),    // ival == 0 (bits of +0.0)
	}
	for i := range vals {
		for j := i + 1; j < len(vals); j++ {
			if vals[i] == vals[j] {
				t.Fatalf("vals[%d] (tag=%d) == vals[%d] (tag=%d)",
					i, vals[i].TType(), j, vals[j].TType())
			}
		}
	}
}

func TestHeapTValuesNeverEqualByContent(t *testing.T) {
	// Two distinct heap objects with the same content must not be ==
	// (pointer identity, not deep equality).
	a1 := ArrayVal(&TypeArray{Value: []TValue{IntVal(1)}})
	a2 := ArrayVal(&TypeArray{Value: []TValue{IntVal(1)}})
	if a1 == a2 {
		t.Fatal("two distinct *Array TValues compared equal via ==")
	}
}
