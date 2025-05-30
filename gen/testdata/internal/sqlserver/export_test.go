package dao

import (
	"context"
	"log"

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

func MockSqlserverBaseDao[T any](r *require.Assertions, table string) (*baseDao[T], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := newBaseDao[T](gdao.NewDaoReq{DB: db}, table)
	gdao.LogCfg(Logger{}, "info", false)
	return dao, mock
}

func WriteCondition[T any](c condition, b *gdao.Builder[T]) {
	cb := getConditionBuilder(b)
	c.write(cb)
}
