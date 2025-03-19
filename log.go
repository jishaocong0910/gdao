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

var Log = func() struct {
	Logger           Logger
	PrintSqlLogLevel logLevel
} {
	return struct {
		Logger           Logger
		PrintSqlLogLevel logLevel
	}{}
}()

func printSql(ctx context.Context, sql string, args []any) {
	if Log.Logger == nil || Log.PrintSqlLogLevel == 0 { // coverage-ignore
		return
	}

	msg := "SQL: " + sql
	var argValues []any
	if len(args) > 0 {
		msg += "\nArgs: %v"
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

	switch Log.PrintSqlLogLevel {
	case LOG_LEVEL_DEBUG:
		Log.Logger.Debugf(ctx, msg, argValues...)
	case LOG_LEVEL_INFO:
		Log.Logger.Infof(ctx, msg, argValues...)
	}
}

func printWarn(ctx context.Context, err error) {
	if Log.Logger == nil || err == nil { // coverage-ignore
		return
	}
	Log.Logger.Warnf(ctx, fmt.Sprintf("%v", err))
}

func printError(ctx context.Context, err error) {
	if Log.Logger == nil || err == nil { // coverage-ignore
		return
	}
	Log.Logger.Errorf(ctx, fmt.Sprintf("%v", err))
}
