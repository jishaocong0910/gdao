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
	"bufio"
	"bytes"
	"database/sql"
	_ "embed"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/jishaocong0910/gdao"
	"github.com/jishaocong0910/gdao/internal"
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

func MappingBase[T gdao.BaseType]() Mapping {
	var t T
	return Mapping{t: t, mt: mappingType_.base}
}

func MappingSlice[T gdao.BaseType](dim int) Mapping {
	var t T
	if dim < 1 {
		dim = 1
	}
	return Mapping{t: t, sliceDim: dim, mt: mappingType_.slice}
}

func MappingConvert[T any]() Mapping {
	var t T
	return Mapping{t: t, mt: mappingType_.convert}
}

// Config 生成配置
type Config struct {
	// 数据库类型
	DbType dbType
	// 数据库连接URL，空字符串时不会生成实体
	Dsn string
	// 相对 [os.Getwd] 的go.mod文件路径
	GoModPath string
	// 相对go.mod文件的生成文件路径，默认为“dao”
	OutPath string
	// 表配置
	TableCfg TableCfg
	// DAO配置
	DaoCfg DaoCfg
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
	// 指定表字段映射实体字段类型，使用函数 [MappingBase]、[MappingSlice] 或 [MappingConvertor] 指定
	Mappers Mappers
	// 指定表忽略的字段，key为表名，value为列名
	Ignores Ignores
}

type Mapping struct {
	t        any
	sliceDim int
	mt       mappingType
}

type Tables []string

type Mappers map[string]Mappings

type Mappings map[string]Mapping

type Ignores map[string]Columns

type Columns []string

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
	if g.cfg.OutPath == "" {
		g.cfg.OutPath = "dao"
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(wd, g.cfg.GoModPath, "go.mod"))
	if err != nil {
		return err
	}
	r := bufio.NewReader(file)
	var moduleName string
	for {
		bs, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line := string(bs)
		spaceIdx := strings.Index(line, " ")
		if spaceIdx == -1 {
			continue
		}
		if line[:spaceIdx] == "module" {
			moduleName = strings.TrimSpace(line[spaceIdx+1:])
			break
		}
	}
	if moduleName == "" {
		return errors.New("module name is empty")
	}
	g.entityPkgPath = moduleName + "/" + g.cfg.OutPath + "/entity"
	goModPath, _ := filepath.Split(file.Name())
	g.dir = filepath.Join(goModPath, g.cfg.OutPath)
	g.entityDir = filepath.Join(g.dir, "entity")
	err = os.MkdirAll(g.dir, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(g.entityDir, os.ModePerm)
	if err != nil {
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
			// 自定义映射
			impts, err := g.mappingFields(table, fields)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			// 创建实体模板参数
			entityName := entityNameMapper.Convert(table)
			e := entityTplParam{
				Table:      table,
				EntityName: entityName,
				Fields:     fields,
				Comment:    comment,
				Imports:    impts,
				dao: daoTplParam{
					Table:         table,
					PkgName:       pkgName,
					DaoName:       daoNameMapper.Convert(table),
					EntityName:    entityName,
					EntityPkgPath: g.entityPkgPath,
				},
			}
			g.entityTplParams = append(g.entityTplParams, e)
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

func (g *Generator) mappingFields(table string, fields []fieldTplParam) ([]string, error) {
	mappings := g.cfg.TableCfg.Mappers[table]
	if mappings == nil {
		return nil, nil
	}

	for i := 0; i < len(fields); i++ {
		f := &fields[i]
		if m, ok := mappings[f.Column]; ok {
			switch m.mt.String() {
			case mappingType_.base.String():
				fieldType := g.determineFieldType(table, reflect.TypeOf(m.t))
				f.FieldType = "*" + fieldType
			case mappingType_.slice.String():
				var fieldType string
				for i := 0; i < m.sliceDim; i++ {
					fieldType += "[]"
				}
				fieldType += g.determineFieldType(table, reflect.TypeOf(m.t))
				f.FieldType = fieldType
			case mappingType_.convert.String():
				ft := reflect.TypeOf(m.t)
				validConvertType := false
				switch ft.Kind() {
				case reflect.Pointer:
					if ft.Elem().Kind() == reflect.Struct {
						validConvertType = true
					}
				case reflect.Struct:
					ft = reflect.New(ft).Type()
					validConvertType = true
				case reflect.Slice, reflect.Map:
					validConvertType = true
				}
				if validConvertType && internal.IsImplementConvert(ft) == 1 {
					f.FieldType = g.determineFieldType(table, ft)
				} else { // coverage-ignore
					return nil, errors.New("the mapping of table \"" + table + "\"'s column \"" + f.Column + "\" is invalid implementing gdao.Convert")
				}
			}
		}
	}
	var impts []string
	for name, path := range g.getPkgNameToPaths(table) {
		if name != path[strings.LastIndex(path, "/")+1:] {
			impts = append(impts, name+" \""+path+"\"")
		} else {
			impts = append(impts, "\""+path+"\"")
		}
	}
	return impts, nil
}

func (g *Generator) getPkgNameToPaths(table string) map[string]string {
	m := g.pkgNameToPaths[table]
	if m == nil {
		m = make(map[string]string)
		g.pkgNameToPaths[table] = m
	}
	return m
}

func (g *Generator) determineFieldType(table string, ft reflect.Type) string {
	pkgNameToPaths := g.getPkgNameToPaths(table)

	var pkgPath string
	if ft.Kind() == reflect.Pointer {
		pkgPath = ft.Elem().PkgPath()
	} else {
		pkgPath = ft.PkgPath()
	}
	arr := strings.SplitN(ft.String(), ".", 2)
	// 基础类型所以没有包名
	if len(arr) == 1 {
		return ft.String()
	}
	pkgName := arr[0]
	if pkgName[:1] == "*" {
		pkgName = pkgName[1:]
	}
	typeName := arr[1]
	pkgName = g.determinePkgName(pkgPath, pkgName, pkgNameToPaths, false)

	pkgNameToPaths[pkgName] = pkgPath
	return pkgName + "." + typeName
}

func (g *Generator) determinePkgName(pkgPath, pkgName string, pkgNameToPaths map[string]string, conflict bool) string {
	if conflict {
		arr := pkgNameRegex.FindSubmatch([]byte(pkgName))
		pkgName = string(arr[1])
		num := string(arr[2])
		if num != "" {
			i, _ := strconv.ParseInt(num, 10, 32)
			pkgName += strconv.Itoa(int(i + 1))
		} else {
			pkgName += "2"
		}
	}
	if path, ok := pkgNameToPaths[pkgName]; ok {
		if path == pkgPath {
			return pkgName
		} else {
			return g.determinePkgName(pkgPath, pkgName, pkgNameToPaths, true)
		}
	} else {
		return pkgName
	}
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
	Table         string
	PkgName       string
	DaoName       string
	EntityName    string
	EntityPkgPath string
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
