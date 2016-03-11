package memusage

import "reflect"

var (
	uintptrSize      = uint64(reflect.TypeOf(uintptr(0)).Size())
	sliceHeaderSize  = uint64(reflect.TypeOf(reflect.SliceHeader{}).Size())
	stringHeaderSize = uint64(reflect.TypeOf(reflect.StringHeader{}).Size())
)

// Bytes calculates the memory used by the object and all of the objects it references, recursively.
func Bytes(obj interface{}) uint64 {
	return size(reflect.ValueOf(obj))
}

func size(v reflect.Value) uint64 {
	sz := uint64(v.Type().Size())
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			sz += size(v.Elem())
		}
	case reflect.String:
		sz += uint64(v.Len())
	case reflect.Array, reflect.Slice:
		if !v.IsNil() {
			for i := 0; i < v.Len(); i++ {
				sz += size(v.Index(i))
			}
		}
	case reflect.Map:
		if !v.IsNil() {
			for _, k := range v.MapKeys() {
				sz += size(k)
				sz += size(v.MapIndex(k))
			}
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			sz += size(v.Field(i))
		}
	case reflect.Interface:
		if !v.IsNil() {
			sz += size(reflect.ValueOf(v.Interface()))
		}
	}
	return sz
}
