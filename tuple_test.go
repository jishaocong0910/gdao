// Copyright 2024-present jishaocong0910
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm_test

import (
	"testing"
	"time"

	orm "github.com/jishaocong0910/cozy-orm"
	"github.com/stretchr/testify/require"
)

func TestTuple(t *testing.T) {
	r := require.New(t)
	tm := time.UnixMilli(1749198596000)
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "level", "attributes", "create_at", "phone"}).
			AddRow(1, tm, 5, `{"key1": "value1","key2": "value2"}`, "phone"))
		tuples, err := db.Query[orm.Tuple4[*int64, *time.Time, *orm.UserLevel, orm.UserAttributes]](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.NoError(err)
		r.Equal(new(int64(1)), tuples[0].Field1)
		r.Equal(new(tm), tuples[0].Field2)
		r.Equal(new(orm.UserLevel("5")), tuples[0].Field3)
		r.Equal(orm.UserAttributes{"key1": "value1", "key2": "value2"}, tuples[0].Field4)
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"name", "level", "attributes", "create_at", "email", "phone"}).
			AddRow("name", tm, 5, `{"source":"unknown","country":"unknown"}`, "aa,bb,cc", `{"key1": "value1","key2": "value2"}`))
		tuples, err := db.Query[orm.Tuple6[string, time.Time, orm.UserLevel, orm.UserProperties, *orm.UserTags, *orm.UserAttributes]](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.NoError(err)
		r.Equal("name", tuples[0].Field1)
		r.Equal(tm, tuples[0].Field2)
		r.Equal(orm.UserLevel("5"), tuples[0].Field3)
		r.Equal(orm.UserProperties{Source: "unknown", Country: "unknown"}, tuples[0].Field4)
		r.Equal(new(orm.UserTags{"aa", "bb", "cc"}), tuples[0].Field5)
		r.Equal(new(orm.UserAttributes{"key1": "value1", "key2": "value2"}), tuples[0].Field6)
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"name", "phone", "properties", "tags"}).
			AddRow(nil, nil, nil, nil))
		tuples, err := db.Query[orm.Tuple4[string, *string, orm.UserProperties, orm.UserTags]](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.NoError(err)
		r.Equal("", tuples[0].Field1)
		r.Nil(tuples[0].Field2)
		r.Equal(orm.UserProperties{}, tuples[0].Field3)
		r.Equal(*new(orm.UserTags), tuples[0].Field4)
		r.Nil(tuples[0].Field4)
	}
}
