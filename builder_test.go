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
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
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
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
		b.Write(b.Pp("$"))
		r.Equal("$1", b.Sql())
		b.Write(b.Pp("$"))
		r.Equal("$1$2", b.Sql())
	}})
}

func TestBuilder_SetOk(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
		r.True(b.Ok())
		b.SetOk(false)
		r.False(b.Ok())
	}})
}

func TestBuilder_SetError(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
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
		Status:  gdao.Ptr[int8](3),
		Level:   gdao.Ptr[int32](10),
		Address: gdao.Ptr("address"),
		Phone:   gdao.Ptr("56789"),
	}}, BuildSql: func(b *gdao.Builder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.Columns, b.Columns(false))
		r.Equal([]string{"id", "name", "age", "email", "status", "level", "create_at"}, b.Columns(false, []string{"address", "phone"}...))
		r.Equal([]string{"address", "phone", "status", "level"}, b.Columns(true))
		r.Equal([]string{"status", "level"}, b.Columns(true, []string{"address", "phone"}...))
	}})
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
		r.Empty(b.Columns(true))
	}})
}

func TestBuilder_AutoColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.AutoIncrementColumns, b.AutoColumns())
	}})
}

func TestBuilder_Entity(t *testing.T) {
	r := require.New(t)
	u := &User{
		Status:  gdao.Ptr[int8](3),
		Level:   gdao.Ptr[int32](10),
		Address: gdao.Ptr("address"),
		Phone:   gdao.Ptr("56789"),
	}
	u2 := &User{
		Status:  gdao.Ptr[int8](2),
		Level:   gdao.Ptr[int32](2),
		Address: gdao.Ptr("addr"),
		Phone:   gdao.Ptr("2325325"),
	}
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{u, u2}, BuildSql: func(b *gdao.Builder[User]) {
		r.Equal(u, b.Entity())
	}})
}

func TestBuilder_EachColumn(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{{
		Status:  gdao.Ptr[int8](3),
		Level:   gdao.Ptr[int32](10),
		Address: gdao.Ptr("address"),
		Phone:   gdao.Ptr("56789"),
	}}, BuildSql: func(b *gdao.Builder[User]) {
		b.EachColumn(b.Entity(), nil, func(column string, value any) bool {
			return column != "phone"
		}, func(n int, column string, value any) {
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
		Status:  gdao.Ptr[int8](3),
		Level:   gdao.Ptr[int32](10),
		Address: gdao.Ptr("address"),
		Phone:   gdao.Ptr("56789"),
	}
	dao, _ := mockUserDao(r)
	dao.Query(gdao.QueryReq[User]{Entities: []*User{u}, BuildSql: func(b *gdao.Builder[User]) {
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
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{Entities: []*Account{a}, BuildSql: func(b *gdao.Builder[Account]) {
		b.Repeat(6, b.SepFix("(", ",", ")", false), func(i int) bool {
			return i != 2 && i != 4
		}, func(n, i int) {
			b.Write(strconv.Itoa(n)).Write("-").Write(strconv.Itoa(i))
		})
		r.Equal("(1-0,2-1,3-3,4-5)", b.Sql())
	}})
}

func TestBuilder_WriteColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.Builder[Account]) {
		b.WriteColumns("id", "", "user_id")
		r.Equal("id, user_id", b.Sql())

		b.WriteColumns()
		r.Equal("id, user_idid, other_id, user_id, status, balance, licence_file", b.Sql())
	}})
}

func TestBuilder_EachEntity(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	a2 := &Account{
		UserId:  gdao.Ptr[int32](2),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](200),
	}
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{Entities: []*Account{a, nil, a2}, BuildSql: func(b *gdao.Builder[Account]) {
		b.EachEntity(b.Sep(","), func(entity *Account) bool {
			return entity != nil
		}, func(n int, entity *Account) {
			switch n {
			case 1:
				r.Equal(a, entity)
			case 2:
				r.Equal(a2, entity)
			}
		})
	}})
}
