/*
Copyright 2024-present jishaocong0910

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gdao

import (
	"errors"
	"github.com/jishaocong0910/gdao/internal"
	"reflect"
)

type Dao[T any] struct {
	*baseDao__
	commaColumns           string
	columns                []string
	columnToFieldIndex     map[string]int
	columnToFieldConvertor map[string]fieldConvertor
	autoIncrementColumns   []string
	autoIncrementStep      int64
	autoIncrementConvert   func(id int64) reflect.Value
}

func (this *Dao[T]) Query(req QueryReq[T]) (first *T, list []*T, err error) {
	list = make([]*T, 0)
	b := newDaoSqlBuilder(this, req.Entities)
	req.BuildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		checkMust(req.Must, err)
		return nil, list, err
	}
	if !b.Ok() { // coverage-ignore
		return
	}
	rows, columns, closeFunc, err := this.query(req.Ctx, b.Sql(), b.Args())
	if err != nil { // coverage-ignore
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), -1, -1, err)
		checkMust(req.Must, err)
		return nil, nil, err
	}
	defer closeFunc()

	switch req.RowAs.String() {
	case RowAs_.RETURNING.String():
		var affected int64
		for i := 0; rows.Next() && i < len(req.Entities); i++ {
			entity := req.Entities[i]
			if entity == nil { // coverage-ignore
				continue
			}
			v := reflect.ValueOf(entity).Elem()
			var fields []any
			for _, c := range columns {
				if fieldIndex, ok := this.columnToFieldIndex[c]; ok {
					field := v.Field(fieldIndex).Addr().Interface()
					fields = append(fields, field)
				}
			}
			if len(fields) > 0 {
				printWarn(req.Ctx, rows.Scan(fields...))
			}
			affected++
		}
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), affected, -1, nil)
	case RowAs_.LAST_ID.String():
		var affected int64
		var id *int64
		if rows.Next() && len(columns) == 1 && len(this.autoIncrementColumns) == 1 {
			err = rows.Scan(&id)
			printWarn(req.Ctx, err)
			if err != nil && rows.Next() { // coverage-ignore
				id = nil
			}
		}
		if id != nil {
			fieldIndex := this.columnToFieldIndex[this.autoIncrementColumns[0]]
			entityLength := len(req.Entities)
			for i := 0; i < entityLength; i++ {
				entity := req.Entities[i]
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(this.autoIncrementConvert(*id - int64(entityLength-1-i)*this.autoIncrementStep))
				affected++
			}
		} else {
			for i := 0; i < len(req.Entities); i++ {
				entity := req.Entities[i]
				if entity != nil { // coverage-ignore
					affected++
				}
			}
		}
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), affected, -1, nil)
	default:
		var rowCounts int64
		for rows.Next() {
			entity := new(T)
			dests, afterScans := this.mappingScanFields(entity, columns)
			err = rows.Scan(dests...)
			if err != nil {
				checkMust(req.Must, err)
				return
			}
			for _, after := range afterScans {
				after()
			}
			list = append(list, entity)
			rowCounts++
		}
		if len(list) > 0 {
			first = list[0]
		}
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), -1, rowCounts, nil)
	}
	return
}

func (this *Dao[T]) Exec(req ExecReq[T]) (affected int64, err error) {
	b := newDaoSqlBuilder(this, req.Entities)
	req.BuildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		checkMust(req.Must, err)
		return 0, err
	}
	if !b.Ok() { // coverage-ignore
		return 0, nil
	}
	result, affected, err := this.exec(req.Ctx, b.Sql(), b.Args())
	printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), affected, -1, err)
	if err != nil { // coverage-ignore
		checkMust(req.Must, err)
		return
	}

	switch req.LastInsertIdAs.String() {
	case LastInsertIdAs_.FIRST_ID.String():
		id, err := result.LastInsertId()
		printWarn(req.Ctx, err)
		if err == nil && len(req.Entities) > 0 && len(this.autoIncrementColumns) == 1 {
			fieldIndex := this.columnToFieldIndex[this.autoIncrementColumns[0]]
			for i, entity := range req.Entities {
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(this.autoIncrementConvert(id + int64(i)*this.autoIncrementStep))
			}
		}
	case LastInsertIdAs_.LAST_ID.String():
		id, err := result.LastInsertId()
		printWarn(req.Ctx, err)
		if err == nil && len(req.Entities) > 0 && len(this.autoIncrementColumns) == 1 {
			fieldIndex := this.columnToFieldIndex[this.autoIncrementColumns[0]]
			entityLength := len(req.Entities)
			for i := 0; i < entityLength; i++ {
				entity := req.Entities[i]
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(this.autoIncrementConvert(id - int64(entityLength-1-i)*this.autoIncrementStep))
			}
		}
	}
	return
}

func (this *Dao[T]) mappingScanFields(entity *T, columns []string) ([]any, []func()) {
	v := reflect.ValueOf(entity).Elem()
	dests := make([]any, 0, len(columns))
	afterScans := make([]func(), 0, len(columns))
	for _, c := range columns {
		if index, ok := this.columnToFieldIndex[c]; ok {
			field := v.Field(index)
			if fc, ok := this.columnToFieldConvertor[c]; ok {
				sc := fc.newScanDest()
				dests = append(dests, sc.dest)
				afterScans = append(afterScans, func() {
					v := sc.getValue()
					f := fc.toField(v)
					if f != nil {
						field.Set(reflect.ValueOf(f))
					}
				})
			} else {
				dests = append(dests, field.Addr().Interface())
			}
		} else {
			dests = append(dests, new(any))
		}
	}
	return dests, afterScans
}

func (this *Dao[T]) registerEntity(req NewDaoReq) error {
	err := checkEntityType[T]()
	if err != nil {
		return err
	}
	t := reflect.TypeOf((*T)(nil)).Elem()
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			if req.AllowInvalidField {
				continue
			}
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" must be exported")
		}
		if !tf.Anonymous {
			ft := tf.Type
			switch internal.IsImplementConvert(ft) {
			case 1:
				fc := getFieldConvertor(ft)
				this.registerField(tf, req.ColumnMapper, &fc)
				continue
			case 2:
				if !req.AllowInvalidField {
					return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" is invalid implementing gdao.Convert")
				}
			}
			if ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice {
				if internal.IsBaseType(ft.Elem()) {
					this.registerField(tf, req.ColumnMapper, nil)
					continue
				}
			}
			if req.AllowInvalidField {
				continue
			}
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" is not supported type")
		}
	}
	return nil
}

func (this *Dao[T]) registerField(tf reflect.StructField, columnMapper *NameMapper, fieldConvertor *fieldConvertor) {
	var column string

	t := parseTag(tf)
	if t.column == "" {
		if columnMapper != nil {
			column = columnMapper.Convert(tf.Name)
		} else { // coverage-ignore
			return
		}
	} else {
		column = t.column
	}

	this.columns = append(this.columns, column)
	if this.commaColumns != "" {
		this.commaColumns += ", "
	}
	this.commaColumns += column
	this.columnToFieldIndex[column] = tf.Index[0]
	if t.isAutoIncrement {
		if convertor := lastInsertIdConvertor_.OfString(tf.Type.Elem().String()); !convertor.IsUndefined() {
			this.autoIncrementColumns = append(this.autoIncrementColumns, column)
			this.autoIncrementStep = t.autoIncrementStep
			this.autoIncrementConvert = convertor.convert
		}
	}
	if fieldConvertor != nil {
		this.columnToFieldConvertor[column] = *fieldConvertor
	}
	return
}

func NewDao[T any](req NewDaoReq) *Dao[T] {
	dao := &Dao[T]{
		columnToFieldIndex:     make(map[string]int),
		columnToFieldConvertor: make(map[string]fieldConvertor),
	}
	dao.baseDao__ = extendBaseDao(dao, req.DB)
	err := dao.registerEntity(req)
	must(err)
	return dao
}
