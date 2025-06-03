package gdao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	pkgErrors "github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

var (
	DEFAULT_DB *sql.DB

	ctx_key_tx = Ptr("")
)

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
	Ctx      context.Context
	RowAs    rowAs
	Entities []*T
	BuildSql func(b *Builder[T])
}

type ExecReq[T any] struct {
	Ctx            context.Context
	LastInsertIdAs lastInsertIdAs
	Entities       []*T
	BuildSql       func(b *Builder[T])
}

type baseDao struct {
	db *sql.DB
}

func (d baseDao) DB() *sql.DB {
	if d.db == nil {
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
	b := newBuilder(d, req.Entities)
	req.BuildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		return nil, list, err
	}
	if !b.Ok() { // coverage-ignore
		return
	}
	rows, columns, closeFunc, err := query(req.Ctx, d.DB(), b.Sql(), b.args)
	if err != nil { // coverage-ignore
		return nil, nil, err
	}
	defer closeFunc()

	switch req.RowAs {
	case ROW_AS_RETURNING:
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
				printError(req.Ctx, rows.Scan(fields...))
			}
		}
	case ROW_AS_LAST_ID:
		var id *int64
		if rows.Next() && len(columns) == 1 && len(d.autoIncrementColumns) == 1 {
			err = rows.Scan(&id)
			printError(req.Ctx, err)
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
			}
		}
	default:
		for rows.Next() {
			entity := new(T)
			fields := d.mappingFields(entity, columns)
			err = rows.Scan(fields...)
			if err != nil {
				return
			}
			list = append(list, entity)
		}
		if len(list) > 0 {
			first = list[0]
		}
	}
	return
}

func (d *Dao[T]) Exec(req ExecReq[T]) (affected int64, err error) {
	b := newBuilder(d, req.Entities)
	req.BuildSql(b)
	err = b.Error()
	if err != nil { // coverage-ignore
		return 0, err
	}
	if !b.Ok() { // coverage-ignore
		return 0, nil
	}
	result, affected, err := exec(req.Ctx, d.DB(), b.Sql(), b.args)
	if err != nil { // coverage-ignore
		return
	}

	switch req.LastInsertIdAs {
	case LAST_INSERT_ID_AS_FIRST_ID:
		id, err := result.LastInsertId()
		printError(req.Ctx, err)
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
		printError(req.Ctx, err)
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
		printSql(ctx, sql, args, -1, err)
		return nil, nil, nil, err
	}
	rows, err = prepare.QueryContext(ctx, args...)
	if err != nil { // coverage-ignore
		printWarn(ctx, prepare.Close())
		printSql(ctx, sql, args, -1, err)
		return nil, nil, nil, err
	}
	closeFunc = func() {
		printWarn(ctx, rows.Close())
		printWarn(ctx, prepare.Close())
	}
	columns, err = rows.Columns()
	if err != nil { // coverage-ignore
		closeFunc()
		printSql(ctx, sql, args, -1, err)
		return nil, nil, nil, err
	}
	printSql(ctx, sql, args, -1, err)
	return rows, columns, closeFunc, nil
}

func exec(ctx context.Context, db *sql.DB, sql string, args []any) (result sql.Result, affected int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	affected = int64(-1)
	prepare, err := createPrepare(ctx, db, sql)
	if err != nil { // coverage-ignore
		printSql(ctx, sql, args, affected, err)
		return nil, 0, err
	}
	defer func() {
		printWarn(ctx, prepare.Close())
	}()
	result, err = prepare.ExecContext(ctx, args...)
	if err != nil { // coverage-ignore
		printSql(ctx, sql, args, affected, err)
		return nil, 0, err
	}
	affected, err = result.RowsAffected()
	printSql(ctx, sql, args, affected, err)
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
		}
	}
	return fields
}

func Ptr[T any](t T) *T {
	return &t
}

func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, ctx_key_tx, tx)
	return ctx
}

func Tx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, do func(ctx context.Context) error) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	tx := getTx(ctx)
	if tx == nil {
		if db == nil { // coverage-ignore
			db = DEFAULT_DB
		}
		if db == nil { // coverage-ignore
			return errors.New(`cannot begin a transaction, parameter "db" and gdao.DEFAULT_DB are nil `)
		}
		tx, err = DEFAULT_DB.BeginTx(ctx, opts)
		if err != nil { // coverage-ignore
			return
		}
		WithTx(ctx, tx)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else if r := recover(); r != nil {
			tx.Rollback()
			err = pkgErrors.WithStack(fmt.Errorf("%v", r))
		} else {
			tx.Commit()
		}
	}()
	err = do(ctx)
	return err
}

func getTx(ctx context.Context) *sql.Tx {
	if ctx != nil {
		if tx, ok := ctx.Value(ctx_key_tx).(*sql.Tx); ok {
			return tx
		}
	}
	return nil
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
				panic("field \"" + tf.Name + "\" is invalid")
			}
		}
		if !tf.Anonymous {
			if ft := tf.Type; ft.Kind() == reflect.Pointer {
				if _, ok := supportedFieldTypes[ft.Elem().String()]; ok {
					registerField[T](dao, tf, req.ColumnMapper)
				}
			} else {
				if req.AllowInvalidField {
					continue
				} else {
					panic("field \"" + tf.Name + "\" is invalid")
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
		} else {
			return
		}
	}
	d.columns = append(d.columns, column)
	if d.commaColumns != "" {
		d.commaColumns += ","
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
