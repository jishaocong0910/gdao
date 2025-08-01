package sqlite

import (
	"github.com/jishaocong0910/gdao/gen"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSqlite(t *testing.T) {
	r := require.New(t)
	gen.GetGenerator(gen.Cfg{
		DbType:  gen.DB_SQLITE,
		Dsn:     "testdata/sqlite.db",
		OutPath: "testdata",
		Package: "dao",
		Tables:  gen.Tables{"sqlite": nil},
		GenDao:  true,
	}).Gen()

	defer os.Remove("testdata/count_dao.go")
	defer os.Remove("testdata/sqlite.go")
	defer os.Remove("testdata/sqlite_dao.go")
	defer os.Remove("testdata/base_dao.go")

	compareFile(r, "testdata/entity.golden", "testdata/sqlite.go")
	compareFile(r, "testdata/dao.golden", "testdata/sqlite_dao.go")
	compareFile(r, "internal/base_dao.go", "testdata/base_dao.go")
}

func compareFile(r *require.Assertions, golden, gen string) {
	goldenEntity, err := os.ReadFile(golden)
	r.NoError(err)
	genEntity, err := os.ReadFile(gen)
	r.NoError(err)
	r.Equal(string(goldenEntity), string(genEntity))
}
