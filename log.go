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

func printWarn(ctx context.Context, err error) {
	if err == nil { // coverage-ignore
		return
	}
	printLog(ctx, LogLevel_.WARN, fmt.Sprintf("%v", err))
}

func printSql(ctx context.Context, logLevel LogLevel, desc string, sql string, args []any, affected, rowCount int64, count int64, err error) {
	if err != nil {
		logLevel = LogLevel_.ERROR
	} else if logLevel.IsUndefined() {
		if cfg.SqlLogLevel.IsUndefined() { // coverage-ignore
			return
		}
		logLevel = cfg.SqlLogLevel
	}

	var msg strings.Builder
	msgArgs := make([]any, 0, 5+len(args))
	if desc != "" {
		msg.WriteString("desc: %s, ")
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

	if count != -1 {
		msg.WriteString(sep)
		sep = ", "
		msg.WriteString("count: %d")
		msgArgs = append(msgArgs, count)
	} else if rowCount != -1 {
		msg.WriteString(sep)
		sep = ", "
		msg.WriteString("rowcount: %d")
		msgArgs = append(msgArgs, rowCount)
	}

	if err != nil {
		msg.WriteString(sep)
		msg.WriteString("error: %+v")
		msgArgs = append(msgArgs, err)
	}

	printLog(ctx, logLevel, msg.String(), msgArgs...)
}

func printLog(ctx context.Context, logLevel LogLevel, msg string, args ...any) {
	if cfg.Logger == nil { // coverage-ignore
		return
	}
	switch logLevel.String() {
	case LogLevel_.DEBUG.String():
		cfg.Logger.Debugf(ctx, msg, args...)
	case LogLevel_.INFO.String():
		cfg.Logger.Infof(ctx, msg, args...)
	case LogLevel_.WARN.String():
		cfg.Logger.Warnf(ctx, msg, args...)
	case LogLevel_.ERROR.String():
		cfg.Logger.Errorf(ctx, msg, args...)
	}
}

func formatSql(sql string) string {
	if cfg.CompressSqlLog {
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
