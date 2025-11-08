package gdao

import (
	"errors"
	"reflect"
	"strconv"
	"time"
)

type Type interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | bool | string | time.Time
}

type Convert[V Type, F any] interface {
	GdaoValue() V
	GdaoField(value V) F
}

func checkImplementConvert(ft reflect.Type) (*fieldConvertor, bool) {
	switch ft.Kind() {
	case reflect.Pointer:
		fte := ft.Elem()
		if !checkSupportedFieldType(fte) && fte.Kind() != reflect.Struct {
			return nil, true
		}
		if fc, ok := fieldConvertors[fte]; ok {
			return &fc, true
		}
	case reflect.Slice, reflect.Map:
		if fc, ok := fieldConvertors[ft]; ok {
			return &fc, true
		}
	default:
		return nil, true
	}
	var vt reflect.Type
	if method, ok := ft.MethodByName("GdaoValue"); ok {
		mt := method.Type
		if mt.NumIn() != 1 || mt.NumOut() != 1 {
			return nil, false
		}
		vt = mt.Out(0)
		if !checkSupportedFieldType(vt) {
			return nil, false
		}
	} else {
		return nil, true
	}
	if method, ok := ft.MethodByName("GdaoField"); ok {
		mt := method.Type
		if mt.NumIn() != 2 || mt.In(1) != vt || mt.NumOut() != 1 || mt.Out(0) != ft {
			return nil, false
		}
	} else {
		return nil, true
	}

	zero := reflect.New(vt).Elem().Interface()
	fc := newFieldConvertor(&zero, ft)
	key := ft
	if key.Kind() == reflect.Pointer {
		key = key.Elem()
	}
	fieldConvertors[key] = *fc
	return fc, true
}

func newFieldConvertor(zero *any, ft reflect.Type) *fieldConvertor {
	ftZeroType := ft
	if ft.Kind() == reflect.Pointer {
		ftZeroType = ft.Elem()
	}
	return &fieldConvertor{
		toValue: func(entity any) any {
			if entity == nil {
				return nil
			}
			ev := reflect.ValueOf(entity)
			return ev.MethodByName("GdaoValue").Call(nil)[0].Interface()
		},
		newScanDest: func() scanDest {
			dest := &zero
			return scanDest{
				dest: &dest,
				getValue: func() any {
					if dest != nil {
						return **dest
					}
					return nil
				}}
		},
		toField: func(value any) any {
			if value == nil {
				return nil
			}
			m := reflect.New(ftZeroType).MethodByName("GdaoField")
			return m.Call([]reflect.Value{reflect.ValueOf(value)})[0].Interface()
		},
	}
}

func checkEntityType[T any]() error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("generics must be struct type")
	}
	return nil
}

// fte 为字段类型的元素类型 ft.Elem()
func checkSupportedFieldType(fte reflect.Type) bool {
	if _, ok := supportedFieldTypes[fte.Kind().String()]; ok {
		return true
	}
	if _, ok := supportedFieldTypes[fte.String()]; ok {
		return true
	}
	return false
}

func convertArgs(args []any) []any {
	for i, a := range args {
		if a == nil {
			continue
		}
		t := reflect.TypeOf(a)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if convert, ok := fieldConvertors[t]; ok {
			args[i] = convert.toValue(a)
		}
	}
	return args
}

type scanDest struct {
	dest     any
	getValue func() any
}

type fieldConvertor struct {
	toValue     func(any) any
	newScanDest func() scanDest
	toField     func(value any) any
}

var (
	supportedFieldTypes = map[string]struct{}{
		"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {}, "float32": {}, "float64": {}, "bool": {}, "string": {}, "time.Time": {},
	}

	lastInsertIdConvertors = map[string]func(id int64) reflect.Value{
		"int":     func(id int64) reflect.Value { i := int(id); return reflect.ValueOf(&i) },
		"int8":    func(id int64) reflect.Value { i := int8(id); return reflect.ValueOf(&i) },
		"int16":   func(id int64) reflect.Value { i := int16(id); return reflect.ValueOf(&i) },
		"int32":   func(id int64) reflect.Value { i := int32(id); return reflect.ValueOf(&i) },
		"int64":   func(id int64) reflect.Value { return reflect.ValueOf(&id) },
		"uint":    func(id int64) reflect.Value { u := uint(id); return reflect.ValueOf(&u) },
		"uint8":   func(id int64) reflect.Value { u := uint8(id); return reflect.ValueOf(&u) },
		"uint16":  func(id int64) reflect.Value { u := uint16(id); return reflect.ValueOf(&u) },
		"uint32":  func(id int64) reflect.Value { u := uint32(id); return reflect.ValueOf(&u) },
		"uint64":  func(id int64) reflect.Value { u := uint64(id); return reflect.ValueOf(&u) },
		"float32": func(id int64) reflect.Value { f := float32(id); return reflect.ValueOf(&f) },
		"float64": func(id int64) reflect.Value { f := float64(id); return reflect.ValueOf(&f) },
		"string":  func(id int64) reflect.Value { s := strconv.FormatInt(id, 10); return reflect.ValueOf(&s) },
	}

	fieldConvertors = map[reflect.Type]fieldConvertor{}
)
