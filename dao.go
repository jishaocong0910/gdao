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
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/jishaocong0910/gdao/internal"
)

type query[T any] struct {
	dao         *Dao[T]
	ctx         context.Context
	must        bool
	sqlLogLevel LogLevel
	desc        string
	rowAs       RowAs
	entities    []*T
	buildSql    func(b *SqlBuilder[T])
}

func (q *query[T]) Ctx(ctx context.Context) *query[T] {
	q.ctx = ctx
	return q
}

func (q *query[T]) Must(must bool) *query[T] {
	q.must = must
	return q
}

func (q *query[T]) SqlLogLevel(logLevel LogLevel) *query[T] {
	q.sqlLogLevel = logLevel
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

func (q *query[T]) BuildSql(buildSql func(b *SqlBuilder[T])) *query[T] {
	q.buildSql = buildSql
	return q
}

func (q *query[T]) Do() (first *T, list []*T, err error) {
	list = make([]*T, 0)
	b := newSqlBuilder(q.dao, q.entities)
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
		printSql(q.ctx, q.sqlLogLevel, q.desc, b.Sql(), b.Args(), -1, -1, err)
		checkMust(q.must, err)
		return nil, nil, err
	}
	defer closeFunc()

	switch q.rowAs.String() {
	case RowAs_.RETURNING.String():
		q.rowAsReturning(b, rows, columns)
	case RowAs_.LAST_ID.String():
		q.rowAsLastId(b, rows, columns)
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
		printSql(q.ctx, q.sqlLogLevel, q.desc, b.Sql(), b.Args(), -1, rowCounts, nil)
	}
	return
}

func (q *query[T]) rowAsReturning(b *SqlBuilder[T], rows *sql.Rows, columns []string) {
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
	printSql(q.ctx, q.sqlLogLevel, q.desc, b.Sql(), b.Args(), affected, -1, nil)
}

func (q *query[T]) rowAsLastId(b *SqlBuilder[T], rows *sql.Rows, columns []string) {
	var affected int64
	var id *int64
	if rows.Next() && len(columns) == 1 && len(q.dao.autoIncrementColumns) == 1 {
		err := rows.Scan(&id)
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
	printSql(q.ctx, q.sqlLogLevel, q.desc, b.Sql(), b.Args(), affected, -1, nil)
}

type exec[T any] struct {
	dao            *Dao[T]
	ctx            context.Context
	must           bool
	sqlLogLevel    LogLevel
	desc           string
	lastInsertIdAs LastInsertIdAs
	entities       []*T
	buildSql       func(b *SqlBuilder[T])
}

func (e *exec[T]) Ctx(ctx context.Context) *exec[T] {
	e.ctx = ctx
	return e
}

func (e *exec[T]) Must(must bool) *exec[T] {
	e.must = must
	return e
}

func (e *exec[T]) SqlLogLevel(logLevel LogLevel) *exec[T] {
	e.sqlLogLevel = logLevel
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

func (e *exec[T]) BuildSql(buildSql func(b *SqlBuilder[T])) *exec[T] {
	e.buildSql = buildSql
	return e
}

func (e *exec[T]) Do() (affected int64, err error) {
	b := newSqlBuilder(e.dao, e.entities)
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
	printSql(e.ctx, e.sqlLogLevel, e.desc, b.Sql(), b.Args(), affected, -1, err)
	if err != nil { // coverage-ignore
		checkMust(e.must, err)
		return
	}

	switch e.lastInsertIdAs.String() {
	case LastInsertIdAs_.FIRST_ID.String():
		e.lastInsertIdAsFirstId(result)
	case LastInsertIdAs_.LAST_ID.String():
		e.lastInsertIdAsLastId(result)
	}
	return
}

func (e *exec[T]) lastInsertIdAsFirstId(result sql.Result) {
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
}

func (e *exec[T]) lastInsertIdAsLastId(result sql.Result) {
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

type count[T any] struct {
	dao         *Dao[T]
	ctx         context.Context
	must        bool
	sqlLogLevel LogLevel
	desc        string
	entities    []*T
	buildSql    func(b *SqlBuilder[T])
}

func (c *count[T]) Ctx(ctx context.Context) *count[T] {
	c.ctx = ctx
	return c
}

func (c *count[T]) Must(must bool) *count[T] {
	c.must = must
	return c
}

func (c *count[T]) SqlLogLevel(logLevel LogLevel) *count[T] {
	c.sqlLogLevel = logLevel
	return c
}

func (c *count[T]) Desc(desc string) *count[T] {
	c.desc = desc
	return c
}

func (c *count[T]) Entities(entities ...*T) *count[T] {
	c.entities = entities
	return c
}

func (c *count[T]) BuildSql(buildSql func(b *SqlBuilder[T])) *count[T] {
	c.buildSql = buildSql
	return c
}

func (c *count[T]) Do() (count *Count, err error) {
	b := newSqlBuilder(c.dao, c.entities)
	c.buildSql(b)
	if !b.Ok() { // coverage-ignore
		return nil, b.Error()
	}
	rows, columns, closeFunc, err := c.dao.query(c.ctx, b.Sql(), b.Args())
	if err != nil { // coverage-ignore
		printSql(c.ctx, c.sqlLogLevel, c.desc, b.Sql(), b.Args(), -1, -1, err)
		checkMust(c.must, err)
		return nil, err
	}
	defer closeFunc()

	var rowCounts int64
	for rows.Next() {
		rowCounts++
		if count != nil { // coverage-ignore
			continue
		}
		count = &Count{}
		if len(columns) > 1 {
			count = nil
			err = errors.New("returns more than one column")
			checkMust(c.must, err)
			return
		}
		err = rows.Scan(&count.Value)
		if err != nil { // coverage-ignore
			count = nil
			checkMust(c.must, err)
			return
		}
	}

	if rowCounts > 1 {
		count = nil
		err = errors.New("returns more than one row")
		printSql(c.ctx, c.sqlLogLevel, c.desc, b.Sql(), b.Args(), -1, rowCounts, err)
		checkMust(c.must, err)
		return count, err
	}
	printSql(c.ctx, c.sqlLogLevel, c.desc, b.Sql(), b.Args(), -1, rowCounts, nil)
	return
}

type Count struct {
	Value *int64
}

func (c *Count) Int() int {
	if c == nil || c.Value == nil {
		return 0
	}
	return int(*c.Value)
}

func (c *Count) Int8() int8 {
	if c == nil || c.Value == nil {
		return 0
	}
	return int8(*c.Value)
}

func (c *Count) Int16() int16 {
	if c == nil || c.Value == nil {
		return 0
	}
	return int16(*c.Value)
}

func (c *Count) Int32() int32 {
	if c == nil || c.Value == nil {
		return 0
	}
	return int32(*c.Value)
}

func (c *Count) Int64() int64 {
	if c == nil || c.Value == nil {
		return 0
	}
	return *c.Value
}

func (c *Count) Bool() bool {
	if c == nil || c.Value == nil {
		return false
	}
	return *c.Value > 0
}

func (c *Count) IntPtr() *int {
	if c == nil {
		return nil
	}
	i := int(*c.Value)
	return &i
}

func (c *Count) Int8Ptr() *int8 {
	if c == nil {
		return nil
	}
	i := int8(*c.Value)
	return &i
}

func (c *Count) Int16Ptr() *int16 {
	if c == nil {
		return nil
	}
	i := int16(*c.Value)
	return &i
}

func (c *Count) Int32Ptr() *int32 {
	if c == nil {
		return nil
	}
	i := int32(*c.Value)
	return &i
}

func (c *Count) Int64Ptr() *int64 {
	if c == nil {
		return nil
	}
	return c.Value
}

func (c *Count) BoolPtr() *bool {
	if c == nil {
		return nil
	}
	b := *c.Value > 0
	return &b
}

type Dao[T any] struct {
	*baseDao
	table                  string
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

func (d *Dao[T]) Count() *count[T] {
	return &count[T]{dao: d}
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
		if tf.Anonymous {
			continue
		}
		tag := parseTag(tf)
		if tag.skip {
			continue
		}
		if !tf.IsExported() {
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" must be exported")
		}
		column := d.determineColumn(tf, tag, b.columnMapper)
		if column == "" {
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" has not specified the column name")
		}
		ft := tf.Type
		switch internal.IsImplementConvert(ft) {
		case 1:
			fc := getFieldConvertor(ft)
			d.registerField(tf, tag, column, &fc)
			continue
		case 2:
			return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" is invalid implementing gdao.Convert")
		}
		if ft.Kind() == reflect.Pointer || ft.Kind() == reflect.Slice {
			if internal.IsBaseType(ft.Elem()) {
				d.registerField(tf, tag, column, nil)
				continue
			}
		}
		return errors.New("field \"" + tf.Name + "\" of \"" + t.String() + "\" is not supported type")
	}
	return nil
}

func (d *Dao[T]) determineColumn(tf reflect.StructField, t tag, columnMapper *NameMapper) string {
	column := t.column
	if t.column == "" && columnMapper != nil {
		column = columnMapper.Convert(tf.Name)
	}
	return column
}

func (d *Dao[T]) registerField(tf reflect.StructField, tag tag, column string, fieldConvertor *fieldConvertor) {
	d.columns = append(d.columns, column)
	if d.commaColumns != "" {
		d.commaColumns += ", "
	}
	d.commaColumns += column
	d.columnToFieldIndex[column] = tf.Index[0]
	d.fieldNameToColumn[tf.Name] = column
	if tag.autoIncrement {
		if convertor := lastInsertIdConvertor_.OfString(tf.Type.Elem().String()); !convertor.IsUndefined() {
			d.autoIncrementColumns = append(d.autoIncrementColumns, column)
			d.autoIncrementStep = tag.autoIncrementStep
			d.autoIncrementConvert = convertor.convert
		}
	}
	if fieldConvertor != nil {
		d.columnToFieldConvertor[column] = *fieldConvertor
	}
}

type SqlBuilder[T any] struct {
	*BaseSqlBuilder
	dao      *Dao[T]
	entities []*T
}

func (b *SqlBuilder[T]) Write(str string, args ...any) *SqlBuilder[T] {
	b.BaseSqlBuilder.Write(str, args...)
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

func (b *SqlBuilder[T]) Columns(onlyAssigned bool, ignoredColumns ...string) (columns []string) {
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
				fieldIndex := b.dao.columnToFieldIndex[column]
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

func (b *SqlBuilder[T]) AutoColumns() []string {
	return b.dao.autoIncrementColumns
}

func (b *SqlBuilder[T]) EntityAt(index int) *T {
	var t *T
	if index < len(b.entities) {
		t = b.entities[index]
	}
	return t
}

func (b *SqlBuilder[T]) Entity() *T {
	return b.EntityAt(0)
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
	b.WritePrefix(sep, n)
	for _, entity := range b.entities {
		n++
		b.WritePrefix(sep, n)
		b.WriteSep(sep, n)
		handle(n, entity)
	}
	b.WriteSuffix(sep, n)
	return b
}

func (b *SqlBuilder[T]) EachColumn(entity *T, sep *Separate, handle func(n int, column string, value any), columns ...string) {
	v := reflect.ValueOf(entity).Elem()
	var n int
	b.WritePrefix(sep, n)
	for _, column := range columns {
		fieldIndex := b.dao.columnToFieldIndex[column]
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}
		n++
		b.WritePrefix(sep, n)
		b.WriteSep(sep, n)

		handle(n, column, value)
	}
	b.WriteSuffix(sep, n)
	return
}

func (b *SqlBuilder[T]) toMap(s []string) map[string]struct{} {
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
	return &SqlBuilder[T]{BaseSqlBuilder: NewBaseSqlBuilder(), dao: d, entities: entities}
}

type daoBuilder[T any] struct {
	db           *sql.DB
	table        string
	columnMapper *NameMapper
}

func (b *daoBuilder[T]) DB(db *sql.DB) *daoBuilder[T] {
	b.db = db
	return b
}

func (b *daoBuilder[T]) Table(table string) *daoBuilder[T] {
	b.table = table
	return b
}

func (b *daoBuilder[T]) ColumnMapper(columnMapper *NameMapper) *daoBuilder[T] {
	b.columnMapper = columnMapper
	return b
}

func (b *daoBuilder[T]) Build() *Dao[T] {
	dao := &Dao[T]{
		baseDao:                newBaseDao(b.db),
		table:                  b.table,
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
