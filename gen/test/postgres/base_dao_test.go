package postgres_test

import (
	dao "github.com/jishaocong0910/gdao/gen/test/postgres/internal"
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
		dao.MockBaseDao[User](r, "")
	})
}

func TestBaseDao_List(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`SELECT id, name FROM user WHERE status = \$1 ORDER BY name ASC, address DESC LIMIT 10 OFFSET 3 FOR UPDATE`).
			ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "lucy").AddRow(2, "nick"))
		list, err := d.List(dao.ListReq{
			SelectColumns: dao.Columns("id", "name"),
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
}

func TestBaseDao_Get(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`SELECT id, phone FROM user WHERE status = \$1 LIMIT 1 FOR UPDATE`).
			ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "lucy"))

		get, err := d.Get(dao.GetReq{
			SelectColumns: dao.Columns("id", "phone"),
			Condition:     dao.And().Eq("status", 4),
			ForUpdate:     true,
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int32(1), *get.Id)
		r.Equal("lucy", *get.Name)
	}
}

func TestBaseDao_Insert(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`INSERT INTO user\(name, age, address, phone, status, create_at, level\) VALUES\(\$1, NULL, NULL, \$2, NULL, NULL, NULL\) RETURNING id`).
			ExpectQuery().WithArgs("abc", "12345").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

		u := &User{
			Name:  gdao.Ptr("abc"),
			Phone: gdao.Ptr("12345"),
			Email: gdao.Ptr("email"),
		}
		err := d.Insert(dao.InsertReq[User]{
			Entity:         u,
			InsertAll:      true,
			SetNullColumns: []string{"level"},
			IgnoredColumns: dao.Columns("email"),
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int32(7), *u.Id)
	}
}

func TestBaseDao_InsertBatch(t *testing.T) {
	r := require.New(t)
	d, mock := dao.MockBaseDao[User](r, "user")
	mock.ExpectPrepare(`INSERT INTO user\(name, phone, email\) VALUES\(\$1, \$2, \$3\), \(\$4, \$5, \$6\) RETURNING id`).
		ExpectQuery().WithArgs("abc", "12345", "email11", "def", "6789", "email22").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8).AddRow(9))

	u := &User{
		Name:  gdao.Ptr("abc"),
		Phone: gdao.Ptr("12345"),
		Email: gdao.Ptr("email11"),
	}
	u2 := &User{
		Name:  gdao.Ptr("def"),
		Phone: gdao.Ptr("6789"),
		Email: gdao.Ptr("email22"),
	}
	err := d.InsertBatch(dao.InsertBatchReq[User]{
		Entities: dao.Entities(u, u2),
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int32(8), *u.Id)
	r.Equal(int32(9), *u2.Id)
}

func TestBaseDao_Update(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET name = \$1, email = NULL, phone = NULL WHERE status = \$2 AND level = \$3 AND age = \$4`).
			ExpectExec().WithArgs("name", 2, 10, 20).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Name:    gdao.Ptr("name"),
			Address: gdao.Ptr("addr"),
			Status:  gdao.Ptr[int8](2),
			Level:   gdao.Ptr[int32](10),
		}
		affected, err := d.Update(dao.UpdateReq[User]{
			Entity:         u,
			SetNullColumns: dao.Columns("email", "phone"),
			IgnoredColumns: dao.Columns("address"),
			WhereColumns:   dao.Columns("status", "level"),
			Condition:      dao.And().Eq("age", 20),
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET name = \$1, age = NULL, address = \$2, phone = NULL, email = NULL, level = \$3, create_at = NULL WHERE id = \$4 AND status IS NULL`).
			ExpectExec().WithArgs("name", "addr", 10, 1).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Id:      gdao.Ptr(int32(1)),
			Name:    gdao.Ptr("name"),
			Address: gdao.Ptr("addr"),
			Level:   gdao.Ptr[int32](10),
		}
		affected, err := d.Update(dao.UpdateReq[User]{
			Entity:       u,
			UpdateAll:    true,
			WhereColumns: dao.Columns("id", "status"),
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
}

func TestBaseDao_UpdateBatch(t *testing.T) {
	r := require.New(t)
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET name = CASE id WHEN \$1 THEN \$2 WHEN \$3 THEN \$4 WHEN \$5 THEN \$6 END, phone = CASE id WHEN \$7 THEN \$8 WHEN \$9 THEN \$10 WHEN \$11 THEN \$12 END, state = NULL, level = NULL WHERE id IN\(\$13, \$14, \$15\)`).
			ExpectExec().WithArgs(1, "name1", 2, "name2", 3, "name3", 1, "phone1", 2, "phone2", 3, "phone3", 1, 2, 3).WillReturnResult(sqlmock.NewResult(0, 3))

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
			SetNullColumns: dao.Columns("state", "level"),
			IgnoredColumns: dao.Columns("email"),
			WhereColumn:    "id",
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`UPDATE user SET name = CASE id WHEN \$1 THEN \$2 WHEN \$3 THEN \$4 WHEN \$5 THEN \$6 END, age = CASE id WHEN \$7 THEN NULL WHEN \$8 THEN NULL WHEN \$9 THEN NULL END, address = CASE id WHEN \$10 THEN NULL WHEN \$11 THEN NULL WHEN \$12 THEN NULL END, phone = CASE id WHEN \$13 THEN \$14 WHEN \$15 THEN \$16 WHEN \$17 THEN \$18 END, email = CASE id WHEN \$19 THEN \$20 WHEN \$21 THEN \$22 WHEN \$23 THEN \$24 END, status = CASE id WHEN \$25 THEN NULL WHEN \$26 THEN NULL WHEN \$27 THEN NULL END, level = CASE id WHEN \$28 THEN NULL WHEN \$29 THEN NULL WHEN \$30 THEN NULL END, create_at = CASE id WHEN \$31 THEN NULL WHEN \$32 THEN NULL WHEN \$33 THEN NULL END WHERE id IN\(\$34, \$35, \$36\)`).
			ExpectExec().WithArgs(1, "name1", 2, "name2", 3, "name3", 1, 2, 3, 1, 2, 3, 1, "phone1", 2, "phone2", 3, "phone3", 1, "email1", 2, "email2", 3, "email3", 1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3).WillReturnResult(sqlmock.NewResult(0, 3))

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
			Entities:    []*User{u, u2, u3},
			UpdateAll:   true,
			WhereColumn: "id",
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
}

func TestBaseDao_Delete(t *testing.T) {
	r := require.New(t)
	d, mock := dao.MockBaseDao[User](r, "user")
	mock.ExpectPrepare(`DELETE FROM user WHERE status = \$1`).
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
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`c1 = \$1 AND c2 = \$2`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c := dao.And().Eq("c1", 1).Group(nil)
			c2 := dao.Or().Eq("c2", 2).Group(nil)
			c = dao.And().Group(c).Group(c2)

			str, args := c.ToStrArgs()
			r.Equal("c1 = $1 AND c2 = $2", str)
			r.Len(args, 2)
			r.Contains(args, 1)
			r.Contains(args, 2)
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`c1 = \$1 AND c2 <> \$2 AND c3 > \$3 AND c4 < \$4 AND c5 >= \$5 AND c6 <= \$6 AND c7 LIKE \$7 AND c8 LIKE \$8 AND c9 LIKE \$9 AND c10 IN\(\$10, \$11, \$12\) AND c11 BETWEEN \$13 AND \$14 AND c12 IS NULL AND c13 IS NOT NULL`).
			ExpectQuery().WithArgs(1, 2, 3, 4, 5, 6, "%abc%", "abc%", "%abc", 1, 2, 3, 1, 3).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c := dao.And().Eq("c1", 1).
				NotEq("c2", 2).
				Gt("c3", 3).
				Lt("c4", 4).
				Ge("c5", 5).
				Le("c6", 6).
				Like("c7", "abc").
				LikeLeft("c8", "abc").
				LikeRight("c9", "abc").
				In("c10", dao.Anys([]int{1, 2, 3}...)).
				Between("c11", 1, 3).
				IsNull("c12").
				IsNotNull("c13")
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`0 = 0 AND NOT c1 = \$1 AND NOT \(c2 = \$2 AND c3 = \$3 AND 1 = 1 and 2 = 2\) AND c4 = \$4 AND \(c5 = \$5 OR c6 = \$6\) AND NOT c7 = \$7 AND NOT \(c8 = \$8 OR c9 = \$9\)`).
			ExpectQuery().WithArgs(1, 2, 3, 4, 5, 6, 7, 8, 9).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c0 := dao.And().Plain("0 = 0")
			c1 := dao.NotAnd().Eq("c1", 1)
			c2 := dao.NotAnd().Eq("c2", 2).Eq("c3", 3).Plain("1 = 1 and 2 = 2")
			c3 := dao.Or().Eq("c4", 4)
			c4 := dao.Or().Eq("c5", 5).Eq("c6", 6)
			c5 := dao.NotOr().Eq("c7", 7)
			c6 := dao.NotOr().Eq("c8", 8).Eq("c9", 9)
			c := dao.And().Group(c0).Group(c1).Group(c2).Group(c3).Group(c4).Group(c5).Group(c6)
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
	{
		d, mock := dao.MockBaseDao[User](r, "user")
		mock.ExpectPrepare(`\(c1 = \$1 OR c2 = \$2\) AND c3 = \$3 AND c4 = \$4`).
			ExpectQuery().WithArgs(1, 2, 3, 4).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
			c1 := dao.Or().Eq("c1", 1).Eq("c2", 2)
			c2 := dao.Or().Eq("c3", 3)
			c3 := dao.Or().Eq("c4", 4)
			c := dao.And().Group(c1).Group(c2).Group(c3)
			dao.WriteCondition(c, b)
		}})
		r.NoError(err)
	}
}
