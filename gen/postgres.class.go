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

package gen

import (
	_ "embed"
	"errors"
	"github.com/jishaocong0910/gdao"
	_ "github.com/lib/pq"
	"strings"
)

//go:embed postgres_base_dao.tpl
var postgresBaseDaoTpl string

type postgresGenerator struct {
	*generator__
	schema   string
	database string
}

func (g postgresGenerator) getDriverName() string {
	return "postgres"
}

func (g postgresGenerator) getBaseDaoTemplate() string {
	return postgresBaseDaoTpl
}

func (g postgresGenerator) getTableInfo(table string) ([]fieldTplParam, string, error) {
	var (
		exists       bool
		fields       []fieldTplParam
		tableComment string
	)

	rows := mustReturn(g.db.Query("SELECT i.column_name,i.udt_name,i.is_identity,i.column_default,p.attndims,p.description FROM information_schema.columns i JOIN (SELECT n.nspname,c.relname,a.attname,a.attndims,d.description FROM pg_namespace n JOIN pg_class c ON n.oid = c.relnamespace JOIN pg_attribute a ON a.attrelid = c.oid LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = a.attnum WHERE a.attnum > 0 AND NOT a.attisdropped) p ON p.nspname=i.table_schema and p.relname=i.table_name and p.attname=i.column_name WHERE i.table_catalog = $1 AND i.table_schema = $2 AND i.table_name = $3 ORDER BY i.ordinal_position", g.database, g.schema, table))
	defer rows.Close()
	for rows.Next() {
		exists = true
		var (
			fieldType string

			column          string
			udtName         string
			isIdentity      string
			columnDefault   *string
			attndims        int
			description     *string
			isAutoIncrement bool
		)
		must(rows.Scan(&column, &udtName, &isIdentity, &columnDefault, &attndims, &description))
		if description == nil {
			description = gdao.P("")
		}

		if strings.EqualFold(isIdentity, "YES") || (columnDefault != nil && strings.HasPrefix(*columnDefault, "nextval(")) {
			isAutoIncrement = true
		}

		if udtName[:1] == "_" {
			udtName = udtName[1:]
		}
		switch udtName {
		case "int8":
			fieldType = "*int64"
		case "int4":
			fieldType = "*int32"
		case "int2":
			fieldType = "*int16"
		case "time", "timetz", "date", "timestamp", "timestamptz":
			fieldType = "*time.Time"
		case "float8", "numeric", "money":
			fieldType = "*float64"
		case "float4":
			fieldType = "*float32"
		case "bool":
			fieldType = "*bool"
		case "bytea":
			fieldType = "[]byte"
		case "bit", "varbit", "box", "bpchar", "varchar", "cidr", "circle", "inet", "interval", "json", "jsonb", "line",
			"lseg", "macaddr", "macaddr8", "path", "pg_lsn", "pg_snapshot", "point", "polygon", "text",
			"tsquery", "tsvector", "txid_snapshot", "uuid", "xml":
			fieldType = "*string"
		}
		for i := 0; i < attndims; i++ {
			if i == 0 {
				fieldType = strings.TrimPrefix(fieldType, "*")
			}
			fieldType = "[]" + fieldType
		}

		f := fieldTplParam{
			Column:          column,
			FieldName:       fieldNameMapper.Convert(column),
			FieldType:       fieldType,
			Comment:         *description,
			Valid:           fieldType != "any",
			IsAutoIncrement: isAutoIncrement,
		}
		fields = append(fields, f)
	}

	rows = mustReturn(g.db.Query("SELECT d.description FROM information_schema.tables i JOIN pg_class c ON c.relname = i.table_name LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = '0' WHERE i.table_catalog = $1 AND i.table_schema = $2 AND i.table_name = $3", g.database, g.schema, table))
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tableComment)
	}
	if !exists {
		return nil, "", errors.New("\"" + table + "\" is not exists")
	}
	return fields, tableComment, nil
}

func newPostgresGenerator(c GenCfg) *postgresGenerator {
	this := &postgresGenerator{}
	this.generator__ = extendGenerator_(this, c)

	if this.db != nil {
		schema := ""
		rows := mustReturn(this.db.Query("SELECT CURRENT_SCHEMA()"))
		if rows.Next() {
			rows.Scan(&schema)
		}
		rows.Close()

		database := ""
		rows = mustReturn(this.db.Query("SELECT CURRENT_DATABASE()"))
		if rows.Next() {
			rows.Scan(&database)
		}
		rows.Close()

		this.schema = schema
		this.database = database
	}
	return this
}
