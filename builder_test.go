package gdao_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

func TestBuilder_Arg(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(t)
	dao.Query(gdao.QueryReq[User]{nil, nil, nil, func(b *gdao.Builder[User]) (ok bool) {
		b.Arg("a")
		r.Equal(1, len(b.Args()))
		r.Contains(b.Args(), "a")

		b.Arg("B")
		r.Equal(2, len(b.Args()))
		r.Contains(b.Args(), "a", "B")
		return false
	}})
}

func TestBuilder_Pp(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(t)
	dao.Query(gdao.QueryReq[User]{nil, nil, nil, func(b *gdao.Builder[User]) (ok bool) {
		b.Write(b.Pp("$"))
		r.Equal("$1", b.Sql())
		b.Write(b.Pp("$"))
		r.Equal("$1$2", b.Sql())
		return false
	}})
}

func TestBuilder_Columns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(t)
	dao.Query(gdao.QueryReq[User]{nil, nil, nil, func(b *gdao.Builder[User]) (ok bool) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.Columns, b.Columns())
		return false
	}})
}

func TestBuilder_AutoColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(t)
	dao.Query(gdao.QueryReq[User]{nil, nil, nil, func(b *gdao.Builder[User]) (ok bool) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.AutoIncrementColumns, b.AutoColumns())
		return false
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
	dao, _ := mockUserDao(t)
	dao.Query(gdao.QueryReq[User]{nil, nil, []*User{u, u2}, func(b *gdao.Builder[User]) (ok bool) {
		r.Equal(u, b.Entity())
		return false
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
	dao, _ := mockUserDao(t)
	dao.Query(gdao.QueryReq[User]{nil, nil, []*User{u}, func(b *gdao.Builder[User]) (ok bool) {
		r.Nil(b.ColumnValue(nil, ""))
		r.Nil(b.ColumnValue(b.Entity(), ""))
		r.Nil(b.ColumnValue(b.Entity(), "name"))
		cvs1, cvs2 := b.ColumnValuesAt(nil, false)
		r.Nil(cvs1)
		r.Nil(cvs2)
		r.Equal(int8(3), reflect.ValueOf(b.ColumnValue(b.Entity(), "status")).Elem().Interface())
		r.Equal("address", reflect.ValueOf(b.ColumnValue(b.Entity(), "address")).Elem().Interface())
		return false
	}})
}

func TestBuilder_ColumnValues(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	dao, _ := mockAccountDao(t)
	dao.Query(gdao.QueryReq[Account]{nil, nil, []*Account{a}, func(b *gdao.Builder[Account]) (ok bool) {
		cvs, _ := b.ColumnValues(false)
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

		cvs2, cvs3 := b.ColumnValues(true, "user_id")
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
		for i, cv := range cvs3 {
			switch i {
			case 0:
				r.Equal("user_id", cv.Column)
				r.Equal(int32(1), reflect.ValueOf(cv.Value).Elem().Interface())
			case 1:
				r.Equal("id", cv.Column)
				r.Equal(nil, cv.Value)
			case 2:
				r.Equal("other_id", cv.Column)
				r.Equal(nil, cv.Value)
			}
		}

		return false
	}})
}

func TestBuilder_Repeat(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	dao, _ := mockAccountDao(t)
	dao.Query(gdao.QueryReq[Account]{nil, nil, []*Account{a}, func(b *gdao.Builder[Account]) (ok bool) {
		b.Repeat(6, b.SepFix("(", ",", ")"), func(i int) bool {
			return i != 2 && i != 4
		}, func(n, i int) {
			b.Write(strconv.Itoa(i))
		})
		r.Equal("(0,1,2,3)", b.Sql())
		return false
	}})
}

func TestBuilder_WriteCommaColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(t)
	dao.Query(gdao.QueryReq[Account]{nil, nil, nil, func(b *gdao.Builder[Account]) (ok bool) {
		b.WriteCommaColumns("id", "", "user_id")
		r.Equal("id,user_id", b.Sql())

		b.WriteCommaColumns()
		r.Equal("id,user_idid,other_id,user_id,status,balance", b.Sql())
		return false
	}})
}

func TestBuilder_EachColumnName(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(t)
	dao.Query(gdao.QueryReq[Account]{nil, nil, nil, func(b *gdao.Builder[Account]) (ok bool) {
		b.EachColumnName([]string{"id", "", "user_id", "status", "balance"}, b.Sep(","), func(n, i int, column string) {
			b.Write(strconv.Itoa(i)).Write("-").Write(column)
		}, "user_id", "balance")
		r.Equal("0-id,1-status", b.Sql())
		return false
	}})
}

func TestBuilder_EachColumnValues(t *testing.T) {
	r := require.New(t)
	a := &Account{
		UserId:  gdao.Ptr[int32](1),
		Status:  gdao.Ptr[int8](1),
		Balance: gdao.Ptr[int64](100),
	}
	dao, _ := mockAccountDao(t)
	dao.Query(gdao.QueryReq[Account]{nil, nil, []*Account{a}, func(b *gdao.Builder[Account]) (ok bool) {
		cvs, _ := b.ColumnValues(false)
		b.EachColumnValues(cvs, b.SepFix("(", ",", ")"), func(column string, value any) {
			b.Write(column)
		})
		r.Equal("(id,other_id,user_id,status,balance)", b.Sql())
		return false
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
	dao, _ := mockAccountDao(t)
	dao.Query(gdao.QueryReq[Account]{nil, nil, []*Account{a, nil, a2}, func(b *gdao.Builder[Account]) (ok bool) {
		b.EachEntity(b.Sep(","), func(n, i int, entity *Account) {
			switch i {
			case 0:
				r.Equal(a, entity)
			case 1:
				r.Equal(a2, entity)
			}
		})
		return false
	}})
}
