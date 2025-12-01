/*
 * Copyright 2024-present jishaocong0910
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gdao

import (
	"errors"
	"reflect"
	"time"
)

type Type interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | bool | string | time.Time
}

type Convert[V Type, F any] interface {
	GdaoValue() V
	GdaoField(value V) F
}

var fieldConvertors = map[reflect.Type]fieldConvertor{}

func getFieldConvertor(ft reflect.Type) fieldConvertor {
	if fc, ok := fieldConvertors[ft]; ok {
		return fc
	}
	method, _ := ft.MethodByName("GdaoValue")
	zero := reflect.New(method.Type.Out(0)).Elem().Interface()
	fc := newFieldConvertor(&zero, ft)
	key := ft
	if key.Kind() == reflect.Pointer {
		key = key.Elem()
	}
	fieldConvertors[key] = fc
	return fc
}

func newFieldConvertor(zero *any, ft reflect.Type) fieldConvertor {
	ftZeroType := ft
	if ft.Kind() == reflect.Pointer {
		ftZeroType = ft.Elem()
	}
	return fieldConvertor{
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
