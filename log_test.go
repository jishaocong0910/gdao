package gdao_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

type MockLogger struct {
	DebugMsg  string
	DebugArgs []any
	InfoMsg   string
	InfoArgs  []any
	WarnMsg   string
	WarnArgs  []any
	ErrorMsg  string
	ErrorArgs []any
}

func (d *MockLogger) Debugf(ctx context.Context, msg string, args ...interface{}) {
	d.DebugMsg = msg
	d.DebugArgs = append(d.DebugArgs, args...)
}

func (d *MockLogger) Infof(ctx context.Context, msg string, args ...interface{}) {
	d.InfoMsg = msg
	d.InfoArgs = append(d.InfoArgs, args...)
}

func (d *MockLogger) Warnf(ctx context.Context, msg string, args ...interface{}) {
	d.WarnMsg = msg
	d.WarnArgs = append(d.WarnArgs, args...)
}

func (d *MockLogger) Errorf(ctx context.Context, msg string, args ...interface{}) {
	d.ErrorMsg = msg
	d.ErrorArgs = append(d.ErrorArgs, args...)
}

func TestPrintSql(t *testing.T) {
	r := require.New(t)
	{
		log := &MockLogger{}
		gdao.Log.Logger = log
		gdao.Log.PrintSqlLogLevel = gdao.LOG_LEVEL_DEBUG
		gdao.PrintSql(nil, "SELECT * FROM user WHERE id=? AND status=? AND level=?", []any{gdao.Ptr(1), 2, nil})
		r.Equal("SQL: SELECT * FROM user WHERE id=? AND status=? AND level=?\nArgs: %v", log.DebugMsg)
		r.Equal(1, log.DebugArgs[0])
		r.Equal(2, log.DebugArgs[1])
		r.Nil(log.DebugArgs[2])
	}
	{
		log := &MockLogger{}
		gdao.Log.Logger = log
		gdao.Log.PrintSqlLogLevel = gdao.LOG_LEVEL_INFO
		gdao.PrintSql(nil, "SELECT * FROM user WHERE id=? AND status=? AND level=?", []any{gdao.Ptr(1), (*int)(nil), nil})
		r.Equal("SQL: SELECT * FROM user WHERE id=? AND status=? AND level=?\nArgs: %v", log.InfoMsg)
		r.Equal(1, log.InfoArgs[0])
		r.Nil(log.InfoArgs[1])
		r.Nil(log.InfoArgs[2])
	}
	{
		log := &MockLogger{}
		gdao.Log.Logger = log
		gdao.Log.PrintSqlLogLevel = gdao.LOG_LEVEL_INFO
		gdao.PrintSql(nil, "SELECT * FROM user WHERE id=1", nil)
		r.Equal("SQL: SELECT * FROM user WHERE id=1", log.InfoMsg)
		r.Nil(log.InfoArgs)
	}
}

func TestPrintWarn(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	gdao.Log.Logger = log
	gdao.PrintWarn(nil, errors.New("warn"))
	r.Equal("warn", log.WarnMsg)
}

func TestPrintError(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	gdao.Log.Logger = log
	gdao.PrintError(nil, errors.New("error"))
	r.Equal("error", log.ErrorMsg)
}
