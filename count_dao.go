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
)

type CountDao struct {
	*baseDao
}

func (d *CountDao) Count() *count {
	return &count{dao: d.baseDao, req: &countReq{}}
}

type countReq struct {
	ctx      context.Context
	must     bool
	logLevel LogLevel
	desc     string
	buildSql func(b *CountBuilder)
}

type count struct {
	dao *baseDao
	req *countReq
}

func (c *count) Ctx(ctx context.Context) *count {
	c.req.ctx = ctx
	return c
}

func (c *count) Must(must bool) *count {
	c.req.must = must
	return c
}

func (c *count) LogLevel(logLevel LogLevel) *count {
	c.req.logLevel = logLevel
	return c
}

func (c *count) Desc(desc string) *count {
	c.req.desc = desc
	return c
}

func (c *count) BuildSql(buildSql func(b *CountBuilder)) *count {
	c.req.buildSql = buildSql
	return c
}

func (c *count) Do() (count *Count, err error) {
	b := &CountBuilder{BaseSqlBuilder: NewBaseSqlBuilder()}
	c.req.buildSql(b)
	if !b.Ok() { // coverage-ignore
		return nil, b.Error()
	}
	rows, columns, closeFunc, err := c.dao.query(c.req.ctx, b.Sql(), b.Args())
	if err != nil { // coverage-ignore
		printSql(c.req.ctx, c.req.logLevel, c.req.desc, b.Sql(), b.Args(), -1, -1, err)
		checkMust(c.req.must, err)
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
			checkMust(c.req.must, err)
			return
		}
		err = rows.Scan(&count.Value)
		if err != nil { // coverage-ignore
			count = nil
			checkMust(c.req.must, err)
			return
		}
	}

	if rowCounts > 1 {
		count = nil
		err = errors.New("returns more than one row")
		printSql(c.req.ctx, c.req.logLevel, c.req.desc, b.Sql(), b.Args(), -1, rowCounts, err)
		checkMust(c.req.must, err)
		return count, err
	}
	printSql(c.req.ctx, c.req.logLevel, c.req.desc, b.Sql(), b.Args(), -1, rowCounts, nil)
	return
}

type CountBuilder struct {
	*BaseSqlBuilder
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
	db *sql.DB
}

func (b *countDaoBuilder) DB(db *sql.DB) *countDaoBuilder {
	b.db = db
	return b
}

func (b *countDaoBuilder) Build() *CountDao {
	return &CountDao{baseDao: newBaseDao(b.db)}
}

func CountDaoBuilder() *countDaoBuilder {
	return &countDaoBuilder{}
}
