package memusage

import (
	"fmt"
	"reflect"
)

var (
	uintptrSize      = uint64(reflect.TypeOf(uintptr(0)).Size())
	sliceHeaderSize  = uint64(reflect.TypeOf(reflect.SliceHeader{}).Size())
	stringHeaderSize = uint64(reflect.TypeOf(reflect.StringHeader{}).Size())
)

// Visitor visits each object in a call graph
type Visitor interface {
	Visit(v reflect.Value, ctx Context) bool
}

// Context represents the context in which an object was visited
type Context struct {
	Parent reflect.Kind         // Parent is the kind of the object that references the object
	Field  *reflect.StructField // Field is non-nil if this object is an element of a struct
	Index  *int                 // Index is non-nil if this object is an element of an array
	Key    interface{}          // Key is non-nil for values (and keys) of maps
}

// Walk visits the object pointed to by the pointer and then
// each of the object's descendants, recursively. It keeps a
// record of which objects have already been visited and avoids
// visiting them a second time. The first parameter must be a
// pointer; this function will not call Visit on the pointer
// itself but it will call Visit on the object it points to,
// and it will also call Visit on any other pointers it
// encounters during the graph walk. If you want the walk to
// include the pointer then pass in a pointer to the pointer.
func Walk(ptr interface{}, visitor Visitor) {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected a pointer but got %T", ptr))
	}
	w := walker{
		visitor: visitor,
		seen:    make(map[reflect.Value]struct{}),
	}
	ctx := Context{Parent: reflect.Ptr}
	w.walk(v.Elem(), ctx)
}

// walker holds the state for walking an object graph
type walker struct {
	visitor Visitor
	seen    map[reflect.Value]struct{}
}

// walk visits v and then each of v's descendants, recursively
func (w *walker) walk(v reflect.Value, context Context) {
	if _, b := w.seen[v]; b {
		return
	}
	w.seen[v] = struct{}{}

	if !w.visitor.Visit(v, context) {
		return
	}

	ctx := Context{Parent: v.Kind()}

	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			w.walk(v.Elem(), ctx)
		}
	case reflect.Slice, reflect.Array:
		if !v.IsNil() {
			for i := 0; i < v.Len(); i++ {
				ctx.Index = &i
				w.walk(v.Index(i), ctx)
			}
		}
	case reflect.Map:
		if !v.IsNil() {
			for _, k := range v.MapKeys() {
				ctx.Key = k
				w.walk(k, ctx)
				w.walk(v.MapIndex(k), ctx)
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			w.walk(v.Field(i), ctx)
		}
	case reflect.Interface:
		if !v.IsNil() {
			w.walk(reflect.ValueOf(v.Interface()), ctx)
		}
	}
}

// Size computes the size in bytes of an object, not including other
// objects it references. For strings the size includes the size of
// the constituent elements. For arrays and slices, the size includes
// the data buffer that holds the array.
func Size(obj interface{}) uint64 {
	return size(reflect.ValueOf(obj))
}

func size(v reflect.Value) uint64 {
	sz := uint64(v.Type().Size())
	switch v.Kind() {
	case reflect.String:
		sz += uint64(v.Len())
	}
	return sz
}

// Profile contains information about memory consumption in an object graph
type Profile struct {
	sizeByType  map[reflect.Type]uint64
	countByType map[reflect.Type]uint64
	TotalBytes  uint64
}

// NewProfile calculates the memory used by an object and
// all of the objects it references.
func NewProfile(obj interface{}) *Profile {
	p := Profile{
		sizeByType:  make(map[reflect.Type]uint64),
		countByType: make(map[reflect.Type]uint64),
	}
	pr := profiler{&p}
	Walk(obj, &pr)
	return &p
}

type profiler struct {
	p *Profile
}

func (pr *profiler) Visit(v reflect.Value, ctx Context) bool {
	var stop bool
	t := v.Type()
	sz := size(v)
	switch v.Kind() {
	case reflect.Array:
		if isScalar(t.Elem()) {
			stop = true
			sz += scalarSize(t.Elem()) * uint64(v.Len())
		} else {
			sz += uint64(t.Size()) * uint64(v.Len())
		}
	case reflect.Slice:
		if isScalar(t.Elem()) {
			stop = true
			sz += scalarSize(t.Elem()) * uint64(v.Cap())
		} else {
			sz += uint64(t.Size()) * uint64(v.Cap())
		}
	}
	pr.p.sizeByType[t] += sz
	pr.p.countByType[t]++

	if ctx.Parent != reflect.Array && ctx.Parent != reflect.Slice && ctx.Parent != reflect.Struct {
		pr.p.TotalBytes += sz
	}

	return !stop
}

func isScalar(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
		return true
	case reflect.Ptr:
		return isScalar(t.Elem())
	}
	return false
}

func scalarSize(t reflect.Type) uint64 {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
		return uint64(t.Size())
	case reflect.Ptr:
		if elemSize := scalarSize(t.Elem()); elemSize > 0 {
			return uint64(t.Size()) + elemSize
		}
	}
	return 0
}
