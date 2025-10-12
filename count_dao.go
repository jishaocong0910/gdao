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
)

type Count struct {
	Value *int64 `gdao:"column=count"`
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

type CountReq struct {
	Ctx      context.Context
	Must     bool
	Desc     string
	BuildSql func(b *CountBuilder)
}

type NewCountDaoReq struct {
	DB *sql.DB
}

type CountDao struct {
	baseDao
}

func (d *CountDao) Count(req CountReq) (first *Count, list []*Count, err error) {
	b := newCountSqlBuilder()
	req.BuildSql(b)
	if !b.Ok() { // coverage-ignore
		return nil, nil, b.err
	}
	rows, columns, closeFunc, err := query(req.Ctx, d.DB(), b.Sql(), b.args)
	if err != nil { // coverage-ignore
		printSql(req.Ctx, req.Desc, b.Sql(), b.args, -1, -1, err)
		checkMust(req.Must, err)
		return nil, nil, err
	}
	defer closeFunc()

	var rowCounts int64
	for rows.Next() {
		c := &Count{}
		if len(columns) == 1 {
			err = rows.Scan(&c.Value)
			if err != nil { // coverage-ignore
				checkMust(req.Must, err)
				return
			}
		} else {
			var fields []any
			for _, column := range columns {
				if column == "count" {
					fields = append(fields, &c.Value)
				} else {
					fields = append(fields, new(any))
				}
			}
			err = rows.Scan(fields...)
			if err != nil { // coverage-ignore
				checkMust(req.Must, err)
				return
			}
		}
		list = append(list, c)
		rowCounts++
	}
	if len(list) > 0 {
		first = list[0]
	}
	printSql(req.Ctx, req.Desc, b.Sql(), b.args, -1, rowCounts, nil)
	return
}

func NewCountDao(req NewCountDaoReq) *CountDao {
	d := &CountDao{}
	d.db = req.DB
	return d
}
