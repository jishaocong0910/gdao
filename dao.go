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
	"reflect"
	"strconv"
	"strings"
)

var DEFAULT_DB *sql.DB

type lastInsertIdAs int

const (
	LAST_INSERT_ID_AS_FIRST_ID lastInsertIdAs = iota + 1
	LAST_INSERT_ID_AS_LAST_ID
)

type rowAs int

const (
	ROW_AS_RETURNING rowAs = iota + 1
	ROW_AS_LAST_ID
)

type NewDaoReq struct {
	DB                *sql.DB
	AllowInvalidField bool
	ColumnMapper      *NameMapper
}

type QueryReq[T any] struct {
	Ctx         context.Context
	Must        bool
	SqlLogLevel SqlLogLevel
	Desc        string
	RowAs       rowAs
	Entities    []*T
	BuildSql    func(b *DaoSqlBuilder[T])
}

type ExecReq[T any] struct {
	Ctx            context.Context
	Must           bool
	SqlLogLevel    SqlLogLevel
	Desc           string
	LastInsertIdAs lastInsertIdAs
	Entities       []*T
	BuildSql       func(b *DaoSqlBuilder[T])
}

type baseDao struct {
	db *sql.DB
}

func (d baseDao) DB() *sql.DB {
	if d.db == nil { // coverage-ignore
		return DEFAULT_DB
	}
	return d.db
}

type Dao[T any] struct {
	baseDao
	commaColumns           string
	columns                []string
	columnToFieldIndexMap  map[string]int
	autoIncrementColumns   []string
	autoIncrementStep      int64
	autoIncrementConvertor func(id int64) reflect.Value
}

func (d *Dao[T]) Query(req QueryReq[T]) (first *T, list []*T, err error) {
	list = make([]*T, 0)
	b := newDaoSqlBuilder(d, req.Entities)
	req.BuildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		checkMust(req.Must, err)
		return nil, list, err
	}
	if !b.Ok() { // coverage-ignore
		return
	}
	rows, columns, closeFunc, err := query(req.Ctx, d.DB(), b.Sql(), b.args)
	if err != nil { // coverage-ignore
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.args, -1, -1, err)
		checkMust(req.Must, err)
		return nil, nil, err
	}
	defer closeFunc()

	switch req.RowAs {
	case ROW_AS_RETURNING:
		var affected int64
		for i := 0; rows.Next() && i < len(req.Entities); i++ {
			entity := req.Entities[i]
			if entity == nil { // coverage-ignore
				continue
			}
			v := reflect.ValueOf(entity).Elem()
			var fields []any
			for _, c := range columns {
				if fieldIndex, ok := d.columnToFieldIndexMap[c]; ok {
					field := v.Field(fieldIndex).Addr().Interface()
					fields = append(fields, field)
				}
			}
			if len(fields) > 0 {
				printWarn(req.Ctx, rows.Scan(fields...))
			}
			affected++
		}
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.args, affected, -1, nil)
	case ROW_AS_LAST_ID:
		var affected int64
		var id *int64
		if rows.Next() && len(columns) == 1 && len(d.autoIncrementColumns) == 1 {
			err = rows.Scan(&id)
			printWarn(req.Ctx, err)
			if err != nil && rows.Next() { // coverage-ignore
				id = nil
			}
		}
		if id != nil {
			fieldIndex := d.columnToFieldIndexMap[d.autoIncrementColumns[0]]
			entityLength := len(req.Entities)
			for i := 0; i < entityLength; i++ {
				entity := req.Entities[i]
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(d.autoIncrementConvertor(*id - int64(entityLength-1-i)*d.autoIncrementStep))
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
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.args, affected, -1, nil)
	default:
		var rowCounts int64
		for rows.Next() {
			entity := new(T)
			fields := d.mappingFields(entity, columns)
			err = rows.Scan(fields...)
			if err != nil {
				checkMust(req.Must, err)
				return
			}
			list = append(list, entity)
			rowCounts++
		}
		if len(list) > 0 {
			first = list[0]
		}
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.args, -1, rowCounts, nil)
	}
	return
}

func (d *Dao[T]) Exec(req ExecReq[T]) (affected int64, err error) {
	b := newDaoSqlBuilder(d, req.Entities)
	req.BuildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		checkMust(req.Must, err)
		return 0, err
	}
	if !b.Ok() { // coverage-ignore
		return 0, nil
	}
	result, affected, err := exec(req.Ctx, d.DB(), b.Sql(), b.args)
	printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.args, affected, -1, err)
	if err != nil { // coverage-ignore
		checkMust(req.Must, err)
		return
	}

	switch req.LastInsertIdAs {
	case LAST_INSERT_ID_AS_FIRST_ID:
		id, err := result.LastInsertId()
		printWarn(req.Ctx, err)
		if err == nil && len(req.Entities) > 0 && len(d.autoIncrementColumns) == 1 {
			fieldIndex := d.columnToFieldIndexMap[d.autoIncrementColumns[0]]
			for i, entity := range req.Entities {
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(d.autoIncrementConvertor(id + int64(i)*d.autoIncrementStep))
			}
		}
	case LAST_INSERT_ID_AS_LAST_ID:
		id, err := result.LastInsertId()
		printWarn(req.Ctx, err)
		if err == nil && len(req.Entities) > 0 && len(d.autoIncrementColumns) == 1 {
			fieldIndex := d.columnToFieldIndexMap[d.autoIncrementColumns[0]]
			entityLength := len(req.Entities)
			for i := 0; i < entityLength; i++ {
				entity := req.Entities[i]
				if entity == nil { // coverage-ignore
					continue
				}
				v := reflect.ValueOf(entity).Elem()
				field := v.Field(fieldIndex)
				field.Set(d.autoIncrementConvertor(id - int64(entityLength-1-i)*d.autoIncrementStep))
			}
		}
	}
	return
}

func query(ctx context.Context, db *sql.DB, sql string, args []any) (rows *sql.Rows, columns []string, closeFunc func(), err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	prepare, err := createPrepare(ctx, db, sql)
	if err != nil { // coverage-ignore
		return nil, nil, nil, err
	}
	rows, err = prepare.QueryContext(ctx, args...)
	if err != nil { // coverage-ignore
		printWarn(ctx, prepare.Close())
		return nil, nil, nil, err
	}
	closeFunc = func() {
		printWarn(ctx, rows.Close())
		printWarn(ctx, prepare.Close())
	}
	columns, err = rows.Columns()
	if err != nil { // coverage-ignore
		closeFunc()
		return nil, nil, nil, err
	}
	return rows, columns, closeFunc, nil
}

func exec(ctx context.Context, db *sql.DB, sql string, args []any) (result sql.Result, affected int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	affected = int64(-1)
	prepare, err := createPrepare(ctx, db, sql)
	if err != nil { // coverage-ignore
		return nil, 0, err
	}
	defer func() {
		printWarn(ctx, prepare.Close())
	}()
	result, err = prepare.ExecContext(ctx, args...)
	if err != nil { // coverage-ignore
		return nil, 0, err
	}
	affected, err = result.RowsAffected()
	return
}

func createPrepare(ctx context.Context, db *sql.DB, _sql string) (*sql.Stmt, error) {
	if tx := getTx(ctx); tx != nil {
		return tx.PrepareContext(ctx, _sql)
	} else {
		if db == nil { // coverage-ignore
			return nil, errors.New("no available *sql.DB variable")
		}
		return db.PrepareContext(ctx, _sql)
	}
}

func (d *Dao[T]) mappingFields(entity *T, columns []string) []any {
	v := reflect.ValueOf(entity).Elem()
	fields := make([]any, 0, len(columns))
	for _, c := range columns {
		if value, ok := d.columnToFieldIndexMap[c]; ok {
			field := v.Field(value)
			fields = append(fields, field.Addr().Interface())
		} else {
			fields = append(fields, new(any))
		}
	}
	return fields
}

func P[T any](t T) *T {
	return &t
}

func V[T any](t *T) T {
	var v T
	if t != nil {
		v = *t
	}
	return v
}

func checkMust(must bool, err error) { // coverage-ignore
	if must && err != nil {
		panic(err)
	}
}

func NewDao[T any](req NewDaoReq) *Dao[T] {
	dao := &Dao[T]{}

	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		panic("generics must be struct type")
	}

	dao.db = req.DB
	dao.columnToFieldIndexMap = make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			if req.AllowInvalidField {
				continue
			} else {
				panic("field \"" + tf.Name + "\" of \"" + t.String() + "\" must be exported")
			}
		}
		if !tf.Anonymous {
			ft := tf.Type
			if ft.Kind() == reflect.Pointer {
				if _, ok := supportedFieldTypes[ft.Elem().String()]; ok {
					registerField[T](dao, tf, req.ColumnMapper)
				} else {
					if req.AllowInvalidField {
						continue
					} else {
						panic("field \"" + tf.Name + "\" of \"" + t.String() + "\" is not supported type")
					}
				}
			} else if ft.Kind() == reflect.Slice {
				if _, ok := supportedFieldTypes[ft.Elem().String()]; ok {
					registerField[T](dao, tf, req.ColumnMapper)
				} else {
					if req.AllowInvalidField {
						continue
					} else {
						panic("field \"" + tf.Name + "\" of \"" + t.String() + "\", its element type is not supported")
					}
				}
			} else {
				if req.AllowInvalidField {
					continue
				} else {
					panic("field \"" + tf.Name + "\" of \"" + t.String() + "\" must be a pointer")
				}
			}
		}
	}
	return dao
}

func registerField[T any](d *Dao[T], tf reflect.StructField, columnMapper *NameMapper) {
	var column string
	var isAutoIncrement bool
	var autoIncrementStep int64 = 1
	if tag, ok := tf.Tag.Lookup("gdao"); ok {
		params := strings.Split(tag, ";")
		for _, p := range params {
			kv := strings.Split(p, "=")
			if len(kv) == 1 {
				p = strings.TrimSpace(p)
				if p == "auto" {
					isAutoIncrement = true
				}
			}
			if len(kv) == 2 {
				k := strings.TrimSpace(kv[0])
				v := strings.TrimSpace(kv[1])
				switch k {
				case "column":
					column = v
				case "auto":
					isAutoIncrement = true
					i, err := strconv.ParseInt(v, 10, 64)
					if err == nil { // coverage-ignore
						autoIncrementStep = i
					}
				}
			}
		}
	}

	if column == "" {
		if columnMapper != nil {
			column = columnMapper.Convert(tf.Name)
		} else { // coverage-ignore
			return
		}
	}
	d.columns = append(d.columns, column)
	if d.commaColumns != "" {
		d.commaColumns += ", "
	}
	d.commaColumns += column
	d.columnToFieldIndexMap[column] = tf.Index[0]
	if isAutoIncrement {
		if convertor, ok := lastInsertIdConvertors[tf.Type.Elem().String()]; ok {
			d.autoIncrementColumns = append(d.autoIncrementColumns, column)
			d.autoIncrementStep = autoIncrementStep
			d.autoIncrementConvertor = convertor
		}
	}
	return
}

var lastInsertIdConvertors = map[string]func(id int64) reflect.Value{
	"int":     func(id int64) reflect.Value { i := int(id); return reflect.ValueOf(&i) },
	"int8":    func(id int64) reflect.Value { i := int8(id); return reflect.ValueOf(&i) },
	"int16":   func(id int64) reflect.Value { i := int16(id); return reflect.ValueOf(&i) },
	"int32":   func(id int64) reflect.Value { i := int32(id); return reflect.ValueOf(&i) },
	"int64":   func(id int64) reflect.Value { return reflect.ValueOf(&id) },
	"uint":    func(id int64) reflect.Value { u := uint(id); return reflect.ValueOf(&u) },
	"uint8":   func(id int64) reflect.Value { u := uint8(id); return reflect.ValueOf(&u) },
	"uint16":  func(id int64) reflect.Value { u := uint16(id); return reflect.ValueOf(&u) },
	"uint32":  func(id int64) reflect.Value { u := uint32(id); return reflect.ValueOf(&u) },
	"uint64":  func(id int64) reflect.Value { u := uint64(id); return reflect.ValueOf(&u) },
	"float32": func(id int64) reflect.Value { f := float32(id); return reflect.ValueOf(&f) },
	"float64": func(id int64) reflect.Value { f := float64(id); return reflect.ValueOf(&f) },
	"string":  func(id int64) reflect.Value { s := strconv.FormatInt(id, 10); return reflect.ValueOf(&s) },
}

var supportedFieldTypes = map[string]struct{}{
	"bool": {}, "string": {}, "time.Time": {}, "float32": {}, "float64": {}, "int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
}
