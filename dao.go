package gdao

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Db is the default db
var Db *sql.DB

type NewDaoReq struct {
	Db           *sql.DB
	ColumnMapper *NameMapper
}

type RawQueryReq struct {
	Ctx  context.Context
	Tx   *sql.Tx
	Sql  string
	Args []any
}

type RawMutationReq struct {
	Ctx  context.Context
	Tx   *sql.Tx
	Sql  string
	Args []any
}

type QueryReq[T any] struct {
	Ctx      context.Context
	Tx       *sql.Tx
	Entities []*T
	BuildSql func(b *Builder[T])
}

type MutationReq[T any] struct {
	Ctx      context.Context
	Tx       *sql.Tx
	Entities []*T
	BuildSql func(b *Builder[T])
}

func Ptr[T any](t T) *T {
	return &t
}

func ExtendDao[T any](req NewDaoReq) (*Dao[T], *DaoProt[T]) {
	dao := NewDao[T](req)
	return dao, dao.p
}

func NewDao[T any](req NewDaoReq) *Dao[T] {
	dao := &Dao[T]{}
	dao.p = newDaoProtected(dao)

	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		panic("generics must be struct type")
	}

	dao.p.Db = req.Db
	dao.p.ColumnToFieldIndexMap = make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			continue
		}
		if !tf.Anonymous {
			if ft := tf.Type; ft.Kind() == reflect.Pointer {
				if _, ok := supportedFieldTypes[ft.Elem().String()]; ok {
					registerField[T](dao, tf, req.ColumnMapper)
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
	d.p.Columns = append(d.p.Columns, column)
	if d.p.CommaColumns != "" {
		d.p.CommaColumns += ","
	}
	d.p.CommaColumns += column
	d.p.ColumnToFieldIndexMap[column] = tf.Index[0]
	if isAutoIncrement {
		if convertor, ok := lastInsertIdConvertors[tf.Type.Elem().String()]; ok {
			d.p.AutoIncrementColumns = append(d.p.AutoIncrementColumns, column)
			d.p.AutoIncrementStep = autoIncrementStep
			d.p.AutoIncrementConvertor = convertor
		}
	}
	return
}

func newDaoProtected[T any](dao *Dao[T]) *DaoProt[T] {
	return &DaoProt[T]{dao: dao}
}

type DaoProt[T any] struct {
	dao                    *Dao[T]
	Db                     *sql.DB
	CommaColumns           string
	Columns                []string
	ColumnToFieldIndexMap  map[string]int
	AutoIncrementColumns   []string
	AutoIncrementStep      int64
	AutoIncrementConvertor func(id int64) reflect.Value
}

func (d DaoProt[T]) NewBuilder(entities ...*T) *Builder[T] {
	return newBuilder(d.dao, entities)
}

func (d DaoProt[T]) GetBuilderProt(b *Builder[T]) *BuilderProt {
	return b.p
}

type Dao[T any] struct {
	p *DaoProt[T]
}

func (d *Dao[T]) SetDb(db *sql.DB) {
	d.p.Db = db
}

func (d *Dao[T]) Db() *sql.DB {
	if d.p.Db == nil {
		return Db
	}
	return d.p.Db
}

func (d *Dao[T]) Query(req QueryReq[T]) (first *T, list []*T, err error) {
	list = make([]*T, 0)
	b := newBuilder(d, req.Entities)
	req.BuildSql(b)
	if !b.Ok() {
		printSqlLog(req.Ctx, "canceled SQL: %s", b.Sql())
		return
	}
	rows, columns, closeFunc, err := d.query(req.Ctx, req.Tx, b.p.String(), b.p.args)
	if err != nil { // coverage-ignore
		return nil, nil, err
	}
	defer closeFunc()
	for rows.Next() {
		entity := new(T)
		fields := d.mappingFields(entity, columns)
		err := rows.Scan(fields...)
		printError(req.Ctx, err)
		list = append(list, entity)
	}
	if len(list) > 0 {
		first = list[0]
	}
	return first, list, nil
}

func (d *Dao[T]) Mutation(req MutationReq[T]) *mutationDao[T] {
	return &mutationDao[T]{Dao: d, req: &req}
}

func (d *Dao[T]) query(ctx context.Context, tx *sql.Tx, sql string, args []any) (rows *sql.Rows, columns []string, closeFunc func(), err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	printSql(ctx, sql)
	printArgs(ctx, args)
	prepare, err := d.createPrepare(ctx, tx, sql)
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

func (d *Dao[T]) exec(ctx context.Context, tx *sql.Tx, sql string, args []any) (result sql.Result, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	printSql(ctx, sql)
	printArgs(ctx, args)
	prepare, err := d.createPrepare(ctx, tx, sql)
	if err != nil { // coverage-ignore
		return nil, err
	}
	defer func() {
		printWarn(ctx, prepare.Close())
	}()
	result, err = prepare.ExecContext(ctx, args...)
	if err != nil {
		affected, err := result.RowsAffected()
		if err != nil {
			printAffected(ctx, affected)
		}
	}
	return
}

func (d *Dao[T]) createPrepare(ctx context.Context, tx *sql.Tx, _sql string) (*sql.Stmt, error) {
	var prepare *sql.Stmt
	var err error
	if tx == nil {
		db := d.Db()
		if db == nil { // coverage-ignore
			return nil, errors.New("the db has not been set")
		}
		prepare, err = db.PrepareContext(ctx, _sql)
		if err != nil { // coverage-ignore
			return nil, err
		}
	} else {
		prepare, err = tx.PrepareContext(ctx, _sql)
		if err != nil { // coverage-ignore
			return nil, err
		}
	}
	return prepare, nil
}

func (d *Dao[T]) mappingFields(entity *T, columns []string) []any {
	v := reflect.ValueOf(entity).Elem()
	fields := make([]any, 0, len(columns))
	for _, c := range columns {
		if value, ok := d.p.ColumnToFieldIndexMap[c]; ok {
			field := v.Field(value)
			fields = append(fields, field.Addr().Interface())
		}
	}
	return fields
}

type mutationDao[T any] struct {
	*Dao[T]
	req *MutationReq[T]
}

func (d *mutationDao[T]) Exec() (affected int64, err error) {
	b := newBuilder(d.Dao, d.req.Entities)
	d.req.BuildSql(b)
	if !b.Ok() { // coverage-ignore
		printSqlLog(d.req.Ctx, "canceled SQL: %s", b.Sql())
		return 0, nil
	}
	result, err := d.exec(d.req.Ctx, d.req.Tx, b.p.String(), b.p.args)
	if err != nil { // coverage-ignore
		return 0, err
	}
	affected, err = result.RowsAffected()
	printError(d.req.Ctx, err)
	return
}

func (d *mutationDao[T]) Insert() (affected int64, err error) {
	b := newBuilder(d.Dao, d.req.Entities)
	d.req.BuildSql(b)
	if !b.Ok() { // coverage-ignore
		printSqlLog(d.req.Ctx, "canceled SQL: %s", b.Sql())
		return 0, nil
	}
	result, err := d.exec(d.req.Ctx, d.req.Tx, b.p.String(), b.p.args)
	if err != nil { // coverage-ignore
		return 0, err
	}
	affected, err = result.RowsAffected()
	printError(d.req.Ctx, err)
	id, err := result.LastInsertId()
	printError(d.req.Ctx, err)
	if err == nil && len(d.req.Entities) > 0 && len(d.p.AutoIncrementColumns) == 1 {
		fieldIndex := d.p.ColumnToFieldIndexMap[d.p.AutoIncrementColumns[0]]
		for i, entity := range d.req.Entities {
			if entity == nil { // coverage-ignore
				continue
			}
			v := reflect.ValueOf(entity).Elem()
			vf := v.Field(fieldIndex)
			vf.Set(d.p.AutoIncrementConvertor(id + int64(i)*d.p.AutoIncrementStep))
		}
	}
	return
}

func (d *mutationDao[T]) Query() (affected int64, err error) {
	b := newBuilder(d.Dao, d.req.Entities)
	d.req.BuildSql(b)
	if !b.Ok() { // coverage-ignore
		printSqlLog(d.req.Ctx, "canceled SQL: %s", b.Sql())
		return 0, nil
	}
	rows, queriedColumns, closeFunc, err := d.query(d.req.Ctx, d.req.Tx, b.p.String(), b.p.args)
	if err != nil { // coverage-ignore
		return 0, err
	}
	defer closeFunc()

	for rows.Next() {
		if len(queriedColumns) > 0 && affected < int64(len(d.req.Entities)) {
			entity := d.req.Entities[affected]
			v := reflect.ValueOf(entity).Elem()

			var queriedFields []any
			for _, c := range queriedColumns {
				if fieldIndex, ok := d.p.ColumnToFieldIndexMap[c]; ok {
					field := v.Field(fieldIndex).Addr().Interface()
					queriedFields = append(queriedFields, field)
				}
			}

			if len(queriedFields) > 0 {
				printError(d.req.Ctx, rows.Scan(queriedFields...))
			}
		}
		affected++
	}
	return affected, nil
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
