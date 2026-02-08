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
	"strconv"
	"strings"
)

// PlainSqlBuilder 无泛型的SQL构造器
type PlainSqlBuilder struct {
	sql    strings.Builder
	args   []any
	argNum int
	cancel bool
	err    error
}

func (b *PlainSqlBuilder) Write(str string, args ...any) *PlainSqlBuilder {
	b.sql.WriteString(str)
	b.SetArgs(args...)
	return b
}

func (b *PlainSqlBuilder) WriteIf(str string, check bool, args ...any) *PlainSqlBuilder {
	if check {
		b.sql.WriteString(str)
		b.SetArgs(args...)
	}
	return b
}

func (b *PlainSqlBuilder) SetArgs(args ...any) *PlainSqlBuilder {
	b.args = append(b.args, args...)
	return b
}

func (b *PlainSqlBuilder) Ph(prefix string) string {
	b.argNum++
	return prefix + strconv.Itoa(b.argNum)
}

func (b *PlainSqlBuilder) Sql() string {
	return b.sql.String()
}

func (b *PlainSqlBuilder) Args() []any {
	return b.args
}

func (b *PlainSqlBuilder) Cancel() bool {
	return b.cancel
}

func (b *PlainSqlBuilder) SetCancel(cancel bool) {
	if b.err == nil {
		b.cancel = cancel
	}
}

func (b *PlainSqlBuilder) Error() error {
	return b.err
}

func (b *PlainSqlBuilder) SetError(err error) {
	if err != nil {
		b.err = err
		b.cancel = true
	}
}

func (b *PlainSqlBuilder) Sep(separator string) *Separate {
	return &Separate{separator: separator}
}

func (b *PlainSqlBuilder) SepFix(prefix, separator, suffix string, omitempty bool) *Separate {
	return &Separate{prefix: prefix, separator: separator, suffix: suffix, omitempty: omitempty}
}

func (b *PlainSqlBuilder) Repeat(num int, sep *Separate, filter func(i int) bool, handle func(n, i int)) {
	var n int
	b.writePrefix(sep, n)
	for i := 0; i < num; i++ {
		if filter != nil && !filter(i) {
			continue
		}
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)
		handle(n, i)
	}
	b.writeSuffix(sep, n)
}

func (b *PlainSqlBuilder) writeSep(s *Separate, n int) {
	if s != nil && n != 1 {
		b.Write(s.separator)
	}
}

func (b *PlainSqlBuilder) writePrefix(s *Separate, n int) {
	if s != nil && s.prefix != "" {
		if n == 0 && !s.omitempty || n == 1 && s.omitempty {
			b.Write(s.prefix)
		}
	}
}

func (b *PlainSqlBuilder) writeSuffix(s *Separate, n int) {
	if s != nil && s.suffix != "" && n != 0 {
		b.Write(s.suffix)
	}
}

type Separate struct {
	prefix, separator, suffix string
	omitempty                 bool
}

type SqlBuilder[T Entity] struct {
	*PlainSqlBuilder
	dao      *Dao[T]
	entities []*T
}

func (b *SqlBuilder[T]) Write(str string, args ...any) *SqlBuilder[T] {
	b.PlainSqlBuilder.Write(str, args...)
	return b
}

func (b *SqlBuilder[T]) WriteIf(str string, check bool, args ...any) *SqlBuilder[T] {
	if check {
		b.sql.WriteString(str)
		b.SetArgs(args...)
	}
	return b
}

func (b *SqlBuilder[T]) WriteTable() *SqlBuilder[T] {
	b.Write(b.dao.table)
	return b
}

func (b *SqlBuilder[T]) WriteColumns(columns ...string) *SqlBuilder[T] {
	if len(columns) == 0 {
		b.Write(b.dao.commaColumns)
	} else {
		for i, c := range columns {
			if c == "" {
				continue
			}
			if i != 0 {
				b.Write(", ")
			}
			b.Write(c)
		}
	}
	return b
}

func (b *SqlBuilder[T]) SetArgs(args ...any) *SqlBuilder[T] {
	b.PlainSqlBuilder.SetArgs(args...)
	return b
}

func (b *SqlBuilder[T]) AutoColumns() []string {
	return b.dao.autoIncrementColumns
}

func (b *SqlBuilder[T]) Columns(ignored []string) (columns []string) {
	if len(ignored) == 0 {
		return b.dao.columns
	}
	ignoredSet := toSet(ignored)
	for _, column := range b.dao.columns {
		if _, ok := ignoredSet[column]; !ok {
			columns = append(columns, column)
		}
	}
	return
}

func (b *SqlBuilder[T]) AssignedColumns(entity *T, fixed []string, ignored []string) (columns []string) {
	if entity == nil { // coverage-ignore
		return
	}
	v := reflect.ValueOf(entity).Elem()
	ignoredSet := toSet(ignored)
	fixedSet := toSet(fixed)
	for _, column := range b.dao.columns {
		if fieldIndex, ok := b.dao.columnToFieldIndex[column]; ok {
			if _, ok := ignoredSet[column]; ok {
				continue
			}
			if v.Field(fieldIndex).IsNil() {
				if _, ok := fixedSet[column]; !ok {
					continue
				}
			}
			columns = append(columns, column)
		}
	}
	return
}

func (b *SqlBuilder[T]) Entity() *T {
	return b.EntityAt(0)
}

func (b *SqlBuilder[T]) EntityAt(index int) *T {
	var t *T
	if index < len(b.entities) {
		t = b.entities[index]
	}
	return t
}

func (b *SqlBuilder[T]) ColumnValue(entity *T, column string) any {
	if entity == nil {
		return nil
	}
	fieldIndex, ok := b.dao.columnToFieldIndex[column]
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

func (b *SqlBuilder[T]) EachEntity(sep *Separate, handle func(n int, entity *T)) *SqlBuilder[T] {
	b.Repeat(len(b.entities), sep, nil, func(n int, i int) {
		handle(n, b.entities[i])
	})
	return b
}

func (b *SqlBuilder[T]) EachColumn(entity *T, sep *Separate, handle func(n int, column string, value any), columns ...string) *SqlBuilder[T] {
	if entity == nil { // coverage-ignore
		return b
	}
	v := reflect.ValueOf(entity).Elem()
	b.Repeat(len(columns), sep, nil, func(n int, i int) {
		column := columns[i]
		var value any
		if fieldIndex, ok := b.dao.columnToFieldIndex[column]; ok {
			field := v.Field(fieldIndex)
			if !field.IsNil() {
				value = field.Interface()
			}
		}
		handle(n, column, value)
	})
	return b
}

func newSqlBuilder[T Entity](d *Dao[T], entities []*T) *SqlBuilder[T] {
	return &SqlBuilder[T]{PlainSqlBuilder: &PlainSqlBuilder{}, dao: d, entities: entities}
}
