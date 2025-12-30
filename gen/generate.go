/*
 * Copyright 2024-present jishaocong0910
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gen

import (
	"bytes"
	"database/sql"
	_ "embed"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"

	"github.com/jishaocong0910/gdao"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/imports"
)

// GetGenerator 创建生成器
func GetGenerator(cfg Config) *Generator {
	var dbInfo dbInfo
	switch cfg.DbType.String() {
	case DbType_.MYSQL.String():
		dbInfo = newMySqlInfo(cfg)
	case DbType_.ORACLE.String():
		dbInfo = newOracleDbInfo(cfg)
	case DbType_.POSTGRES.String():
		dbInfo = newPostgresGenerator(cfg)
	case DbType_.SQLSERVER.String():
		dbInfo = newSqlServerInfo(cfg)
	case DbType_.SQLITE.String():
		dbInfo = newSqliteGenerator(cfg)
	default: // coverage-ignore
		panic("not support this db type yet")
	}
	return newGenerator(cfg, dbInfo)
}

// Config 生成配置
type Config struct {
	// 数据库类型
	DbType dbType
	// 数据库连接URL，空字符串时不会生成实体
	Dsn string
	// 相对生成器运行目录的go.mod文件路径
	GoModPath string
	// 相对go.mod文件的生成文件路径，默认为“dao”
	OutPath string
	// 表配置
	TableCfg TableCfg
	// DAO配置
	DaoCfg DaoCfg
	// 逻辑删除配置
	LogicalDelCfg LogicalDelCfg
}

func (c *Config) getDB() *sql.DB {
	var db *sql.DB
	if c.Dsn != "" {
		db = mustReturn(sql.Open(c.DbType.driverName, c.Dsn))
	}
	return db
}

type DaoCfg struct {
	// 不覆盖BaseDao
	NotCoverBaseDao bool
}

type TableCfg struct {
	// 需要生成的表
	Tables Tables
	// 指定表忽略的字段，key为表名，value为列名
	Ignores Ignores
}

type Tables []string

type Ignores map[string]Columns

type Columns []string

type LogicalDelCfg struct {
	Mode       logicalDelMode
	FlagColumn string
	IdColumn   string
	QueryValue any
}

type Generator struct {
	cfg             Config
	dbInfo          dbInfo
	dir             string
	entityDir       string
	entityPkgPath   string
	entityTpl       *template.Template
	daoTpl          *template.Template
	countDaoTpl     *template.Template
	entityTplParams []entityTplParam
	baseDaoTplParam baseDaoTplParam
	pkgNameToPaths  map[string]map[string]string
}

func (g *Generator) Gen() {
	err := g.checkDir()
	if err != nil { // coverage-ignore
		log.Printf("%v", err)
		return
	}
	log.Println("start generating...")
	log.Printf("full output path: %s", g.dir)
	g.queryTplParams()
	g.genBaseDao()
	g.genEntity()
	log.Println("finish generating")
}

func (g *Generator) checkDir() error {
	if g.cfg.OutPath == "" { // coverage-ignore
		g.cfg.OutPath = "dao"
	}
	file, err := os.Open(filepath.Join(g.cfg.GoModPath, "go.mod"))
	if err != nil { // coverage-ignore
		return err
	}
	defer file.Close()
	bs, err := io.ReadAll(file)
	if err != nil { // coverage-ignore
		return err
	}
	modFile, err := modfile.Parse("", bs, nil)
	if err != nil { // coverage-ignore
		return err
	}
	g.entityPkgPath = modFile.Module.Mod.Path + "/" + g.cfg.OutPath + "/entity"
	goModPath, _ := filepath.Split(file.Name())
	g.dir = filepath.Join(goModPath, g.cfg.OutPath)
	g.entityDir = filepath.Join(g.dir, "entity")
	err = os.MkdirAll(g.dir, os.ModePerm)
	if err != nil { // coverage-ignore
		return err
	}
	err = os.MkdirAll(g.entityDir, os.ModePerm)
	if err != nil { // coverage-ignore
		return err
	}
	return nil
}

func (g *Generator) queryTplParams() {
	_, pkgName := filepath.Split(g.cfg.OutPath)
	if g.cfg.Dsn != "" {
		for _, table := range g.cfg.TableCfg.Tables {
			// 获取表信息
			fields, comment, err := g.dbInfo.getTableInfo(table)
			if err != nil { // coverage-ignore
				log.Println(err.Error())
				continue
			}
			// 过滤字段
			fields = g.ignoreFields(table, fields)
			// 创建逻辑删参数
			ldtp := g.buildLogicalDelTplParam(fields)
			// 创建DAO模板参数
			dtp := g.buildDaoTplParam(table, pkgName, ldtp)
			// 创建实体模板参数
			etp := g.buildEntityTplParam(table, fields, comment, dtp)
			// 收集
			g.entityTplParams = append(g.entityTplParams, etp)
		}
	}
	g.baseDaoTplParam = baseDaoTplParam{
		PkgName: pkgName,
	}
}

func (g *Generator) ignoreFields(table string, fields []fieldTplParam) []fieldTplParam {
	ignoreColumns := g.cfg.TableCfg.Ignores[table]
	if ignoreColumns != nil {
		var temp []fieldTplParam
		for _, f := range fields {
			isIgnored := false
			for _, column := range ignoreColumns {
				if column == f.Column {
					isIgnored = true
				}
			}
			if !isIgnored {
				temp = append(temp, f)
			}
		}
		fields = temp
	}
	return fields
}

func (g *Generator) buildDaoTplParam(table string, pkgName string, ldtp logicalDelTplParam) daoTplParam {
	return daoTplParam{
		Table:              table,
		PkgName:            pkgName,
		DaoName:            daoNameMapper.Convert(table),
		EntityName:         entityNameMapper.Convert(table),
		EntityPkgPath:      g.entityPkgPath,
		LogicalDelTplParam: ldtp,
	}
}

func (g *Generator) buildEntityTplParam(table string, fields []fieldTplParam, comment string, dtp daoTplParam) entityTplParam {
	return entityTplParam{
		Table:      table,
		EntityName: entityNameMapper.Convert(table),
		Fields:     fields,
		Comment:    comment,
		dao:        dtp,
	}
}

func (g *Generator) buildLogicalDelTplParam(fields []fieldTplParam) (param logicalDelTplParam) {
	if g.cfg.LogicalDelCfg.Mode.IsUndefined() {
		return
	}

	hasFlagColumn := false
	for _, field := range fields {
		if field.Column == g.cfg.LogicalDelCfg.FlagColumn {
			hasFlagColumn = true
			break
		}
	}
	if !hasFlagColumn { // coverage-ignore
		return
	}

	if LogicalDelMode_.SET_ID.Is(g.cfg.LogicalDelCfg.Mode) {
		hasIdColumn := false
		for _, field := range fields {
			if field.Column == g.cfg.LogicalDelCfg.FlagColumn {
				hasIdColumn = true
				break
			}
		}
		if !hasIdColumn { // coverage-ignore
			return
		}
	}

	if g.cfg.LogicalDelCfg.QueryValue == nil { // coverage-ignore
		return
	}

	queryValue := ""
	switch g.cfg.LogicalDelCfg.QueryValue.(type) {
	case int:
		queryValue = strconv.Itoa(g.cfg.LogicalDelCfg.QueryValue.(int))
	case string:
		queryValue = "\"" + g.cfg.LogicalDelCfg.QueryValue.(string) + "\""
	default: // coverage-ignore
		return
	}

	param.Mode = g.cfg.LogicalDelCfg.Mode.code
	param.FlagColumn = g.cfg.LogicalDelCfg.FlagColumn
	param.IdColumn = g.cfg.LogicalDelCfg.IdColumn
	param.QueryValue = queryValue
	return param
}

func (g *Generator) genBaseDao() {
	baseDaoTpl := mustReturn(template.New("").Parse(g.dbInfo.getBaseDaoTemplate()))
	generated, err := g.createFile(g.dir, "base_dao.go", !g.cfg.DaoCfg.NotCoverBaseDao, baseDaoTpl, g.baseDaoTplParam)
	if err != nil { // coverage-ignore
		log.Printf("create base dao fail: %+v\n", err)
	} else if generated {
		log.Println("create base dao success")
	}
}

func (g *Generator) genEntity() {
	for _, e := range g.entityTplParams {
		generated, err := g.createFile(g.entityDir, entityFileNameMapper.Convert(e.Table), true, g.entityTpl, e)
		if err != nil { // coverage-ignore
			log.Printf("create entity of table \"%s\" fail, error: %+v\n", e.Table, err)
		} else if generated {
			log.Printf("create entity of table \"%s\" success\n", e.Table)
		}
		generated, err = g.createFile(g.dir, daoFileNameMapper.Convert(e.Table), false, g.daoTpl, e.dao)
		if err != nil { // coverage-ignore
			log.Printf("create dao of table \"%s\" fail, error: %+v\n", e.Table, err)
		} else if generated {
			log.Printf("create dao of table \"%s\" success\n", e.Table)
		}
	}
}

func (g *Generator) createFile(outPath, fileName string, cover bool, tpl *template.Template, param any) (bool, error) {
	path := filepath.Join(outPath, fileName)
	if !cover {
		_, err := os.Stat(path)
		if err == nil { // coverage-ignore
			return false, nil
		}
	}
	var buf bytes.Buffer
	err := tpl.Execute(&buf, param)
	if err != nil { // coverage-ignore
		return false, err
	}
	content, importErr := imports.Process("", buf.Bytes(), nil)
	if importErr != nil { // coverage-ignore
		content = buf.Bytes()
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil { // coverage-ignore
		return false, err
	}
	return true, importErr
}

func newGenerator(cfg Config, dbInfo dbInfo) *Generator {
	return &Generator{
		cfg:            cfg,
		dbInfo:         dbInfo,
		entityTpl:      mustReturn(template.New("").Parse(entityTpl)),
		daoTpl:         mustReturn(template.New("").Parse(daoTpl)),
		countDaoTpl:    mustReturn(template.New("").Parse(countDaoTpl)),
		pkgNameToPaths: make(map[string]map[string]string),
	}
}

type dbInfo interface {
	getTableInfo(table string) ([]fieldTplParam, string, error)
	getBaseDaoTemplate() string
}

type baseDaoTplParam struct {
	PkgName string
}

type entityTplParam struct {
	Table      string
	EntityName string
	Fields     []fieldTplParam
	Comment    string
	Imports    []string

	dao daoTplParam
}

type daoTplParam struct {
	Table              string
	PkgName            string
	DaoName            string
	EntityName         string
	EntityPkgPath      string
	LogicalDelTplParam logicalDelTplParam
}

type fieldTplParam struct {
	Column            string
	FieldName         string
	FieldType         string
	IsAutoIncrement   bool
	IsNotNull         bool
	HasDefaultValue   bool
	AutoIncrementStep int
	Comment           string
	Valid             bool
}

type logicalDelTplParam struct {
	Mode       int
	FlagColumn string
	IdColumn   string
	QueryValue string
}

//go:embed entity.tpl
var entityTpl string

//go:embed dao.tpl
var daoTpl string

//go:embed count_dao.tpl
var countDaoTpl string

var entityNameMapper = gdao.NewNameMapper().UpperCamelCase()
var fieldNameMapper = gdao.NewNameMapper().UpperCamelCase()
var daoNameMapper = gdao.NewNameMapper().UpperCamelCase()
var entityFileNameMapper = gdao.NewNameMapper().LowerSnakeCase().AddSuffix(".go")
var daoFileNameMapper = gdao.NewNameMapper().LowerSnakeCase().AddSuffix(".go")

var pkgNameRegex = regexp.MustCompile(`^([a-zA-Z_](?:\w*[a-zA-Z_])*)(\d*)$`)

func must(err error) {
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
