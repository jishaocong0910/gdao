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
	"context"
	"database/sql"
	"errors"
	"github.com/jishaocong0910/gdao/internal"
	"reflect"
	"strings"
)

type Dao[T any] struct {
	*baseDao
	commaColumns           string
	columns                []string
	columnToFieldIndex     map[string]int
	columnToFieldConvertor map[string]fieldConvertor
	fieldNameToColumn      map[string]string
	autoIncrementColumns   []string
	autoIncrementStep      int64
	autoIncrementConvert   func(id int64) reflect.Value
}

func (d *Dao[T]) Query() *query[T] {
	return &query[T]{dao: d}
}

func (d *Dao[T]) Exec() *exec[T] {
	return &exec[T]{dao: d}
}

func (d *Dao[T]) NameMap() map[string]string {
	return d.fieldNameToColumn
}

func (d *Dao[T]) mappingScanFields(entity *T, columns []string) ([]any, []func()) {
	v := reflect.ValueOf(entity).Elem()
	dests := make([]any, 0, len(columns))
	afterScans := make([]func(), 0, len(columns))
	for _, c := range columns {
		if index, ok := d.columnToFieldIndex[c]; ok {
			field := v.Field(index)
			if fc, ok := d.columnToFieldConvertor[c]; ok {
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

func (d *Dao[T]) registerEntity(b *daoBuilder[T]) error {
	err := checkEntityType[T]()
	if err != nil {
		return err
	}
	t := reflect.TypeOf((*T)(nil)).Elem()
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			if b.allowInvalidField {
				continue
			}
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" must be exported")
		}
		if !tf.Anonymous {
			ft := tf.Type
			switch internal.IsImplementConvert(ft) {
			case 1:
				fc := getFieldConvertor(ft)
				d.registerField(tf, b.columnMapper, &fc)
				continue
			case 2:
				if !b.allowInvalidField {
					return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" is invalid implementing gdao.Convert")
				}
			}
			if ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice {
				if internal.IsBaseType(ft.Elem()) {
					d.registerField(tf, b.columnMapper, nil)
					continue
				}
			}
			if b.allowInvalidField {
				continue
			}
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" is not supported type")
		}
	}
	return nil
}

func (d *Dao[T]) registerField(tf reflect.StructField, columnMapper *NameMapper, fieldConvertor *fieldConvertor) {
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

	d.columns = append(d.columns, column)
	if d.commaColumns != "" {
		d.commaColumns += ", "
	}
	d.commaColumns += column
	d.columnToFieldIndex[column] = tf.Index[0]
	d.fieldNameToColumn[tf.Name] = column
	if t.isAutoIncrement {
		if convertor := lastInsertIdConvertor_.OfString(tf.Type.Elem().String()); !convertor.IsUndefined() {
			d.autoIncrementColumns = append(d.autoIncrementColumns, column)
			d.autoIncrementStep = t.autoIncrementStep
			d.autoIncrementConvert = convertor.convert
		}
	}
	if fieldConvertor != nil {
		d.columnToFieldConvertor[column] = *fieldConvertor
	}
	return
}

type query[T any] struct {
	dao      *Dao[T]
	ctx      context.Context
	must     bool
	logLevel LogLevel
	desc     string
	rowAs    RowAs
	entities []*T
	buildSql func(b *DaoSqlBuilder[T])
}

func (q *query[T]) Ctx(ctx context.Context) *query[T] {
	q.ctx = ctx
	return q
}

func (q *query[T]) Must(must bool) *query[T] {
	q.must = must
	return q
}

func (q *query[T]) LogLevel(logLevel LogLevel) *query[T] {
	q.logLevel = logLevel
	return q
}

func (q *query[T]) Desc(desc string) *query[T] {
	q.desc = desc
	return q
}

func (q *query[T]) RowAs(rowAs RowAs) *query[T] {
	q.rowAs = rowAs
	return q
}

func (q *query[T]) Entities(entities ...*T) *query[T] {
	q.entities = entities
	return q
}

func (q *query[T]) BuildSql(buildSql func(b *DaoSqlBuilder[T])) *query[T] {
	q.buildSql = buildSql
	return q
}

func (q *query[T]) Do() (first *T, list []*T, err error) {
	list = make([]*T, 0)
	b := newDaoSqlBuilder(q.dao, q.entities)
	q.buildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		checkMust(q.must, err)
		return nil, list, err
	}
	if !b.Ok() { // coverage-ignore
		return
	}
	rows, columns, closeFunc, err := q.dao.query(q.ctx, b.Sql(), b.Args())
	if err != nil { // coverage-ignore
		printSql(q.ctx, q.logLevel, q.desc, b.Sql(), b.Args(), -1, -1, err)
		checkMust(q.must, err)
		return nil, nil, err
	}
	defer closeFunc()

	switch q.rowAs.String() {
	case RowAs_.RETURNING.String():
		var affected int64
		for i := 0; rows.Next() && i < len(q.entities); i++ {
			entity := q.entities[i]
			if entity == nil { // coverage-ignore
				continue
			}
			v := reflect.ValueOf(entity).Elem()
			var fields []any
			for _, c := range columns {
				if fieldIndex, ok := q.dao.columnToFieldIndex[c]; ok {
					field := v.Field(fieldIndex).Addr().Interface()
					fields = append(fields, field)
				}
			}
			if len(fields) > 0 {
				printWarn(q.ctx, rows.Scan(fields...))
			}
			affected++
		}
		printSql(q.ctx, q.logLevel, q.desc, b.Sql(), b.Args(), affected, -1, nil)
	case RowAs_.LAST_ID.String():
		var affected int64
		var id *int64
		if rows.Next() && len(columns) == 1 && len(q.dao.autoIncrementColumns) == 1 {
			err = rows.Scan(&id)
			printWarn(q.ctx, err)
			if err != nil && rows.Next() { // coverage-ignore
				id = nil
			}
		}
		if id != nil {
			fieldIndex := q.dao.columnToFieldIndex[q.dao.autoIncrementColumns[0]]
			entityLength := len(q.entities)
			for i := 0; i < entityLength; i++ {
				entity := q.entities[i]
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(q.dao.autoIncrementConvert(*id - int64(entityLength-1-i)*q.dao.autoIncrementStep))
				affected++
			}
		} else {
			for i := 0; i < len(q.entities); i++ {
				entity := q.entities[i]
				if entity != nil { // coverage-ignore
					affected++
				}
			}
		}
		printSql(q.ctx, q.logLevel, q.desc, b.Sql(), b.Args(), affected, -1, nil)
	default:
		var rowCounts int64
		for rows.Next() {
			entity := new(T)
			dests, afterScans := q.dao.mappingScanFields(entity, columns)
			err = rows.Scan(dests...)
			if err != nil {
				checkMust(q.must, err)
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
		printSql(q.ctx, q.logLevel, q.desc, b.Sql(), b.Args(), -1, rowCounts, nil)
	}
	return
}

type exec[T any] struct {
	dao            *Dao[T]
	ctx            context.Context
	must           bool
	logLevel       LogLevel
	desc           string
	lastInsertIdAs LastInsertIdAs
	entities       []*T
	buildSql       func(b *DaoSqlBuilder[T])
}

func (e *exec[T]) Ctx(ctx context.Context) *exec[T] {
	e.ctx = ctx
	return e
}

func (e *exec[T]) Must(must bool) *exec[T] {
	e.must = must
	return e
}

func (e *exec[T]) LogLevel(logLevel LogLevel) *exec[T] {
	e.logLevel = logLevel
	return e
}

func (e *exec[T]) Desc(desc string) *exec[T] {
	e.desc = desc
	return e
}

func (e *exec[T]) LastInsertIdAs(lastInsertIdAs LastInsertIdAs) *exec[T] {
	e.lastInsertIdAs = lastInsertIdAs
	return e
}

func (e *exec[T]) Entities(entities ...*T) *exec[T] {
	e.entities = entities
	return e
}

func (e *exec[T]) BuildSql(buildSql func(b *DaoSqlBuilder[T])) *exec[T] {
	e.buildSql = buildSql
	return e
}

func (e *exec[T]) Do() (affected int64, err error) {
	b := newDaoSqlBuilder(e.dao, e.entities)
	e.buildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		checkMust(e.must, err)
		return 0, err
	}
	if !b.Ok() { // coverage-ignore
		return 0, nil
	}
	result, affected, err := e.dao.exec(e.ctx, b.Sql(), b.Args())
	printSql(e.ctx, e.logLevel, e.desc, b.Sql(), b.Args(), affected, -1, err)
	if err != nil { // coverage-ignore
		checkMust(e.must, err)
		return
	}

	switch e.lastInsertIdAs.String() {
	case LastInsertIdAs_.FIRST_ID.String():
		id, err := result.LastInsertId()
		printWarn(e.ctx, err)
		if err == nil && len(e.entities) > 0 && len(e.dao.autoIncrementColumns) == 1 {
			fieldIndex := e.dao.columnToFieldIndex[e.dao.autoIncrementColumns[0]]
			for i, entity := range e.entities {
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(e.dao.autoIncrementConvert(id + int64(i)*e.dao.autoIncrementStep))
			}
		}
	case LastInsertIdAs_.LAST_ID.String():
		id, err := result.LastInsertId()
		printWarn(e.ctx, err)
		if err == nil && len(e.entities) > 0 && len(e.dao.autoIncrementColumns) == 1 {
			fieldIndex := e.dao.columnToFieldIndex[e.dao.autoIncrementColumns[0]]
			entityLength := len(e.entities)
			for i := 0; i < entityLength; i++ {
				entity := e.entities[i]
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(e.dao.autoIncrementConvert(id - int64(entityLength-1-i)*e.dao.autoIncrementStep))
			}
		}
	}
	return
}

type DaoSqlBuilder[T any] struct {
	*BaseSqlBuilder
	dao      *Dao[T]
	entities []*T
}

func (this *DaoSqlBuilder[T]) Write(str string, args ...any) *DaoSqlBuilder[T] {
	this.BaseSqlBuilder.Write(str, args...)
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
				fieldIndex := this.dao.columnToFieldIndex[column]
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
	fieldIndex, ok := this.dao.columnToFieldIndex[column]
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

func (this *DaoSqlBuilder[T]) EachEntity(sep *Separate, handle func(n int, entity *T)) *DaoSqlBuilder[T] {
	var n int
	this.WritePrefix(sep, n)
	for _, entity := range this.entities {
		n++
		this.WritePrefix(sep, n)
		this.WriteSep(sep, n)
		handle(n, entity)
	}
	this.WriteSuffix(sep, n)
	return this
}

func (this *DaoSqlBuilder[T]) EachColumn(entity *T, sep *Separate, handle func(n int, column string, value any), columns ...string) {
	v := reflect.ValueOf(entity).Elem()
	var n int
	this.WritePrefix(sep, n)
	for _, column := range columns {
		fieldIndex := this.dao.columnToFieldIndex[column]
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}
		n++
		this.WritePrefix(sep, n)
		this.WriteSep(sep, n)

		handle(n, column, value)
	}
	this.WriteSuffix(sep, n)
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

func newDaoSqlBuilder[T any](d *Dao[T], entities []*T) *DaoSqlBuilder[T] {
	return &DaoSqlBuilder[T]{BaseSqlBuilder: NewBaseSqlBuilder(), dao: d, entities: entities}
}

type daoBuilder[T any] struct {
	db                *sql.DB
	allowInvalidField bool
	columnMapper      *NameMapper
}

func (b *daoBuilder[T]) DB(db *sql.DB) *daoBuilder[T] {
	b.db = db
	return b
}

func (b *daoBuilder[T]) AllowInvalidField(allowInvalidField bool) *daoBuilder[T] {
	b.allowInvalidField = allowInvalidField
	return b
}

func (b *daoBuilder[T]) ColumnMapper(columnMapper *NameMapper) *daoBuilder[T] {
	b.columnMapper = columnMapper
	return b
}

func (b *daoBuilder[T]) Build() *Dao[T] {
	dao := &Dao[T]{
		baseDao:                newBaseDao(b.db),
		columnToFieldIndex:     make(map[string]int),
		columnToFieldConvertor: make(map[string]fieldConvertor),
		fieldNameToColumn:      make(map[string]string),
	}
	err := dao.registerEntity(b)
	must(err)
	return dao
}

func DaoBuilder[T any]() *daoBuilder[T] {
	return &daoBuilder[T]{}
}
