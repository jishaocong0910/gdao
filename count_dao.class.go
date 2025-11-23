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

import "errors"

type CountDao struct {
	*baseDao__
}

func (d *CountDao) Count(req CountReq) (count *Count, err error) {
	b := newCountSqlBuilder()
	req.BuildSql(b)
	if !b.Ok() { // coverage-ignore
		return nil, b.Error()
	}
	rows, columns, closeFunc, err := d.query(req.Ctx, b.Sql(), b.Args())
	if err != nil { // coverage-ignore
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), -1, -1, err)
		checkMust(req.Must, err)
		return nil, err
	}
	defer closeFunc()

	var rowCounts int64
	for rows.Next() {
		rowCounts++
		if count != nil { // coverage-ignore
			continue
		}
		c := &Count{}
		if len(columns) > 1 {
			err = errors.New("returns more than one column")
			checkMust(req.Must, err)
			return
		}
		err = rows.Scan(&c.Value)
		if err != nil { // coverage-ignore
			checkMust(req.Must, err)
			return
		}
		count = c
	}

	if rowCounts > 1 {
		count = nil
		err = errors.New("returns more than one row")
		printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), -1, rowCounts, err)
		checkMust(req.Must, err)
		return count, err
	}
	printSql(req.Ctx, req.SqlLogLevel, req.Desc, b.Sql(), b.Args(), -1, rowCounts, nil)
	return
}

func NewCountDao(req NewCountDaoReq) *CountDao {
	d := &CountDao{}
	d.baseDao__ = extendBaseDao(d, req.DB)
	return d
}
