package internal

import (
	"reflect"
)

var baseTypes = map[string]struct{}{
	"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {}, "float32": {}, "float64": {}, "bool": {}, "string": {}, "time.Time": {},
}

func IsImplementConvert(ft reflect.Type) int {
	kind := ft.Kind()
	if kind == reflect.Pointer {
		fte := ft.Elem()
		if !IsBaseType(fte) && fte.Kind() != reflect.Struct {
			return 0
		}
	} else if kind != reflect.Slice && kind != reflect.Map {
		return 0
	}
	var vt reflect.Type
	if method, ok := ft.MethodByName("GdaoValue"); ok {
		mt := method.Type
		if mt.NumIn() != 1 || mt.NumOut() != 1 { // coverage-ignore
			return 2
		}
		vt = mt.Out(0)
		if !IsBaseType(vt) { // coverage-ignore
			return 2
		}
	} else {
		return 0
	}
	if method, ok := ft.MethodByName("GdaoField"); ok {
		mt := method.Type
		if mt.NumIn() != 2 || mt.In(1) != vt || mt.NumOut() != 1 || mt.Out(0) != ft {
			return 2
		}
	} else { // coverage-ignore
		return 2
	}
	return 1
}

// IsBaseType 判断fte是否为基本类型，fte 为字段类型的元素类型 ft.Elem()
func IsBaseType(fte reflect.Type) bool {
	if _, ok := baseTypes[fte.Kind().String()]; ok {
		return true
	}
	if _, ok := baseTypes[fte.String()]; ok {
		return true
	}
	return false
}
