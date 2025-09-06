package gen

import (
	"bytes"
	_ "embed"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jishaocong0910/gdao"
	"golang.org/x/tools/imports"
)

type dbType int

const (
	DB_MYSQL dbType = iota + 1
	DB_ORACLE
	DB_POSTGRES
	DB_SQLSERVER
	DB_SQLITE
)

type dialectGenerator interface {
	getTableInfo(table string) (bool, []*field, string)
	getBaseDaoTemplate() *template.Template
}

type Generator struct {
	c Cfg
	d dialectGenerator
}

func (g Generator) Gen() {
	log.Println("start generating...")
	log.Printf("full output path: %s", g.c.OutPath)
	var entities []entity
	var tableNamedCount string
	for _, table := range g.c.TableCfg.Tables {
		// 表名为count时会和生成CountDao冲突
		if strings.EqualFold(table, "count") {
			tableNamedCount = table
		}
		// 获取表信息
		exists, fields, comment := g.d.getTableInfo(table)
		if !exists {
			log.Printf("table \"%s\" is not exists", table)
			continue
		}
		// 过滤字段
		if columns, ok := g.c.TableCfg.IgnoreColumns[table]; ok {
			var filtered []*field
			for _, f := range fields {
				isIgnored := false
				for _, column := range columns {
					if column == f.Column {
						isIgnored = true
					}
				}
				if !isIgnored {
					filtered = append(filtered, f)
				}
			}
			fields = filtered
		}
		// 强制指定表字段映射类型
		if mappingTypes, ok := g.c.TableCfg.MappingTypes[table]; ok {
			for column, fieldType := range mappingTypes {
				for _, f := range fields {
					if f.Column == column {
						f.FieldType = fieldType
						if strings.HasPrefix(f.FieldType, "[]") {
							if _, ok := supportedFieldTypes[strings.TrimLeft(f.FieldType, "[]")]; !ok {
								f.Valid = false
							}
						} else {
							if _, ok := supportedFieldTypes[f.FieldType[1:]]; !ok {
								f.Valid = false
							}
						}
					}
				}
			}
		}
		// 创建实体模板参数
		e := entity{
			DbType:            g.c.DbType,
			Table:             table,
			EntityFileName:    entityFileNameMapper.Convert(table),
			Package:           g.c.Package,
			EntityName:        entityNameMapper.Convert(table),
			Fields:            fields,
			Comment:           comment,
			GenDao:            g.c.DaoCfg.GenDao,
			DaoFileName:       daoFileNameMapper.Convert(table),
			DaoName:           daoNameMapper.Convert(table),
			AllowInvalidField: g.c.DaoCfg.AllowInvalidField,
		}
		entities = append(entities, e)
	}

	if len(entities) == 0 {
		return
	}

	g.createOutPath()
	if g.c.DaoCfg.GenDao {
		b := baseDao{
			DbType:  g.c.DbType,
			Package: g.c.Package,
		}
		err := g.createFile("base_dao.go", g.c.DaoCfg.CoverBaseDao, g.d.getBaseDaoTemplate(), b)
		if err != nil {
			log.Printf("create base dao fail: %+v\n", err)
		} else {
			log.Println("create base dao success")
		}
		if tableNamedCount != "" {
			log.Printf("create count dao fail because exists table named \"%s\"", tableNamedCount)
		} else {
			err = g.createFile("count_dao.go", false, tplCountDao, b)
			if err != nil {
				log.Printf("create count dao fail: %+v\n", err)
			} else {
				log.Println("create count dao success")
			}
		}
	}
	for _, e := range entities {
		err := g.createFile(e.EntityFileName, true, tplEntity, e)
		if err != nil {
			log.Printf("create entity of table \"%s\" is fail, error: %+v\n", e.Table, err)
		} else {
			log.Printf("create entity of table \"%s\" is success\n", e.Table)
		}
		if g.c.DaoCfg.GenDao {
			err := g.createFile(e.DaoFileName, false, tplDao, e)
			if err != nil {
				log.Printf("create dao of table \"%s\" is fail, error: %+v\n", e.Table, err)
			} else {
				log.Printf("create dao of table \"%s\" is success\n", e.Table)
			}
		}
	}
	log.Println("finish generating")
}

func (g Generator) createOutPath() {
	mustNoError(os.MkdirAll(g.c.OutPath, os.ModePerm))
}

func (g Generator) createFile(fileName string, cover bool, tpl *template.Template, param any) error {
	path := filepath.Join(g.c.OutPath, fileName)
	if !cover {
		_, err := os.Stat(path)
		if err == nil {
			return errors.New("file already exists")
		}
	}
	var buf bytes.Buffer
	err := tpl.Execute(&buf, param)
	if err != nil {
		return err
	}
	content, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil {
		return err
	}
	return nil
}

type baseDao struct {
	DbType  dbType
	Package string
}

type entity struct {
	DbType            dbType
	Table             string
	EntityFileName    string
	Package           string
	EntityName        string
	Fields            []*field
	Comment           string
	GenDao            bool
	DaoFileName       string
	DaoName           string
	AllowInvalidField bool
}

type field struct {
	Column            string
	FieldName         string
	FieldType         string
	IsAutoIncrement   bool
	AutoIncrementStep int
	Comment           string
	Valid             bool
}

// Cfg 生成配置
type Cfg struct {
	// 数据库类型
	DbType dbType
	// 数据库连接URL
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

type Tables []string

type MappingTypes map[string]Mapper

type Mapper map[string]string

type IgnoreColumns map[string]Columns

type Columns []string

// GetGenerator 创建生成器
func GetGenerator(c Cfg) Generator {
	checkCfg(&c)
	var d dialectGenerator
	switch c.DbType {
	case DB_MYSQL:
		d = newMySqlGenerator(c)
	case DB_ORACLE:
		d = newOracleGenerator(c)
	case DB_POSTGRES:
		d = newPostgresGenerator(c)
	case DB_SQLSERVER:
		d = newSqlServerGenerator(c)
	case DB_SQLITE:
		d = newSqliteGenerator(c)
	default:
		panic("not support this db type yet")
	}
	return Generator{c: c, d: d}
}

func checkCfg(c *Cfg) {
	fullOutPath := mustReturn(os.Getwd())
	if c.OutPath != "" {
		fullOutPath = filepath.Join(fullOutPath, c.OutPath)
	}
	c.OutPath = fullOutPath

	if c.Package == "" {
		_, p := filepath.Split(c.OutPath)
		c.Package = p
	}
}

var entityNameMapper = gdao.NewNameMapper().UpperCamelCase()
var fieldNameMapper = gdao.NewNameMapper().UpperCamelCase()
var daoNameMapper = gdao.NewNameMapper().UpperCamelCase().AddSuffix("Dao")
var entityFileNameMapper = gdao.NewNameMapper().LowerSnakeCase().AddSuffix(".go")
var daoFileNameMapper = gdao.NewNameMapper().LowerSnakeCase().AddSuffix("_dao.go")

var supportedFieldTypes = map[string]struct{}{
	"bool": {}, "string": {}, "time.Time": {}, "float32": {}, "float64": {}, "int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {}, "uint32": {}, "uint64": {},
}

//go:embed entity.tpl
var entityTpl string

//go:embed dao.tpl
var daoTpl string

//go:embed count_dao.tpl
var countDaoTpl string

var tplEntity *template.Template
var tplDao *template.Template
var tplCountDao *template.Template

func init() {
	tplEntity = mustReturn(template.New("").Parse(entityTpl))
	tplDao = mustReturn(template.New("").Parse(daoTpl))
	tplCountDao = mustReturn(template.New("").Parse(countDaoTpl))
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
