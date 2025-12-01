/*
 * Copyright 2024-present jishaocong0910
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gdao_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
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
	Id          *int32 `gdao:"auto=2"`
	OtherId     *int32
	UserId      *int32
	Status      *int8
	Balance     *int64
	LicenceFile []uint8
}

type Product struct {
	Id         *int64 `gdao:"auto"`
	Tags       MyStringSlice
	Status     *MyStatus
	Properties *Properties
	Attributes *Attributes
}

type Properties struct {
	Unit   string `json:"unit"`
	Weight int    `json:"weight"`
}

func (p Properties) GdaoValue() string {
	str, _ := json.Marshal(p)
	return string(str)
}

func (Properties) GdaoField(value string) *Properties {
	var p *Properties
	json.Unmarshal([]byte(value), &p)
	return p
}

type Attributes struct {
	Size  []int    `json:"size"`
	Color []string `json:"color"`
}

func (a Attributes) GdaoValue() string {
	str, _ := json.Marshal(a)
	return string(str)
}

func (*Attributes) GdaoField(value string) *Attributes {
	var a *Attributes
	json.Unmarshal([]byte(value), &a)
	return a
}

type MyStringSlice []string

func (c MyStringSlice) GdaoValue() string {
	return strings.Join(c, ",")
}

func (MyStringSlice) GdaoField(value string) MyStringSlice {
	return strings.Split(value, ",")
}

type MyStatus int

type InvalidField struct {
	field *string `gdao:"column=field"`
}

type InvalidField2 struct {
	Field string `gdao:"column=field"`
}

type InvalidField3 struct {
	Field *any `gdao:"column=field"`
}

type InvalidField4 struct {
	Field []any `gdao:"column=field"`
}

type InvalidField5 struct {
	Field *InvalidImplementConvert `gdao:"column=field"`
}

type InvalidImplementConvert struct {
}

func (InvalidImplementConvert) GdaoValue() int {
	return 0
}

func (InvalidImplementConvert) GdaoField(int) InvalidImplementConvert {
	return InvalidImplementConvert{}
}

func mockUserDao(r *require.Assertions) (*gdao.Dao[User], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.DaoBuilder[User]().Build()
	gdao.Config(gdao.Cfg{DefaultDB: db})
	return dao, mock
}

func mockAccountDao(r *require.Assertions) (*gdao.Dao[Account], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.DaoBuilder[Account]().DB(db).ColumnMapper(gdao.NewNameMapper().LowerSnakeCase()).Build()
	return dao, mock
}

func mockProductDao(r *require.Assertions) (*gdao.Dao[Product], sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	r.NoError(err)
	dao := gdao.DaoBuilder[Product]().DB(db).ColumnMapper(gdao.NewNameMapper().LowerSnakeCase()).Build()
	return dao, mock
}

func TestNewDao(t *testing.T) {
	r := require.New(t)
	{
		dao, _ := mockUserDao(r)
		export := gdao.ExportDao(dao)
		r.Equal("id, name, age, address, phone, email, status, level, create_at", export.ColumnsWithComma)
		r.Equal([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at"}, export.Columns)
		r.Len(export.ColumnToFieldIndex, 9)
		checkMap(r, map[string]int{"id": 0, "name": 1, "age": 2, "address": 3, "phone": 4, "email": 5, "status": 6, "level": 7, "create_at": 8}, export.ColumnToFieldIndex)
		checkMap(r, map[string]string{"Id": "id", "Name": "name", "Age": "age", "Address": "address", "Phone": "phone", "Email": "email", "Status": "status", "Level": "level", "CreateAt": "create_at"}, dao.NameMap())
		r.Contains(export.AutoIncrementColumns, "id")
		r.Equal(int64(1), export.AutoIncrementStep)
		r.NotNil(export.AutoIncrementConvertor)
	}
	{
		dao, _ := mockAccountDao(r)
		export := gdao.ExportDao(dao)
		r.Equal("id, other_id, user_id, status, balance, licence_file", export.ColumnsWithComma)
		r.Equal([]string{"id", "other_id", "user_id", "status", "balance", "licence_file"}, export.Columns)
		r.Len(export.ColumnToFieldIndex, 6)
		checkMap(r, map[string]int{"id": 0, "other_id": 1, "user_id": 2, "status": 3, "balance": 4, "licence_file": 5}, export.ColumnToFieldIndex)
		checkMap(r, map[string]string{"Id": "id", "OtherId": "other_id", "UserId": "user_id", "Status": "status", "Balance": "balance", "LicenceFile": "licence_file"}, dao.NameMap())
		r.Contains(export.AutoIncrementColumns, "id")
		r.Equal(int64(2), export.AutoIncrementStep)
		r.NotNil(export.AutoIncrementConvertor)
	}
	{
		dao, _ := mockProductDao(r)
		export := gdao.ExportDao(dao)
		r.Equal("id, tags, status, properties, attributes", export.ColumnsWithComma)
		r.Equal([]string{"id", "tags", "status", "properties", "attributes"}, export.Columns)
		r.Len(export.ColumnToFieldIndex, 5)
		checkMap(r, map[string]int{"id": 0, "tags": 1, "status": 2, "properties": 3, "attributes": 4}, export.ColumnToFieldIndex)
		checkMap(r, map[string]string{"Id": "id", "Tags": "tags", "Status": "status", "Properties": "properties", "Attributes": "attributes"}, dao.NameMap())
		r.Len(export.ColumnToFieldConvertor, 3)
		checkMapKeys(r, []string{"tags", "properties", "attributes"}, export.ColumnToFieldConvertor)
		r.Contains(export.AutoIncrementColumns, "id")
		r.Equal(int64(1), export.AutoIncrementStep)
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
		user, users, err := dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("SELECT ")
			b.WriteColumns()
			b.Write(" FROM user")
			b.Write(" WHERE id=? AND status=?", 1, 2)
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(user, users[0])
		r.Equal(int32(1), gdao.V(user.Id))
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
			Status:   gdao.P[int8](1),
			CreateAt: gdao.P(createAt),
			Level:    gdao.P[int32](2),
		}
		mock.ExpectPrepare(`SELECT id, name, age, address, phone, email, status, level, create_at FROM user WHERE status=\$1 AND level=\$2 AND create_at=\$3`).
			ExpectQuery().WithArgs(user.Status, user.Level, user.CreateAt).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt))
		_, _, err := dao.Query().Entities(user).BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("SELECT ")
			b.WriteColumns()
			b.Write(" FROM user")
			b.Write(" WHERE ")
			e := b.Entity()
			b.Write("status=")
			b.Write(b.Pp("$"), e.Status)
			b.Write(" AND ")
			b.Write("level=")
			b.Write(b.Pp("$"), e.Level)
			b.Write(" AND ")
			b.Write("create_at=")
			b.Write(b.Pp("$"), e.CreateAt)
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		// 测试查询字段比实体字段多的情况
		dao, mock := mockUserDao(r)
		mock.ExpectPrepare(`SELECT id, name, age, address, phone, email, status, level, create_at, valid FROM user WHERE id=\? AND status=\?`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "name", "age", "address", "phone", "email", "status", "level", "create_at", "valid"}).
			AddRow(1, "foo", 1, "bar", "123456", "foo@gmail", 2, 1, createAt, true))
		user, users, err := dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("SELECT ")
			b.WriteColumns()
			b.Write(", valid")
			b.Write(" FROM user")
			b.Write(" WHERE id=? AND status=?", 1, 2)
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(user, users[0])
		r.Equal(int32(1), gdao.V(user.Id))
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
			UserId:  gdao.P[int32](1),
			Status:  gdao.P[int8](1),
			Balance: gdao.P[int64](100),
		},
		{
			UserId:  gdao.P[int32](2),
			Status:  gdao.P[int8](1),
			Balance: gdao.P[int64](200),
		},
	}
	mock.ExpectPrepare(`INSERT account\(user_id,status,balance\) VALUES\(\?,\?,\?\),\(\?,\?,\?\) RETURNING id`).
		ExpectQuery().WithArgs(accounts[0].UserId, accounts[0].Status, accounts[0].Balance,
		accounts[1].UserId, accounts[1].Status, accounts[1].Balance).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow(2001).AddRow(2002))
	_, _, err := dao.Query().Entities(accounts...).RowAs(gdao.RowAs_.RETURNING).
		BuildSql(func(b *gdao.DaoSqlBuilder[Account]) {
			b.Write("INSERT account")
			b.Write("(user_id,status,balance) VALUES")
			b.EachEntity(b.Sep(","), func(n int, entity *Account) {
				b.Write("(?,?,?)", entity.UserId, entity.Status, entity.Balance)
			})
			b.Write(" RETURNING id")
		}).Do()

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
			UserId:  gdao.P[int32](1),
			Status:  gdao.P[int8](1),
			Balance: gdao.P[int64](100),
		},
		{
			UserId:  gdao.P[int32](2),
			Status:  gdao.P[int8](1),
			Balance: gdao.P[int64](200),
		},
	}
	mock.ExpectPrepare(`INSERT account\(user_id,status,balance\) VALUES\(\?,\?,\?\),\(\?,\?,\?\); SELECT ID=convert\(bigint, SCOPE_IDENTITY\(\)\)`).
		ExpectQuery().WithArgs(accounts[0].UserId, accounts[0].Status, accounts[0].Balance,
		accounts[1].UserId, accounts[1].Status, accounts[1].Balance).
		WillReturnRows(mock.NewRows([]string{"ID"}).AddRow(1234))
	_, _, err := dao.Query().Entities(accounts...).RowAs(gdao.RowAs_.LAST_ID).
		BuildSql(func(b *gdao.DaoSqlBuilder[Account]) {
			b.Write("INSERT account")
			b.Write("(user_id,status,balance) VALUES")
			b.EachEntity(b.Sep(","), func(n int, entity *Account) {
				b.Write("(?,?,?)", entity.UserId, entity.Status, entity.Balance)
			})
			b.Write("; SELECT ID=convert(bigint, SCOPE_IDENTITY())")
		}).Do()

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
			Status:  gdao.P[int8](3),
			Level:   gdao.P[int32](10),
			Address: gdao.P("address"),
			Phone:   gdao.P("56789"),
		}
		mock.ExpectPrepare(`UPDATE user SET address=\?,phone=\?,status=\?,level=\? WHERE id=\?`).
			ExpectExec().WithArgs(user.Address, user.Phone, user.Status, user.Level, 1001).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := dao.Exec().Entities(user).BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("UPDATE user SET ")
			columns := b.Columns(true)
			b.EachColumn(b.Entity(), b.Sep(","), func(_ int, column string, value any) {
				b.Write(column)
				b.Write("=?", value)
			}, columns...)
			b.Write(" WHERE id=?", 1001)
		}).Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		dao, mock := mockUserDao(r)
		user := &User{
			Name:    gdao.P("foo"),
			Status:  gdao.P[int8](3),
			Level:   gdao.P[int32](10),
			Address: gdao.P("address"),
			Phone:   gdao.P("56789"),
		}
		mock.ExpectPrepare(`INSERT user\(name, address, phone, status, level\) VALUES\(\?,\?,\?,\?,\?\)`).
			ExpectExec().
			WithArgs(user.Name, user.Address, user.Phone, user.Status, user.Level).
			WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := dao.Exec().Entities(user).
			BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
				b.Write("INSERT user")
				columns := b.Columns(true)
				b.Write("(")
				b.WriteColumns(columns...)
				b.Write(")")
				b.EachColumn(b.Entity(), b.SepFix(" VALUES(", ",", ")", true), func(n int, column string, value any) {
					b.Write("?", value)
				}, columns...)
			}).Do()

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
			Status:  gdao.P[int8](3),
			Level:   gdao.P[int32](10),
			Address: gdao.P("address"),
			Phone:   gdao.P("56789"),
		},
		{
			Status:  gdao.P[int8](2),
			Level:   gdao.P[int32](2),
			Address: gdao.P("addr"),
			Phone:   gdao.P("2325325"),
		},
	}
	mock.ExpectPrepare(`INSERT user\(`+export.ColumnsWithComma+`\) VALUES\(\?,\?,\?,\?,\?,\?,\?,\?,\?\),\(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
		ExpectExec().
		WithArgs(users[0].Id, users[0].Name, users[0].Age, users[0].Address, users[0].Phone, users[0].Email, users[0].Status, users[0].Level, users[0].CreateAt,
			users[1].Id, users[1].Name, users[1].Age, users[1].Address, users[1].Phone, users[1].Email, users[1].Status, users[1].Level, users[1].CreateAt).
		WillReturnResult(sqlmock.NewResult(1001, 2))
	affected, err := dao.Exec().Entities(users...).LastInsertIdAs(gdao.LastInsertIdAs_.FIRST_ID).
		BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("INSERT user")
			b.Write("(")
			b.WriteColumns()
			b.Write(") VALUES")
			b.EachEntity(b.Sep(","), func(n int, entity *User) {
				b.EachColumn(entity, b.SepFix("(", ",", ")", false), func(n int, column string, value any) {
					b.Write("?", value)
				}, b.Columns(false)...)
			})
		}).Do()

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
			Status:  gdao.P[int8](3),
			Level:   gdao.P[int32](10),
			Address: gdao.P("address"),
			Phone:   gdao.P("56789"),
		},
		{
			Status:  gdao.P[int8](2),
			Level:   gdao.P[int32](2),
			Address: gdao.P("addr"),
			Phone:   gdao.P("2325325"),
		},
	}
	mock.ExpectPrepare(`INSERT user\(`+export.ColumnsWithComma+`\) VALUES\(\?,\?,\?,\?,\?,\?,\?,\?,\?\),\(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
		ExpectExec().
		WithArgs(users[0].Id, users[0].Name, users[0].Age, users[0].Address, users[0].Phone, users[0].Email, users[0].Status, users[0].Level, users[0].CreateAt,
			users[1].Id, users[1].Name, users[1].Age, users[1].Address, users[1].Phone, users[1].Email, users[1].Status, users[1].Level, users[1].CreateAt).
		WillReturnResult(sqlmock.NewResult(1001, 2))
	affected, err := dao.Exec().Entities(users...).LastInsertIdAs(gdao.LastInsertIdAs_.LAST_ID).
		BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
			b.Write("INSERT user")
			b.Write("(")
			b.WriteColumns()
			b.Write(") VALUES")
			b.EachEntity(b.Sep(","), func(n int, entity *User) {
				b.EachColumn(entity, b.SepFix("(", ",", ")", false), func(n int, column string, value any) {
					b.Write("?", value)
				}, b.Columns(false)...)
			})
		}).Do()

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(1000), *users[0].Id)
	r.Equal(int32(1001), *users[1].Id)
}

func TestDao_Query_FieldConvert(t *testing.T) {
	r := require.New(t)
	{
		dao, mock := mockProductDao(r)
		mock.ExpectPrepare(`SELECT \* FROM product WHERE status = \? AND tag = \?`).
			ExpectQuery().WithArgs(MyStatus(2), "a,b,c").WillReturnRows(mock.NewRows([]string{"id", "tags", "status", "properties", "attributes"}).
			AddRow(1, "a,b,c", 2, "{\"unit\": \"kg\",\"weight\": 10}", "{\"size\": [56, 57, 58],\"color\": [\"red\",\"green\",\"blue\"]}"))
		product, _, err := dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[Product]) {
			b.Write("SELECT * FROM product WHERE status = ? AND tag = ?", MyStatus(2), MyStringSlice{"a", "b", "c"})
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), *product.Id)
		r.Equal(MyStringSlice{"a", "b", "c"}, product.Tags)
		r.Equal(Properties{Unit: "kg", Weight: 10}, *product.Properties)
		r.Equal(Attributes{Size: []int{56, 57, 58}, Color: []string{"red", "green", "blue"}}, *product.Attributes)
	}
}

func TestNewDaoPanic(t *testing.T) {
	r := require.New(t)
	r.PanicsWithError("generics must be struct type", func() {
		gdao.DaoBuilder[*User]().Build()
	})
	r.NotPanics(func() {
		gdao.DaoBuilder[User]().Build()
	})
	r.PanicsWithError(`field "field" of "gdao_test.InvalidField" must be exported`, func() {
		gdao.DaoBuilder[InvalidField]().Build()
	})
	r.NotPanics(func() {
		gdao.DaoBuilder[InvalidField]().AllowInvalidField(true).Build()
	})
	r.PanicsWithError(`field "Field" of "gdao_test.InvalidField2" is not supported type`, func() {
		gdao.DaoBuilder[InvalidField2]().Build()
	})
	r.NotPanics(func() {
		gdao.DaoBuilder[InvalidField2]().AllowInvalidField(true).Build()
	})
	r.PanicsWithError(`field "Field" of "gdao_test.InvalidField3" is not supported type`, func() {
		gdao.DaoBuilder[InvalidField3]().Build()
	})
	r.NotPanics(func() {
		gdao.DaoBuilder[InvalidField3]().AllowInvalidField(true).Build()
	})
	r.PanicsWithError(`field "Field" of "gdao_test.InvalidField4" is not supported type`, func() {
		gdao.DaoBuilder[InvalidField4]().Build()
	})
	r.NotPanics(func() {
		gdao.DaoBuilder[InvalidField4]().AllowInvalidField(true).Build()
	})
	r.PanicsWithError(`field "Field" of "gdao_test.InvalidField5" is invalid implementing gdao.Convert`, func() {
		gdao.DaoBuilder[InvalidField5]().Build()
	})
	r.NotPanics(func() {
		gdao.DaoBuilder[InvalidField5]().AllowInvalidField(true).Build()
	})
}

func TestLastInsertIdConvertors(t *testing.T) {
	r := require.New(t)
	id := int64(123)
	r.Equal(123, gdao.ConvertLastInsertId("int", id))
	r.Equal(int8(123), gdao.ConvertLastInsertId("int8", id))
	r.Equal(int16(123), gdao.ConvertLastInsertId("int16", id))
	r.Equal(int32(123), gdao.ConvertLastInsertId("int32", id))
	r.Equal(int64(123), gdao.ConvertLastInsertId("int64", id))
	r.Equal(uint(123), gdao.ConvertLastInsertId("uint", id))
	r.Equal(uint8(123), gdao.ConvertLastInsertId("uint8", id))
	r.Equal(uint16(123), gdao.ConvertLastInsertId("uint16", id))
	r.Equal(uint32(123), gdao.ConvertLastInsertId("uint32", id))
	r.Equal(uint64(123), gdao.ConvertLastInsertId("uint64", id))
	r.Equal(float32(123), gdao.ConvertLastInsertId("float32", id))
	r.Equal(float64(123), gdao.ConvertLastInsertId("float64", id))
	r.Equal("123", gdao.ConvertLastInsertId("string", id))
}

func checkMap[K comparable, V any](r *require.Assertions, expected map[K]V, actual map[K]V) {
	r.Equal(expected, actual)
}

func checkMapKeys[K comparable, V any](r *require.Assertions, keys []K, actual map[K]V) {
	r.Len(keys, len(actual))
	for _, key := range keys {
		r.Contains(actual, key)
	}
}

//================================================
//=============== Test Builder ===================
//================================================

func TestBuilder_Arg(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		b.SetArgs("a")
		r.Equal(1, len(b.Args()))
		r.Contains(b.Args(), "a")

		b.SetArgs("B")
		r.Equal(2, len(b.Args()))
		r.Contains(b.Args(), "a", "B")
	}).Do()
}

func TestBuilder_Pp(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		b.Write(b.Pp("$"))
		r.Equal("$1", b.Sql())
		b.Write(b.Pp("$"))
		r.Equal("$1$2", b.Sql())
	}).Do()
}

func TestBuilder_SetOk(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		r.True(b.Ok())
		b.SetOk(false)
		r.False(b.Ok())
	}).Do()
}

func TestBuilder_SetError(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		b.SetError(errors.New("this is an error"))
		r.EqualError(b.Error(), "this is an error")
		r.False(b.Ok())
		b.SetOk(true)
		r.False(b.Ok())
	}).Do()
}

func TestBuilder_Columns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().Entities(&User{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
		Phone:   gdao.P("56789"),
	}).BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.Columns, b.Columns(false))
		r.Equal([]string{"id", "name", "age", "email", "status", "level", "create_at"}, b.Columns(false, []string{"address", "phone"}...))
		r.Equal([]string{"address", "phone", "status", "level"}, b.Columns(true))
		r.Equal([]string{"status", "level"}, b.Columns(true, []string{"address", "phone"}...))
	}).Do()
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		r.Empty(b.Columns(true))
	}).Do()
}

func TestBuilder_AutoColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		exportDao := gdao.ExportDao(dao)
		r.Equal(exportDao.AutoIncrementColumns, b.AutoColumns())
	}).Do()
}

func TestBuilder_Entity(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	u := &User{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
		Phone:   gdao.P("56789"),
	}
	dao.Query().Entities(u, &User{
		Status:  gdao.P[int8](2),
		Level:   gdao.P[int32](2),
		Address: gdao.P("addr"),
		Phone:   gdao.P("2325325"),
	}).BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		r.Equal(u, b.Entity())
	}).Do()
}

func TestBuilder_EachColumn(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().Entities(&User{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
	}).BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
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
	}).Do()
}

func TestBuilder_ColumnValue(t *testing.T) {
	r := require.New(t)
	dao, _ := mockUserDao(r)
	dao.Query().Entities(&User{
		Status:  gdao.P[int8](3),
		Level:   gdao.P[int32](10),
		Address: gdao.P("address"),
		Phone:   gdao.P("56789"),
	}).BuildSql(func(b *gdao.DaoSqlBuilder[User]) {
		r.Nil(b.ColumnValue(nil, ""))
		r.Nil(b.ColumnValue(b.Entity(), ""))
		r.Nil(b.ColumnValue(b.Entity(), "name"))
		r.Equal(int8(3), reflect.ValueOf(b.ColumnValue(b.Entity(), "status")).Elem().Interface())
		r.Equal("address", reflect.ValueOf(b.ColumnValue(b.Entity(), "address")).Elem().Interface())
	}).Do()
}

func TestBuilder_Repeat(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(r)
	dao.Query().Entities(&Account{
		UserId:  gdao.P[int32](1),
		Status:  gdao.P[int8](1),
		Balance: gdao.P[int64](100),
	}).BuildSql(func(b *gdao.DaoSqlBuilder[Account]) {
		b.Repeat(6, b.SepFix("(", ",", ")", false), func(i int) bool {
			return i != 2 && i != 4
		}, func(n, i int) {
			b.Write(strconv.Itoa(n))
			b.Write("-")
			b.Write(strconv.Itoa(i))
		})
		r.Equal("(1-0,2-1,3-3,4-5)", b.Sql())
	}).Do()
}

func TestBuilder_WriteColumns(t *testing.T) {
	r := require.New(t)
	dao, _ := mockAccountDao(r)
	dao.Query().BuildSql(func(b *gdao.DaoSqlBuilder[Account]) {
		b.WriteColumns("id", "", "user_id")
		r.Equal("id, user_id", b.Sql())

		b.WriteColumns()
		r.Equal("id, user_idid, other_id, user_id, status, balance, licence_file", b.Sql())
	}).Do()
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
	dao.Query().Entities(a, a2).BuildSql(func(b *gdao.DaoSqlBuilder[Account]) {
		b.EachEntity(b.Sep(","), func(n int, entity *Account) {
			switch n {
			case 1:
				r.Equal(a, entity)
			case 2:
				r.Equal(a2, entity)
			}
		})
	}).Do()
}
