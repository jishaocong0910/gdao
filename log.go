package gdao

import (
	"context"
	"fmt"
	"reflect"
)

type logLevel int

const (
	LOG_LEVEL_DEBUG logLevel = iota + 1
	LOG_LEVEL_INFO
)

type Logger interface {
	Debugf(context.Context, string, ...any)
	Infof(context.Context, string, ...any)
	Warnf(context.Context, string, ...any)
	Errorf(context.Context, string, ...any)
}

func LogConf(log Logger, level logLevel) {
	logger = log
	printSqlLogLevel = level
}

var logger Logger
var printSqlLogLevel logLevel

func printSql(ctx context.Context, sql string) {
	printSqlLog(ctx, "SQL: %s", sql)
}

func printArgs(ctx context.Context, args []any) {
	msg := ""
	var argValues []any
	if len(args) > 0 {
		msg = "Args: %v"
		argValues = make([]any, 0, len(args))
		for _, a := range args {
			if a == nil {
				argValues = append(argValues, nil)
			} else {
				v := reflect.ValueOf(a)
				if v.Kind() == reflect.Pointer {
					if v.IsNil() {
						argValues = append(argValues, nil)
					} else {
						argValues = append(argValues, v.Elem().Interface())
					}
				} else {
					argValues = append(argValues, a)
				}
			}
		}
	}
	printSqlLog(ctx, msg, argValues)
}

func printAffected(ctx context.Context, affected int64) {
	printSqlLog(ctx, "Affected: %d", affected)
}

func printSqlLog(ctx context.Context, msg string, args ...any) {
	if logger == nil || printSqlLogLevel == 0 || msg == "" { // coverage-ignore
		return
	}
	switch printSqlLogLevel {
	case LOG_LEVEL_DEBUG:
		logger.Debugf(ctx, msg, args...)
	case LOG_LEVEL_INFO:
		logger.Infof(ctx, msg, args...)
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
