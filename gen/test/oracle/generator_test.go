package oracle_test

import (
	"context"
	"database/sql"
	"github.com/jishaocong0910/gdao/gen"
	go_ora "github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"testing"
	"time"
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

	gen.GetGenerator(gen.Cfg{
		DbType:  gen.DB_ORACLE,
		Dsn:     dsn,
		OutPath: "testdata",
		Package: "dao",
		TableCfg: gen.TableCfg{
			Tables: gen.Tables{"TEST_TABLE"},
		},
		DaoCfg: gen.DaoCfg{
			GenDao:            true,
			CoverBaseDao:      true,
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
