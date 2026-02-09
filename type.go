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

package orm

import (
	"reflect"
	"time"
)

type Entity interface {
	Table() string
}

type Convert[T BaseType, F any] interface {
	OrmValue() T
	OrmField(value T) F
}

type BaseType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~bool | ~string | time.Time
}

// 基本类型
var baseTypes = map[string]struct{}{
	"int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {}, "float32": {}, "float64": {}, "bool": {}, "string": {}, "time.Time": {},
}

// 判断是否为基本类型指针
func isBaseTypePointer(ft reflect.Type) bool {
	if ft.Kind() == reflect.Pointer {
		return isBaseType(ft.Elem())
	}
	return false
}

// 判断切片元素是否基本类型，多维数组将递归查找
func isValidSliceType(ft reflect.Type, dive int) bool {
	if ft.Kind() != reflect.Slice {
		if dive == 0 {
			return false
		}
		return isBaseType(ft)
	}
	return isValidSliceType(ft.Elem(), dive+1)
}

// 判断是否为基本类型
func isBaseType(fte reflect.Type) bool {
	// 兼容基础类型定义
	if _, ok := baseTypes[fte.Kind().String()]; ok {
		return true
	}
	// 兼容非基础类型，如：time.Time
	if _, ok := baseTypes[fte.String()]; ok {
		return true
	}
	return false
}

// 判断实体字段是否实现了Convert接口，返回0为未实现，1为实现，2为实现错误
func isImplementConvert(ft reflect.Type) int {
	kind := ft.Kind()
	if kind == reflect.Pointer {
		fte := ft.Elem()
		if !isBaseType(fte) && fte.Kind() != reflect.Struct {
			return 0
		}
	} else if kind != reflect.Slice && kind != reflect.Map {
		return 0
	}
	var vt reflect.Type
	if method, ok := ft.MethodByName("OrmValue"); ok {
		mt := method.Type
		if mt.NumIn() != 1 || mt.NumOut() != 1 { // coverage-ignore
			return 2
		}
		vt = mt.Out(0)
		if !isBaseType(vt) && !isValidSliceType(vt, 0) { // coverage-ignore
			return 2
		}
	} else {
		return 0
	}
	if method, ok := ft.MethodByName("OrmField"); ok {
		mt := method.Type
		if mt.NumIn() != 2 || mt.In(1) != vt || mt.NumOut() != 1 || mt.Out(0) != ft {
			return 2
		}
	} else { // coverage-ignore
		return 2
	}
	return 1
}
