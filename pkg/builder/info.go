package builder

import "reflect"

func typeName(source reflect.Type) string {
	if source.Kind() == reflect.Ptr {
		return source.Elem().Name()
	}
	return source.Name()
}

func isSequence(source reflect.Value) bool {
	return source.Kind() == reflect.Array || source.Kind() == reflect.Slice
}

func isPtr(source reflect.Value) bool {
	return source.Kind() == reflect.Ptr
}
