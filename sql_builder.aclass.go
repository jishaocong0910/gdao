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
	"strconv"
	"strings"
)

type sqlBuilder_ interface {
	sqlBuilder_() *sqlBuilder__
}

type sqlBuilder__ struct {
	i      sqlBuilder_
	sql    strings.Builder
	args   []any
	argNum int
	ok     bool
	err    error
}

func (this *sqlBuilder__) sqlBuilder_() *sqlBuilder__ { // coverage-ignore
	return this
}

func (this *sqlBuilder__) Write(str string, args ...any) *sqlBuilder__ {
	this.sql.WriteString(str)
	this.SetArgs(args...)
	return this
}

func (this *sqlBuilder__) SetArgs(args ...any) {
	this.args = append(this.args, args...)
}

func (this *sqlBuilder__) Pp(prefix string) string {
	this.argNum++
	return prefix + strconv.Itoa(this.argNum)
}

func (this *sqlBuilder__) Sql() string {
	return this.sql.String()
}

func (this *sqlBuilder__) Args() []any {
	return this.args
}

func (this *sqlBuilder__) SetError(err error) {
	if err != nil {
		this.err = err
		this.ok = false
	}
}

func (this *sqlBuilder__) Error() error {
	return this.err
}

func (this *sqlBuilder__) SetOk(ok bool) {
	if this.err == nil {
		this.ok = ok
	}
}

func (this *sqlBuilder__) Ok() bool {
	return this.ok
}

func (this *sqlBuilder__) Sep(separator string) *separate {
	return &separate{separator: separator}
}

func (this *sqlBuilder__) SepFix(prefix, separator, suffix string, writeFixIfEmpty bool) *separate {
	return &separate{prefix: prefix, separator: separator, suffix: suffix, writeFixIfEmpty: writeFixIfEmpty}
}

func (this *sqlBuilder__) Repeat(num int, sep *separate, filter func(i int) bool, handle func(n, i int)) {
	var n int
	this.writePrefix(sep, n)
	for i := 0; i < num; i++ {
		if filter != nil && !filter(i) {
			continue
		}
		n++
		this.writePrefix(sep, n)
		this.writeSep(sep, n)
		handle(n, i)
	}
	this.writeSuffix(sep, n)
}

func (this *sqlBuilder__) writePrefix(s *separate, n int) *sqlBuilder__ {
	if s != nil {
		if n == 0 && s.writeFixIfEmpty || n == 1 && !s.writeFixIfEmpty {
			this.Write(s.prefix)
		}
	}
	return this
}

func (this *sqlBuilder__) writeSep(s *separate, n int) *sqlBuilder__ {
	if s != nil && n != 1 {
		this.Write(s.separator)
	}
	return this
}

func (this *sqlBuilder__) writeSuffix(s *separate, n int) *sqlBuilder__ {
	if s != nil && n != 0 {
		this.Write(s.suffix)
	}
	return this
}

func extendSqlBuilder(i sqlBuilder_) *sqlBuilder__ {
	return &sqlBuilder__{i: i, ok: true}
}
