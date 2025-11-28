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

package oracle_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jishaocong0910/gdao/gen"
	go_ora "github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOracle(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "gvenzl/oracle-free:slim-faststart",
		ExposedPorts: []string{"1521/tcp"},
		Env: map[string]string{
			"APP_USER":               "test",
			"APP_USER_PASSWORD":      "12345678",
			"ORACLE_RANDOM_PASSWORD": "yes",
		},
		WaitingFor: wait.ForLog("DATABASE IS READY TO USE!").WithStartupTimeout(time.Minute * 5),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	r.NoError(err)
	containerPort, err := container.MappedPort(ctx, "1521/tcp")
	r.NoError(err)
	host, err := container.Host(ctx)
	r.NoError(err)

	dsn := go_ora.BuildUrl(host, containerPort.Int(), "FREEPDB1", "test", "12345678", nil)
	db, err := sql.Open("oracle", dsn)
	r.NoError(err)

	script, err := os.ReadFile("testdata/init_script.sql")
	r.NoError(err)
	sqls := strings.Split(string(script), ";")
	for _, s := range sqls {
		s = strings.TrimSpace(s)
		if s != "" {
			_, err = db.Exec(s)
			r.NoError(err)
		}
	}
	db.Close()

	gen.GetGenerator(gen.GenCfg{
		DbType:    gen.DbType_.ORACLE,
		Dsn:       dsn,
		GoModPath: "../../..",
		OutPath:   "gen/test/oracle/testdata",
		TableCfg: gen.TableCfg{
			Tables: gen.Tables{"TEST_TABLE"},
		},
		DaoCfg: gen.DaoCfg{
			CoverBaseDao:      true,
			AllowInvalidField: true,
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
