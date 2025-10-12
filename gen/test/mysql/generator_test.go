/*
Copyright 2024-present jishaocong0910

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mysql_test

import (
	"context"
	"github.com/jishaocong0910/gdao/gen"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"testing"
	"time"
)

func TestMySql(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	mysqlContainer, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0.36"),
		mysql.WithDatabase("test"),
		mysql.WithUsername("root"),
		mysql.WithPassword("12345678"),
		mysql.WithScripts("testdata/init_script.sql"),
		testcontainers.WithWaitStrategyAndDeadline(time.Minute*5, wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(time.Minute*5)),
	)
	r.NoError(err)

	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	dsn, err := mysqlContainer.ConnectionString(ctx)
	r.NoError(err)

	gen.GetGenerator(gen.GenCfg{
		DbType:  gen.DbType_.MYSQL,
		Dsn:     dsn,
		OutPath: "testdata",
		Package: "dao",
		TableCfg: gen.TableCfg{
			Tables: gen.Tables{"test_table"},
			MappingTypes: gen.MappingTypes{
				"test_table": gen.Mapper{
					"other":  "any",
					"other2": "*rune",
					"other3": "string",
					"other4": "[]string",
				},
			},
			IgnoreColumns: gen.IgnoreColumns{
				"test_table": gen.Columns{
					"other5",
				},
			},
		},
		DaoCfg: gen.DaoCfg{
			GenDao:            true,
			CoverBaseDao:      true,
			GenCountDao:       true,
			AllowInvalidField: true,
		},
	}).Gen()

	defer os.Remove("testdata/test_table.go")
	defer os.Remove("testdata/test_table_dao.go")
	defer os.Remove("testdata/base_dao.go")
	defer os.Remove("testdata/count_dao.go")

	compareFile(r, "testdata/entity.golden", "testdata/test_table.go")
	compareFile(r, "testdata/dao.golden", "testdata/test_table_dao.go")
	compareFile(r, "internal/base_dao.go", "testdata/base_dao.go")
	compareFile(r, "testdata/count_dao.golden", "testdata/count_dao.go")
}

func compareFile(r *require.Assertions, golden, gen string) {
	goldenEntity, err := os.ReadFile(golden)
	r.NoError(err)
	genEntity, err := os.ReadFile(gen)
	r.NoError(err)
	r.Equal(string(goldenEntity), string(genEntity))
}
