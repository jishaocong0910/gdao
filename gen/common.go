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
	"github.com/jishaocong0910/gdao"
)

// GenCfg 生成配置
type GenCfg struct {
	// 数据库类型
	DbType dbType
	// 数据库连接URL，空字符串时不会生成实体
	Dsn string
	// 生成文件相对路径，绝对路径为"os.Getwd()/OutPath"
	OutPath string
	// 包名，默认为目录名
	Package string
	// 表配置
	TableCfg TableCfg
	// DAO配置
	DaoCfg DaoCfg
}

type DaoCfg struct {
	// 是否生成DAO
	GenDao bool
	// 覆盖BaseDao
	CoverBaseDao bool
	// 是否生成CountDao
	GenCountDao bool
	// 覆盖CountDao
	CoverCountDao bool
	// 是否允许非法字段，如字段未导出、未使用指针等。若为false，实体中有非法字段将会在程序初始化时panic
	AllowInvalidField bool
}

type TableCfg struct {
	// 需要生成的表
	Tables Tables
	// 指定表字段映射实体字段类型，key为表名，value的key为表字段名，value为实体字段类型
	MappingTypes MappingTypes
	// 指定表忽略的字段，key为表名，value为列名
	IgnoreColumns IgnoreColumns
}

type baseDaoTplParam struct {
	Package string
}

type entityTplParam struct {
	Table             string
	EntityFileName    string
	Package           string
	EntityName        string
	Fields            []*fieldTplParam
	Comment           string
	GenDao            bool
	DaoFileName       string
	DaoName           string
	AllowInvalidField bool
}

type fieldTplParam struct {
	Column            string
	FieldName         string
	FieldType         string
	IsAutoIncrement   bool
	AutoIncrementStep int
	Comment           string
	Valid             bool
}

type Tables []string

type MappingTypes map[string]Mapper

type Mapper map[string]string

type IgnoreColumns map[string]Columns

type Columns []string

//go:embed entity.tpl
var entityTpl string

//go:embed dao.tpl
var daoTpl string

//go:embed count_dao.tpl
var countDaoTpl string

var entityNameMapper = gdao.NewNameMapper().UpperCamelCase()
var fieldNameMapper = gdao.NewNameMapper().UpperCamelCase()
var daoNameMapper = gdao.NewNameMapper().UpperCamelCase().AddSuffix("Dao")
var entityFileNameMapper = gdao.NewNameMapper().LowerSnakeCase().AddSuffix(".go")
var daoFileNameMapper = gdao.NewNameMapper().LowerSnakeCase().AddSuffix("_dao.go")

var supportedFieldTypes = map[string]struct{}{
	"bool": {}, "string": {}, "time.Time": {}, "float32": {}, "float64": {}, "int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
}

func mustNoError(err error) {
	if err != nil { // coverage-ignore
		panic(err)
	}
}

func mustReturn[T any](t T, err error) T {
	if err != nil { // coverage-ignore
		panic(err)
	}
	return t
}
