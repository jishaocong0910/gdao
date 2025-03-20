package gdao_test

import (
	"database/sql"
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
	// test invalid field
	field1 *string `gdao:"column=field1"`
	Field2 string  `gdao:"column=field2"`
	Field3 *string
}

type Account struct {
	Id      *int32 `gdao:"auto=2"`
	OtherId *int32
	UserId  *int32
	Status  *int8
	Balance *int64
}

func mockUserDao(t *testing.T) (gdao.Dao[User], sqlmock.Sqlmock) {
	r := require.New(t)
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao, err := gdao.NewDao[User](gdao.NewDaoReq{Db: db, Table: "user"})
	r.NoError(err)
	return dao, mock
}

func mockAccountDao(t *testing.T) (gdao.Dao[Account], sqlmock.Sqlmock) {
	r := require.New(t)
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao, err := gdao.NewDao[Account](gdao.NewDaoReq{Db: db, Table: "account", ColumnMapper: gdao.NewNameMapper().LowerSnakeCase()})
	r.NoError(err)
	return dao, mock
}

func TestNewDao(t *testing.T) {
	r := require.New(t)
	{
		dao, _ := mockUserDao(t)
		export := gdao.ExportDao(dao)
		r.Equal("user", export.Table)
		r.Equal("id,name,age,address,phone,email,status,level,create_at", export.ColumnsWithComma)
		r.Contains(export.Columns, "id", "name", "age", "address", "phone", "email", "status", "level", "create_at")
		r.Equal(9, len(export.ColumnToFieldIndexMap))
		r.Contains(export.ColumnToFieldIndexMap, "id", "name", "age", "address", "phone", "email", "status", "level", "create_at")
		r.Equal("id", export.AutoIncrementColumn)
		r.Equal(int64(1), export.AutoIncrementOffset)
		r.NotNil(export.AutoIncrementConvertor)
	}
	{
		dao, _ := mockAccountDao(t)
		export := gdao.ExportDao(dao)
		r.Equal("account", export.Table)
		r.Equal("id,other_id,user_id,status,balance", export.ColumnsWithComma)
		r.Contains(export.Columns, "id", "other_id", "user_id", "status", "balance")
		r.Equal(5, len(export.ColumnToFieldIndexMap))
		r.Contains(export.ColumnToFieldIndexMap, "id", "other_id", "user_id", "status", "balance")
		r.Equal("id", export.AutoIncrementColumn)
		r.Equal(int64(2), export.AutoIncrementOffset)
		r.NotNil(export.AutoIncrementConvertor)
	}
}

func TestDao_RawQuery(t *testing.T) {
	r := require.New(t)
	dao, mock := mockUserDao(t)
	mock.ExpectPrepare(`SELECT \* FROM user WHERE id=\?`).ExpectQuery().WithArgs(1).WillReturnRows(
		mock.NewRows([]string{"id", "create_at", "name", "status", "age", "address", "phone", "email", "level"}).
			AddRow(1, time.UnixMilli(1703659380000), "foo", 1, 1, "bar", 123, "foo@gmail", 1))
	rows, cl, err := dao.RawQuery(gdao.RawQueryReq{Sql: "SELECT * FROM user WHERE id=?", Args: []any{1}})
	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	defer cl()
	user := User{}
	if rows.Next() {
		rows.Scan(&user.Id, &user.CreateAt, &user.Name, &user.Status, &user.Age, &user.Address, &user.Phone, &user.Email, &user.Level)
	}
	r.Equal(int32(1), *user.Id)
	r.Equal(time.UnixMilli(1703659380000), *user.CreateAt)
	r.Equal("foo", *user.Name)
	r.Equal(int32(1), *user.Age)
	r.Equal(int8(1), *user.Status)
	r.Equal("bar", *user.Address)
	r.Equal("123", *user.Phone)
	r.Equal("foo@gmail", *user.Email)
	r.Equal(int32(1), *user.Level)
}

func TestDao_RawMutation(t *testing.T) {
	r := require.New(t)
	dao, mock := mockUserDao(t)
	mock.ExpectPrepare(`UPDATE user set level=\? WHERE id=\?`).ExpectExec().WithArgs(2, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	result, err := dao.RawMutation(gdao.RawMutationReq{Sql: "UPDATE user set level=? WHERE id=?", Args: []any{2, 1}})
	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	affected, err := result.RowsAffected()
	r.NoError(err)
	r.Equal(int64(1), affected)
}

func TestDao_Query(t *testing.T) {
	r := require.New(t)
	createAt := time.UnixMilli(1703659380000)
	{
		dao, mock := mockUserDao(t)
		mock.ExpectPrepare(`SELECT id,name,age,address,phone,email,status,level,create_at FROM user WHERE id=\? AND status=\?`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt))
		user, users, err := dao.Query(gdao.QueryReq[User]{
			BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
				b.Write("SELECT ").Write(b.Columns()).
					Write(" FROM ").Write(b.Table()).
					Write(" WHERE id=? AND status=?").AddArgs(1, 2)
				return b.String(), b.Args()
			},
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(user, users[0])
		r.Equal(int32(1), *user.Id)
		r.Equal("foo", *user.Name)
		r.Equal(int8(2), *user.Status)
		r.Equal(int32(1), *user.Age)
		r.Equal("bar", *user.Address)
		r.Equal("123456", *user.Phone)
		r.Equal("foo@gmail", *user.Email)
		r.Equal(int32(1), *user.Level)
		r.Equal(createAt, *user.CreateAt)
	}
	{
		dao, mock := mockUserDao(t)
		user := &User{
			Status:   gdao.Ptr[int8](1),
			CreateAt: gdao.Ptr(createAt),
			Level:    gdao.Ptr[int32](2),
		}
		mock.ExpectPrepare(`SELECT id,name,age,address,phone,email,status,level,create_at FROM user WHERE status=\$1 AND level=\$2 AND create_at=\$3`).
			ExpectQuery().WithArgs(user.Status, user.Level, user.CreateAt).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt))
		_, _, err := dao.Query(gdao.QueryReq[User]{
			Entities: []*User{user},
			BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
				b.Write("SELECT ").Write(b.Columns())
				b.Write(" FROM ").Write(b.Table())
				b.Write(" WHERE ")
				e := b.Entity()
				b.Write("status=").Write(b.ArgN("$")).AddArgs(e.Status)
				b.Write(" AND ").Write("level=").Write(b.ArgN("$")).AddArgs(e.Level)
				b.Write(" AND ").Write("create_at=").Write(b.ArgN("$")).AddArgs(e.CreateAt)
				return b.String(), b.Args()
			},
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestMutationDao_Exec(t *testing.T) {
	r := require.New(t)
	{
		dao, mock := mockUserDao(t)
		user := &User{
			Status:  gdao.Ptr[int8](3),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("address"),
			Phone:   gdao.Ptr("56789"),
		}
		mock.ExpectPrepare(`UPDATE user SET address=\?,phone=\?,status=\?,level=\? WHERE id=\?`).
			ExpectExec().WithArgs(user.Address, user.Phone, user.Status, user.Level, 1001).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := dao.Mutation(gdao.MutationReq[User]{
			Entities: []*User{user},
			BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
				b.Write("UPDATE ").Write(b.Table()).Write(" SET ")
				b.EachAssignedColumn(b.Separate("", ",", ""), func(i int, column string, value any) {
					b.Write(column).Write("=?").AddArgs(value)
				})
				b.Write(" WHERE id=?").AddArgs(1001)
				return b.String(), b.Args()
			},
		}).Exec()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		dao, mock := mockUserDao(t)
		user := &User{
			Name:    gdao.Ptr("foo"),
			Status:  gdao.Ptr[int8](3),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("address"),
			Phone:   gdao.Ptr("56789"),
		}
		mock.ExpectPrepare(`INSERT user\(name,address,phone,status,level\) VALUES\(\?,\?,\?,\?,\?\)`).
			ExpectExec().
			WithArgs(user.Name, user.Address, user.Phone, user.Status, user.Level).
			WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := dao.Mutation(gdao.MutationReq[User]{
			Entities: []*User{user},
			BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
				b.Write("INSERT ").Write(b.Table())
				b.EachAssignedColumn(b.Separate("(", ",", ")"), func(i int, column string, value any) {
					b.Write(column)
				})
				b.EachAssignedColumn(b.Separate(" VALUES(", ",", ")"), func(i int, column string, value any) {
					b.Write("?").AddArgs(value)
				})
				return b.String(), b.Args()
			},
		}).Exec()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Nil(user.Id)
	}
	{
		dao, mock := mockUserDao(t)
		export := gdao.ExportDao(dao)
		user := &User{
			Status:  gdao.Ptr[int8](3),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("address"),
			Phone:   gdao.Ptr("56789"),
		}
		mock.ExpectPrepare(`INSERT user\(`+export.ColumnsWithComma+`\) VALUES\(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
			ExpectExec().
			WithArgs(user.Id, user.Name, user.Age, user.Address, user.Phone, user.Email, user.Status, user.Level, user.CreateAt).
			WillReturnResult(sqlmock.NewResult(0, 2))
		affected, err := dao.Mutation(gdao.MutationReq[User]{
			Entities: []*User{user},
			BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
				b.Write("INSERT ").Write(b.Table()).Write("(").Write(b.Columns()).Write(") VALUES(")
				b.EachColumn(b.Separate("", ",", ""),
					func(i int, column string, value any) {
						b.Write("?").AddArgs(value)
					})
				b.Write(")")
				return b.String(), b.Args()
			},
		}).Insert()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(2), affected)
	}
}

func TestMutationDao_Insert(t *testing.T) {
	r := require.New(t)
	dao, mock := mockUserDao(t)
	export := gdao.ExportDao(dao)
	users := []*User{
		{
			Status:  gdao.Ptr[int8](3),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("address"),
			Phone:   gdao.Ptr("56789"),
		},
		{
			Status:  gdao.Ptr[int8](2),
			Level:   gdao.Ptr[int32](2),
			Address: gdao.Ptr("addr"),
			Phone:   gdao.Ptr("2325325"),
		},
	}
	mock.ExpectPrepare(`INSERT user\(`+export.ColumnsWithComma+`\) VALUES\(\?,\?,\?,\?,\?,\?,\?,\?,\?\),\(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
		ExpectExec().
		WithArgs(users[0].Id, users[0].Name, users[0].Age, users[0].Address, users[0].Phone, users[0].Email, users[0].Status, users[0].Level, users[0].CreateAt,
			users[1].Id, users[1].Name, users[1].Age, users[1].Address, users[1].Phone, users[1].Email, users[1].Status, users[1].Level, users[1].CreateAt).
		WillReturnResult(sqlmock.NewResult(1001, 2))
	affected, err := dao.Mutation(gdao.MutationReq[User]{
		Entities: users,
		BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
			b.Write("INSERT ").Write(b.Table()).Write("(").Write(b.Columns()).Write(") VALUES")
			b.EachEntity(b.Separate("", ",", ""), func(i int) {
				b.Write("(")
				b.EachColumnAt(i, b.Separate("", ",", ""), func(i int, column string, value any) {
					b.Write("?").AddArgs(value)
				})
				b.Write(")")
			})
			return b.String(), b.Args()
		},
	}).Insert()

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(1001), *users[0].Id)
	r.Equal(int32(1002), *users[1].Id)
}

func TestMutationDao_Query(t *testing.T) {
	r := require.New(t)
	dao, mock := mockAccountDao(t)
	accounts := []*Account{
		{
			UserId:  gdao.Ptr[int32](1),
			Status:  gdao.Ptr[int8](1),
			Balance: gdao.Ptr[int64](100),
		},
		{
			UserId:  gdao.Ptr[int32](2),
			Status:  gdao.Ptr[int8](1),
			Balance: gdao.Ptr[int64](200),
		},
	}
	mock.ExpectPrepare(`INSERT account\(user_id,status,balance\) VALUES\(\?,\?,\?\),\(\?,\?,\?\) RETURNING id`).
		ExpectQuery().WithArgs(accounts[0].UserId, accounts[0].Status, accounts[0].Balance,
		accounts[1].UserId, accounts[1].Status, accounts[1].Balance).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(2001).AddRow(2002))
	affected, err := dao.Mutation(gdao.MutationReq[Account]{
		Entities: accounts,
		BuildSql: func(b gdao.Builder[Account]) (sql string, args []any) {
			b.Write("INSERT ").Write(b.Table()).Write("(user_id,status,balance) VALUES")
			b.EachEntity(b.Separate("", ",", ""), func(i int) {
				e := b.EntityAt(i)
				b.Write("(?,?,?)").AddArgs(e.UserId, e.Status, e.Balance)
			})
			b.Write(" RETURNING id")
			return b.String(), b.Args()
		},
	}).Query()

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(2001), *accounts[0].Id)
	r.Equal(int32(2002), *accounts[1].Id)
}

func TestMustNewDao(t *testing.T) {
	r := require.New(t)
	{
		r.PanicsWithError("db must not be nil", func() {
			gdao.MustNewDao[User](gdao.NewDaoReq{})
		})
	}
	{
		r.PanicsWithError("table must not be blank", func() {
			gdao.MustNewDao[User](gdao.NewDaoReq{Db: &sql.DB{}})
		})
	}
	{
		r.PanicsWithError("generics must be struct", func() {
			gdao.MustNewDao[*User](gdao.NewDaoReq{Db: &sql.DB{}, Table: "user"})
		})
	}
	{
		r.NotPanics(func() {
			gdao.MustNewDao[User](gdao.NewDaoReq{Db: &sql.DB{}, Table: "user"})
		})
	}
}

func TestTx(t *testing.T) {
	r := require.New(t)
	{
		userDao, mock := mockUserDao(t)
		mock.ExpectBegin()
		mock.ExpectPrepare(`SELECT \* FROM user WHERE id=\?`).ExpectQuery()
		mock.ExpectCommit()
		tx, err := userDao.Db().Begin()
		r.NoError(err)
		userDao.RawQuery(gdao.RawQueryReq{Tx: tx, Sql: "SELECT * FROM user WHERE id=?", Args: []any{1}})
		tx.Commit()
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestLastInsertIdConvertors(t *testing.T) {
	r := require.New(t)
	id := int64(123)
	r.Equal(123, gdao.LastInsertIdConvertors["int"](id).Elem().Interface())
	r.Equal(int8(123), gdao.LastInsertIdConvertors["int8"](id).Elem().Interface())
	r.Equal(int16(123), gdao.LastInsertIdConvertors["int16"](id).Elem().Interface())
	r.Equal(int32(123), gdao.LastInsertIdConvertors["int32"](id).Elem().Interface())
	r.Equal(int64(123), gdao.LastInsertIdConvertors["int64"](id).Elem().Interface())
	r.Equal(uint(123), gdao.LastInsertIdConvertors["uint"](id).Elem().Interface())
	r.Equal(uint8(123), gdao.LastInsertIdConvertors["uint8"](id).Elem().Interface())
	r.Equal(uint16(123), gdao.LastInsertIdConvertors["uint16"](id).Elem().Interface())
	r.Equal(uint32(123), gdao.LastInsertIdConvertors["uint32"](id).Elem().Interface())
	r.Equal(uint64(123), gdao.LastInsertIdConvertors["uint64"](id).Elem().Interface())
	r.Equal(float32(123), gdao.LastInsertIdConvertors["float32"](id).Elem().Interface())
	r.Equal(float64(123), gdao.LastInsertIdConvertors["float64"](id).Elem().Interface())
	r.Equal("123", gdao.LastInsertIdConvertors["string"](id).Elem().Interface())
}
