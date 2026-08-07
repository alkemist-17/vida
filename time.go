package vida

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alkemist-17/vida/token"
)

type Time time.Time

func (t Time) Boolean() Bool {
	return True
}

func (t Time) Prefix(ctx *Context, op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, ErrPrefixOpNotDefined
	}
}

func (t Time) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.OR):
		return t, nil
	case uint64(token.IN):
		return isMemberOf(ctx, t, rhs)
	default:
		return Nil, ErrBinaryOpNotDefined
	}
}

func (t Time) Get(ctx *Context, index Value) Value {
	return Nil
}

func (t Time) Set(index, val Value) error {
	return ErrValueNotIndexable
}

func (t Time) Equals(ctx *Context, other Value) Bool {
	if o, ok := other.(Time); ok {
		return Bool(time.Time(t).Equal(time.Time(o)))
	}
	return false
}

func (t Time) IsIterable() Bool {
	return false
}

func (t Time) IsCallable() Bool {
	return false
}

func (t Time) Call(ctx *Context, args ...Value) (Value, error) {
	return Nil, ErrNotImplemented
}

func (t Time) Iterator(ctx *Context) Value {
	return Nil
}

func (t Time) String() string {
	return time.Time(t).String()
}

func (t Time) ObjectKey() string {
	return t.String()
}

func (t Time) GetVTable(ctx *Context) Value {
	if ctx.vtables[timeT] == nil {
		ctx.loadTimeVT()
	}
	return ctx.vtables[timeT]
}

func (t Time) LookUp(ctx *Context, message Value) Value {
	if ctx.vtables[timeT] == nil {
		ctx.loadTimeVT()
	}
	if vtable, ok := ctx.vtables[timeT]; ok {
		return vtable.Get(ctx, message)
	}
	return Nil
}

func (t Time) Type() string {
	return timeT
}

func (t Time) Clone() Value {
	return t
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func loadFoundationTime() Value {
	m := &Object{Value: make(map[string]Value, 51)}
	// Unix Time
	m.Value["unixNano"] = NativeFunction(timestampNano)
	m.Value["unixMilli"] = NativeFunction(timestampMilli)
	m.Value["unixMicro"] = NativeFunction(timestampMicro)
	m.Value["unixSec"] = NativeFunction(timestamp)
	// Time
	m.Value["now"] = NativeFunction(timeNow)
	m.Value["date"] = NativeFunction(timeDate)
	m.Value["format"] = NativeFunction(timeFormat)
	// Extract info from Time
	m.Value["getYear"] = NativeFunction(timeGetYear)
	m.Value["getMonth"] = NativeFunction(timeGetMonth)
	m.Value["getDay"] = NativeFunction(timeGetDay)
	m.Value["getHours"] = NativeFunction(timeGetHours)
	m.Value["getMinutes"] = NativeFunction(timeGetMinutes)
	m.Value["getSeconds"] = NativeFunction(timeGetSeconds)
	m.Value["getNanoseconds"] = NativeFunction(timeGetNanoseconds)
	m.Value["getLocation"] = NativeFunction(timeGetLocation)
	m.Value["toUnixNano"] = NativeFunction(timeToUnixNano)
	// Time Sleep
	m.Value["sleep"] = NativeFunction(timeSleep)
	// Time Units
	m.Value["millisecond"] = Integer(time.Millisecond)
	m.Value["nanosecond"] = Integer(time.Nanosecond)
	m.Value["microsecond"] = Integer(time.Microsecond)
	m.Value["hour"] = Integer(time.Hour)
	m.Value["minute"] = Integer(time.Minute)
	m.Value["second"] = Integer(time.Second)
	// Time Formats
	m.Value["RFC3339"] = &String{Value: time.RFC3339}
	m.Value["RFC3339Nano"] = &String{Value: time.RFC3339Nano}
	m.Value["RFC1123"] = &String{Value: time.RFC1123}
	m.Value["RFC1123Z"] = &String{Value: time.RFC1123Z}
	m.Value["RFC822"] = &String{Value: time.RFC822}
	m.Value["RFC822Z"] = &String{Value: time.RFC822Z}
	m.Value["RFC850"] = &String{Value: time.RFC850}
	m.Value["Unix"] = &String{Value: time.UnixDate}
	m.Value["ANSIC"] = &String{Value: time.ANSIC}
	m.Value["RubyDate"] = &String{Value: time.RubyDate}
	m.Value["Kitchen"] = &String{Value: time.Kitchen}
	// Time Stamps
	m.Value["Stamp"] = &String{Value: time.Stamp}
	m.Value["StampMicro"] = &String{Value: time.StampMicro}
	m.Value["StampMilli"] = &String{Value: time.StampMilli}
	m.Value["StampNano"] = &String{Value: time.StampNano}
	m.Value["DateTime"] = &String{Value: time.DateTime}
	m.Value["DateOnly"] = &String{Value: time.DateOnly}
	m.Value["TimeOnly"] = &String{Value: time.TimeOnly}
	// Time ops with TimeZones
	m.Value["nowIn"] = NativeFunction(timeIn)
	m.Value["dateIn"] = NativeFunction(timeDateIn)
	// Basic TimeZones
	m.Value["Local"] = &String{Value: time.Local.String()}
	m.Value["UTC"] = &String{Value: time.UTC.String()}
	// Parse Time
	m.Value["parse"] = NativeFunction(timeParse)
	// Time operations
	m.Value["since"] = NativeFunction(timeSince)
	m.Value["add"] = NativeFunction(timeAddDuration)
	m.Value["sub"] = NativeFunction(timeSub)
	m.Value["after"] = NativeFunction(timeAfter)
	m.Value["before"] = NativeFunction(timeBefore)
	return m
}

func timeSleep(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.sleep", "expected an integer duration (nanoseconds) argument")
	}
	val, ok := args[0].(Integer)
	if !ok {
		return argError("time.sleep", fmt.Sprintf("expected an integer duration argument, got %s", args[0].Type()))
	}
	time.Sleep(time.Duration(val))
	return Nil, nil
}

func timestampNano(ctx *Context, args ...Value) (Value, error) {
	return Integer(time.Now().UnixNano()), nil
}

func timestampMilli(ctx *Context, args ...Value) (Value, error) {
	return Integer(time.Now().UnixMilli()), nil
}

func timestampMicro(ctx *Context, args ...Value) (Value, error) {
	return Integer(time.Now().UnixMicro()), nil
}

func timestamp(ctx *Context, args ...Value) (Value, error) {
	return Integer(time.Now().Unix()), nil
}

func timeNow(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 0:
		return Time(time.Now()), nil
	case 1:
		f, ok := args[0].(*String)
		if !ok {
			return argError("time.now", fmt.Sprintf("expected a string argument, got %s", args[0].Type()))
		}
		if f.Value == time.Local.String() {
			return Time(time.Now().Local()), nil
		} else if f.Value == time.UTC.String() {
			return Time(time.Now().UTC()), nil
		}
		return &String{Value: time.Now().Format(f.Value)}, nil
	case 2:
		f, ok := args[0].(*String)
		if !ok {
			return argError("time.now", fmt.Sprintf("argument 1 (format) must be a string, got %s", args[0].Type()))
		}
		l, ok := args[1].(*String)
		if !ok {
			return argError("time.now", fmt.Sprintf("argument 2 (location) must be a string, got %s", args[1].Type()))
		}
		switch l.Value {
		case time.Local.String():
			return &String{Value: time.Now().Local().Format(f.Value)}, nil
		case time.UTC.String():
			return &String{Value: time.Now().UTC().Format(f.Value)}, nil
		default:
			return argError("time.now", fmt.Sprintf("unsupported location %q: expected %q or %q", l.Value, time.Local.String(), time.UTC.String()))
		}
	default:
		return argError("time.now", fmt.Sprintf("expected 0, 1, or 2 arguments, got %d", len(args)))
	}
}

func timeDate(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 0:
		return Time(time.Now()), nil
	case 8:
		y, ok_0 := args[0].(Integer)
		if !ok_0 {
			return argError("time.date", fmt.Sprintf("argument 1 (year) must be an integer, got %s", args[0].Type()))
		}
		m, ok_1 := args[1].(Integer)
		if !ok_1 {
			return argError("time.date", fmt.Sprintf("argument 2 (month) must be an integer, got %s", args[1].Type()))
		}
		d, ok_2 := args[2].(Integer)
		if !ok_2 {
			return argError("time.date", fmt.Sprintf("argument 3 (day) must be an integer, got %s", args[2].Type()))
		}
		h, ok_3 := args[3].(Integer)
		if !ok_3 {
			return argError("time.date", fmt.Sprintf("argument 4 (hour) must be an integer, got %s", args[3].Type()))
		}
		min, ok_4 := args[4].(Integer)
		if !ok_4 {
			return argError("time.date", fmt.Sprintf("argument 5 (minute) must be an integer, got %s", args[4].Type()))
		}
		sec, ok_5 := args[5].(Integer)
		if !ok_5 {
			return argError("time.date", fmt.Sprintf("argument 6 (second) must be an integer, got %s", args[5].Type()))
		}
		nsec, ok_6 := args[6].(Integer)
		if !ok_6 {
			return argError("time.date", fmt.Sprintf("argument 7 (nanosecond) must be an integer, got %s", args[6].Type()))
		}
		loc, ok_7 := args[7].(*String)
		if !ok_7 {
			return argError("time.date", fmt.Sprintf("argument 8 (location) must be a string, got %s", args[7].Type()))
		}
		if loc.Value == time.Local.String() {
			return Time(time.Date(int(y), time.Month(m), int(d), int(h), int(min), int(sec), int(nsec), time.Local)), nil
		} else if loc.Value == time.UTC.String() {
			return Time(time.Date(int(y), time.Month(m), int(d), int(h), int(min), int(sec), int(nsec), time.UTC)), nil
		}
		return argError("time.date", fmt.Sprintf("unsupported location %q: expected %q or %q", loc.Value, time.Local.String(), time.UTC.String()))
	default:
		return argError("time.date", fmt.Sprintf("expected 0 or 8 arguments (year, month, day, hour, minute, second, nanosecond, location), got %d", len(args)))
	}
}

func timeFormat(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.format", fmt.Sprintf("expected 2 arguments (time, layout), got %d", len(args)))
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.format", fmt.Sprintf("argument 1 must be a time value, got %s", args[0].Type()))
	}
	f, ok := args[1].(*String)
	if !ok {
		return argError("time.format", fmt.Sprintf("argument 2 (layout) must be a string, got %s", args[1].Type()))
	}
	return &String{Value: time.Time(t).Format(f.Value)}, nil
}

func timeGetYear(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getYear", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getYear", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Year()), nil
}

func timeGetMonth(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getMonth", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getMonth", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Month()), nil
}

func timeGetDay(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getDay", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getDay", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Day()), nil
}

func timeGetHours(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getHours", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getHours", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Hour()), nil
}

func timeGetMinutes(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getMinutes", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getMinutes", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Minute()), nil
}

func timeGetSeconds(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getSeconds", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getSeconds", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Second()), nil
}

func timeGetNanoseconds(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getNanoseconds", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getNanoseconds", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).Nanosecond()), nil
}

func timeGetLocation(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.getLocation", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.getLocation", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return &String{Value: time.Time(t).Location().String()}, nil
}

func timeIn(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return Time(time.Now().UTC()), nil
	}
	zone, ok := args[0].(*String)
	if !ok {
		return argError("time.nowIn", fmt.Sprintf("expected a string timezone argument, got %s", args[0].Type()))
	}
	location, e := time.LoadLocation(zone.Value)
	if e != nil {
		return argError("time.nowIn", e.Error())
	}
	return Time(time.Now().In(location)), nil
}

func timeDateIn(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.dateIn", fmt.Sprintf("expected 2 arguments (time, timezone), got %d", len(args)))
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.dateIn", fmt.Sprintf("argument 1 must be a time value, got %s", args[0].Type()))
	}
	zone, ok := args[1].(*String)
	if !ok {
		return argError("time.dateIn", fmt.Sprintf("argument 2 (timezone) must be a string, got %s", args[1].Type()))
	}
	location, e := time.LoadLocation(zone.Value)
	if e != nil {
		return argError("time.dateIn", e.Error())
	}
	return Time(time.Time(t).In(location)), nil
}

func timeToUnixNano(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.toUnixNano", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.toUnixNano", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return Integer(time.Time(t).UnixNano()), nil
}

func timeParse(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.parse", fmt.Sprintf("expected 2 arguments (layout, value), got %d", len(args)))
	}
	f, ok := args[0].(*String)
	if !ok {
		return argError("time.parse", fmt.Sprintf("argument 1 (layout) must be a string, got %s", args[0].Type()))
	}
	dt, ok := args[1].(*String)
	if !ok {
		return argError("time.parse", fmt.Sprintf("argument 2 (value) must be a string, got %s", args[1].Type()))
	}
	t, err := time.Parse(f.Value, dt.Value)
	if err != nil {
		return &VidaError{Message: &String{Value: err.Error()}}, nil
	}
	return Time(t), nil
}

func timeSince(ctx *Context, args ...Value) (Value, error) {
	if len(args) == 0 {
		return argError("time.since", "expected a time argument")
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.since", fmt.Sprintf("expected a time argument, got %s", args[0].Type()))
	}
	return timeCreateDuration(time.Since(time.Time(t))), nil
}

func timeAddDuration(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.add", fmt.Sprintf("expected 2 arguments (time, durationNanos), got %d", len(args)))
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.add", fmt.Sprintf("argument 1 must be a time value, got %s", args[0].Type()))
	}
	duration, ok := args[1].(Integer)
	if !ok {
		return argError("time.add", fmt.Sprintf("argument 2 (duration) must be an integer, got %s", args[1].Type()))
	}
	return Time(time.Time(t).Add(time.Duration(duration))), nil
}

func timeSub(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.sub", fmt.Sprintf("expected 2 time arguments, got %d", len(args)))
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.sub", fmt.Sprintf("argument 1 must be a time value, got %s", args[0].Type()))
	}
	u, ok := args[1].(Time)
	if !ok {
		return argError("time.sub", fmt.Sprintf("argument 2 must be a time value, got %s", args[1].Type()))
	}
	return timeCreateDuration(time.Time(t).Sub(time.Time(u))), nil
}

func timeAfter(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.after", fmt.Sprintf("expected 2 time arguments, got %d", len(args)))
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.after", fmt.Sprintf("argument 1 must be a time value, got %s", args[0].Type()))
	}
	u, ok := args[1].(Time)
	if !ok {
		return argError("time.after", fmt.Sprintf("argument 2 must be a time value, got %s", args[1].Type()))
	}
	return Bool(time.Time(t).After(time.Time(u))), nil
}

func timeBefore(ctx *Context, args ...Value) (Value, error) {
	if len(args) < 2 {
		return argError("time.before", fmt.Sprintf("expected 2 time arguments, got %d", len(args)))
	}
	t, ok := args[0].(Time)
	if !ok {
		return argError("time.before", fmt.Sprintf("argument 1 must be a time value, got %s", args[0].Type()))
	}
	u, ok := args[1].(Time)
	if !ok {
		return argError("time.before", fmt.Sprintf("argument 2 must be a time value, got %s", args[1].Type()))
	}
	return Bool(time.Time(t).Before(time.Time(u))), nil
}

func timeCreateDuration(v time.Duration) *Object {
	o := &Object{Value: make(map[string]Value, 7)}
	o.Value["hours"] = Float(v.Hours())
	o.Value["minutes"] = Float(v.Minutes())
	o.Value["seconds"] = Float(v.Seconds())
	o.Value["microseconds"] = Integer(v.Microseconds())
	o.Value["milliseconds"] = Integer(v.Milliseconds())
	o.Value["nanoseconds"] = Integer(v.Nanoseconds())
	o.Value["description"] = &String{Value: v.String()}
	return o
}
