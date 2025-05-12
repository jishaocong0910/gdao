package gdao_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

type MockLogger struct {
	msg  string
	args []any
}

func (d *MockLogger) Debugf(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func TestPrintSql(t *testing.T) {
	r := require.New(t)
	{
		log := &MockLogger{}
		gdao.LogCfg(log, gdao.LOG_LEVEL_DEBUG)
		gdao.PrintSql(nil, "SELECT * FROM user WHERE id=? AND status=? AND level=?")
		r.Equal("SQL: %s", log.msg)
		r.Equal("SELECT * FROM user WHERE id=? AND status=? AND level=?", log.args[0])
		gdao.PrintArgs(nil, []any{gdao.Ptr(1), 2, nil, (*int)(nil)})
		r.Equal("Args: %v", log.msg)
		args := log.args[0].([]any)
		r.Equal(4, len(args))
		r.Equal(1, args[0])
		r.Equal(2, args[1])
		r.Equal(nil, args[2])
		r.Equal(nil, args[3])
		gdao.PrintAffected(nil, 2)
		r.Equal("Affected: %d", log.msg)

	}
	{
		log := &MockLogger{}
		gdao.LogCfg(log, gdao.LOG_LEVEL_INFO)
		gdao.PrintSql(nil, "test")
		r.Equal("SQL: %s", log.msg)
		r.Equal("test", log.args[0])
	}
	{
		log := &MockLogger{}
		gdao.LogCfg(log, gdao.LOG_LEVEL_INFO)
		gdao.PrintAffected(nil, 2)
		r.Equal("Affected: %d", log.msg)
		r.Equal(int64(2), log.args[0])
	}
}

func TestPrintWarn(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	gdao.LogCfg(log, 0)
	gdao.PrintWarn(nil, errors.New("warn"))
	r.Equal("warn", log.msg)
}

func TestPrintError(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	gdao.LogCfg(log, 0)
	gdao.PrintError(nil, errors.New("error"))
	r.Equal("error", log.msg)
}
