package vida

import "fmt"

type Method func(self Value, args ...Value) (Value, error)

type VTable struct {
	VType   string
	methods map[string]Method
	parent  *VTable
}

func NewVTable(vtype string) *VTable {
	return &VTable{
		VType:   vtype,
		methods: make(map[string]Method),
	}
}

func (vtable *VTable) Extend(name string) *VTable {
	return &VTable{
		VType:   name,
		methods: make(map[string]Method),
		parent:  vtable,
	}
}

func (vtable *VTable) AddMethod(name string, m Method) {
	vtable.methods[name] = m
}

func (vtable *VTable) Lookup(name string) (Method, bool) {
	for cur := vtable; cur != nil; cur = cur.parent {
		if m, ok := cur.methods[name]; ok {
			return m, true
		}
	}
	return nil, false
}

func (vtable *VTable) Implements(name string) bool {
	_, ok := vtable.Lookup(name)
	return ok
}

type TypedObject struct {
	ReferenceSemanticsImpl
	VTable *VTable
	Fields map[string]Value
}

func NewTypedObject(vtable *VTable) *TypedObject {
	return &TypedObject{
		VTable: vtable,
		Fields: make(map[string]Value),
	}
}

func (o *TypedObject) Get(name string) (Value, bool) {
	v, ok := o.Fields[name]
	return v, ok
}

func (o *TypedObject) Set(name string, v Value) {
	o.Fields[name] = v
}

type NativeObject struct {
	ReferenceSemanticsImpl
	VTable *VTable
	Native any
}

func NewNativeObject(vt *VTable, native any) *NativeObject {
	return &NativeObject{VTable: vt, Native: native}
}

func Dispatch(receiver Value, name string, args ...Value) (Value, error) {
	vt := vtableOf(receiver)
	if vt == nil {
		return nil, fmt.Errorf("%s has no vtable — cannot call :%s", typeName(receiver), name)
	}

	method, ok := vt.Lookup(name)
	if !ok {
		return nil, fmt.Errorf("%s has no method :%s", vt.VType, name)
	}

	return method(receiver, args...)
}
func vtableOf(v Value) *VTable {
	switch o := v.(type) {
	case *TypedObject:
		return o.VTable
	case *NativeObject:
		return o.VTable
	}
	return nil
}

func typeName(v Value) string {
	if vt := vtableOf(v); vt != nil {
		return vt.VType
	}
	return fmt.Sprintf("%T", v)
}

// Experimental VTable String Object
type Text struct {
	Runes  []rune
	Value  string
	VTable *VTable
}
