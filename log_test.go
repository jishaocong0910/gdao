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

package orm_test

import (
	"context"
	"errors"
	"testing"

	orm "github.com/jishaocong0910/cozy-orm"
	"github.com/stretchr/testify/require"
)

type MockLogger struct {
	msg  string
	args []any
}

func (d *MockLogger) Debugf(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Infof(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Warnf(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func (d *MockLogger) Errorf(ctx context.Context, msg string, args ...interface{}) {
	d.msg = msg
	d.args = args
}

func TestPrintSql(t *testing.T) {
	r := require.New(t)
	{
		log := &MockLogger{}
		orm.Config(orm.Cfg{nil, log, orm.LogLevel_.DEBUG, false, nil})
		orm.PrintSql(nil, orm.LogLevel_.Undefined(), "update a user", "UPDATE user SET status=?,phone=?,email=? WHERE level=?)", []any{2, nil, (*int)(nil), orm.P("abc")}, 15, -1, -1, errors.New("error"))
		r.Equal(`desc: %s, SQL: %s; args: %v, affected: %d, error: %+v`, log.msg)
		r.Len(log.args, 5)
		r.Equal("update a user", log.args[0])
		r.Equal("UPDATE user SET status=?,phone=?,email=? WHERE level=?)", log.args[1])
		args := log.args[2].([]any)
		r.Len(args, 4)
		r.Equal(2, args[0])
		r.Equal(nil, args[1])
		r.Equal(nil, args[2])
		r.Equal(`"abc"`, args[3])
		r.Equal(int64(15), log.args[3])
		r.EqualError(log.args[4].(error), "error")
	}
	{
		log := &MockLogger{}
		orm.Config(orm.Cfg{nil, log, orm.LogLevel_.DEBUG, true, nil})
		orm.PrintSql(nil, orm.LogLevel_.Undefined(),
			"", `  
SELECT *
  FROM
user`, nil, -1, 10, -1, nil)
		r.Equal("SQL: %s; rowcount: %d", log.msg)
		r.Equal("SELECT *  FROM user", log.args[0])
	}
	{
		log := &MockLogger{}
		orm.Config(orm.Cfg{nil, log, orm.LogLevel_.DEBUG, true, nil})
		orm.PrintSql(nil, orm.LogLevel_.Undefined(),
			"", `SELECT COUNT(*) FROM user`, nil, -1, -1, 2, nil)
		r.Equal("SQL: %s; count: %d", log.msg)
		r.Equal("SELECT COUNT(*) FROM user", log.args[0])
	}
}

func TestPrintWarn(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	orm.Config(orm.Cfg{nil, log, orm.LogLevel_.DEBUG, false, nil})
	orm.PrintWarn(nil, errors.New("warn"))
	r.Equal("warn", log.msg)
}
