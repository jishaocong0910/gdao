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

package gen

import "strings"

type stringTokenizer struct {
	text             string
	chars            []rune
	len              int
	c                rune
	pos              int
	end              bool
	token            string
	tokenIsDelimiter bool
	delimiters       map[rune]struct{}
}

func (st *stringTokenizer) nextDelimiterAndToken() bool {
	if st.end {
		return false
	}
	st.token = ""
	st.tokenIsDelimiter = false
	if _, ok := st.delimiters[st.c]; ok {
		st.token = string(st.c)
		st.tokenIsDelimiter = true
		st.nextChar()
	} else {
		begin := st.pos
		for st.pos < st.len && st.nextChar() {
			if _, ok := st.delimiters[st.c]; ok {
				break
			}
		}
		st.token = string(st.chars[begin:st.pos])
	}
	return true
}

func (st *stringTokenizer) nextToken() bool {
	for st.nextDelimiterAndToken() {
		cs := []rune(st.token)
		if len(cs) != 1 {
			return true
		}
		if _, ok := st.delimiters[cs[0]]; !ok { // coverage-ignore
			return true
		}
	}
	return false
}

func (st *stringTokenizer) nextUntilDelimiterAndToken(token string) bool {
	for {
		if strings.EqualFold(st.token, token) {
			return true
		}
		if !st.nextDelimiterAndToken() { // coverage-ignore
			break
		}
	}
	return false
}

func (st *stringTokenizer) nextChar() bool {
	if st.end { // coverage-ignore
		return false
	}
	st.c = 0
	st.pos++
	if st.pos == st.len {
		st.end = true
		return false
	}
	st.c = st.chars[st.pos]
	return true
}

func (st *stringTokenizer) reset() {
	st.c = 0
	st.pos = -1
	st.end = false
	st.token = ""
	st.tokenIsDelimiter = false
	st.nextChar()
	st.nextDelimiterAndToken()
}

func newStringTokenizer(text string, delimiters []rune) *stringTokenizer {
	chars := []rune(text)
	m := make(map[rune]struct{})
	for _, c := range delimiters {
		m[c] = struct{}{}
	}
	st := stringTokenizer{text: text, chars: chars, len: len(chars), pos: -1, delimiters: m}
	st.nextChar()
	st.nextDelimiterAndToken()
	return &st
}
