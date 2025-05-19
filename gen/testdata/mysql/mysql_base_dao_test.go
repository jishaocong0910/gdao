package mysql_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jishaocong0910/gdao"
	"github.com/jishaocong0910/gdao/gen/testdata/mysql"
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

func TestBaseDao_List(t *testing.T) {
	r := require.New(t)
	dao, mock := mysql.MockMysqlBaseDao[User](t)

	mock.ExpectPrepare(`SELECT id,name FROM user WHERE status=\? ORDER BY name ASC,address DESC LIMIT 10,10 FOR UPDATE`).
		ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
		AddRow(1, "lucy").AddRow(2, "nick"))

	list, err := dao.List(mysql.ListReq{
		SelectColumns: []string{"id", "name"},
		Condition:     mysql.And().Eq("status", 4),
		OrderBy:       mysql.OrderBy().Asc("name").Desc("address"),
		Pagination:    mysql.Page(2, 10),
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

func TestBaseDao_Get(t *testing.T) {
	r := require.New(t)
	dao, mock := mysql.MockMysqlBaseDao[User](t)

	mock.ExpectPrepare(`SELECT id,phone FROM user WHERE status=\? FOR UPDATE`).
		ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
		AddRow(1, "lucy"))

	get, err := dao.Get(mysql.GetReq{
		SelectColumns: []string{"id", "phone"},
		Condition:     mysql.And().Eq("status", 4),
		ForUpdate:     true,
	})
	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int32(1), *get.Id)
	r.Equal("lucy", *get.Name)
}

func TestBaseDao_Insert(t *testing.T) {
	r := require.New(t)
	dao, mock := mysql.MockMysqlBaseDao[User](t)

	mock.ExpectPrepare(`INSERT INTO user\(name,phone,email,level\) VALUES\(\?,\?,\?,NULL\)`).
		ExpectExec().WithArgs("abc", "12345", "email").WillReturnResult(sqlmock.NewResult(7, 1))

	u := &User{
		Name:  gdao.Ptr("abc"),
		Phone: gdao.Ptr("12345"),
		Email: gdao.Ptr("email"),
	}
	affected, err := dao.Insert(mysql.InsertReq[User]{
		Entity:         u,
		SetNullColumns: []string{"level"},
	})
	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(1), affected)
	r.Equal(int32(7), *u.Id)
}

func TestBaseDao_InsertBatch(t *testing.T) {
	r := require.New(t)
	dao, mock := mysql.MockMysqlBaseDao[User](t)

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
	affected, err := dao.InsertBatch(mysql.InsertBatchReq[User]{
		Entities:       []*User{u, u2},
		IgnoredColumns: []string{"id", "age", "address", "status", "level", "create_at"},
	})
	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(8), *u.Id)
	r.Equal(int32(9), *u2.Id)
}
