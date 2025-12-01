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
	e "github.com/jishaocong0910/enum"
	"reflect"
	"strconv"
)

type LogLevel struct {
	*e.EnumElem__
}

type _LogLevel struct {
	*e.Enum__[LogLevel]
	OFF,
	DEBUG,
	INFO LogLevel
}

var LogLevel_ = e.NewEnum[LogLevel](_LogLevel{})

type LastInsertIdAs struct {
	*e.EnumElem__
}

type _LastInsertIdAs struct {
	*e.Enum__[LastInsertIdAs]
	FIRST_ID,
	LAST_ID LastInsertIdAs
}

var LastInsertIdAs_ = e.NewEnum[LastInsertIdAs](_LastInsertIdAs{})

type RowAs struct {
	*e.EnumElem__
}

type _RowAs struct {
	*e.Enum__[RowAs]
	RETURNING,
	LAST_ID RowAs
}

var RowAs_ = e.NewEnum[RowAs](_RowAs{})

type lastInsertIdConvertor struct {
	*e.EnumElem__
	convert func(id int64) reflect.Value
}

type _lastInsertIdConvertor struct {
	*e.Enum__[lastInsertIdConvertor]
	int, int8, int16, int32, int64,
	uint, uint8, uint16, uint32, uint64,
	float32, float64, string lastInsertIdConvertor
}

var lastInsertIdConvertor_ = e.NewEnum[lastInsertIdConvertor](_lastInsertIdConvertor{
	int:     lastInsertIdConvertor{convert: func(id int64) reflect.Value { i := int(id); return reflect.ValueOf(&i) }},
	int8:    lastInsertIdConvertor{convert: func(id int64) reflect.Value { i := int8(id); return reflect.ValueOf(&i) }},
	int16:   lastInsertIdConvertor{convert: func(id int64) reflect.Value { i := int16(id); return reflect.ValueOf(&i) }},
	int32:   lastInsertIdConvertor{convert: func(id int64) reflect.Value { i := int32(id); return reflect.ValueOf(&i) }},
	int64:   lastInsertIdConvertor{convert: func(id int64) reflect.Value { return reflect.ValueOf(&id) }},
	uint:    lastInsertIdConvertor{convert: func(id int64) reflect.Value { u := uint(id); return reflect.ValueOf(&u) }},
	uint8:   lastInsertIdConvertor{convert: func(id int64) reflect.Value { u := uint8(id); return reflect.ValueOf(&u) }},
	uint16:  lastInsertIdConvertor{convert: func(id int64) reflect.Value { u := uint16(id); return reflect.ValueOf(&u) }},
	uint32:  lastInsertIdConvertor{convert: func(id int64) reflect.Value { u := uint32(id); return reflect.ValueOf(&u) }},
	uint64:  lastInsertIdConvertor{convert: func(id int64) reflect.Value { u := uint64(id); return reflect.ValueOf(&u) }},
	float32: lastInsertIdConvertor{convert: func(id int64) reflect.Value { f := float32(id); return reflect.ValueOf(&f) }},
	float64: lastInsertIdConvertor{convert: func(id int64) reflect.Value { f := float64(id); return reflect.ValueOf(&f) }},
	string:  lastInsertIdConvertor{convert: func(id int64) reflect.Value { s := strconv.FormatInt(id, 10); return reflect.ValueOf(&s) }},
})
