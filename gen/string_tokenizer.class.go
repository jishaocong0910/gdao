package gen

import (
	"strings"

	o "github.com/jishaocong0910/go-object"
)

type stringTokenizer struct {
	chars      []rune
	len        int
	pos        int
	token      string
	delimiters *o.Set[rune]
}

func (this *stringTokenizer) nextIncludeDelimiterToken() bool {
	this.token = ""
	if this.pos == this.len {
		return false
	}
	if this.delimiters.Contains(this.chars[this.pos]) {
		this.token = string(this.chars[this.pos])
		this.pos++
		return true
	}
	begin := this.pos
	for ; this.pos < this.len; this.pos++ {
		if this.delimiters.Contains(this.chars[this.pos]) {
			break
		}
	}
	this.token = string(this.chars[begin:this.pos])
	return true
}

func (this *stringTokenizer) nextToken() bool {
	for this.nextIncludeDelimiterToken() {
		cs := []rune(this.token)
		if len(cs) != 1 {
			return true
		}
		if !this.delimiters.Contains(cs[0]) {
			return true
		}
	}
	return false
}

func (this *stringTokenizer) nextUntilToken(token string) bool {
	for {
		if strings.EqualFold(this.token, token) {
			return true
		}
		if !this.nextIncludeDelimiterToken() {
			break
		}
	}
	return false
}

func newStringTokenizer(str string) *stringTokenizer {
	chars := []rune(str)
	st := &stringTokenizer{chars: chars, len: len(chars), delimiters: o.NewSet(' ', '\n', '\t', ',', '(')}
	st.nextIncludeDelimiterToken()
	return st
}
