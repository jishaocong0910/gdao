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
	"database/sql"
	"reflect"
)

func Config(c Cfg) {
	cfg = c
}

type Cfg struct {
	// DefaultDB is the default database connection
	DefaultDB *sql.DB
	// Logger is the logger
	Logger Logger
	// SqlLogLevel is the log level of sql, e.g. orm.LogLevel_.DEBUG
	SqlLogLevel LogLevel
	// CompressSqlLog is whether to compress the SQL to one line in the log
	CompressSqlLog bool
	// ColumnMapper is the global default column name mapper, create by orm.NewNameMapper()
	ColumnMapper *NameMapper
}

var (
	cfg             Cfg
	fieldConvertors = map[reflect.Type]fieldConvertor{}
)

type fieldConvertor struct {
	newScanDest func() scanDest
	toValue     func(any) any
	toField     func(value any) any
}

type scanDest struct {
	dest     any
	getValue func() any
}

func getFieldConvertor(ft reflect.Type) fieldConvertor {
	if fc, ok := fieldConvertors[ft]; ok {
		return fc
	}
	method, _ := ft.MethodByName("OrmValue")
	fc := newFieldConvertor(method.Type.Out(0), ft)
	key := ft
	if key.Kind() == reflect.Pointer {
		key = key.Elem()
	}
	fieldConvertors[key] = fc
	return fc
}

func newFieldConvertor(vt reflect.Type, ft reflect.Type) fieldConvertor {
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	vt = reflect.New(vt).Type()
	return fieldConvertor{
		newScanDest: func() scanDest {
			value := reflect.New(vt)
			return scanDest{
				dest: value.Interface(),
				getValue: func() any {
					if !value.Elem().IsNil() {
						return value.Elem().Elem().Interface()
					}
					return nil
				}}
		},
		toValue: func(entity any) any {
			if entity == nil {
				return nil
			}
			ev := reflect.ValueOf(entity)
			return ev.MethodByName("OrmValue").Call(nil)[0].Interface()
		},
		toField: func(value any) any {
			if value == nil {
				return nil
			}
			m := reflect.New(ft).MethodByName("OrmField")
			return m.Call([]reflect.Value{reflect.ValueOf(value)})[0].Interface()
		},
	}
}
