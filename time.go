package vida

import (
	"time"

	"github.com/alkemist-17/vida/token"
	"github.com/alkemist-17/vida/verror"
)

type Time time.Time

func (t Time) Boolean() Bool {
	return True
}

func (t Time) Prefix(op uint64) (Value, error) {
	switch op {
	case uint64(token.NOT):
		return False, nil
	default:
		return Nil, verror.ErrPrefixOpNotDefined
	}
}

func (t Time) Binop(ctx *Context, op uint64, rhs Value) (Value, error) {
	switch op {
	case uint64(token.AND):
		return rhs, nil
	case uint64(token.OR):
		return t, nil
	case uint64(token.IN):
		return IsMemberOf(t, rhs)
	default:
		return Nil, verror.ErrBinaryOpNotDefined
	}
}

func (t Time) Get(ctx *Context, index Value) Value {
	return Nil
}

func (t Time) Set(index, val Value) error {
	return verror.ErrValueNotIndexable
}

func (t Time) Equals(other Value) Bool {
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
	return Nil, verror.ErrNotImplemented
}

func (t Time) Iterator() Value {
	return Nil
}

func (t Time) String() string {
	return time.Time(t).String()
}

func (t Time) ObjectKey() string {
	return time.Time(t).String()
}

func (t Time) LookUp(ctx *Context, message Value) Value {
	return Nil
}

func (t Time) Type() string {
	return "time"
}

func (t Time) Clone() Value {
	return t
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
	if len(args) > 0 {
		val, ok := args[0].(Integer)
		if ok {
			time.Sleep(time.Duration(val))
		}
	}
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
		if f, ok := args[0].(*String); ok && f.Value == time.Local.String() {
			return Time(time.Now().Local()), nil
		} else if ok && f.Value == time.UTC.String() {
			return Time(time.Now().UTC()), nil
		} else {
			r := time.Now().Format(f.Value)
			if len(r) > 0 {
				return &String{Value: r}, nil
			}
		}
	case 2:
		if f, ok := args[0].(*String); ok {
			if l, ok := args[1].(*String); ok {
				switch l.Value {
				case time.Local.String():
					return &String{Value: time.Now().Local().Format(f.Value)}, nil
				case time.UTC.String():
					return &String{Value: time.Now().UTC().Format(f.Value)}, nil
				}
			}
		}
	}
	return Nil, nil
}

func timeDate(ctx *Context, args ...Value) (Value, error) {
	switch len(args) {
	case 0:
		return Time(time.Now()), nil
	case 8:
		y, ok_0 := args[0].(Integer)
		m, ok_1 := args[1].(Integer)
		d, ok_2 := args[2].(Integer)
		h, ok_3 := args[3].(Integer)
		min, ok_4 := args[4].(Integer)
		sec, ok_5 := args[5].(Integer)
		nsec, ok_6 := args[6].(Integer)
		loc, ok_7 := args[7].(*String)
		if ok_0 && ok_1 && ok_2 && ok_3 && ok_4 && ok_5 && ok_6 && ok_7 {
			if loc.Value == time.Local.String() {
				return Time(time.Date(int(y), time.Month(m), int(d), int(h), int(min), int(sec), int(nsec), time.Local)), nil
			} else if loc.Value == time.UTC.String() {
				return Time(time.Date(int(y), time.Month(m), int(d), int(h), int(min), int(sec), int(nsec), time.UTC)), nil
			}
		}
	}
	return Nil, nil
}

func timeFormat(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if t, ok := args[0].(Time); ok {
			if f, ok := args[1].(*String); ok {
				return &String{Value: time.Time(t).Format(f.Value)}, nil
			}
		}
	}
	return Nil, nil
}

func timeGetYear(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Year()), nil
		}
	}
	return Nil, nil

}

func timeGetMonth(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Month()), nil
		}
	}
	return Nil, nil

}

func timeGetDay(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Day()), nil
		}
	}
	return Nil, nil

}

func timeGetHours(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Hour()), nil
		}
	}
	return Nil, nil

}

func timeGetMinutes(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Minute()), nil
		}
	}
	return Nil, nil

}

func timeGetSeconds(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Second()), nil
		}
	}
	return Nil, nil

}

func timeGetNanoseconds(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).Nanosecond()), nil
		}
	}
	return Nil, nil

}

func timeGetLocation(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return &String{Value: time.Time(t).Location().String()}, nil
		}
	}
	return Nil, nil
}

func timeIn(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if zone, ok := args[0].(*String); ok {
			location, e := time.LoadLocation(zone.Value)
			if e != nil {
				return Nil, nil
			}
			return Time(time.Now().In(location)), nil
		}
	}
	return Time(time.Now().UTC()), nil
}

func timeDateIn(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if t, ok := args[0].(Time); ok {
			if zone, ok := args[1].(*String); ok {
				location, e := time.LoadLocation(zone.Value)
				if e != nil {
					return Nil, nil
				}
				return Time(time.Time(t).In(location)), nil
			}
		}
	}
	return Nil, nil
}

func timeToUnixNano(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return Integer(time.Time(t).UnixNano()), nil
		}
	}
	return Nil, nil
}

func timeParse(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if f, ok := args[0].(*String); ok {
			if dt, ok := args[1].(*String); ok {
				t, err := time.Parse(f.Value, dt.Value)
				if err != nil {
					return &VidaError{Message: &String{Value: err.Error()}}, nil
				}
				return Time(t), nil
			}
		}
	}
	return Nil, nil
}

func timeSince(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 0 {
		if t, ok := args[0].(Time); ok {
			return timeCreateDuration(time.Since(time.Time(t))), nil
		}
	}
	return Nil, nil

}

func timeAddDuration(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if t, ok := args[0].(Time); ok {
			if duration, ok := args[1].(Integer); ok {
				return Time(time.Time(t).Add(time.Duration(duration))), nil
			}
		}
	}
	return Nil, nil
}

func timeSub(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if t, ok := args[0].(Time); ok {
			if u, ok := args[1].(Time); ok {
				return timeCreateDuration(time.Time(t).Sub(time.Time(u))), nil
			}
		}
	}
	return Nil, nil
}

func timeAfter(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if t, ok := args[0].(Time); ok {
			if u, ok := args[1].(Time); ok {
				return Bool(time.Time(t).After(time.Time(u))), nil
			}
		}
	}
	return Nil, nil
}

func timeBefore(ctx *Context, args ...Value) (Value, error) {
	if len(args) > 1 {
		if t, ok := args[0].(Time); ok {
			if u, ok := args[1].(Time); ok {
				return Bool(time.Time(t).Before(time.Time(u))), nil
			}
		}
	}
	return Nil, nil
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
