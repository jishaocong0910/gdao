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

	gen.GetGenerator(gen.Cfg{
		DbType:       gen.DB_MYSQL,
		Dsn:          dsn,
		OutPath:      "testdata",
		Package:      "dao",
		CoverBaseDao: true,
		Tables: gen.Tables{"mysql": gen.FieldTypes{
			"other":  "[]string",
			"other2": "[]rune",
			"other3": "*string",
			"other4": "string",
			"other5": "rune",
		}},
		GenDao: true,
	}).Gen()

	defer os.Remove("testdata/mysql.go")
	defer os.Remove("testdata/mysql_dao.go")
	defer os.Remove("testdata/base_dao.go")
	defer os.Remove("testdata/count_dao.go")

	compareFile(r, "testdata/entity.golden", "testdata/mysql.go")
	compareFile(r, "testdata/dao.golden", "testdata/mysql_dao.go")
	compareFile(r, "internal/mysql_base_dao.go", "testdata/base_dao.go")
	compareFile(r, "testdata/count_dao.golden", "testdata/count_dao.go")
}

func compareFile(r *require.Assertions, golden, gen string) {
	goldenEntity, err := os.ReadFile(golden)
	r.NoError(err)
	genEntity, err := os.ReadFile(gen)
	r.NoError(err)
	r.Equal(string(goldenEntity), string(genEntity))
}
