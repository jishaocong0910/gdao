package gdao

import (
	"reflect"
	"strconv"
	"strings"
)

type Builder[T any] struct {
	dao      *Dao[T]
	entities []*T
	sql      strings.Builder
	args     []any
	argNum   int
	ok       bool
}

func (b *Builder[T]) Write(str string, args ...any) *Builder[T] {
	b.sql.WriteString(str)
	b.args = append(b.args, args...)
	return b
}

func (b *Builder[T]) Arg(a any) *Builder[T] {
	b.args = append(b.args, a)
	return b
}

func (b *Builder[T]) WriteColumns(columns ...string) *Builder[T] {
	if len(columns) == 0 {
		b.Write(b.dao.commaColumns)
	} else {
		for i, c := range columns {
			if c == "" {
				continue
			}
			if i != 0 {
				b.Write(",")
			}
			b.Write(c)
		}
	}
	return b
}

func (b *Builder[T]) Columns() []string {
	return b.dao.columns
}

func (b *Builder[T]) AutoColumns() []string {
	return b.dao.autoIncrementColumns
}

func (b *Builder[T]) Pp(prefix string) string {
	b.argNum++
	return prefix + strconv.Itoa(b.argNum)
}

func (b *Builder[T]) Sql() string {
	return b.sql.String()
}

func (b *Builder[T]) Args() []any {
	return b.args
}

func (b *Builder[T]) SetOk(ok bool) {
	b.ok = ok
}

func (b *Builder[T]) Ok() bool {
	return b.ok
}

func (b *Builder[T]) EntityAt(index int) *T {
	var t *T
	if index < len(b.entities) {
		t = b.entities[index]
	}
	return t
}

func (b *Builder[T]) Entity() *T {
	return b.EntityAt(0)
}

func (b *Builder[T]) ColumnValuesAt(entity *T, onlyAssigned bool, filterColumns ...string) []ColumnValue {
	if entity == nil { // coverage-ignore
		return nil
	}
	v := reflect.ValueOf(entity).Elem()

	filterColumnMap := make(map[string]struct{}, len(filterColumns))
	if len(filterColumns) > 0 {
		for _, column := range filterColumns {
			filterColumnMap[column] = struct{}{}
		}
	}

	var cvs []ColumnValue
	for _, column := range b.dao.columns {
		fieldIndex := b.dao.columnToFieldIndexMap[column]
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}
		if onlyAssigned && value == nil {
			continue
		}
		_, c := filterColumnMap[column]
		if c {
			continue
		}
		cvs = append(cvs, ColumnValue{Column: column, Value: value})
	}
	return cvs
}

func (b *Builder[T]) ColumnValues(onlyAssigned bool, filterColumns ...string) []ColumnValue {
	return b.ColumnValuesAt(b.Entity(), onlyAssigned, filterColumns...)
}

func (b *Builder[T]) ColumnValue(entity *T, column string) any {
	if entity == nil {
		return nil
	}
	fieldIndex, ok := b.dao.columnToFieldIndexMap[column]
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

func (b *Builder[T]) EachColumnName(columnNames []string, sep *separate, handle func(n, i int, column string), filterColumns ...string) {
	filterColumnMap := make(map[string]struct{}, len(filterColumns))
	if len(filterColumns) > 0 {
		for _, column := range filterColumns {
			column = strings.TrimSpace(column)
			filterColumnMap[column] = struct{}{}
		}
	}
	var n int
	b.writePrefix(sep, n)
	for i, column := range columnNames {
		column = strings.TrimSpace(column)
		if column == "" {
			continue
		}
		_, c := filterColumnMap[column]
		if c || column == "" {
			continue
		}
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)
		handle(n, i, column)
	}
	b.writeSuffix(sep, n)
}

func (b *Builder[T]) EachEntity(sep *separate, handle func(n, i int, entity *T)) *Builder[T] {
	var n int
	b.writePrefix(sep, n)
	for i := range b.entities {
		entity := b.EntityAt(i)
		if entity == nil {
			continue
		}
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)
		handle(n, i, entity)
	}
	b.writeSuffix(sep, n)
	return b
}

func (b *Builder[T]) EachColumnValues(cvs []ColumnValue, sep *separate, handle columnValueHandle) *Builder[T] {
	var n int
	b.writePrefix(sep, n)
	for _, cv := range cvs {
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)
		handle(cv.Column, cv.Value)
	}
	b.writeSuffix(sep, n)
	return b
}

func (b *Builder[T]) Repeat(num int, sep *separate, filter func(i int) bool, handle func(n, i int)) *Builder[T] {
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
	return b
}

func (b *Builder[T]) Sep(separator string) *separate {
	return &separate{separator: separator}
}

func (b *Builder[T]) SepFix(prefix, separator, suffix string, writeFixIfEmpty bool) *separate {
	return &separate{prefix: prefix, separator: separator, suffix: suffix, writeFixIfEmpty: writeFixIfEmpty}
}

func (b *Builder[T]) writePrefix(s *separate, n int) {
	if s != nil {
		if n == 0 && s.writeFixIfEmpty || n == 1 && !s.writeFixIfEmpty {
			b.Write(s.prefix)
		}
	}
}

func (b *Builder[T]) writeSep(s *separate, n int) {
	if s != nil && n != 1 {
		b.Write(s.separator)
	}
}

func (b *Builder[T]) writeSuffix(s *separate, n int) {
	if s != nil && n != 0 {
		b.Write(s.suffix)
	}
}

type ColumnValue struct {
	Column string
	Value  any
}

type separate struct {
	prefix, separator, suffix string
	writeFixIfEmpty           bool
}

type columnValueHandle func(column string, value any)

func newBuilder[T any](d *Dao[T], entities []*T) *Builder[T] {
	return &Builder[T]{dao: d, entities: entities, ok: true}
}
