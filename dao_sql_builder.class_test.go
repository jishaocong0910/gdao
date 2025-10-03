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
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Arg(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		b.SetArgs("a")
		r.Equal(1, len(b.Args()))
		r.Contains(b.Args(), "a")

		b.SetArgs("B")
		r.Equal(2, len(b.Args()))
		r.Contains(b.Args(), "a", "B")
	}})
}

func TestBuilder_Pp(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		b.Write(b.Pp("$"))
		r.Equal("$1", b.Sql())
		b.Write(b.Pp("$"))
		r.Equal("$1$2", b.Sql())
	}})
}

func TestBuilder_SetOk(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		r.True(b.Ok())
		b.SetOk(false)
		r.False(b.Ok())
	}})
}

func TestBuilder_SetError(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		b.SetError(errors.New("this is an error"))
		r.EqualError(b.Error(), "this is an error")
		r.False(b.Ok())
		b.SetOk(true)
		r.False(b.Ok())
	}})
}

func TestBuilder_Columns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
		Phone:   gdao.P("56789"),
	}}, BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.Columns, b.Columns(false))
		r.Equal([]string{"id", "name", "age", "email", "status", "level", "create_at"}, b.Columns(false, []string{"address", "phone"}...))
		r.Equal([]string{"address", "phone", "status", "level"}, b.Columns(true))
		r.Equal([]string{"status", "level"}, b.Columns(true, []string{"address", "phone"}...))
	}})
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		r.Empty(b.Columns(true))
	}})
}

func TestBuilder_AutoColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.AutoIncrementColumns, b.AutoColumns())
	}})
}

func TestBuilder_Entity(t *testing.T) {
	r := require.New(t)
	u := &User{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
		Phone:   gdao.P("56789"),
	}
	u2 := &User{
		Status:  gdao.P[int8](2),
		Level:   gdao.P[int32](2),
		Address: gdao.P("addr"),
		Phone:   gdao.P("2325325"),
	}
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{u, u2}, BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		r.Equal(u, b.Entity())
	}})
}

func TestBuilder_EachColumn(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
	}}, BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		b.EachColumn(b.Entity(), nil, func(n int, column string, value any) {
			switch n {
			case 1:
				r.Equal("address", reflect.ValueOf(value).Elem().Interface())
			case 2:
				r.Equal(int8(3), reflect.ValueOf(value).Elem().Interface())
			case 3:
				r.Equal(int32(10), reflect.ValueOf(value).Elem().Interface())
			}
		}, b.Columns(true)...)
	}})
}

func TestBuilder_ColumnValue(t *testing.T) {
	r := require.New(t)
	u := &User{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
		Phone:   gdao.P("56789"),
	}
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{u}, BuildSql: func(b *gdao.DaoSqlBuilder[User]) {
		r.Nil(b.ColumnValue(nil, ""))
		r.Nil(b.ColumnValue(b.Entity(), ""))
		r.Nil(b.ColumnValue(b.Entity(), "name"))
		r.Equal(int8(3), reflect.ValueOf(b.ColumnValue(b.Entity(), "status")).Elem().Interface())
		r.Equal("address", reflect.ValueOf(b.ColumnValue(b.Entity(), "address")).Elem().Interface())
	}})
}

func TestBuilder_Repeat(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.P[int32](1),
		Status:  gdao.P[int8](1),
		Balance: gdao.P[int64](100),
	}
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{Entities: []*Account{a}, BuildSql: func(b *gdao.DaoSqlBuilder[Account]) {
		b.Repeat(6, b.SepFix("(", ",", ")", false), func(i int) bool {
			return i != 2 && i != 4
		}, func(n, i int) {
			b.Write(strconv.Itoa(n))
			b.Write("-")
			b.Write(strconv.Itoa(i))
		})
		r.Equal("(1-0,2-1,3-3,4-5)", b.Sql())
	}})
}

func TestBuilder_WriteColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.DaoSqlBuilder[Account]) {
		b.WriteColumns("id", "", "user_id")
		r.Equal("id, user_id", b.Sql())

		b.WriteColumns()
		r.Equal("id, user_idid, other_id, user_id, status, balance, licence_file", b.Sql())
	}})
}

func TestBuilder_EachEntity(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.P[int32](1),
		Status:  gdao.P[int8](1),
		Balance: gdao.P[int64](100),
	}
	a2 := &Account{
		UserId:  gdao.P[int32](2),
		Status:  gdao.P[int8](1),
		Balance: gdao.P[int64](200),
	}
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{Entities: []*Account{a, a2}, BuildSql: func(b *gdao.DaoSqlBuilder[Account]) {
		b.EachEntity(b.Sep(","), func(n int, entity *Account) {
			switch n {
			case 1:
				r.Equal(a, entity)
			case 2:
				r.Equal(a2, entity)
			}
		})
	}})
}
