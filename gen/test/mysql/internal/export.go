/*
Copyright 2024-present jishaocong0910

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

func (d Logger) Debugf(ctx context.Context, msg string, args ...interface{}) { // coverage-ignore
	log.Printf(msg, args...)
}

func (d Logger) Infof(ctx context.Context, msg string, args ...interface{}) { // coverage-ignore
	log.Printf(msg, args...)
}

func (d Logger) Warnf(ctx context.Context, msg string, args ...interface{}) { // coverage-ignore
	log.Printf(msg, args...)
}

func (d Logger) Errorf(ctx context.Context, msg string, args ...interface{}) { // coverage-ignore
	log.Printf(msg, args...)
}

func MockBaseDao[T any](r *require.Assertions, table string) (*baseDao[T], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := newBaseDao[T](gdao.NewDaoReq{}, table)
	gdao.Config(gdao.Cfg{DefaultDB: db, Logger: Logger{}, SqlLogLevel: gdao.SqlLogLevel_.INFO})
	return dao, mock
}

func WriteCondition[T any](c condition, b *gdao.DaoSqlBuilder[T]) {
	c.write(b.BaseSqlBuilder__)
}
