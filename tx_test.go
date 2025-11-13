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

package gdao_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

func TestTx(t *testing.T) {
	r := require.New(t)
	{
		userDao, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE user set status=1 WHERE id=\?`).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		tx, err := userDao.DB().Begin()
		r.NoError(err)
		affected, err := userDao.Exec(gdao.ExecReq[User]{Ctx: gdao.SetTx(nil, tx), BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("UPDATE user set status=1 WHERE id=?", 1)
		}})
		tx.Commit()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		userDao, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE user set status=1 WHERE id=\?`).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			_, err := userDao.Exec(gdao.ExecReq[User]{
				Ctx: ctx,
				BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
					b.Write("UPDATE user set status=1 WHERE id=?", 1)
				},
			})
			return err
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		_, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			return errors.New("error")
		})
		r.EqualError(err, "error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		_, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			panic(errors.New("panic error"))
		})
		r.EqualError(err, "panic error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		_, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			panic(1)
		})
		r.EqualError(err, "1")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		_, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectRollback()
		r.PanicsWithError(`test panic`, func() {
			gdao.Tx(nil, func(ctx context.Context) error {
				panic(errors.New("test panic"))
			}, gdao.WithMust())
		})
	}

}
