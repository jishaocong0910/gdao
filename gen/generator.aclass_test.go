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
	"github.com/sijms/go-ora/v2"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
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

	paths := gen.Generator(gen.Config{
		DbType:  gen.DbTypes.MYSQL,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables: gen.Tables{"mysql": gen.FieldTypes{
			"other":  "[]string",
			"other2": "[]rune",
			"other3": "*string",
			"other4": "string",
			"other5": "rune",
		}},
	}).Gen()
	r.Len(paths, 1)
	genFile := paths[0]
	defer os.Remove(genFile)

	golden, err := os.ReadFile("testdata/mysql.golden")
	r.NoError(err)
	gen, err := os.ReadFile(genFile)
	r.NoError(err)
	r.Equal(string(golden), string(gen))
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

	paths := gen.Generator(gen.Config{
		DbType:  gen.DbTypes.POSTGRES,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables:  gen.Tables{"postgres": nil},
	}).Gen()
	r.Len(paths, 1)
	genFile := paths[0]
	defer os.Remove(genFile)

	golden, err := os.ReadFile("testdata/postgres.golden")
	r.NoError(err)
	gen, err := os.ReadFile(genFile)
	r.NoError(err)
	r.Equal(string(golden), string(gen))
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

	paths := gen.Generator(gen.Config{
		DbType:  gen.DbTypes.ORACLE,
		Dsn:     dsn,
		OutPath: "testdata",
		Tables:  gen.Tables{"ORACLE": nil},
	}).Gen()
	r.Len(paths, 1)
	genFile := paths[0]
	defer os.Remove(genFile)

	golden, err := os.ReadFile("testdata/oracle.golden")
	r.NoError(err)
	gen, err := os.ReadFile(genFile)
	r.NoError(err)
	r.Equal(string(golden), string(gen))
}

//func TestSqlServer(t *testing.T) {
//	r := require.New(t)
//	ctx := context.Background()
//
//	mssqlContainer, err := mssql.RunContainer(ctx,
//		testcontainers.WithImage("mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04"),
//		mssql.WithAcceptEULA(),
//		mssql.WithPassword("SuperStrong@Passw0rd"),
//	)
//	r.NoError(err)
//
//	defer func() {
//		if err := mssqlContainer.Terminate(ctx); err != nil {
//			log.Fatalf("failed to terminate container: %s", err)
//		}
//	}()
//
//	dsn, err := mssqlContainer.ConnectionString(ctx)
//	r.NoError(err)
//
//	path := gen.Generator(gen.Config{
//		DbType:  dbTypes.SQLSERVER,
//		Dsn:     dsn,
//		OutPath: "testdata",
//		Tables:  gen.Tables{"sqlserver": nil},
//	}).Gen()
//	r.Len(path, 1)
//	genFile := path[0]
//	defer os.Remove(genFile)
//
//	golden, err := os.ReadFile("testdata/sqlserver.golden")
//	r.NoError(err)
//	gen, err := os.ReadFile(genFile)
//	r.NoError(err)
//	r.Equal(string(golden), string(gen))
//}

//func TestSqlite(t *testing.T) {
//	r := require.New(t)
//	paths := gen.Generator(gen.Config{
//		DbType:  gen.DbTypes.SQLITE,
//		Dsn:     "testdata/sqlite.db",
//		OutPath: "testdata",
//		Tables:  gen.Tables{"sqlite": nil},
//	}).Gen()
//	r.Len(paths, 1)
//	genFile := paths[0]
//	defer os.Remove(genFile)
//
//}

func TestUnsupportedDb(t *testing.T) {
	r := require.New(t)
	r.PanicsWithValue("not support this db type yet", func() {
		gen.Generator(gen.Config{
			DbType: gen.DbType{},
		})
	})
}
