package gdao_test

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
	"testing"
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
		affected, err := userDao.Exec(gdao.ExecReq[User]{Ctx: gdao.SetTx(nil, tx), BuildSql: func(b *gdao.Builder[User]) {
			b.Write("UPDATE user set status=1 WHERE id=?", 1)
		}})
		tx.Commit()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE user set status=1 WHERE id=\?`).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			_, err := userDao.Exec(gdao.ExecReq[User]{
				Ctx: ctx,
				BuildSql: func(b *gdao.Builder[User]) {
					b.Write("UPDATE user set status=1 WHERE id=?", 1)
				},
			})
			return err
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			return errors.New("error")
		})
		r.EqualError(err, "error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			panic(errors.New("panic error"))
		})
		r.EqualError(err, "panic error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, func(ctx context.Context) error {
			panic(1)
		})
		r.EqualError(err, "1")
		r.NoError(mock.ExpectationsWereMet())
	}
}
