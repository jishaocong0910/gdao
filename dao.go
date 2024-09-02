package gdao

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"

	o "github.com/jishaocong0910/go-object"
)

type NewDaoReq struct {
	Db                  *sql.DB
	Table               string
	ColumnMapper        *NameMapper
	ColumnCaseSensitive bool
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
	BuildSql func(b Builder[T]) (sql string, args []any)
}

type MutationReq[T any] struct {
	Ctx      context.Context
	Tx       *sql.Tx
	Entities []*T
	BuildSql func(b Builder[T]) (sql string, args []any)
}

func Ptr[T any](t T) *T {
	return &t
}

func MustNewDao[T any](req NewDaoReq) Dao[T] {
	dao, err := NewDao[T](req)
	if err != nil { // coverage-ignore
		panic(err)
	}
	return dao
}

func NewDao[T any](req NewDaoReq) (Dao[T], error) {
	dao := Dao[T]{}
	if req.Db == nil {
		return dao, errors.New(`db must not be nil`)
	}
	if req.Table == "" {
		return dao, errors.New(`table must not be blank`)
	}
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return dao, errors.New("generics must be struct")
	}

	dao.db = req.Db
	dao.table = req.Table
	dao.columnToFieldIndexMap = o.NewStrKeyMap[int](req.ColumnCaseSensitive)
	dao.columnCaseSensitive = req.ColumnCaseSensitive

	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if !tf.IsExported() {
			continue
		}
		if !tf.Anonymous {
			if ft := tf.Type; ft.Kind() == reflect.Pointer {
				if SupportedFieldTypes.ContainsName(ft.Elem().String()) {
					registerField[T](&dao, tf, req.ColumnMapper)
				}
			}
		}
	}
	return dao, nil
}

func registerField[T any](d *Dao[T], tf reflect.StructField, columnMapper *NameMapper) {
	var column string
	var isAutoIncrement bool
	var autoIncrementOffset int64 = 1
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
						autoIncrementOffset = i
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
	if d.columnsWithComma != "" {
		d.columnsWithComma += ","
	}
	d.columnsWithComma += column
	d.columnToFieldIndexMap.Put(column, tf.Index[0])
	if isAutoIncrement {
		if convertor, ok := lastInsertIdConvertors[tf.Type.Elem().String()]; ok {
			d.autoIncrementColumn = column
			d.autoIncrementOffset = autoIncrementOffset
			d.autoIncrementConvertor = convertor
		}
	}
	return
}

type Dao[T any] struct {
	db                     *sql.DB
	table                  string
	columnsWithComma       string
	columns                []string
	columnToFieldIndexMap  *o.StrKeyMap[int]
	columnCaseSensitive    bool
	autoIncrementColumn    string
	autoIncrementOffset    int64
	autoIncrementConvertor func(id int64) reflect.Value
}

func (d Dao[T]) Db() *sql.DB {
	return d.db
}

func (d Dao[T]) RawQuery(req RawQueryReq) (rows *sql.Rows, closeFunc func(), err error) {
	rows, _, closeFunc, err = d.query(req.Ctx, req.Tx, req.Sql, req.Args)
	return
}

func (d Dao[T]) RawMutation(req RawMutationReq) (result sql.Result, err error) {
	return d.exec(req.Ctx, req.Tx, req.Sql, req.Args)
}

func (d Dao[T]) Query(req QueryReq[T]) (first *T, list []*T, err error) {
	sql, args := req.BuildSql(newBuilder(&d, req.Entities))
	rows, columns, closeFunc, err := d.query(req.Ctx, req.Tx, sql, args)
	if err != nil { // coverage-ignore
		return nil, nil, err
	}
	defer closeFunc()
	list = make([]*T, 0)
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

func (d Dao[T]) Mutation(req MutationReq[T]) *mutationDao[T] {
	return &mutationDao[T]{Dao: &d, req: &req}
}

func (d Dao[T]) query(ctx context.Context, tx *sql.Tx, sql string, args []any) (rows *sql.Rows, columns []string, closeFunc func(), err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	prepare, err := d.createPrepare(ctx, tx, sql)
	if err != nil { // coverage-ignore
		return nil, nil, nil, err
	}
	printSql(ctx, sql, args)
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

func (d Dao[T]) exec(ctx context.Context, tx *sql.Tx, sql string, args []any) (result sql.Result, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	prepare, err := d.createPrepare(ctx, tx, sql)
	if err != nil { // coverage-ignore
		return nil, err
	}
	defer func() {
		printWarn(ctx, prepare.Close())
	}()
	printSql(ctx, sql, args)
	return prepare.ExecContext(ctx, args...)
}

func (d Dao[T]) createPrepare(ctx context.Context, tx *sql.Tx, _sql string) (*sql.Stmt, error) {
	var prepare *sql.Stmt
	var err error
	if tx == nil {
		prepare, err = d.db.PrepareContext(ctx, _sql)
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

func (d Dao[T]) mappingFields(entity *T, columns []string) []any {
	v := reflect.ValueOf(entity).Elem()
	fields := make([]any, 0, len(columns))
	for _, c := range columns {
		if entry := d.columnToFieldIndexMap.GetEntry(c); entry != nil {
			field := v.Field(entry.Value)
			fields = append(fields, field.Addr().Interface())
		}
	}
	return fields
}

type mutationDao[T any] struct {
	*Dao[T]
	req *MutationReq[T]
}

func (m *mutationDao[T]) Exec() (affected int64, err error) {
	sql, args := m.req.BuildSql(newBuilder(m.Dao, m.req.Entities))
	result, err := m.exec(m.req.Ctx, m.req.Tx, sql, args)
	if err != nil { // coverage-ignore
		return 0, err
	}
	affected, err = result.RowsAffected()
	printError(m.req.Ctx, err)
	return
}

func (m *mutationDao[T]) Insert() (affected int64, err error) {
	sql, args := m.req.BuildSql(newBuilder(m.Dao, m.req.Entities))
	result, err := m.exec(m.req.Ctx, m.req.Tx, sql, args)
	if err != nil { // coverage-ignore
		return 0, err
	}
	affected, err = result.RowsAffected()
	printError(m.req.Ctx, err)
	id, err := result.LastInsertId()
	printError(m.req.Ctx, err)
	if err == nil && len(m.req.Entities) > 0 && m.autoIncrementColumn != "" {
		fieldIndex := m.columnToFieldIndexMap.Get(m.autoIncrementColumn)
		for i, entity := range m.req.Entities {
			v := reflect.ValueOf(entity).Elem()
			vf := v.Field(fieldIndex)
			vf.Set(m.autoIncrementConvertor(id + int64(i)*m.autoIncrementOffset))
		}
	}
	return
}

func (m *mutationDao[T]) Query() (affected int64, err error) {
	sql, args := m.req.BuildSql(newBuilder(m.Dao, m.req.Entities))
	rows, queriedColumns, closeFunc, err := m.query(m.req.Ctx, m.req.Tx, sql, args)
	if err != nil { // coverage-ignore
		return 0, err
	}
	defer closeFunc()

	for rows.Next() {
		if len(queriedColumns) > 0 && affected < int64(len(m.req.Entities)) {
			entity := m.req.Entities[affected]
			v := reflect.ValueOf(entity).Elem()

			var queriedFields []any
			for _, c := range queriedColumns {
				if fieldIndex := m.columnToFieldIndexMap.GetEntry(c); fieldIndex != nil {
					field := v.Field(fieldIndex.Value).Addr().Interface()
					queriedFields = append(queriedFields, field)
				}
			}

			if len(queriedFields) > 0 {
				printError(m.req.Ctx, rows.Scan(queriedFields...))
			}
		}
		affected++
	}
	return affected, nil
}

func newBuilder[T any](d *Dao[T], entities []*T) Builder[T] {
	return Builder[T]{dao: d, entities: entities, sql: &strings.Builder{}}
}

type Builder[T any] struct {
	dao      *Dao[T]
	entities []*T
	sql      *strings.Builder
	argNum   int
	args     []any
}

func (b *Builder[T]) Table() string {
	return b.dao.table
}

func (b *Builder[T]) Columns() string {
	return b.dao.columnsWithComma
}

func (b *Builder[T]) Write(str string) *Builder[T] {
	b.sql.WriteString(str)
	return b
}

func (b *Builder[T]) Separate(start, separator, end string) *separate {
	return &separate{start: start, separator: separator, end: end}
}

func (b *Builder[T]) AddArgs(args ...any) *Builder[T] {
	for _, a := range args {
		b.args = append(b.args, a)
	}
	return b
}

func (b *Builder[T]) ArgN(prefix string) string {
	b.argNum++
	return prefix + strconv.FormatInt(int64(b.argNum), 10)
}

func (b *Builder[T]) Args() []any {
	return b.args
}

func (b *Builder[T]) String() string {
	return b.sql.String()
}

func (b *Builder[T]) Entity() *T {
	return b.EntityAt(0)
}

func (b *Builder[T]) EntityAt(i int) *T {
	var t *T
	if i < len(b.entities) {
		t = b.entities[i]
	}
	return t
}

func (b *Builder[T]) EachEntity(separate *separate, handle func(i int)) {
	for i := range b.entities {
		if i != 0 && separate != nil {
			b.writeSeparator(separate)
		}
		b.writeStart(separate)
		handle(i)
		b.writeEnd(separate)
	}
}

func (b *Builder[T]) EachColumn(separate *separate, handle func(i int, column string, value any)) {
	b.EachColumnAt(0, separate, handle)
}

func (b *Builder[T]) EachColumnAt(i int, separate *separate, handle func(i int, column string, value any)) {
	b.iterateColumnAt(i, separate, func(i int, column string, value any) bool {
		return true
	}, handle)
}

func (b *Builder[T]) EachAssignedColumn(separate *separate, handle func(i int, column string, value any)) {
	b.iterateColumnAt(0, separate, func(i int, column string, value any) bool {
		if value == nil {
			return false
		}
		return true
	}, handle)
}

func (b *Builder[T]) iterateColumnAt(i int, separate *separate, canHandle func(i int, column string, value any) bool, handle func(i int, column string, value any)) {
	if !(i < len(b.entities)) || handle == nil { // coverage-ignore
		return
	}

	entity := b.entities[i]
	v := reflect.ValueOf(entity).Elem()

	var columnIndex int
	b.writeStart(separate)
	for _, column := range b.dao.columns {
		fieldIndex := b.dao.columnToFieldIndexMap.Get(column)
		field := v.Field(fieldIndex)
		var value any
		if !field.IsNil() {
			value = field.Interface()
		}

		if canHandle(columnIndex, column, value) {
			if columnIndex != 0 && separate != nil {
				b.writeSeparator(separate)
			}
			handle(columnIndex, column, value)
			columnIndex++
		}
	}
	b.writeEnd(separate)
}
func (b *Builder[T]) writeStart(s *separate) {
	if s != nil {
		b.Write(s.start)
	}
}

func (b *Builder[T]) writeSeparator(s *separate) {
	if s != nil {
		b.Write(s.separator)
	}
}

func (b *Builder[T]) writeEnd(s *separate) {
	if s != nil {
		b.Write(s.end)
	}
}

type separate struct {
	start, separator, end string
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
