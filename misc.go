// Copyright 2024-present jishaocong0910
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

import (
	"runtime/debug"
	"slices"
)

func checkMust(must bool, err error) error {
	if must && err != nil {
		panic(err)
	}
	return err
}

func deferStack() []byte {
	stack := debug.Stack()
	line := 0
	begin := 0
	for i, b := range stack {
		if b == '\n' {
			line++
		}
		if line == 9 {
			begin = i + 1
			break
		}
	}
	return stack[begin:]
}

type set[T comparable] map[T]struct{}

func (s set[T]) contain(e T) bool {
	_, ok := s[e]
	return ok
}

func (s set[T]) add(e_ ...T) {
	for _, e := range e_ {
		s[e] = struct{}{}
	}
}

func (s set[T]) concat(source_ ...set[T]) set[T] {
	source_ = slices.DeleteFunc(source_, func(s set[T]) bool {
		return len(s) == 0
	})
	if len(source_) == 0 {
		return s
	}

	set_ := make([]set[T], 0, 1+len(source_))
	set_ = append(set_, s)
	size := len(s)
	for _, source := range source_ {
		set_ = append(set_, source)
		size += len(source)
	}

	m := make(set[T], size)
	for _, s2 := range set_ {
		for e := range s2 {
			m[e] = struct{}{}
		}
	}
	return m
}

func newSet[T comparable](e_ ...T) set[T] {
	m := make(set[T], len(e_))
	if len(e_) > 0 {
		for _, e := range e_ {
			m[e] = struct{}{}
		}
	}
	return m
}

func newSetWithCap[T comparable](cap int) set[T] {
	return make(set[T], cap)
}
