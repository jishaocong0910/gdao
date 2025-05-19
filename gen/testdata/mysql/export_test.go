package mysql

import (
	"context"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

type Logger struct {
}

func (d Logger) Debugf(ctx context.Context, msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (d Logger) Infof(ctx context.Context, msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (d Logger) Warnf(ctx context.Context, msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func (d Logger) Errorf(ctx context.Context, msg string, args ...interface{}) {
	log.Printf(msg, args...)
}

func MockMysqlBaseDao[T any](t *testing.T) (*baseDao[T], sqlmock.Sqlmock) {
	r := require.New(t)
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := newBaseDao[T](gdao.NewDaoReq{DB: db}, "user")
	gdao.LogCfg(Logger{}, "info")
	return dao, mock
}
