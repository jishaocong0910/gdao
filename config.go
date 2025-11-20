package gdao

import "database/sql"

type Cfg struct {
	DefaultDB      *sql.DB
	Logger         Logger
	SqlLogLevel    SqlLogLevel
	CompressSqlLog bool
}

var global Cfg

func Config(cfg Cfg) {
	global = cfg
}
