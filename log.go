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

func LogCfg(log Logger, level string) {
	logger = log
	printSqlLevel = level
}

var logger Logger
var printSqlLevel string

func printSql(ctx context.Context, sql string, args []any, affected int64, err error) {
	var msg strings.Builder
	msg.WriteString("SQL: %s")
	msgArgs := make([]any, 0, 3)
	msgArgs = append(msgArgs, sql)

	if len(args) > 0 {
		msg.WriteString(", args: %v")
		var values = make([]any, 0, len(args))
		for _, a := range args {
			if a == nil {
				values = append(values, nil)
			} else {
				v := reflect.ValueOf(a)
				if v.Kind() == reflect.Pointer {
					if v.IsNil() {
						values = append(values, nil)
					} else {
						values = append(values, v.Elem().Interface())
					}
				} else {
					values = append(values, a)
				}
			}
		}
		msgArgs = append(msgArgs, values)
	}

	if affected != -1 {
		msg.WriteString(", affected: %d")
		msgArgs = append(msgArgs, affected)
	}

	if err != nil {
		msg.WriteString(", error: %+v")
		msgArgs = append(msgArgs, err)
	}

	printSqlLog(ctx, msg.String(), msgArgs...)
}

func printSqlCanceled(ctx context.Context, sql string) {
	printSqlLog(ctx, "SQL canceled: %s", sql)
}

func printSqlLog(ctx context.Context, msg string, args ...any) {
	if logger == nil { // coverage-ignore
		return
	}
	switch strings.ToLower(printSqlLevel) {
	case "info":
		logger.Infof(ctx, msg, args...)
	default:
		logger.Debugf(ctx, msg, args...)
	}
}

func printWarn(ctx context.Context, err error) {
	if logger == nil || err == nil { // coverage-ignore
		return
	}
	logger.Warnf(ctx, fmt.Sprintf("%v", err))
}

func printError(ctx context.Context, err error) {
	if logger == nil || err == nil { // coverage-ignore
		return
	}
	logger.Errorf(ctx, fmt.Sprintf("%v", err))
}
