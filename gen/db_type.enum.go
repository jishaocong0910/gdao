package gen

import o "github.com/jishaocong0910/go-object"

type DbType struct {
	*o.M_EnumValue
	DriveName string
}

var DbTypes = o.NewEnum[DbType](struct {
	*o.M_Enum[DbType]
	MYSQL, ORACLE, POSTGRES, SQLSERVER, SQLITE DbType
}{
	MYSQL:     DbType{DriveName: "mysql"},
	ORACLE:    DbType{DriveName: "oracle"},
	POSTGRES:  DbType{DriveName: "postgres"},
	SQLSERVER: DbType{DriveName: "mssql"},
	SQLITE:    DbType{DriveName: "sqlite3"},
})
