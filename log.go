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
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Logger interface {
	Debugf(ctx context.Context, msg string, args ...any)
	Infof(ctx context.Context, msg string, args ...any)
	Warnf(ctx context.Context, msg string, args ...any)
	Errorf(ctx context.Context, msg string, args ...any)
}

func LogCfg(log Logger, printSqlLevel string, compressSql bool) {
	_logger = log
	_printSqlLevel = strings.ToLower(printSqlLevel)
	_compressSql = compressSql
}

var _logger Logger
var _printSqlLevel string
var _compressSql bool

func formatSql(sql string) string {
	if _compressSql {
		sql = strings.TrimSpace(sql)
		var line strings.Builder
		chars := []rune(sql)
		var prevC rune
		for i, c := range chars {
			if c == '\n' {
				if prevC != ' ' && i != len(chars)-1 && chars[i+1] != ' ' {
					line.WriteRune(' ')
				}
				continue
			}
			line.WriteRune(c)
			prevC = c
		}
		sql = line.String()
	}
	return sql
}

func printSql(ctx context.Context, desc string, sql string, args []any, affected, rowCounts int64, err error) {
	var msg strings.Builder
	msgArgs := make([]any, 0, 5+len(args))
	if desc != "" {
		msg.WriteString("Desc: %s, ")
		msgArgs = append(msgArgs, desc)
	}
	msg.WriteString("SQL: %s;")
	msgArgs = append(msgArgs, formatSql(sql))

	sep := " "
	if len(args) > 0 {
		sep = ", "
		msg.WriteString(" args: %v")
		var values = make([]any, 0, len(args))
		for _, a := range args {
			if a == nil {
				values = append(values, nil)
			} else {
				v := reflect.ValueOf(a)
				if v.Kind() == reflect.Pointer {
					if v.IsNil() {
						a = nil
					} else {
						a = v.Elem().Interface()
					}
				}
				if a != nil {
					if s, ok := a.(string); ok {
						a = "\"" + s + "\""
					} else if t, ok := a.(time.Time); ok {
						a = "time.Time(" + t.String() + ")"
					}
				}
				values = append(values, a)
			}
		}
		msgArgs = append(msgArgs, values)
	}

	if affected != -1 {
		msg.WriteString(sep)
		sep = ", "
		msg.WriteString("affected: %d")
		msgArgs = append(msgArgs, affected)
	}

	if rowCounts != -1 {
		msg.WriteString(sep)
		sep = ", "
		msg.WriteString("row counts: %d")
		msgArgs = append(msgArgs, rowCounts)
	}

	if err != nil {
		msg.WriteString(sep)
		msg.WriteString("error: %+v")
		msgArgs = append(msgArgs, err)
	}

	printSqlLog(ctx, err != nil, msg.String(), msgArgs...)
}

func printSqlLog(ctx context.Context, hasError bool, msg string, args ...any) {
	if _logger == nil { // coverage-ignore
		return
	}
	if hasError {
		_logger.Errorf(ctx, msg, args...)
	} else {
		switch _printSqlLevel {
		case "info":
			_logger.Infof(ctx, msg, args...)
		default:
			_logger.Debugf(ctx, msg, args...)
		}
	}
}

func printWarn(ctx context.Context, err error) {
	if _logger == nil || err == nil { // coverage-ignore
		return
	}
	_logger.Warnf(ctx, fmt.Sprintf("%v", err))
}
