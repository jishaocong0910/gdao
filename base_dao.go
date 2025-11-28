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
	"strconv"
	"strings"
)

type baseDao struct {
	db *sql.DB
}

func (d baseDao) DB() *sql.DB {
	if d.db == nil { // coverage-ignore
		return global.DefaultDB
	}
	return d.db
}

func (d baseDao) query(ctx context.Context, sql string, args []any) (rows *sql.Rows, columns []string, closeFunc func(), err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	prepare, err := d.createPrepare(ctx, sql)
	if err != nil { // coverage-ignore
		return nil, nil, nil, err
	}
	args = convertArgs(args)
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

func (d baseDao) exec(ctx context.Context, sql string, args []any) (result sql.Result, affected int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	affected = int64(-1)
	prepare, err := d.createPrepare(ctx, sql)
	if err != nil { // coverage-ignore
		return nil, 0, err
	}
	defer func() {
		printWarn(ctx, prepare.Close())
	}()
	args = convertArgs(args)
	result, err = prepare.ExecContext(ctx, args...)
	if err != nil { // coverage-ignore
		return nil, 0, err
	}
	affected, err = result.RowsAffected()
	return
}

func (d baseDao) createPrepare(ctx context.Context, _sql string) (*sql.Stmt, error) {
	if tx := getTx(ctx); tx != nil {
		return tx.PrepareContext(ctx, _sql)
	} else {
		db := d.DB()
		if db == nil { // coverage-ignore
			return nil, errors.New("no available *sql.DB variable")
		}
		return db.PrepareContext(ctx, _sql)
	}
}

func newBaseDao(db *sql.DB) *baseDao {
	return &baseDao{db: db}
}

type Separate struct {
	prefix, separator, suffix string
	writeFixIfEmpty           bool
}

type BaseSqlBuilder struct {
	sql    strings.Builder
	args   []any
	argNum int
	ok     bool
	err    error
}

func (b *BaseSqlBuilder) Write(str string, args ...any) *BaseSqlBuilder {
	b.sql.WriteString(str)
	b.SetArgs(args...)
	return b
}

func (b *BaseSqlBuilder) SetArgs(args ...any) {
	b.args = append(b.args, args...)
}

func (b *BaseSqlBuilder) Pp(prefix string) string {
	b.argNum++
	return prefix + strconv.Itoa(b.argNum)
}

func (b *BaseSqlBuilder) Sql() string {
	return b.sql.String()
}

func (b *BaseSqlBuilder) Args() []any {
	return b.args
}

func (b *BaseSqlBuilder) SetError(err error) {
	if err != nil {
		b.err = err
		b.ok = false
	}
}

func (b *BaseSqlBuilder) Error() error {
	return b.err
}

func (b *BaseSqlBuilder) SetOk(ok bool) {
	if b.err == nil {
		b.ok = ok
	}
}

func (b *BaseSqlBuilder) Ok() bool {
	return b.ok
}

func (b *BaseSqlBuilder) Sep(separator string) *Separate {
	return &Separate{separator: separator}
}

func (b *BaseSqlBuilder) SepFix(prefix, separator, suffix string, writeFixIfEmpty bool) *Separate {
	return &Separate{prefix: prefix, separator: separator, suffix: suffix, writeFixIfEmpty: writeFixIfEmpty}
}

func (b *BaseSqlBuilder) Repeat(num int, sep *Separate, filter func(i int) bool, handle func(n, i int)) {
	var n int
	b.WritePrefix(sep, n)
	for i := 0; i < num; i++ {
		if filter != nil && !filter(i) {
			continue
		}
		n++
		b.WritePrefix(sep, n)
		b.WriteSep(sep, n)
		handle(n, i)
	}
	b.WriteSuffix(sep, n)
}

func (b *BaseSqlBuilder) WritePrefix(s *Separate, n int) *BaseSqlBuilder {
	if s != nil {
		if n == 0 && s.writeFixIfEmpty || n == 1 && !s.writeFixIfEmpty {
			b.Write(s.prefix)
		}
	}
	return b
}

func (b *BaseSqlBuilder) WriteSep(s *Separate, n int) *BaseSqlBuilder {
	if s != nil && n != 1 {
		b.Write(s.separator)
	}
	return b
}

func (b *BaseSqlBuilder) WriteSuffix(s *Separate, n int) *BaseSqlBuilder {
	if s != nil && n != 0 {
		b.Write(s.suffix)
	}
	return b
}

func NewBaseSqlBuilder() *BaseSqlBuilder {
	return &BaseSqlBuilder{ok: true}
}
