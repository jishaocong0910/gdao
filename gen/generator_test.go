package gen_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jishaocong0910/gdao/gen"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMySql(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	mysqlContainer, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0.36"),
		mysql.WithDatabase("test"),
		mysql.WithUsername("root"),
		mysql.WithPassword("12345678"),
		mysql.WithScripts(filepath.Join("testdata", "mysql.sql")),
	)
	r.NoError(err)

	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	dsn, err := mysqlContainer.ConnectionString(ctx)
	r.NoError(err)

	gen.GetGenerator(gen.Conf{
		DbType:  gen.DB_MYSQL,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables: gen.Tables{"mysql": gen.FieldTypes{
			"other":  "[]string",
			"other2": "[]rune",
			"other3": "*string",
			"other4": "string",
			"other5": "rune",
		}},
		GenDao: true,
	}).Gen()

	defer os.Remove("testdata/base_dao.go")
	compareFile(r, "testdata/mysql_entity.golden", "testdata/mysql.go")
	compareFile(r, "testdata/mysql_dao.golden", "testdata/mysql_dao.go")
}

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
		WaitingFor: wait.ForLog("DATABASE IS READY TO USE!").WithStartupTimeout(time.Minute * 3),
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

	script, err := os.ReadFile("testdata/oracle.sql")
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

	gen.GetGenerator(gen.Conf{
		DbType:  gen.DB_ORACLE,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables:  gen.Tables{"ORACLE": nil},
		GenDao:  true,
	}).Gen()

	defer os.Remove("testdata/base_dao.go")
	compareFile(r, "testdata/oracle_entity.golden", "testdata/oracle.go")
	compareFile(r, "testdata/oracle_dao.golden", "testdata/oracle_dao.go")
}

func TestPostgres(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("12345678"),
		postgres.WithInitScripts(filepath.Join("testdata", "postgres.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	r.NoError(err)

	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	r.NoError(err)

	gen.GetGenerator(gen.Conf{
		DbType:  gen.DB_POSTGRES,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables:  gen.Tables{"postgres": nil},
		GenDao:  true,
	}).Gen()

	defer os.Remove("testdata/base_dao.go")
	compareFile(r, "testdata/postgres_entity.golden", "testdata/postgres.go")
	compareFile(r, "testdata/postgres_dao.golden", "testdata/postgres_dao.go")
}

// 由于容器镜像只支持Intel芯片，此用例只能在Intel芯片执行
func TestSqlServer(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()

	mssqlContainer, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
		mssql.WithPassword("SuperStrong@PassWord"),
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

	script, err := os.ReadFile("testdata/sqlserver.sql")
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

	gen.GetGenerator(gen.Conf{
		DbType:  gen.DB_SQLSERVER,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables:  gen.Tables{"sqlserver": nil},
		GenDao:  true,
	}).Gen()

	defer os.Remove("testdata/base_dao.go")
	compareFile(r, "testdata/sqlserver_entity.golden", "testdata/sqlserver.go")
	compareFile(r, "testdata/sqlserver_dao.golden", "testdata/sqlserver_dao.go")
}

func TestSqlite(t *testing.T) {
	r := require.New(t)
	gen.GetGenerator(gen.Conf{
		DbType:  gen.DB_SQLITE,
		Dsn:     "testdata/sqlite.db",
		OutPath: "testdata",
		Tables:  gen.Tables{"sqlite": nil},
		GenDao:  true,
	}).Gen()

	defer os.Remove("testdata/base_dao.go")
	compareFile(r, "testdata/sqlite_entity.golden", "testdata/sqlite.go")
	compareFile(r, "testdata/sqlite_dao.golden", "testdata/sqlite_dao.go")
}

func compareFile(r *require.Assertions, golden, gen string) {
	defer os.Remove(gen)
	goldenEntity, err := os.ReadFile(golden)
	r.NoError(err)
	genEntity, err := os.ReadFile(gen)
	r.NoError(err)
	r.Equal(string(goldenEntity), string(genEntity))
}

func TestUnsupportedDb(t *testing.T) {
	r := require.New(t)
	r.PanicsWithValue("not support this db type yet", func() {
		gen.GetGenerator(gen.Conf{DbType: gen.DbType(999)})
	})
}
