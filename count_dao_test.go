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
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
	"testing"
)

func mockCountDao(r *require.Assertions) (*gdao.CountDao, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.CountDaoBuilder().DB(db).Build()
	return dao, mock
}

func TestCountDao_Count(t *testing.T) {
	r := require.New(t)
	{
		dao, mock := mockCountDao(r)
		mock.ExpectPrepare(`SELECT count\(\*\) FROM user`).ExpectQuery().WillReturnRows(mock.NewRows([]string{"c"}).AddRow(20))

		count, err := dao.Count().BuildSql(func(b *gdao.CountBuilder) {
			b.Write("SELECT count(*) FROM user")
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(20, count.Int())
	}
	{
		dao, mock := mockCountDao(r)
		mock.ExpectPrepare(`SELECT count\(\*\) count FROM user GROUP BY id`).ExpectQuery().WillReturnRows(mock.NewRows([]string{"count"}).AddRow(12).AddRow(20))

		_, err := dao.Count().BuildSql(func(b *gdao.CountBuilder) {
			b.Write("SELECT count(*) count FROM user GROUP BY id")
		}).Do()
		r.NoError(mock.ExpectationsWereMet())
		r.Error(err, "returns more than one row")
	}
	{
		dao, mock := mockCountDao(r)
		mock.ExpectPrepare(`SELECT id, count\(\*\) count FROM user GROUP BY id`).ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "count"}).AddRow(1, 12).AddRow(2, 20))

		_, err := dao.Count().BuildSql(func(b *gdao.CountBuilder) {
			b.Write("SELECT id, count(*) count FROM user GROUP BY id")
		}).Do()
		r.NoError(mock.ExpectationsWereMet())
		r.Error(err, "returns more than one column")
	}
}

func TestCount(t *testing.T) {
	r := require.New(t)
	count := &gdao.Count{Value: gdao.P(int64(20))}

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
