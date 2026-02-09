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

package sqlite_test

import (
	"testing"
	"time"

	"github.com/jishaocong0910/cozy-orm/dial/test/sqlite/internal"

	"github.com/DATA-DOG/go-sqlmock"
	orm "github.com/jishaocong0910/cozy-orm"
	"github.com/stretchr/testify/require"
)

type User struct {
	Id       *int32     `orm:"column=id;auto"`
	Name     *string    `orm:"column=name"`
	Age      *int32     `orm:"column=age"`
	Address  *string    `orm:"column=address"`
	Phone    *string    `orm:"column=phone"`
	Email    *string    `orm:"column=email"`
	Status   *int8      `orm:"column=status"`
	Level    *int32     `orm:"column=level"`
	CreateAt *time.Time `orm:"column=create_at"`
}

func (u User) Table() string {
	return "user"
}

type Product struct {
	Id      *int32  `orm:"column=id;auto"`
	Name    *string `orm:"column=name"`
	Channel *int8   `orm:"column=channel"`
	Status  *int8   `orm:"column=status"`
	Valid   *string `orm:"column=valid"`
}

func (p Product) Table() string {
	return "product"
}

type Sku struct {
	Id        *int32  `orm:"column=id;auto"`
	ProductId *int32  `orm:"column=product_id"`
	ItemNo    *string `orm:"column=item_no"`
	Channel   *int8   `orm:"column=channel"`
	Status    *int8   `orm:"column=status"`
	Deleted   *int32  `orm:"column=Valid"`
}

func (s Sku) Table() string {
	return "sku"
}

func TestBaseDao_Get(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`SELECT id, name FROM user WHERE status = \? ORDER BY name ASC, id DESC FOR UPDATE`).
			ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "lucy"))

		get, err := d.Get().Select("id", "name").Condition(testdata.And().Eq("status", 4)).
			Order(testdata.Order().Asc("name").Desc("id")).
			ForUpdate(true).
			Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int32(1), *get.Id)
		r.Equal("lucy", *get.Name)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"name"}).
			AddRow("lucy").AddRow("jack"))
		_, err := d.Get().Select("name").Condition(testdata.And().Eq("status", 1)).Do()
		r.NoError(err)

		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"name"}).
			AddRow("lucy").AddRow("jack"))
		_, err = d.Get().Select("name").CheckOne(true).Condition(testdata.And().Eq("status", 1)).Do()
		r.EqualError(err, "return more than one row")
	}
}

func TestBaseDao_List(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`SELECT id, name FROM user WHERE status = \? ORDER BY name ASC, address DESC LIMIT 10 OFFSET 3 FOR UPDATE`).
			ExpectQuery().WithArgs(4).WillReturnRows(mock.NewRows([]string{"id", "name"}).
			AddRow(1, "lucy").AddRow(2, "nick"))
		list, err := d.List().Select("id", "Name").Condition(testdata.And().Eq("status", 4)).
			OrderBy(testdata.Order().Asc("name").Desc("address")).
			Page(testdata.Page(3, 10)).
			ForUpdate(true).
			Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Len(list, 2)
		r.Equal(int32(1), *list[0].Id)
		r.Equal("lucy", *list[0].Name)
		r.Equal(int32(2), *list[1].Id)
		r.Equal("nick", *list[1].Name)
	}
}

func TestBaseDao_Insert(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`INSERT INTO user\(name, phone, email, status\) VALUES \(\?, \?, \?, NULL\)`).
			ExpectExec().WithArgs("abc", "12345", "email").WillReturnResult(sqlmock.NewResult(7, 1))

		u := &User{
			Name:  orm.P("abc"),
			Phone: orm.P("12345"),
			Email: orm.P("email"),
		}
		affected, err := d.Insert().Entity(u).Nullable("email", "status").Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Equal(int32(7), *u.Id)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`INSERT INTO user\(name, age, address, phone, status, level, create_at\) VALUES \(\?, NULL, NULL, \?, NULL, NULL, NULL\)`).
			ExpectExec().WithArgs("abc", "12345").WillReturnResult(sqlmock.NewResult(7, 1))

		u := &User{
			Name:  orm.P("abc"),
			Phone: orm.P("12345"),
			Email: orm.P("email"),
		}
		affected, err := d.Insert().Entity(u).All(true).Ignore("email").Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Equal(int32(7), *u.Id)
	}
}

func TestBaseDao_InsertBatch(t *testing.T) {
	r := require.New(t)
	d, mock := testdata.MockBaseDao[User](r, nil)
	mock.ExpectPrepare(`INSERT INTO user\(name, phone, email, level\) VALUES \(\?, \?, \?, \?\), \(\?, \?, \?, NULL\)`).
		ExpectExec().WithArgs("abc", "12345", "email11", 2, "def", "6789", "email22").WillReturnResult(sqlmock.NewResult(8, 2))

	u := &User{
		Name:  orm.P("abc"),
		Phone: orm.P("12345"),
		Email: orm.P("email11"),
		Level: orm.P[int32](2),
	}
	u2 := &User{
		Name:  orm.P("def"),
		Phone: orm.P("6789"),
		Email: orm.P("email22"),
	}
	affected, err := d.InsertBatch().Entities(u, u2).Nullable("level").Do()

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(2), affected)
	r.Equal(int32(7), *u.Id)
	r.Equal(int32(8), *u2.Id)
}

func TestBaseDao_Update(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`UPDATE user SET name = \?, phone = \?, email = NULL WHERE status = \? AND level = \? AND age = \?`).
			ExpectExec().WithArgs("name", "123", 2, 10, 20).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Name:   orm.P("name"),
			Status: orm.P[int8](2),
			Phone:  orm.P("123"),
			Level:  orm.P[int32](10),
		}
		affected, err := d.Update().Entity(u).Nullable("email", "phone", "status").Where("status", "level").
			Condition(testdata.And().Eq("age", 20)).Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`UPDATE user SET name = \?, age = NULL, address = \?, phone = NULL, level = \?, create_at = NULL WHERE id = \? AND status IS NULL`).
			ExpectExec().WithArgs("name", "addr", 10, 1).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Id:      orm.P(int32(1)),
			Name:    orm.P("name"),
			Address: orm.P("addr"),
			Level:   orm.P[int32](10),
		}
		affected, err := d.Update().Entity(u).All(true).Ignore("email").Where("id", "status").Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
}

func TestBaseDao_UpdateBatch(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`UPDATE user SET name = CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END, phone = CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END, level = CASE id WHEN \? THEN NULL WHEN \? THEN NULL WHEN \? THEN NULL END WHERE id IN\(\?, \?, \?\) AND status = \?`).
			ExpectExec().WithArgs(1, "name1", 2, "name2", 3, "name3", 1, "phone1", 2, "phone2", 3, "phone3", 1, 2, 3, 1, 2, 3, 1).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Id:    orm.P[int32](1),
			Name:  orm.P("name1"),
			Phone: orm.P("phone1"),
		}
		u2 := &User{
			Id:    orm.P[int32](2),
			Name:  orm.P("name2"),
			Phone: orm.P("phone2"),
		}
		u3 := &User{
			Id:    orm.P[int32](3),
			Name:  orm.P("name3"),
			Phone: orm.P("phone3"),
		}
		affected, err := d.UpdateBatch().Entities(u, u2, u3).Nullable("phone", "level").Where("id").
			Condition(testdata.And().Eq("Status", 1)).Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`UPDATE user SET name = CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END, age = CASE id WHEN \? THEN NULL WHEN \? THEN NULL WHEN \? THEN NULL END, address = CASE id WHEN \? THEN NULL WHEN \? THEN NULL WHEN \? THEN NULL END, phone = CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END, email = CASE id WHEN \? THEN \? WHEN \? THEN \? WHEN \? THEN \? END, create_at = CASE id WHEN \? THEN NULL WHEN \? THEN NULL WHEN \? THEN NULL END WHERE id IN\(\?, \?, \?\)`).
			ExpectExec().WithArgs(1, "name1", 2, "name2", 3, "name3", 1, 2, 3, 1, 2, 3, 1, "phone1", 2, "phone2", 3, "phone3", 1, "email1", 2, "email2", 3, "email3", 1, 2, 3, 1, 2, 3).WillReturnResult(sqlmock.NewResult(0, 3))

		u := &User{
			Id:     orm.P[int32](1),
			Name:   orm.P("name1"),
			Phone:  orm.P("phone1"),
			Email:  orm.P("email1"),
			Status: orm.P[int8](1),
		}
		u2 := &User{
			Id:     orm.P[int32](2),
			Name:   orm.P("name2"),
			Phone:  orm.P("phone2"),
			Email:  orm.P("email2"),
			Status: orm.P[int8](1),
		}
		u3 := &User{
			Id:     orm.P[int32](3),
			Name:   orm.P("name3"),
			Phone:  orm.P("phone3"),
			Email:  orm.P("email3"),
			Status: orm.P[int8](2),
		}
		affected, err := d.UpdateBatch().Entities(u, u2, u3).All(true).Ignore("Status", "level").Where("id").Do()

		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), affected)
	}
}

func TestBaseDao_Delete(t *testing.T) {
	r := require.New(t)
	d, mock := testdata.MockBaseDao[User](r, nil)
	mock.ExpectPrepare(`DELETE FROM user WHERE status = \?`).
		ExpectExec().WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 3))

	affected, err := d.Delete().Condition(testdata.And().Eq("status", 1)).Do()

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(3), affected)
}

func TestBaseDao_Count(t *testing.T) {
	r := require.New(t)
	d, mock := testdata.MockBaseDao[User](r, nil)

	mock.ExpectPrepare(`SELECT COUNT\(\*\) FROM user WHERE status = \?`).
		ExpectQuery().WithArgs(1).WillReturnRows(mock.NewRows([]string{"count"}).
		AddRow(8))

	count, err := d.Count().Condition(testdata.And().Eq("status", 1)).Do()

	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
	r.Equal(int64(8), count.Int64())
}

func TestLogicalDel_SetId(t *testing.T) {
	r := require.New(t)
	d, mock := testdata.MockBaseDao[Sku](r, &testdata.LogicalDelCfg{Column: "deleted", IdColumn: "id", QueryValue: 0})
	{
		mock.ExpectPrepare(`SELECT id, item_no FROM sku WHERE \(channel = \? OR status = \?\) AND deleted = \?`).
			ExpectQuery().WithArgs(1, 2, 0).WillReturnRows(mock.NewRows([]string{"id", "item_no"}))
		_, err := d.List().Select("id", "item_no").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`SELECT id, item_no FROM sku WHERE channel = \? OR status = \? `).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "item_no"}))
		_, err = d.List().Select("id", "item_no").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`SELECT id, item_no FROM sku WHERE \(channel = \? OR status = \?\) AND deleted = \?`).
			ExpectQuery().WithArgs(1, 2, 0).WillReturnRows(mock.NewRows([]string{"id", "item_no"}))
		_, err := d.Get().Select("id", "item_no").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`SELECT id, item_no FROM sku WHERE channel = \? OR status = \? `).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "item_no"}))
		_, err = d.Get().Select("id", "item_no").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).
			Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`UPDATE sku SET item_no = \? WHERE channel = \? AND status = \? AND deleted = \?`).
			ExpectExec().WithArgs("1001", 1, 2, 0).WillReturnResult(sqlmock.NewResult(0, 2))
		_, err := d.Update().Entity(&Sku{ItemNo: orm.P("1001")}).Condition(testdata.And().Eq("channel", 1).Eq("status", 2)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`UPDATE sku SET item_no = \? WHERE channel = \? AND status = \?`).
			ExpectExec().WithArgs("1001", 1, 2).WillReturnResult(sqlmock.NewResult(0, 2))
		_, err = d.Update().Entity(&Sku{ItemNo: orm.P("1001")}).Condition(testdata.And().Eq("channel", 1).Eq("status", 2)).
			Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`UPDATE sku SET item_no = CASE id WHEN \? THEN \? WHEN \? THEN \? END WHERE id IN\(\?, \?\) AND status = \? AND deleted = \?`).
			ExpectExec().WithArgs(1, "1001", 2, "1002", 1, 2, 5, 0).WillReturnResult(sqlmock.NewResult(0, 2))
		_, err := d.UpdateBatch().Entities(&Sku{Id: orm.P[int32](1), ItemNo: orm.P("1001")}, &Sku{Id: orm.P[int32](2), ItemNo: orm.P("1002")}).
			Where("id").Condition(testdata.Or().Eq("status", 5)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`UPDATE sku SET item_no = CASE id WHEN \? THEN \? WHEN \? THEN \? END WHERE id IN\(\?, \?\) AND status = \?`).
			ExpectExec().WithArgs(1, "1001", 2, "1002", 1, 2, 5).WillReturnResult(sqlmock.NewResult(0, 2))
		_, err = d.UpdateBatch().Entities(&Sku{Id: orm.P[int32](1), ItemNo: orm.P("1001")}, &Sku{Id: orm.P[int32](2), ItemNo: orm.P("1002")}).
			Where("id").Condition(testdata.Or().Eq("status", 5)).Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`SELECT COUNT\(\*\) FROM sku WHERE status = \? AND deleted = \?`).
			ExpectQuery().WithArgs(3, 0).WillReturnRows(sqlmock.NewRows([]string{"c"}))
		_, err := d.Count().Condition(testdata.And().Eq("status", 3)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`SELECT COUNT\(\*\) FROM sku WHERE status = \?`).
			ExpectQuery().WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"c"}))
		_, err = d.Count().Condition(testdata.And().Eq("status", 3)).Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`UPDATE sku SET deleted = id WHERE id = \? AND deleted = ?`).
			ExpectExec().WithArgs(1, 0).WillReturnResult(sqlmock.NewResult(0, 1))
		_, err := d.LogicalDelete().Condition(testdata.And().Eq("id", 1)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestLogicalDel_SetNull(t *testing.T) {
	r := require.New(t)
	d, mock := testdata.MockBaseDao[Product](r, &testdata.LogicalDelCfg{Column: "valid", QueryValue: "Y"})
	{
		mock.ExpectPrepare(`SELECT id, name FROM product WHERE \(channel = \? OR status = \?\) AND valid = \?`).
			ExpectQuery().WithArgs(1, 2, "Y").WillReturnRows(mock.NewRows([]string{"id", "name"}))
		_, err := d.List().Select("id", "name").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`SELECT id, name FROM product WHERE channel = \? OR status = \? `).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "name"}))
		_, err = d.List().Select("id", "name").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`SELECT id, name FROM product WHERE \(channel = \? OR status = \?\) AND valid = \?`).
			ExpectQuery().WithArgs(1, 2, "Y").WillReturnRows(mock.NewRows([]string{"id", "name"}))
		_, err := d.Get().Select("id", "name").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`SELECT id, name FROM product WHERE channel = \? OR status = \? `).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows([]string{"id", "name"}))
		_, err = d.Get().Select("id", "name").Condition(testdata.Or().Eq("channel", 1).Eq("status", 2)).
			Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`UPDATE product SET name = \? WHERE channel = \? AND status = \? AND valid = \?`).
			ExpectExec().WithArgs("shirt", 1, 2, "Y").WillReturnResult(sqlmock.NewResult(0, 2))
		_, err := d.Update().Entity(&Product{Name: orm.P("shirt")}).Condition(testdata.And().Eq("channel", 1).Eq("status", 2)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`UPDATE product SET name = \? WHERE channel = \? AND status = \?`).
			ExpectExec().WithArgs("shirt", 1, 2).WillReturnResult(sqlmock.NewResult(0, 2))
		_, err = d.Update().Entity(&Product{Name: orm.P("shirt")}).Condition(testdata.And().Eq("channel", 1).Eq("status", 2)).
			Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`UPDATE product SET name = CASE id WHEN \? THEN \? WHEN \? THEN \? END WHERE id IN\(\?, \?\) AND status = \? AND valid = \?`).
			ExpectExec().WithArgs(1, "shirt", 2, "dress", 1, 2, 5, "Y").WillReturnResult(sqlmock.NewResult(0, 2))
		_, err := d.UpdateBatch().Entities(&Product{Id: orm.P[int32](1), Name: orm.P("shirt")}, &Product{Id: orm.P[int32](2), Name: orm.P("dress")}).
			Where("id").Condition(testdata.Or().Eq("status", 5)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`UPDATE product SET name = CASE id WHEN \? THEN \? WHEN \? THEN \? END WHERE id IN\(\?, \?\) AND status = \?`).
			ExpectExec().WithArgs(1, "shirt", 2, "dress", 1, 2, 5).WillReturnResult(sqlmock.NewResult(0, 2))
		_, err = d.UpdateBatch().Entities(&Product{Id: orm.P[int32](1), Name: orm.P("shirt")}, &Product{Id: orm.P[int32](2), Name: orm.P("dress")}).
			Where("id").Condition(testdata.Or().Eq("status", 5)).Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`SELECT COUNT\(\*\) FROM product WHERE status = \? AND valid = \?`).
			ExpectQuery().WithArgs(3, "Y").WillReturnRows(sqlmock.NewRows([]string{"c"}))
		_, err := d.Count().Condition(testdata.And().Eq("status", 3)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare(`SELECT COUNT\(\*\) FROM product WHERE status = \?`).
			ExpectQuery().WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"c"}))
		_, err = d.Count().Condition(testdata.And().Eq("status", 3)).Deletion(true).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		mock.ExpectPrepare(`UPDATE product SET valid = NULL WHERE id = \? AND valid = ?`).
			ExpectExec().WithArgs(1, "Y").WillReturnResult(sqlmock.NewResult(0, 1))
		_, err := d.LogicalDelete().Condition(testdata.And().Eq("id", 1)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestCondition(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`c1 = \? AND c2 = \?`).
			ExpectQuery().WithArgs(1, 2).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query().BuildSql(func(b *orm.SqlBuilder[User]) {
			c := testdata.And().Eq("c1", 1).Add(nil)
			c2 := testdata.Or().Eq("c2", 2).Add(nil)
			c = testdata.And().Add(c).Add(c2)

			str, args := c.ToStrArgs(nil)
			r.Equal("c1 = ? AND c2 = ?", str)
			r.Len(args, 2)
			r.Contains(args, 1)
			r.Contains(args, 2)
			testdata.WriteCondition(c, b)
		}).Do()
		r.NoError(err)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`c1 = \? AND c2 <> \? AND c3 > \? AND c4 < \? AND c5 >= \? AND c6 <= \? AND c7 LIKE \? AND c8 LIKE \? AND c9 LIKE \? AND c10 IN\(\?, \?, \?\) AND c11 BETWEEN \? AND \? AND c12 IS NULL AND c13 IS NOT NULL`).
			ExpectQuery().WithArgs(1, 2, 3, 4, 5, 6, "%abc%", "abc%", "%abc", 1, 2, 3, 1, 3).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query().BuildSql(func(b *orm.SqlBuilder[User]) {
			c := testdata.And().Eq("c1", 1).
				Ne("c2", 2).
				Gt("c3", 3).
				Lt("c4", 4).
				Ge("c5", 5).
				Le("c6", 6).
				Like("c7", "abc").
				LikeLeft("c8", "abc").
				LikeRight("c9", "abc").
				In("c10", testdata.InArgs([]int{1, 2, 3}...)).
				Between("c11", 1, 3).
				IsNull("c12").
				IsNotNull("c13")
			testdata.WriteCondition(c, b)
		}).Do()
		r.NoError(err)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`0 = 0 AND NOT c1 = \? AND NOT \(c2 = \? AND c3 = \? AND 1 = 1 and 2 = 2\) AND c4 = \? AND NOT \(c5 = \? OR NOT c6 = \?\) AND NOT c7 = \? AND NOT \(c8 = \? OR c9 = \?\)`).
			ExpectQuery().WithArgs(1, 2, 3, 4, 5, 6, 7, 8, 9).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query().BuildSql(func(b *orm.SqlBuilder[User]) {
			c0 := testdata.And().Plain("0 = 0")
			c1 := testdata.Not().And().Eq("c1", 1)
			c2 := testdata.Not().And().Eq("c2", 2).Eq("c3", 3).Plain("1 = 1 and 2 = 2")
			c3 := testdata.Or().Eq("c4", 4)
			c4 := testdata.Or().Eq("c5", 5).Not().Eq("c6", 6)
			c5 := testdata.Not().Or().Eq("c7", 7)
			c6 := testdata.Not().Or().Eq("c8", 8).Eq("c9", 9)
			c := testdata.And().Add(c0).Add(c1).Add(c2).Add(c3).Not().Add(c4).Add(c5).Add(c6)
			testdata.WriteCondition(c, b)
		}).Do()
		r.NoError(err)
	}
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`\(c1 = \? OR c2 = \?\) AND c3 = \? AND c4 = \?`).
			ExpectQuery().WithArgs(1, 2, 3, 4).WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query().BuildSql(func(b *orm.SqlBuilder[User]) {
			c1 := testdata.Or().Eq("c1", 1).Eq("c2", 2)
			c2 := testdata.Or().Eq("c3", 3)
			c3 := testdata.Or().Eq("c4", 4)
			c := testdata.And().Add(c1).Add(c2).Add(c3)
			testdata.WriteCondition(c, b)
		}).Do()
		r.NoError(err)
	}
}

func TestCondOpt(t *testing.T) {
	r := require.New(t)
	{
		d, mock := testdata.MockBaseDao[User](r, nil)
		mock.ExpectPrepare(`1 = 1`).
			ExpectQuery().WillReturnRows(mock.NewRows(nil))

		_, _, err := d.Query().BuildSql(func(b *orm.SqlBuilder[User]) {
			c := testdata.And().Plain("1 = 1").
				Eq("c1", nil, testdata.WithIfPresent()).
				Eq("c1", 1, testdata.WithIfPredicate(func() bool { return false })).
				Ne("c1", nil, testdata.WithIfPresent()).
				Ne("c1", 1, testdata.WithIfPredicate(func() bool { return false })).
				Gt("c1", nil, testdata.WithIfPresent()).
				Gt("c1", 1, testdata.WithIfPredicate(func() bool { return false })).
				Lt("c1", nil, testdata.WithIfPresent()).
				Lt("c1", 1, testdata.WithIfPredicate(func() bool { return false })).
				Ge("c1", nil, testdata.WithIfPresent()).
				Ge("c1", 1, testdata.WithIfPredicate(func() bool { return false })).
				Le("c1", nil, testdata.WithIfPresent()).
				Le("c1", 1, testdata.WithIfPredicate(func() bool { return false })).
				Like("c1", "", testdata.WithIfPresent()).
				Like("c1", "1", testdata.WithIfPredicate(func() bool { return false })).
				LikeLeft("c1", "", testdata.WithIfPresent()).
				LikeLeft("c1", "1", testdata.WithIfPredicate(func() bool { return false })).
				LikeRight("c1", "", testdata.WithIfPresent()).
				LikeRight("c1", "1", testdata.WithIfPredicate(func() bool { return false })).
				In("c1", nil, testdata.WithIfPresent()).
				In("c1", testdata.InArgs(1, 2), testdata.WithIfPredicate(func() bool { return false })).
				Between("c1", nil, 1, testdata.WithIfPresent()).
				Between("c1", 1, nil, testdata.WithIfPresent()).
				Between("c1", nil, nil, testdata.WithIfPresent()).
				Between("c1", 1, 1, testdata.WithIfPredicate(func() bool { return false }))
			testdata.WriteCondition(c, b)
		}).Do()
		r.NoError(err)
	}
}
