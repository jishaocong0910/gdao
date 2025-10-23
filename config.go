package gdao

import (
	"database/sql"
)

var cfg Cfg

type Cfg struct {
	DefaultDB      *sql.DB
	Logger         Logger
	SqlLogLevel    SqlLogLevel
	CompressSqlLog bool
}

func Config(c Cfg) {
	cfg = c
}
