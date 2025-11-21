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

type baseSqlBuilder_ interface {
	baseSqlBuilder_()
}

type BaseSqlBuilder__ struct {
	i      baseSqlBuilder_
	sql    strings.Builder
	args   []any
	argNum int
	ok     bool
	err    error
}

func (this *BaseSqlBuilder__) baseSqlBuilder_() { // coverage-ignore
}

func (this *BaseSqlBuilder__) Write(str string, args ...any) *BaseSqlBuilder__ {
	this.sql.WriteString(str)
	this.SetArgs(args...)
	return this
}

func (this *BaseSqlBuilder__) SetArgs(args ...any) {
	this.args = append(this.args, args...)
}

func (this *BaseSqlBuilder__) Pp(prefix string) string {
	this.argNum++
	return prefix + strconv.Itoa(this.argNum)
}

func (this *BaseSqlBuilder__) Sql() string {
	return this.sql.String()
}

func (this *BaseSqlBuilder__) Args() []any {
	return this.args
}

func (this *BaseSqlBuilder__) SetError(err error) {
	if err != nil {
		this.err = err
		this.ok = false
	}
}

func (this *BaseSqlBuilder__) Error() error {
	return this.err
}

func (this *BaseSqlBuilder__) SetOk(ok bool) {
	if this.err == nil {
		this.ok = ok
	}
}

func (this *BaseSqlBuilder__) Ok() bool {
	return this.ok
}

func (this *BaseSqlBuilder__) Sep(separator string) *Separate {
	return &Separate{separator: separator}
}

func (this *BaseSqlBuilder__) SepFix(prefix, separator, suffix string, writeFixIfEmpty bool) *Separate {
	return &Separate{prefix: prefix, separator: separator, suffix: suffix, writeFixIfEmpty: writeFixIfEmpty}
}

func (this *BaseSqlBuilder__) Repeat(num int, sep *Separate, filter func(i int) bool, handle func(n, i int)) {
	var n int
	this.WritePrefix(sep, n)
	for i := 0; i < num; i++ {
		if filter != nil && !filter(i) {
			continue
		}
		n++
		this.WritePrefix(sep, n)
		this.WriteSep(sep, n)
		handle(n, i)
	}
	this.WriteSuffix(sep, n)
}

func (this *BaseSqlBuilder__) WritePrefix(s *Separate, n int) *BaseSqlBuilder__ {
	if s != nil {
		if n == 0 && s.writeFixIfEmpty || n == 1 && !s.writeFixIfEmpty {
			this.Write(s.prefix)
		}
	}
	return this
}

func (this *BaseSqlBuilder__) WriteSep(s *Separate, n int) *BaseSqlBuilder__ {
	if s != nil && n != 1 {
		this.Write(s.separator)
	}
	return this
}

func (this *BaseSqlBuilder__) WriteSuffix(s *Separate, n int) *BaseSqlBuilder__ {
	if s != nil && n != 0 {
		this.Write(s.suffix)
	}
	return this
}

func ExtendBaseSqlBuilder(i baseSqlBuilder_) *BaseSqlBuilder__ {
	return &BaseSqlBuilder__{i: i, ok: true}
}
