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
	"reflect"
	"strings"
)

type DaoSqlBuilder[T any] struct {
	*sqlBuilder__
	dao      *Dao[T]
	entities []*T
}

func (this *DaoSqlBuilder[T]) Write(str string, args ...any) *DaoSqlBuilder[T] {
	this.sqlBuilder__.Write(str, args...)
	return this
}

func (this *DaoSqlBuilder[T]) WriteColumns(columns ...string) *DaoSqlBuilder[T] {
	if len(columns) == 0 {
		this.Write(this.dao.commaColumns)
	} else {
		for i, c := range columns {
			if c == "" {
				continue
			}
			if i != 0 {
				this.Write(", ")
			}
			this.Write(c)
		}
	}
	return this
}

func (this *DaoSqlBuilder[T]) Columns(onlyAssigned bool, ignoredColumns ...string) (columns []string) {
	if !onlyAssigned {
		if len(ignoredColumns) == 0 {
			return this.dao.columns
		}
		ignoredColumnMap := this.toMap(ignoredColumns)
		for _, column := range this.dao.columns {
			if _, ok := ignoredColumnMap[column]; !ok {
				columns = append(columns, column)
			}
		}
		return
	} else {
		entity := this.Entity()
		if entity != nil {
			v := reflect.ValueOf(entity).Elem()
			ignoredColumnMap := this.toMap(ignoredColumns)
			for _, column := range this.dao.columns {
				fieldIndex := this.dao.columnToFieldIndexMap[column]
				field := v.Field(fieldIndex)
				if field.IsNil() {
					continue
				}
				if _, ok := ignoredColumnMap[column]; ok {
					continue
				}
				columns = append(columns, column)
			}
			return
		}
		return
	}
}

func (this *DaoSqlBuilder[T]) AutoColumns() []string {
	return this.dao.autoIncrementColumns
}

func (this *DaoSqlBuilder[T]) EntityAt(index int) *T {
	var t *T
	if index < len(this.entities) {
		t = this.entities[index]
	}
	return t
}

func (this *DaoSqlBuilder[T]) Entity() *T {
	return this.EntityAt(0)
}

func (this *DaoSqlBuilder[T]) ColumnValue(entity *T, column string) any {
	if entity == nil {
		return nil
	}
	fieldIndex, ok := this.dao.columnToFieldIndexMap[column]
	if !ok {
		return nil
	}
	v := reflect.ValueOf(entity).Elem()
	vf := v.Field(fieldIndex)
	if vf.IsNil() {
		return nil
	}
	return vf.Interface()
}

func (this *DaoSqlBuilder[T]) EachEntity(sep *separate, handle func(n int, entity *T)) *DaoSqlBuilder[T] {
	var n int
	this.writePrefix(sep, n)
	for _, entity := range this.entities {
		n++
		this.writePrefix(sep, n)
		this.writeSep(sep, n)
		handle(n, entity)
	}
	this.writeSuffix(sep, n)
	return this
}

func (this *DaoSqlBuilder[T]) EachColumn(entity *T, sep *separate, handle func(n int, column string, value any), columns ...string) {
	v := reflect.ValueOf(entity).Elem()
	var n int
	this.writePrefix(sep, n)
	for _, column := range columns {
		fieldIndex := this.dao.columnToFieldIndexMap[column]
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}
		n++
		this.writePrefix(sep, n)
		this.writeSep(sep, n)

		handle(n, column, value)
	}
	this.writeSuffix(sep, n)
	return
}

func (this *DaoSqlBuilder[T]) toMap(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	if len(s) > 0 {
		for _, column := range s {
			column = strings.TrimSpace(column)
			m[column] = struct{}{}
		}
	}
	return m
}

type separate struct {
	prefix, separator, suffix string
	writeFixIfEmpty           bool
}

func newDaoSqlBuilder[T any](d *Dao[T], entities []*T) *DaoSqlBuilder[T] {
	b := &DaoSqlBuilder[T]{dao: d, entities: entities}
	b.sqlBuilder__ = extendSqlBuilder(b)
	return b
}
