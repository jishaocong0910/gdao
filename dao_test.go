package gdao_test

import (
	"context"
	"database/sql"
	"errors"
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

type Account struct {
	Id      *int32 `gdao:"auto=2"`
	OtherId *int32
	UserId  *int32
	Status  *int8
	Balance *int64
}

type InvalidField struct {
	field *string `gdao:"column=field"`
}

type InvalidField2 struct {
	Field string `gdao:"column=field"`
}

func mockUserDao(r *require.Assertions) (*gdao.Dao[User], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.NewDao[User](gdao.NewDaoReq{DB: db})
	return dao, mock
}

func mockAccountDao(r *require.Assertions) (*gdao.Dao[Account], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.NewDao[Account](gdao.NewDaoReq{DB: db, ColumnMapper: gdao.NewNameMapper().LowerSnakeCase()})
	return dao, mock
}

func TestNewDao(t *testing.T) {
	r := require.New(t)
	{
		dao, _ := mockUserDao(r)
		export := gdao.ExportDao(dao)
		r.Equal("id, name, age, address, phone, email, status, level, create_at", export.ColumnsWithComma)
		r.Contains(export.Columns, "id", "name", "age", "address", "phone", "email", "status", "level", "create_at")
		r.Equal(9, len(export.ColumnToFieldIndexMap))
		r.Contains(export.ColumnToFieldIndexMap, "id", "name", "age", "address", "phone", "email", "status", "level", "create_at")
		r.Contains(export.AutoIncrementColumns, "id")
		r.Equal(int64(1), export.AutoIncrementStep)
		r.NotNil(export.AutoIncrementConvertor)
	}
	{
		dao, _ := mockAccountDao(r)
		export := gdao.ExportDao(dao)
		r.Equal("id, other_id, user_id, status, balance", export.ColumnsWithComma)
		r.Contains(export.Columns, "id", "other_id", "user_id", "status", "balance")
		r.Equal(5, len(export.ColumnToFieldIndexMap))
		r.Contains(export.ColumnToFieldIndexMap, "id", "other_id", "user_id", "status", "balance")
		r.Contains(export.AutoIncrementColumns, "id")
		r.Equal(int64(2), export.AutoIncrementStep)
		r.NotNil(export.AutoIncrementConvertor)
	}
}

func TestDao_Query(t *testing.T) {
	r := require.New(t)
	createAt := time.UnixMilli(1703659380000)
	{
		dao, mock := mockUserDao(r)
		mock.ExpectPrepare(`SELECT id, name, age, address, phone, email, status, level, create_at FROM user WHERE id=\? AND status=\?`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt))
		user, users, err := dao.Query(gdao.QueryReq[User]{
			BuildSql: func(b *gdao.Builder[User]) {
				b.Write("SELECT ").WriteColumns().
					Write(" FROM user").
					Write(" WHERE id=? AND status=?", 1, 2)
			},
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(user, users[0])
		r.Equal(int32(1), gdao.PtrToValue(user.Id))
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
		dao, mock := mockUserDao(r)
		user := &User{
			Status:   gdao.Ptr[int8](1),
			CreateAt: gdao.Ptr(createAt),
			Level:    gdao.Ptr[int32](2),
		}
		mock.ExpectPrepare(`SELECT id, name, age, address, phone, email, status, level, create_at FROM user WHERE status=\$1 AND level=\$2 AND create_at=\$3`).
			ExpectQuery().WithArgs(user.Status, user.Level, user.CreateAt).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt))
		_, _, err := dao.Query(gdao.QueryReq[User]{
			Entities: []*User{user},
			BuildSql: func(b *gdao.Builder[User]) {
				b.Write("SELECT ").WriteColumns()
				b.Write(" FROM user")
				b.Write(" WHERE ")
				e := b.Entity()
				b.Write("status=").Write(b.Pp("$"), e.Status)
				b.Write(" AND ").Write("level=").Write(b.Pp("$"), e.Level)
				b.Write(" AND ").Write("create_at=").Write(b.Pp("$"), e.CreateAt)
			},
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		// 测试查询字段比实体字段多的情况
		dao, mock := mockUserDao(r)
		mock.ExpectPrepare(`SELECT id, name, age, address, phone, email, status, level, create_at, valid FROM user WHERE id=\? AND status=\?`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at", "valid"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt, true))
		user, users, err := dao.Query(gdao.QueryReq[User]{
			BuildSql: func(b *gdao.Builder[User]) {
				b.Write("SELECT ").WriteColumns().
					Write(", valid").
					Write(" FROM user").
					Write(" WHERE id=? AND status=?", 1, 2)
			},
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(user, users[0])
		r.Equal(int32(1), gdao.PtrToValue(user.Id))
		r.Equal("foo", *user.Name)
		r.Equal(int8(2), *user.Status)
		r.Equal(int32(1), *user.Age)
		r.Equal("bar", *user.Address)
		r.Equal("123456", *user.Phone)
		r.Equal("foo@gmail", *user.Email)
		r.Equal(int32(1), *user.Level)
		r.Equal(createAt, *user.CreateAt)
	}
}

func TestDao_Query_RowAsReturning(t *testing.T) {
	r := require.New(t)
	dao, mock := mockAccountDao(r)
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
	_, _, err := dao.Query(gdao.QueryReq[Account]{
		Entities: accounts,
		RowAs:    gdao.ROW_AS_RETURNING,
		BuildSql: func(b *gdao.Builder[Account]) {
			b.Write("INSERT account").Write("(user_id,status,balance) VALUES")
			b.EachEntity(b.Sep(","), func(n, i int, entity *Account) {
				b.Write("(?,?,?)", entity.UserId, entity.Status, entity.Balance)
			})
			b.Write(" RETURNING id")
		},
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int32(2001), *accounts[0].Id)
	r.Equal(int32(2002), *accounts[1].Id)
}

func TestDao_Query_RowAsLastId(t *testing.T) {
	r := require.New(t)
	dao, mock := mockAccountDao(r)
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
	mock.ExpectPrepare(`INSERT account\(user_id,status,balance\) VALUES\(\?,\?,\?\),\(\?,\?,\?\); SELECT ID=convert\(bigint, SCOPE_IDENTITY\(\)\)`).
		ExpectQuery().WithArgs(accounts[0].UserId, accounts[0].Status, accounts[0].Balance,
		accounts[1].UserId, accounts[1].Status, accounts[1].Balance).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1234))
	_, _, err := dao.Query(gdao.QueryReq[Account]{
		Entities: accounts,
		RowAs:    gdao.ROW_AS_LAST_ID,
		BuildSql: func(b *gdao.Builder[Account]) {
			b.Write("INSERT account").Write("(user_id,status,balance) VALUES")
			b.EachEntity(b.Sep(","), func(n, i int, entity *Account) {
				b.Write("(?,?,?)", entity.UserId, entity.Status, entity.Balance)
			})
			b.Write("; SELECT ID=convert(bigint, SCOPE_IDENTITY())")
		},
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int32(1232), *accounts[0].Id)
	r.Equal(int32(1234), *accounts[1].Id)
}

func TestDao_Exec(t *testing.T) {
	r := require.New(t)
	{
		dao, mock := mockUserDao(r)
		user := &User{
			Status:  gdao.Ptr[int8](3),
			Level:   gdao.Ptr[int32](10),
			Address: gdao.Ptr("address"),
			Phone:   gdao.Ptr("56789"),
		}
		mock.ExpectPrepare(`UPDATE user SET address=\?,phone=\?,status=\?,level=\? WHERE id=\?`).
			ExpectExec().WithArgs(user.Address, user.Phone, user.Status, user.Level, 1001).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := dao.Exec(gdao.ExecReq[User]{
			Entities: []*User{user},
			BuildSql: func(b *gdao.Builder[User]) {
				b.Write("UPDATE user SET ")
				cvs := b.ColumnValues(true)
				b.EachColumnValues(cvs, b.Sep(","), func(column string, value any) {
					b.Write(column).Write("=?", value)
				})
				b.Write(" WHERE id=?", 1001)
			},
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		dao, mock := mockUserDao(r)
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
		affected, err := dao.Exec(gdao.ExecReq[User]{
			Entities: []*User{user},
			BuildSql: func(b *gdao.Builder[User]) {
				b.Write("INSERT user")
				cvs := b.ColumnValues(true)
				b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(column string, value any) {
					b.Write(column)
				})
				b.EachColumnValues(cvs, b.SepFix(" VALUES(", ",", ")", false), func(column string, value any) {
					b.Write("?", value)
				})
			},
		})

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Nil(user.Id)
	}
}

func TestDao_Exec_LastInsertIdAsFirstId(t *testing.T) {
	r := require.New(t)
	dao, mock := mockUserDao(r)
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
	affected, err := dao.Exec(gdao.ExecReq[User]{
		Entities:       users,
		LastInsertIdAs: gdao.LAST_INSERT_ID_AS_FIRST_ID,
		BuildSql: func(b *gdao.Builder[User]) {
			b.Write("INSERT user").Write("(").WriteColumns().Write(") VALUES")
			b.EachEntity(b.Sep(","), func(n, i int, entity *User) {
				cvs := b.ColumnValuesAt(entity, false)
				b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(column string, value any) {
					b.Write("?", value)
				})
			})
		},
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(1001), *users[0].Id)
	r.Equal(int32(1002), *users[1].Id)
}

func TestDao_Exec_LastInsertIdAsLastId(t *testing.T) {
	r := require.New(t)
	dao, mock := mockUserDao(r)
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
	affected, err := dao.Exec(gdao.ExecReq[User]{
		Entities:       users,
		LastInsertIdAs: gdao.LAST_INSERT_ID_AS_LAST_ID,
		BuildSql: func(b *gdao.Builder[User]) {
			b.Write("INSERT user").Write("(").WriteColumns().Write(") VALUES")
			b.EachEntity(b.Sep(","), func(n, i int, entity *User) {
				cvs := b.ColumnValuesAt(entity, false)
				b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(column string, value any) {
					b.Write("?", value)
				})
			})
		},
	})

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(1000), *users[0].Id)
	r.Equal(int32(1001), *users[1].Id)
}

func TestNewDaoPanic(t *testing.T) {
	r := require.New(t)
	{
		r.PanicsWithValue("generics must be struct type", func() {
			gdao.NewDao[*User](gdao.NewDaoReq{DB: &sql.DB{}})
		})
		r.NotPanics(func() {
			gdao.NewDao[User](gdao.NewDaoReq{DB: &sql.DB{}})
		})
	}
	{
		r.PanicsWithValue(`field "field" is invalid, the entity's field must be a pointer and exported`, func() {
			gdao.NewDao[InvalidField](gdao.NewDaoReq{DB: &sql.DB{}})
		})
		r.NotPanics(func() {
			gdao.NewDao[InvalidField](gdao.NewDaoReq{DB: &sql.DB{}, AllowInvalidField: true})
		})
	}
	{
		r.PanicsWithValue(`field "Field" is invalid, the entity's field must be a pointer and exported`, func() {
			gdao.NewDao[InvalidField2](gdao.NewDaoReq{DB: &sql.DB{}})
		})
		r.NotPanics(func() {
			gdao.NewDao[InvalidField2](gdao.NewDaoReq{DB: &sql.DB{}, AllowInvalidField: true})
		})
	}
}

func TestTx(t *testing.T) {
	r := require.New(t)
	{
		userDao, mock := mockUserDao(r)
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE user set status=1 WHERE id=\?`).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		tx, err := userDao.DB().Begin()
		r.NoError(err)
		affected, err := userDao.Exec(gdao.ExecReq[User]{Ctx: gdao.WithTx(nil, tx), BuildSql: func(b *gdao.Builder[User]) {
			b.Write("UPDATE user set status=1 WHERE id=?", 1)
		}})
		tx.Commit()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectPrepare(`UPDATE user set status=1 WHERE id=\?`).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		err := gdao.Tx(nil, nil, nil, func(ctx context.Context) error {
			_, err := userDao.Exec(gdao.ExecReq[User]{
				Ctx: ctx,
				BuildSql: func(b *gdao.Builder[User]) {
					b.Write("UPDATE user set status=1 WHERE id=?", 1)
				},
			})
			return err
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, nil, nil, func(ctx context.Context) error {
			return errors.New("error")
		})
		r.EqualError(err, "error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, nil, nil, func(ctx context.Context) error {
			panic(errors.New("panic error"))
		})
		r.EqualError(err, "panic error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		userDao, mock := mockUserDao(r)
		gdao.DEFAULT_DB = userDao.DB()
		mock.ExpectBegin()
		mock.ExpectRollback()
		err := gdao.Tx(nil, nil, nil, func(ctx context.Context) error {
			panic(1)
		})
		r.EqualError(err, "1")
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
