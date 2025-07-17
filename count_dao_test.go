package gdao_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
	"testing"
)

func mockCountDao(r *require.Assertions) (*gdao.CountDao, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.NewCountDao(gdao.NewCountDaoReq{DB: db})
	return dao, mock
}

func TestCountDao_Count(t *testing.T) {
	r := require.New(t)
	{
		dao, mock := mockCountDao(r)
		mock.ExpectPrepare(`SELECT count\(\*\) FROM user`).ExpectQuery().WillReturnRows(mock.NewRows([]string{"c"}).AddRow(20))

		count, _, err := dao.Count(gdao.CountReq{BuildSql: func(b *gdao.CountBuilder) {
			b.Write("SELECT count(*) FROM user")
		}})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(20, count.Int())
	}
	{
		dao, mock := mockCountDao(r)
		mock.ExpectPrepare(`SELECT id, count\(\*\) count FROM user GROUP BY id`).ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "count"}).AddRow(1, 12).AddRow(2, 20))

		_, list, err := dao.Count(gdao.CountReq{BuildSql: func(b *gdao.CountBuilder) {
			b.Write("SELECT id, count(*) count FROM user GROUP BY id")
		}})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Len(list, 2)
		r.Equal(12, list[0].Int())
		r.Equal(20, list[1].Int())
	}
}

func TestCount(t *testing.T) {
	r := require.New(t)
	count := &gdao.Count{Value: gdao.Ptr(int64(20))}

	r.Equal(int8(20), count.Int8())
	r.Equal(int16(20), count.Int16())
	r.Equal(int32(20), count.Int32())
	r.Equal(int64(20), count.Int64())
	r.Equal(true, count.Bool())
	r.Equal(20, *count.IntPtr())
	r.Equal(int8(20), *count.Int8Ptr())
	r.Equal(int16(20), *count.Int16Ptr())
	r.Equal(int32(20), *count.Int32Ptr())
	r.Equal(int64(20), *count.Int64Ptr())
	r.Equal(true, *count.BoolPtr())

	count = nil
	r.Equal(0, count.Int())
	r.Equal(int8(0), count.Int8())
	r.Equal(int16(0), count.Int16())
	r.Equal(int32(0), count.Int32())
	r.Equal(int64(0), count.Int64())
	r.Equal(false, count.Bool())
	r.Nil(count.IntPtr())
	r.Nil(count.Int8Ptr())
	r.Nil(count.Int16Ptr())
	r.Nil(count.Int32Ptr())
	r.Nil(count.Int64Ptr())
	r.Nil(count.BoolPtr())

}
