package gdao

import (
	"reflect"
	"strconv"
	"strings"
)

type baseBuilder[T any] struct {
	subBuilder *T
	sql        strings.Builder
	args       []any
	argNum     int
	ok         bool
	err        error
}

func (b *baseBuilder[T]) Write(str string, args ...any) *T {
	b.sql.WriteString(str)
	b.SetArgs(args...)
	return b.subBuilder
}

func (b *baseBuilder[T]) SetArgs(args ...any) *T {
	b.args = append(b.args, args...)
	return b.subBuilder
}

func (b *baseBuilder[T]) Pp(prefix string) string {
	b.argNum++
	return prefix + strconv.Itoa(b.argNum)
}

func (b *baseBuilder[T]) Sql() string {
	return b.sql.String()
}

func (b *baseBuilder[T]) Args() []any {
	return b.args
}

func (b *baseBuilder[T]) SetError(err error) {
	if err != nil {
		b.err = err
		b.ok = false
	}
}

func (b *baseBuilder[T]) Error() error {
	return b.err
}

func (b *baseBuilder[T]) SetOk(ok bool) {
	if b.err == nil {
		b.ok = ok
	}
}

func (b *baseBuilder[T]) Ok() bool {
	return b.ok
}

func (b *baseBuilder[T]) Sep(separator string) *separate {
	return &separate{separator: separator}
}

func (b *baseBuilder[T]) SepFix(prefix, separator, suffix string, writeFixIfEmpty bool) *separate {
	return &separate{prefix: prefix, separator: separator, suffix: suffix, writeFixIfEmpty: writeFixIfEmpty}
}

func (b *baseBuilder[T]) Repeat(num int, sep *separate, filter func(i int) bool, handle func(n, i int)) *T {
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
	return b.subBuilder
}

func (b *baseBuilder[T]) writePrefix(s *separate, n int) {
	if s != nil {
		if n == 0 && s.writeFixIfEmpty || n == 1 && !s.writeFixIfEmpty {
			b.Write(s.prefix)
		}
	}
}

func (b *baseBuilder[T]) writeSep(s *separate, n int) {
	if s != nil && n != 1 {
		b.Write(s.separator)
	}
}

func (b *baseBuilder[T]) writeSuffix(s *separate, n int) {
	if s != nil && n != 0 {
		b.Write(s.suffix)
	}
}

type Builder[T any] struct {
	*baseBuilder[Builder[T]]
	dao      *Dao[T]
	entities []*T
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
				b.Write(", ")
			}
			b.Write(c)
		}
	}
	return b
}

func (b *Builder[T]) Columns(onlyAssigned bool, ignoredColumns ...string) (columns []string) {
	if !onlyAssigned {
		if len(ignoredColumns) == 0 {
			return b.dao.columns
		}
		ignoredColumnMap := b.toMap(ignoredColumns)
		for _, column := range b.dao.columns {
			if _, ok := ignoredColumnMap[column]; !ok {
				columns = append(columns, column)
			}
		}
		return
	} else {
		entity := b.Entity()
		if entity != nil {
			v := reflect.ValueOf(entity).Elem()
			ignoredColumnMap := b.toMap(ignoredColumns)
			for _, column := range b.dao.columns {
				fieldIndex := b.dao.columnToFieldIndexMap[column]
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

func (b *Builder[T]) AutoColumns() []string {
	return b.dao.autoIncrementColumns
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

func (b *Builder[T]) EachEntity(sep *separate, filter func(entity *T) bool, handle func(n int, entity *T)) *Builder[T] {
	var n int
	b.writePrefix(sep, n)
	for _, entity := range b.entities {
		if filter != nil && !filter(entity) {
			continue
		}
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)
		handle(n, entity)
	}
	b.writeSuffix(sep, n)
	return b
}

func (b *Builder[T]) EachColumn(entity *T, sep *separate, filter func(column string, value any) bool, handle func(n int, column string, value any), columns ...string) {
	v := reflect.ValueOf(entity).Elem()
	var n int
	b.writePrefix(sep, n)
	for _, column := range columns {
		fieldIndex := b.dao.columnToFieldIndexMap[column]
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}
		if filter != nil && !filter(column, value) {
			continue
		}
		n++
		b.writePrefix(sep, n)
		b.writeSep(sep, n)

		handle(n, column, value)
	}
	b.writeSuffix(sep, n)
	return
}

func (b *Builder[T]) toMap(s []string) map[string]struct{} {
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

type columnValueHandle func(column string, value any)

func newBaseBuilder[T any](t *T) *baseBuilder[T] {
	return &baseBuilder[T]{subBuilder: t, ok: true}
}

func newBuilder[T any](d *Dao[T], entities []*T) *Builder[T] {
	b := &Builder[T]{dao: d, entities: entities}
	b.baseBuilder = newBaseBuilder(b)
	return b
}
