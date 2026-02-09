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

package sqlite

import (
	"os"
	"testing"

	"github.com/jishaocong0910/cozy-orm/dial"
	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {
	r := require.New(t)
	dial.GetGenerator(dial.GenCfg{
		DbType:    dial.DbType_.SQLITE,
		Dsn:       "testdata/sqlite.db",
		GoModPath: "../../..",
		PathCfg: dial.PathCfg{
			OutPath: "dial/test/sqlite/testdata",
		},
		TableCfg: dial.TableCfg{
			Tables: dial.Tables{"test_table"},
		},
	}).Gen()

	defer os.RemoveAll("testdata/entity")
	defer os.Remove("testdata/test_table.go")
	defer os.Remove("testdata/base_dao.go")

	compareFile(r, "testdata/entity.golden", "testdata/entity/test_table.go")
	compareFile(r, "internal/base_dao.go", "testdata/base_dao.go")
}

func compareFile(r *require.Assertions, golden, gen string) {
	goldenEntity, err := os.ReadFile(golden)
	r.NoError(err)
	genEntity, err := os.ReadFile(gen)
	r.NoError(err)
	r.Equal(string(goldenEntity), string(genEntity))
}
