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

func formatSql(sql string) string {
	if global.CompressSqlLog {
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

func printSql(ctx context.Context, sqlLogLevel SqlLogLevel, desc string, sql string, args []any, affected, rowCounts int64, err error) {
	if sqlLogLevel.IsUndefined() {
		sqlLogLevel = global.SqlLogLevel
	}
	if SqlLogLevel_.Not(sqlLogLevel, SqlLogLevel_.DEBUG, SqlLogLevel_.INFO) { // coverage-ignore
		return
	}
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

	printSqlLog(ctx, sqlLogLevel, err != nil, msg.String(), msgArgs...)
}

func printSqlLog(ctx context.Context, sqlLogLevel SqlLogLevel, hasError bool, msg string, args ...any) {
	if global.Logger == nil { // coverage-ignore
		return
	}
	if hasError {
		global.Logger.Errorf(ctx, msg, args...)
	} else {
		switch sqlLogLevel.String() {
		case SqlLogLevel_.DEBUG.String():
			global.Logger.Debugf(ctx, msg, args...)
		case SqlLogLevel_.INFO.String():
			global.Logger.Infof(ctx, msg, args...)
		}
	}
}

func printWarn(ctx context.Context, err error) {
	if global.Logger == nil || err == nil { // coverage-ignore
		return
	}
	global.Logger.Warnf(ctx, fmt.Sprintf("%v", err))
}
