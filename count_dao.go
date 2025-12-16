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
)

type CountDao struct {
	*baseDao
	table string
}

func (d *CountDao) Count() *count {
	return &count{dao: d}
}

type count struct {
	dao         *CountDao
	ctx         context.Context
	must        bool
	sqlLogLevel LogLevel
	desc        string
	buildSql    func(b *CountSqlBuilder)
}

func (c *count) Ctx(ctx context.Context) *count {
	c.ctx = ctx
	return c
}

func (c *count) Must(must bool) *count {
	c.must = must
	return c
}

func (c *count) SqlLogLevel(logLevel LogLevel) *count {
	c.sqlLogLevel = logLevel
	return c
}

func (c *count) Desc(desc string) *count {
	c.desc = desc
	return c
}

func (c *count) BuildSql(buildSql func(b *CountSqlBuilder)) *count {
	c.buildSql = buildSql
	return c
}

func (c *count) Do() (count *Count, err error) {
	b := newCountSqlBuilder(c.dao)
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

type CountSqlBuilder struct {
	*BaseSqlBuilder
	dao *CountDao
}

func (b *CountSqlBuilder) Write(str string, args ...any) *CountSqlBuilder {
	b.BaseSqlBuilder.Write(str, args...)
	return b
}

func (b *CountSqlBuilder) WriteTable() *CountSqlBuilder {
	b.Write(b.dao.table)
	return b
}

func newCountSqlBuilder(d *CountDao) *CountSqlBuilder {
	return &CountSqlBuilder{BaseSqlBuilder: NewBaseSqlBuilder(), dao: d}
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

type countDaoBuilder struct {
	db    *sql.DB
	table string
}

func (b *countDaoBuilder) DB(db *sql.DB) *countDaoBuilder {
	b.db = db
	return b
}

func (b *countDaoBuilder) Table(table string) *countDaoBuilder {
	b.table = table
	return b
}

func (b *countDaoBuilder) Build() *CountDao {
	return &CountDao{baseDao: newBaseDao(b.db), table: b.table}
}

func CountDaoBuilder() *countDaoBuilder {
	return &countDaoBuilder{}
}
