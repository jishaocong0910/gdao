package gdao

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Logger interface {
	Debugf(ctx context.Context, msg string, args ...any)
	Infof(ctx context.Context, msg string, args ...any)
	Warnf(ctx context.Context, msg string, args ...any)
	Errorf(ctx context.Context, msg string, args ...any)
}

func LogCfg(log Logger, printSqlLevel string, singleLineSql bool) {
	_logger = log
	_printSqlLevel = strings.ToLower(printSqlLevel)
	_singleLineSql = singleLineSql
}

var _logger Logger
var _printSqlLevel string
var _singleLineSql bool

func formatSql(sql string) string {
	if _singleLineSql {
		var line strings.Builder
		chars := []rune(sql)
		var prevC rune
		for i, c := range chars {
			if c == '\n' {
				if prevC == 0 {
					continue
				}
				if prevC != ' ' && i < len(chars)-1 && chars[i+1] != ' ' {
					line.WriteRune(' ')
				}
				continue
			}
			if c == ' ' {
				if prevC == 0 || prevC == ' ' {
					continue
				}
			}
			line.WriteRune(c)
			prevC = c
		}
		sql = line.String()
	}
	return sql
}

func printSql(ctx context.Context, sql string, args []any, affected, rowCounts int64, err error) {
	var msg strings.Builder
	msg.WriteString("SQL: %s;")
	msgArgs := make([]any, 0, 3)
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
				if s, ok := a.(string); ok {
					a = "\"" + s + "\""
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

func printError(ctx context.Context, err error) {
	if _logger == nil || err == nil { // coverage-ignore
		return
	}
	_logger.Errorf(ctx, fmt.Sprintf("%v", err))
}
