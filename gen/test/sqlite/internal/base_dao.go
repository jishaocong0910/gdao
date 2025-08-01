// Code generated by https://github.com/jishaocong0910/gdao. FOR SQLite. DO NOT EDIT.

package dao

import (
	"context"
	"strconv"
	"strings"

	"github.com/jishaocong0910/gdao"
)

func getConditionBuilder[T any](b *gdao.Builder[T]) conditionBuilder {
	return conditionBuilder{
		write: func(str string, args ...any) {
			b.Write(str, args...)
		},
	}
}

func parenthesizeGroup(c condition) {
	if cg, ok := c.(*conditionGroup); ok {
		if len(cg.cs) > 1 && cg.or {
			cg.parenthesized = true
		}
	}
}

type conditionBuilder struct {
	write func(str string, args ...any)
}

type condition interface {
	write(b conditionBuilder)
	empty() bool
}

type baseCondition struct {
	not           bool
	parenthesized bool
}

func (bc baseCondition) empty() bool {
	return false
}

func (bc baseCondition) doWrite(b conditionBuilder, write func()) {
	if bc.not {
		b.write("NOT ")
	}
	if bc.parenthesized {
		b.write("(")
	}
	write()
	if bc.parenthesized {
		b.write(")")
	}
}

type conditionGroup struct {
	baseCondition
	or bool
	cs []condition
}

func (cg *conditionGroup) empty() bool {
	return cg == nil || len(cg.cs) == 0
}

func (cg *conditionGroup) write(b conditionBuilder) {
	cg.doWrite(b, func() {
		for i, cond := range cg.cs {
			if i != 0 {
				if cg.or {
					b.write(" OR ")
				} else {
					b.write(" AND ")
				}
			}
			cond.write(b)
		}
	})
}

func (cg *conditionGroup) addCondition(other condition) *conditionGroup {
	if other != nil && !other.empty() {
		if len(cg.cs) > 0 {
			if cg.not == true {
				cg.parenthesized = true
			}
			parenthesizeGroup(other)
			if len(cg.cs) == 1 {
				parenthesizeGroup(cg.cs[0])
			}
		}
		cg.cs = append(cg.cs, other)
	}
	return cg
}

func (cg *conditionGroup) ToStrArgs() (string, []any) {
	var str strings.Builder
	var args []any
	cb := conditionBuilder{
		write: func(s string, a ...any) {
			str.WriteString(s)
			args = append(args, a...)
		},
	}
	cg.write(cb)
	return str.String(), args
}

func (cg *conditionGroup) Group(other *conditionGroup) *conditionGroup {
	return cg.addCondition(other)
}

func (cg *conditionGroup) Plain(sql string, args ...any) *conditionGroup {
	return cg.addCondition(&conditionPlain{sql: sql, args: args})
}

func (cg *conditionGroup) Eq(column string, arg any) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "=", arg: arg})
}

func (cg *conditionGroup) NotEq(column string, arg any) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "<>", arg: arg})
}

func (cg *conditionGroup) Gt(column string, arg any) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: ">", arg: arg})
}

func (cg *conditionGroup) Lt(column string, arg any) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "<", arg: arg})
}

func (cg *conditionGroup) GtEq(column string, arg any) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: ">=", arg: arg})
}

func (cg *conditionGroup) LtEq(column string, arg any) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "<=", arg: arg})
}

func (cg *conditionGroup) Like(column string, arg string) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "LIKE", arg: "%" + arg + "%"})
}

func (cg *conditionGroup) LikeLeft(column string, arg string) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "LIKE", arg: arg + "%"})
}

func (cg *conditionGroup) LikeRight(column string, arg string) *conditionGroup {
	return cg.addCondition(&conditionBinOp{column: column, op: "LIKE", arg: "%" + arg})
}

// you can use Anys to convert non-any type slice to an any type slice.
func (cg *conditionGroup) In(column string, args []any) *conditionGroup {
	return cg.addCondition(&conditionIn{column: column, args: args})
}

func (cg *conditionGroup) Between(column string, min, max any) *conditionGroup {
	return cg.addCondition(&conditionBetween{column: column, min: min, max: max})
}

func (cg *conditionGroup) IsNull(column string) *conditionGroup {
	return cg.addCondition(&conditionIsNull{column: column})
}

func (cg *conditionGroup) IsNotNull(column string) *conditionGroup {
	return cg.addCondition(&conditionIsNull{column: column, notNull: true})
}

type conditionPlain struct {
	baseCondition
	sql  string
	args []any
}

func (c *conditionPlain) write(b conditionBuilder) {
	c.doWrite(b, func() {
		b.write(c.sql, c.args...)
	})
}

type conditionBinOp struct {
	baseCondition
	column string
	op     string
	arg    any
}

func (c *conditionBinOp) write(b conditionBuilder) {
	c.doWrite(b, func() {
		b.write(c.column)
		b.write(" ")
		b.write(c.op)
		b.write(" ")
		b.write("?", c.arg)
	})
}

type conditionIn struct {
	baseCondition
	column string
	args   []any
}

func (c *conditionIn) write(b conditionBuilder) {
	c.doWrite(b, func() {
		b.write(c.column)
		b.write(" IN(")
		for i := 0; i < len(c.args); i++ {
			if i != 0 {
				b.write(", ")
			}
			b.write("?")
		}
		b.write(")", c.args...)
	})
}

type conditionBetween struct {
	baseCondition
	column   string
	min, max any
}

func (c *conditionBetween) write(b conditionBuilder) {
	c.doWrite(b, func() {
		b.write(c.column)
		b.write(" BETWEEN ? AND ?", c.min, c.max)
	})
}

type conditionIsNull struct {
	baseCondition
	notNull bool
	column  string
}

func (c *conditionIsNull) write(b conditionBuilder) {
	c.doWrite(b, func() {
		b.write(c.column)
		b.write(" IS")
		if c.notNull {
			b.write(" NOT")
		}
		b.write(" NULL")
	})
}

type ListReq struct {
	Ctx context.Context
	// specify the columns which in the select column list, default is all columns.
	SelectColumns []string
	// conditions of the WHERE clause，create by function And, Or, NotAnd and NotOr.
	Condition condition
	// ORDER BY clause，create by function OrderBy.
	OrderBy *orderBy
	// paging query，create by function Page.
	Pagination *pagination
	// FOR UPDATE clause
	ForUpdate bool
}

type GetReq struct {
	Ctx context.Context
	// specify the columns which in the select column list, default is all columns.
	SelectColumns []string
	// conditions of the WHERE clause，create by function And, Or, NotAnd and NotOr.
	Condition condition
	// FOR UPDATE clause
	ForUpdate bool
}

type InsertReq[T any] struct {
	Ctx context.Context
	// the non-nil fields will be saved, and the auto generated keys will be set in it.
	Entity *T
	// if true, all fields will be saved, otherwise, save non-nil fields.
	InsertAll bool
	// specify the columns which set a null value
	SetNullColumns []string
	// specify the columns which be not set
	IgnoredColumns []string
}

type InsertBatchReq[T any] struct {
	Ctx context.Context
	// each element corresponds to a record to be saved, and the auto generated keys will be set in them.
	Entities []*T
	// if true, all fields will be saved, otherwise, save non-nil fields.
	InsertAll bool
	// specify the columns which set a null value
	SetNullColumns []string
	// specify the columns which be not set
	IgnoredColumns []string
}

type UpdateReq[T any] struct {
	Ctx context.Context
	// uses to update values or as the WHERE clause conditions.
	Entity *T
	// if true, all fields will be updated, otherwise, update non-nil fields.
	UpdateAll bool
	// specify the columns which set a null value
	SetNullColumns []string
	// specify the columns which be not set
	IgnoredColumns []string
	// specify the non-nil fields in the entity used as conditions.
	WhereColumns []string
	// conditions of the WHERE clause，create by function And, Or, NotAnd and NotOr..
	Condition condition
}

type UpdateBatchReq[T any] struct {
	Ctx context.Context
	// each element corresponds to a record to be updated.
	Entities []*T
	// if true, all fields will be updated, otherwise, update non-nil fields.
	UpdateAll bool
	// specify the columns which set a null value
	SetNullColumns []string
	// specify the columns which be not set
	IgnoredColumns []string
	// specify the column which used as a condition.
	WhereColumn string
	// conditions of the WHERE clause，create by function And, Or, NotAnd and NotOr..
	Condition condition
}

type DeleteReq struct {
	Ctx context.Context
	// conditions of the WHERE clause，create by function And, Or, NotAnd and NotOr.
	Condition condition
}

type orderBy struct {
	items []orderByItem
}

func (o *orderBy) Asc(column string) *orderBy {
	o.items = append(o.items, orderByItem{column: column, sequence: asc})
	return o
}

func (o *orderBy) Desc(column string) *orderBy {
	o.items = append(o.items, orderByItem{column: column, sequence: desc})
	return o
}

type orderBySequence string

const (
	asc  orderBySequence = "ASC"
	desc                 = "DESC"
)

type orderByItem struct {
	column   string
	sequence orderBySequence
}

type pagination struct {
	offset, pageSize int
}

type baseDao[T any] struct {
	*gdao.Dao[T]
	table string
}

// List queries records of the conditions.
func (d baseDao[T]) List(req ListReq) ([]*T, error) {
	_, list, err := d.Query(gdao.QueryReq[T]{Ctx: req.Ctx, BuildSql: func(b *gdao.Builder[T]) {
		b.Write("SELECT ").WriteColumns(req.SelectColumns...).Write(" FROM ").Write(d.table)
		if req.Condition != nil && !req.Condition.empty() {
			b.Write(" WHERE ")
			cb := getConditionBuilder(b)
			req.Condition.write(cb)
		}
		if req.OrderBy != nil {
			b.Repeat(len(req.OrderBy.items), b.SepFix(" ORDER BY ", ", ", "", false), nil, func(_, i int) {
				item := req.OrderBy.items[i]
				b.Write(item.column).Write(" ")
				b.Write(string(item.sequence))
			})
		}
		if req.Pagination != nil {
			b.Write(" LIMIT ")
			b.Write(strconv.FormatInt(int64(req.Pagination.pageSize), 10))
			b.Write(" OFFSET ")
			b.Write(strconv.FormatInt(int64(req.Pagination.offset), 10))
		}
		if req.ForUpdate {
			b.Write(" FOR UPDATE")
		}
	}})
	return list, err
}

// Get queries a record of the conditions.
func (d baseDao[T]) Get(req GetReq) (*T, error) {
	first, _, err := d.Query(gdao.QueryReq[T]{Ctx: req.Ctx, BuildSql: func(b *gdao.Builder[T]) {
		b.Write("SELECT ").WriteColumns(req.SelectColumns...).Write(" FROM ").Write(d.table)
		if req.Condition != nil && !req.Condition.empty() {
			b.Write(" WHERE ")
			cb := getConditionBuilder(b)
			req.Condition.write(cb)
		}
		b.Write(" LIMIT 1")
		if req.ForUpdate {
			b.Write(" FOR UPDATE")
		}
	}})
	return first, err
}

// Insert saves a record and return the auto generated keys.
func (d baseDao[T]) Insert(req InsertReq[T]) (int64, error) {
	return d.InsertBatch(InsertBatchReq[T]{
		Ctx:            req.Ctx,
		Entities:       []*T{req.Entity},
		InsertAll:      req.InsertAll,
		SetNullColumns: req.SetNullColumns,
		IgnoredColumns: req.IgnoredColumns,
	})
}

// InsertBatch saves records and return the auto generated keys.
func (d baseDao[T]) InsertBatch(req InsertBatchReq[T]) (int64, error) {
	return d.Exec(gdao.ExecReq[T]{Ctx: req.Ctx, LastInsertIdAs: gdao.LAST_INSERT_ID_AS_LAST_ID, Entities: req.Entities,
		BuildSql: func(b *gdao.Builder[T]) {
			var setColumnNum, setNullColumnNum int
			var allIgnoredColumns []string
			allIgnoredColumns = append(allIgnoredColumns, req.SetNullColumns...)
			allIgnoredColumns = append(allIgnoredColumns, req.IgnoredColumns...)
			allIgnoredColumns = append(allIgnoredColumns, b.AutoColumns()...)

			b.Write("INSERT INTO ").Write(d.table)

			columns := b.Columns(!req.InsertAll, allIgnoredColumns...)
			b.Repeat(len(columns), b.SepFix("(", ", ", "", true), nil, func(_, i int) {
				setColumnNum++
				b.Write(columns[i])
			})
			if len(req.SetNullColumns) > 0 {
				if setColumnNum > 0 {
					b.Write(", ")
				}
				b.Repeat(len(req.SetNullColumns), b.Sep(", "), nil, func(_, i int) {
					setNullColumnNum++
					b.Write(req.SetNullColumns[i])
				})
			}
			b.Write(")")

			b.Write(" VALUES")
			b.EachEntity(b.Sep(", "), nil, func(_ int, entity *T) {
				b.EachColumn(entity, b.SepFix("(", ", ", "", true), nil, func(_ int, column string, value any) {
					if value != nil {
						b.Write("?", value)
					} else {
						b.Write("NULL")
					}
				}, columns...)
				if len(req.SetNullColumns) > 0 {
					if setColumnNum > 0 {
						b.Write(", ")
					}
					b.Repeat(len(req.SetNullColumns), b.Sep(", "), nil, func(_, i int) {
						b.Write("NULL")
					})
				}
				b.Write(")")
			})
		}})
}

// Update modifies a record.
func (d baseDao[T]) Update(req UpdateReq[T]) (int64, error) {
	return d.Exec(gdao.ExecReq[T]{Ctx: req.Ctx, Entities: []*T{req.Entity},
		BuildSql: func(b *gdao.Builder[T]) {
			var setColumnNum, setNullColumnNum int
			var allIgnoredColumns []string
			allIgnoredColumns = append(allIgnoredColumns, req.SetNullColumns...)
			allIgnoredColumns = append(allIgnoredColumns, req.IgnoredColumns...)
			allIgnoredColumns = append(allIgnoredColumns, req.WhereColumns...)

			b.Write("UPDATE ").Write(d.table).Write(" SET ")

			columns := b.Columns(!req.UpdateAll, allIgnoredColumns...)
			b.EachColumn(b.Entity(), b.SepFix("", ", ", "", true), nil, func(_ int, column string, value any) {
				setColumnNum++
				b.Write(column).Write(" = ")
				if value != nil {
					b.Write("?").SetArgs(value)
				} else {
					b.Write("NULL")
				}
			}, columns...)
			if len(req.SetNullColumns) > 0 {
				if setColumnNum > 0 {
					b.Write(", ")
				}
				b.Repeat(len(req.SetNullColumns), b.Sep(", "), nil, func(_, i int) {
					setNullColumnNum++
					b.Write(req.SetNullColumns[i]).Write(" = NULL")
				})
			}

			cond := And()
			if len(req.WhereColumns) > 0 {
				b.EachColumn(b.Entity(), nil, nil, func(_ int, column string, value any) {
					if value == nil {
						cond.IsNull(column)
					} else {
						cond.Eq(column, value)
					}
				}, req.WhereColumns...)
			}
			cond.addCondition(req.Condition)
			if !cond.empty() {
				b.Write(" WHERE ")
				cb := getConditionBuilder(b)
				cond.write(cb)
			}
		}})
}

// UpdateBatch modifies multiple records by a SQL.
func (d baseDao[T]) UpdateBatch(req UpdateBatchReq[T]) (int64, error) {
	return d.Exec(gdao.ExecReq[T]{Ctx: req.Ctx, Entities: req.Entities, BuildSql: func(b *gdao.Builder[T]) {
		var setColumnNum, setNullColumnNum int
		var allIgnoredColumns []string
		allIgnoredColumns = append(allIgnoredColumns, req.SetNullColumns...)
		allIgnoredColumns = append(allIgnoredColumns, req.IgnoredColumns...)
		allIgnoredColumns = append(allIgnoredColumns, req.WhereColumn)

		b.Write("UPDATE ").Write(d.table).Write(" SET ")

		columns := b.Columns(!req.UpdateAll, allIgnoredColumns...)
		b.EachColumn(b.Entity(), b.SepFix("", ", ", "", true), nil, func(_ int, column string, value any) {
			setColumnNum++
			b.Write(column).Write(" = CASE ").Write(req.WhereColumn)
			b.EachEntity(nil, nil, func(_ int, entity *T) {
				b.Write(" WHEN ").Write("?").SetArgs(b.ColumnValue(entity, req.WhereColumn)).Write(" THEN ")
				if value != nil {
					b.Write("?").SetArgs(b.ColumnValue(entity, column))
				} else {
					b.Write("NULL")
				}
			})
			b.Write(" END")
		}, columns...)
		if len(req.SetNullColumns) > 0 {
			if setColumnNum > 0 {
				b.Write(", ")
			}
			b.Repeat(len(req.SetNullColumns), b.Sep(", "), nil, func(_, i int) {
				setNullColumnNum++
				b.Write(req.SetNullColumns[i]).Write(" = NULL")
			})
		}

		b.Write(" WHERE ")
		cond := And()
		whereColumnValues := make([]any, 0, len(req.Entities))
		b.EachEntity(nil, nil, func(_ int, entity *T) {
			whereColumnValues = append(whereColumnValues, b.ColumnValue(entity, req.WhereColumn))
		})
		cond.In(req.WhereColumn, Anys(whereColumnValues...))
		cond.addCondition(req.Condition)
		cb := getConditionBuilder(b)
		cond.write(cb)
	}})
}

// Delete removes records.
func (d baseDao[T]) Delete(req DeleteReq) (int64, error) {
	return d.Exec(gdao.ExecReq[T]{Ctx: req.Ctx, BuildSql: func(b *gdao.Builder[T]) {
		b.Write("DELETE FROM ").Write(d.table)
		if req.Condition != nil && !req.Condition.empty() {
			b.Write(" WHERE ")
			cb := getConditionBuilder(b)
			req.Condition.write(cb)
		}
	}})
}

func newBaseDao[T any](req gdao.NewDaoReq, table string) *baseDao[T] {
	dao := gdao.NewDao[T](req)
	table = strings.TrimSpace(table)
	if table == "" {
		panic(`parameter "table" must not be blank`)
	}
	return &baseDao[T]{Dao: dao, table: table}
}

func OrderBy() *orderBy {
	return &orderBy{}
}

func Page(offset, pageSize int) *pagination {
	return &pagination{offset: offset, pageSize: pageSize}
}

func And() *conditionGroup {
	return &conditionGroup{or: false}
}

func Or() *conditionGroup {
	return &conditionGroup{or: true}
}

func NotAnd() *conditionGroup {
	return &conditionGroup{baseCondition: baseCondition{not: true}, or: false}
}

func NotOr() *conditionGroup {
	return &conditionGroup{baseCondition: baseCondition{not: true}, or: true}
}

func Anys[T any](source ...T) (target []any) {
	for _, s := range source {
		target = append(target, s)
	}
	return
}

func Entities[T any](entities ...*T) []*T {
	return entities
}

func Columns(columns ...string) []string {
	return columns
}
