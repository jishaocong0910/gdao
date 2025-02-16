package gdao

import (
	"strings"
	"unicode"
)

type NameMapper struct {
	mapping      []nameMapping
	hasCamelCase bool
}

func (n *NameMapper) Convert(str string) string {
	for _, m := range n.mapping {
		str = m(str)
	}
	return str
}
func NewNameMapper() *NameMapper {
	return &NameMapper{}
}

type nameMapping func(str string) string

func (n *NameMapper) SubPrefix(prefix string) *NameMapper {
	n.mapping = append(n.mapping, subPrefix(prefix))
	return n
}

func (n *NameMapper) SubSuffix(suffix string) *NameMapper {
	n.mapping = append(n.mapping, subSuffix(suffix))
	return n
}

func (n *NameMapper) AddPrefix(prefix string) *NameMapper {
	n.mapping = append(n.mapping, addPrefix(prefix))
	return n
}

func (n *NameMapper) AddSuffix(suffix string) *NameMapper {
	n.mapping = append(n.mapping, addSuffix(suffix))
	return n
}

func (n *NameMapper) LowerCamelCase() *NameMapper {
	n.hasCamelCase = true
	n.mapping = append(n.mapping, lowerCamelCase)
	return n
}

func (n *NameMapper) LowerSnakeCase() *NameMapper {
	if !n.hasCamelCase {
		n.mapping = append(n.mapping, lowerCamelCase)
	}
	n.mapping = append(n.mapping, lowerSnakeCase)
	return n
}

func (n *NameMapper) LowerFirstLiteral() *NameMapper {
	n.mapping = append(n.mapping, lowerFirstLiteral)
	return n
}

func (n *NameMapper) UpperCamelCase() *NameMapper {
	n.hasCamelCase = true
	n.mapping = append(n.mapping, upperCamelCase)
	return n
}

func (n *NameMapper) UpperSnakeCase() *NameMapper {
	if !n.hasCamelCase {
		n.mapping = append(n.mapping, upperCamelCase)
	}
	n.mapping = append(n.mapping, upperSnakeCase)
	return n
}

func (n *NameMapper) UpperFirstLiteral() *NameMapper {
	n.mapping = append(n.mapping, upperFirstLiteral)
	return n
}

var subPrefix = func(prefix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return strings.TrimPrefix(str, prefix)
	}
}

var subSuffix = func(suffix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return strings.TrimRight(str, suffix)
	}
}

var addPrefix = func(prefix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return prefix + str
	}
}

var addSuffix = func(suffix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return str + suffix
	}
}

var lowerCamelCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	chars := []rune(str)
	builder.WriteRune(unicode.ToLower(chars[0]))
	up := false
	for i := 1; i < len(chars); i++ {
		c := chars[i]
		switch c {
		case '_', '-', ' ':
			up = true
			continue
		default:
			if up {
				up = false
				builder.WriteRune(unicode.ToUpper(c))
			} else if unicode.IsUpper(c) && unicode.IsLower(chars[i-1]) {
				builder.WriteRune(c)
			} else {
				builder.WriteRune(unicode.ToLower(c))
			}
		}
	}
	return builder.String()
}

var lowerSnakeCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	chars := []rune(str)
	builder.WriteRune(unicode.ToLower(chars[0]))
	for i := 1; i < len(chars); i++ {
		c := chars[i]
		if unicode.IsUpper(c) && chars[i-1] != '-' && chars[i-1] != ' ' {
			builder.WriteRune('_')
			builder.WriteRune(unicode.ToLower(c))
		} else {
			builder.WriteRune(unicode.ToLower(c))
		}
	}
	return builder.String()
}

var lowerFirstLiteral = func(str string) string {
	if str == "" {
		return str
	}
	return strings.ToLower(str[:1]) + str[1:]
}

var upperCamelCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	chars := []rune(str)
	builder.WriteRune(unicode.ToUpper(chars[0]))
	up := false
	for i := 1; i < len(chars); i++ {
		c := chars[i]
		switch c {
		case '_', '-', ' ':
			up = true
			continue
		default:
			if up {
				up = false
				builder.WriteRune(unicode.ToUpper(c))
			} else if unicode.IsUpper(c) && unicode.IsLower(chars[i-1]) {
				builder.WriteRune(c)
			} else {
				builder.WriteRune(unicode.ToLower(c))
			}
		}
	}
	return builder.String()
}

var upperSnakeCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	chars := []rune(str)
	builder.WriteRune(unicode.ToUpper(chars[0]))
	for i := 1; i < len(chars); i++ {
		c := chars[i]
		if unicode.IsUpper(c) && chars[i-1] != '-' && chars[i-1] != ' ' {
			builder.WriteRune('_')
			builder.WriteRune(unicode.ToUpper(c))
		} else {
			builder.WriteRune(unicode.ToUpper(c))
		}
	}
	return builder.String()
}

var upperFirstLiteral = func(str string) string {
	if str == "" {
		return str
	}
	return strings.ToUpper(str[:1]) + str[1:]
}
