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

package orm

import (
	"errors"
	"reflect"
	"runtime/debug"
)

func P[T any](t T) *T {
	return &t
}

func V[T any](t *T) T {
	var v T
	if t != nil {
		v = *t
	}
	return v
}

func toSet[T comparable](slice []T) map[T]struct{} {
	m := make(map[T]struct{}, len(slice))
	if len(slice) > 0 {
		for _, e := range slice {
			m[e] = struct{}{}
		}
	}
	return m
}

func checkEntityType[T Entity]() error {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("generics must be struct type")
	}
	return nil
}

func checkMust(must bool, err error) { // coverage-ignore
	if must && err != nil {
		panic(err)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func deferStack() []byte {
	stack := debug.Stack()
	skip := 9
	line := 0
	begin := 0
	for i, b := range stack {
		if b == '\n' {
			line++
		}
		if line == skip {
			begin = i + 1
			break
		}
	}
	return stack[begin:]
}
