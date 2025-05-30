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

func (d *MockLogger) Infof(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Warnf(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Errorf(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func TestPrintSql(t *testing.T) {
	r := require.New(t)
	{
		log := &MockLogger{}
		gdao.LogCfg(log, "debug", false)
		gdao.PrintSql(nil, "UPDATE user SET status=?,phone=?,email=? WHERE level=?)", []any{2, nil, (*int)(nil), gdao.Ptr(1)}, 15, errors.New("error"))
		r.Equal(`SQL: %s, args: %v, affected: %d, error: %+v`, log.msg)
		r.Len(log.args, 4)
		r.Equal("UPDATE user SET status=?,phone=?,email=? WHERE level=?)", log.args[0])
		args := log.args[1].([]any)
		r.Len(args, 4)
		r.Equal(2, args[0])
		r.Equal(nil, args[1])
		r.Equal(nil, args[2])
		r.Equal(1, args[3])
		r.Equal(int64(15), log.args[2])
		r.EqualError(log.args[3].(error), "error")
	}
	{
		log := &MockLogger{}
		gdao.LogCfg(log, "debug", true)
		gdao.PrintSql(nil, `  
SELECT *
FROM
 user`, nil, -1, nil)
		r.Equal("SQL: %s", log.msg)
		r.Equal("  SELECT * FROM user", log.args[0])
	}
}

func TestPrintWarn(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	gdao.LogCfg(log, "debug", false)
	gdao.PrintWarn(nil, errors.New("warn"))
	r.Equal("warn", log.msg)
}

func TestPrintError(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	gdao.LogCfg(log, "debug", false)
	gdao.PrintError(nil, errors.New("error"))
	r.Equal("error", log.msg)
}
