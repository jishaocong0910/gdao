package gdao

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

func (b *PlainSqlBuilder) SepFix(prefix, separator, suffix string, writeFixIfEmpty bool) *Separate {
	return &Separate{prefix: prefix, separator: separator, suffix: suffix, writeFixIfEmpty: writeFixIfEmpty}
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
	if s != nil {
		if n == 0 && s.writeFixIfEmpty || n == 1 && !s.writeFixIfEmpty {
			b.Write(s.prefix)
		}
	}
}

func (b *PlainSqlBuilder) writeSuffix(s *Separate, n int) {
	if s != nil && n != 0 {
		b.Write(s.suffix)
	}
}

type Separate struct {
	prefix, separator, suffix string
	writeFixIfEmpty           bool
}

type SqlBuilder[T any] struct {
	*PlainSqlBuilder
	dao      *Dao[T]
	entities []*T
}

func (b *SqlBuilder[T]) Write(str string, args ...any) *SqlBuilder[T] {
	b.PlainSqlBuilder.Write(str, args...)
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

func (b *SqlBuilder[T]) Columns(onlyAssigned bool, ignoredColumns ...string) (columns []string) {
	if !onlyAssigned {
		if len(ignoredColumns) == 0 {
			return b.dao.columns
		}
		ignoredColumnSet := b.toSet(ignoredColumns)
		for _, column := range b.dao.columns {
			if _, ok := ignoredColumnSet[column]; !ok {
				columns = append(columns, column)
			}
		}
		return
	} else {
		entity := b.Entity()
		if entity != nil {
			v := reflect.ValueOf(entity).Elem()
			ignoredColumnSet := b.toSet(ignoredColumns)
			for _, column := range b.dao.columns {
				fieldIndex := b.dao.columnToFieldIndex[column]
				field := v.Field(fieldIndex)
				if field.IsNil() {
					continue
				}
				if _, ok := ignoredColumnSet[column]; ok {
					continue
				}
				columns = append(columns, column)
			}
			return
		}
		return
	}
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
	var n int
	b.writePrefix(sep, n)
	for _, entity := range b.entities {
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)
		handle(n, entity)
	}
	b.writeSuffix(sep, n)
	return b
}

func (b *SqlBuilder[T]) EachColumn(entity *T, sep *Separate, handle func(n int, column string, value any), columns ...string) {
	v := reflect.ValueOf(entity).Elem()
	var n int
	b.writePrefix(sep, n)
	for _, column := range columns {
		fieldIndex := b.dao.columnToFieldIndex[column]
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)

		handle(n, column, value)
	}
	b.writeSuffix(sep, n)
	return
}

func (b *SqlBuilder[T]) toSet(s []string) map[string]struct{} {
	m := make(map[string]struct{}, len(s))
	if len(s) > 0 {
		for _, column := range s {
			column = strings.TrimSpace(column)
			m[column] = struct{}{}
		}
	}
	return m
}

func newSqlBuilder[T any](d *Dao[T], entities []*T) *SqlBuilder[T] {
	return &SqlBuilder[T]{PlainSqlBuilder: &PlainSqlBuilder{}, dao: d, entities: entities}
}
