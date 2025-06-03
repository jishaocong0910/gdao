package dao_test

import (
	"github.com/jishaocong0910/gdao/gen/testdata/internal/sqlite"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/stretchr/testify/require"
)

type User struct {
	Id       *int32     `gdao:"column=id;auto"`
	Name     *string    `gdao:"column=name"`
	Age      *int32     `gdao:"column=age"`
	Address  *string    `gdao:"column=address"`
	Phone    *string    `gdao:"column=phone"`
	Email    *string    `gdao:"column=email"`
	Status   *int8      `gdao:"column=status"`
	Level    *int32     `gdao:"column=level"`
	CreateAt *time.Time `gdao:"column=create_at"`
}

func TestNewBaseDaoPanic(t *testing.T) {
	r := require.New(t)
	r.PanicsWithValue(`parameter "table" must not be blank`, func() {
		dao.MockSqliteBaseDao[User](r, "")
	})
}

func TestBaseDao_List(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`SELECT id,name FROM user WHERE status=\? ORDER BY name ASC,address DESC LIMIT 10 OFFSET 20 FOR UPDATE`).
			ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "lucy").AddRow(2, "nick"))
		list, err := d.List(dao.ListReq{
			SelectColumns: []string{"id", "name"},
			Condition:     dao.And().Eq("status", 4),
			OrderBy:       dao.OrderBy().Asc("name").Desc("address"),
			Pagination:    dao.Page(3, 10),
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Len(list, 2)
		r.Equal(int32(1), *list[0].Id)
		r.Equal("lucy", *list[0].Name)
		r.Equal(int32(2), *list[1].Id)
		r.Equal("nick", *list[1].Name)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		list, err := d.List(dao.ListReq{
			SelectColumns: []string{"id", "name"},
			Condition:     dao.And(),
			OrderBy:       dao.OrderBy().Asc("name").Desc("address"),
			Pagination:    dao.Page(3, 10),
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NotNil(list)
		r.Len(list, 0)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")

		list, err := d.List(dao.ListReq{
			SelectColumns: []string{"id", "name"},
			OrderBy:       dao.OrderBy().Asc("name").Desc("address"),
			Pagination:    dao.Page(3, 10),
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NotNil(list)
		r.Len(list, 0)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestBaseDao_Get(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`SELECT id,phone FROM user WHERE status=\? LIMIT 1 FOR UPDATE`).
			ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "lucy"))

		get, err := d.Get(dao.GetReq{
			SelectColumns: []string{"id", "phone"},
			Condition:     dao.And().Eq("status", 4),
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int32(1), *get.Id)
		r.Equal("lucy", *get.Name)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")

		get, err := d.Get(dao.GetReq{
			SelectColumns: []string{"id", "phone"},
			Condition:     dao.And(),
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Nil(get)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")

		get, err := d.Get(dao.GetReq{
			SelectColumns: []string{"id", "phone"},
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Nil(get)
	}
}

func TestBaseDao_Insert(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`INSERT INTO user\(name,phone,email,level\) VALUES\(\?,\?,\?,NULL\)`).
			ExpectExec().WithArgs("abc", "12345", "email").WillReturnResult(sqlmock.NewResult(7, 1))

		u := &User{
			Name:  gdao.Ptr("abc"),
			Phone: gdao.Ptr("12345"),
			Email: gdao.Ptr("email"),
		}
		affected, err := d.Insert(dao.InsertReq[User]{
			Entity:         u,
			SetNullColumns: []string{"level"},
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Equal(int32(7), *u.Id)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`INSERT INTO user\(id\,name\,age\,address\,phone\,email\,status\,level\,create_at\) VALUES\(\?\,\?\,\?\,\?\,\?\,\?\,\?\,\?\,\?\)`).
			ExpectExec().WithArgs(nil, "abc", nil, nil, "12345", "email", nil, nil, nil).WillReturnResult(sqlmock.NewResult(7, 1))

		u := &User{
			Name:  gdao.Ptr("abc"),
			Phone: gdao.Ptr("12345"),
			Email: gdao.Ptr("email"),
		}
		affected, err := d.Insert(dao.InsertReq[User]{
			Entity:         u,
			InsertAll:      true,
			SetNullColumns: []string{"level"},
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Equal(int32(7), *u.Id)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")

		u := &User{}
		affected, err := d.Insert(dao.InsertReq[User]{
			Entity: u,
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(0), affected)
	}
}

func TestBaseDao_InsertBatch(t *testing.T) {
	r := require.New(t)
	d, mock := dao.MockSqliteBaseDao[User](r, "user")
	mock.ExpectPrepare(`INSERT INTO user\(name,phone,email\) VALUES\(\?,\?,\?\),\(\?,\?,\?\)`).
		ExpectExec().WithArgs("abc", "12345", nil, "def", "6789", nil).WillReturnResult(sqlmock.NewResult(8, 2))

	u := &User{
		Name:  gdao.Ptr("abc"),
		Phone: gdao.Ptr("12345"),
	}
	u2 := &User{
		Name:  gdao.Ptr("def"),
		Phone: gdao.Ptr("6789"),
	}
	affected, err := d.InsertBatch(dao.InsertBatchReq[User]{
		Entities:       []*User{u, u2},
		IgnoredColumns: []string{"id", "age", "address", "status", "level", "create_at"},
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(7), *u.Id)
	r.Equal(int32(8), *u2.Id)
}

func TestBaseDao_Update(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET address=\?,status=\?,level=\?,email=NULL,phone=NULL WHERE age=\?`).
			ExpectExec().WithArgs("addr", 2, 10, 20).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Status:  gdao.Ptr[int8](2),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("addr"),
		}
		affected, err := d.Update(dao.UpdateReq[User]{
			Entity:         u,
			SetNullColumns: []string{"email", "phone"},
			Condition:      dao.And().Eq("age", 20),
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET id=\?,name=\?,address=\?,phone=\?,email=\?,level=\?,create_at=\? WHERE status=\? AND age IS NULL AND address IS NULL AND level=\?`).
			ExpectExec().WithArgs(nil, nil, "addr", nil, nil, 10, nil, 2, 9).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Status:  gdao.Ptr[int8](2),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("addr"),
		}
		affected, err := d.Update(dao.UpdateReq[User]{
			Entity:         u,
			UpdateAll:      true,
			SetNullColumns: []string{"email", "phone"},
			WhereColumns:   []string{"status", "age"},
			Condition:      dao.And().IsNull("address").Eq("level", 9),
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")

		u := &User{
			Status:  gdao.Ptr[int8](2),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("addr"),
		}
		affected, err := d.Update(dao.UpdateReq[User]{
			Entity:         u,
			UpdateAll:      true,
			SetNullColumns: []string{"email", "phone"},
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(0), affected)
	}
}

func TestBaseDao_UpdateBatch(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET name=CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END,address=CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END,phone=CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END,email=CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END WHERE id IN\(\?,\?,\?\)`).
			ExpectExec().WithArgs(1, "name1", 2, "name2", 3, "name3", 1, nil, 2, nil, 3, nil, 1, "phone1", 2, "phone2", 3, "phone3", 1, "email1", 2, "email2", 3, "email3", 1, 2, 3).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Id:    gdao.Ptr[int32](1),
			Name:  gdao.Ptr("name1"),
			Phone: gdao.Ptr("phone1"),
			Email: gdao.Ptr("email1"),
		}
		u2 := &User{
			Id:    gdao.Ptr[int32](2),
			Name:  gdao.Ptr("name2"),
			Phone: gdao.Ptr("phone2"),
			Email: gdao.Ptr("email2"),
		}
		u3 := &User{
			Id:    gdao.Ptr[int32](3),
			Name:  gdao.Ptr("name3"),
			Phone: gdao.Ptr("phone3"),
			Email: gdao.Ptr("email3"),
		}
		affected, err := d.UpdateBatch(dao.UpdateBatchReq[User]{
			Entities:       []*User{u, u2, u3},
			IgnoredColumns: []string{"status", "age", "level", "create_at"},
			WhereColumn:    "id",
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET name=CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END WHERE id IN\(\?,\?,\?\)`).
			ExpectExec().WithArgs(1, "name1", 2, "name2", 3, "name3", 1, 2, 3).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Id:    gdao.Ptr[int32](1),
			Name:  gdao.Ptr("name1"),
			Phone: gdao.Ptr("phone1"),
			Email: gdao.Ptr("email1"),
		}
		u2 := &User{
			Id:    gdao.Ptr[int32](2),
			Name:  gdao.Ptr("name2"),
			Phone: gdao.Ptr("phone2"),
			Email: gdao.Ptr("email2"),
		}
		u3 := &User{
			Id:    gdao.Ptr[int32](3),
			Name:  gdao.Ptr("name3"),
			Phone: gdao.Ptr("phone3"),
			Email: gdao.Ptr("email3"),
		}
		affected, err := d.UpdateBatch(dao.UpdateBatchReq[User]{
			Entities:       []*User{u, u2, u3},
			SetColumns:     []string{"name", "phone"},
			IgnoredColumns: []string{"phone"},
			WhereColumn:    "id",
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")

		u := &User{
			Id:    gdao.Ptr[int32](1),
			Name:  gdao.Ptr("name1"),
			Phone: gdao.Ptr("phone1"),
			Email: gdao.Ptr("email1"),
		}
		affected, err := d.UpdateBatch(dao.UpdateBatchReq[User]{
			Entities:       []*User{u},
			SetColumns:     []string{"name", "phone"},
			IgnoredColumns: []string{"phone"},
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(0), affected)
	}
}

func TestBaseDao_Delete(t *testing.T) {
	r := require.New(t)
	d, mock := dao.MockSqliteBaseDao[User](r, "user")
	mock.ExpectPrepare(`DELETE FROM user WHERE status=\?`).
		ExpectExec().WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 3))

	affected, err := d.Delete(dao.DeleteReq{
		Condition: dao.And().Eq("status", 1),
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(3), affected)
}

func TestCondition(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`c1=\? AND c2=\?`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c := dao.And().Eq("c1", 1).And(nil)
			c2 := dao.Or().Eq("c2", 2).Or(nil)
			c = c.And(c2)
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`c1=\? AND c2<>\? AND c3>\? AND c4<\? AND c5>=\? AND c6<=\? AND c7 LIKE \? AND c8 LIKE \? AND c9 LIKE \? AND c10 IN\(\?,\?,\?\) AND c11 BETWEEN \? AND \? AND c12 IS NULL AND c13 IS NOT NULL`).
			ExpectQuery().WithArgs(1, 2, 3, 4, 5, 6, "%abc%", "abc%", "%abc", 1, 2, 3, 1, 3).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c := dao.And().Eq("c1", 1).
				NotEq("c2", 2).
				Gt("c3", 3).
				Lt("c4", 4).
				GtEq("c5", 5).
				LtEq("c6", 6).
				Like("c7", "abc").
				LikeLeft("c8", "abc").
				LikeRight("c9", "abc").
				In("c10", 1, 2, 3).
				Between("c11", 1, 3).
				IsNull("c12").
				IsNotNull("c13")
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`NOT c1=\? AND NOT \(c2=\? AND c3=\?\) AND c4=\? AND \(c5=\? OR c6=\?\) AND NOT c7=\? AND NOT \(c8=\? OR c9=\?\)`).
			ExpectQuery().WithArgs(1, 2, 3, 4, 5, 6, 7, 8, 9).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c := dao.NotAnd().Eq("c1", 1)
			c2 := dao.NotAnd().Eq("c2", 2).Eq("c3", 3)
			c3 := dao.Or().Eq("c4", 4)
			c4 := dao.Or().Eq("c5", 5).Eq("c6", 6)
			c5 := dao.NotOr().Eq("c7", 7)
			c6 := dao.NotOr().Eq("c8", 8).Eq("c9", 9)
			c = c.And(c2).And(c3).And(c4).And(c5).And(c6)
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
	{
		d, mock := dao.MockSqliteBaseDao[User](r, "user")
		mock.ExpectPrepare(`\(c1=\? OR c2=\?\) AND c3=\? AND c4=\?`).
			ExpectQuery().WithArgs(1, 2, 3, 4).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c := dao.Or().Eq("c1", 1).Eq("c2", 2)
			c2 := dao.Or().Eq("c3", 3)
			c3 := dao.Or().Eq("c4", 4)
			c = c.And(c2).Or(c3)
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
}
