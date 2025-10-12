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

package postgres

import (
	"context"
	"github.com/jishaocong0910/gdao/gen"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"testing"
	"time"
)

func TestPostgres(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("12345678"),
		postgres.WithInitScripts("testdata/init_script.sql"),
		testcontainers.WithWaitStrategyAndDeadline(time.Minute*5, wait.ForLog("database system is ready to accept connections").WithStartupTimeout(time.Minute*5)),
	)
	r.NoError(err)

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	time.Sleep(time.Second * 5)

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	r.NoError(err)

	gen.GetGenerator(gen.GenCfg{
		DbType:  gen.DbType_.POSTGRES,
		Dsn:     dsn,
		OutPath: "testdata",
		Package: "dao",
		TableCfg: gen.TableCfg{
			Tables: gen.Tables{"test_table"},
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
	compareFile(r, "internal/base_dao.go", "testdata/base_dao.go")
}

func compareFile(r *require.Assertions, golden, gen string) {
	goldenEntity, err := os.ReadFile(golden)
	r.NoError(err)
	genEntity, err := os.ReadFile(gen)
	r.NoError(err)
	r.Equal(string(goldenEntity), string(genEntity))
}
