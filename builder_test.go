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
		b.Arg("a")
		r.Equal(1, len(b.Args()))
		r.Contains(b.Args(), "a")

		b.Arg("B")
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
	dao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.Columns, b.Columns())
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
		cvs := b.ColumnValuesAt(nil, false)
		r.Nil(cvs)
		r.Equal(int8(3), reflect.ValueOf(b.ColumnValue(b.Entity(), "status")).Elem().Interface())
		r.Equal("address", reflect.ValueOf(b.ColumnValue(b.Entity(), "address")).Elem().Interface())
	}})
}

func TestBuilder_ColumnValues(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{Entities: []*Account{a}, BuildSql: func(b *gdao.Builder[Account]) {
		cvs := b.ColumnValues(false)
		for i, cv := range cvs {
			switch i {
			case 0:
				r.Equal("id", cv.Column)
				r.Equal(nil, cv.Value)
			case 1:
				r.Equal("other_id", cv.Column)
				r.Equal(nil, cv.Value)
			case 2:
				r.Equal("user_id", cv.Column)
				r.Equal(int32(1), reflect.ValueOf(cv.Value).Elem().Interface())
			case 3:
				r.Equal("status", cv.Column)
				r.Equal(int8(1), reflect.ValueOf(cv.Value).Elem().Interface())
			case 4:
				r.Equal("balance", cv.Column)
				r.Equal(int64(100), reflect.ValueOf(cv.Value).Elem().Interface())
			}
		}

		cvs2 := b.ColumnValues(true, "user_id")
		for i, cv := range cvs2 {
			switch i {
			case 0:
				r.Equal("status", cv.Column)
				r.Equal(int8(1), reflect.ValueOf(cv.Value).Elem().Interface())
			case 1:
				r.Equal("balance", cv.Column)
				r.Equal(int64(100), reflect.ValueOf(cv.Value).Elem().Interface())
			}
		}
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
		r.Equal("id, user_idid, other_id, user_id, status, balance", b.Sql())
	}})
}

func TestBuilder_EachColumnName(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.Builder[Account]) {
		b.EachColumnName([]string{"id", "", "user_id", "status", "balance"}, b.Sep(","), func(n, i int, column string) {
			b.Write(strconv.Itoa(n)).Write("-").Write(strconv.Itoa(i)).Write("-").Write(column)
		}, "user_id", "balance")
		r.Equal("1-0-id,2-3-status", b.Sql())
	}})
}

func TestBuilder_EachColumnValues(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	dao, _ := mockAccountDao(r)
	dao.Query(gdao.QueryReq[Account]{Entities: []*Account{a}, BuildSql: func(b *gdao.Builder[Account]) {
		cvs := b.ColumnValues(false)
		b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(column string, value any) {
			b.Write(column)
		})
		r.Equal("(id,other_id,user_id,status,balance)", b.Sql())
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
		b.EachEntity(b.Sep(","), func(n, i int, entity *Account) {
			switch i {
			case 0:
				r.Equal(a, entity)
			case 1:
				r.Equal(a2, entity)
			}
		})
	}})
}
