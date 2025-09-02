package sqlserver

import (
	"context"
	"database/sql"
	"github.com/jishaocong0910/gdao/gen"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// 由于容器镜像只支持Intel芯片，此用例只能在Intel芯片执行
func TestSqlServer(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	mssqlContainer, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
		mssql.WithPassword("SuperStrong@PassWord"),
		testcontainers.WithWaitStrategyAndDeadline(time.Minute*5, wait.ForLog("Recovery is complete.").WithStartupTimeout(time.Minute*5)),
	)
	r.NoError(err)
	defer func() {
		if err := testcontainers.TerminateContainer(mssqlContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	dsn, err := mssqlContainer.ConnectionString(ctx)
	r.NoError(err)
	db, err := sql.Open("mssql", dsn)
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
		DbType:       gen.DB_SQLSERVER,
		Dsn:          dsn,
		OutPath:      "testdata",
		Package:      "dao",
		CoverBaseDao: true,
		Tables:       gen.Tables{"test_table"},
		GenDao:       true,
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
