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
	for table, fieldTypes := range g.c.Tables {
		if strings.EqualFold(table, "count") {
			tableNamedCount = table
		}
		exists, fields, comment := g.d.getTableInfo(table)
		if !exists {
			log.Printf("table \"%s\" is not exists", table)
			continue
		}
		e := entity{
			DbType:         g.c.DbType,
			Table:          table,
			EntityFileName: entityFileNameMapper.Convert(table),
			Package:        g.c.Package,
			EntityName:     entityNameMapper.Convert(table),
			Fields:         fields,
			Comment:        comment,
			GenDao:         g.c.GenDao,
			DaoFileName:    daoFileNameMapper.Convert(table),
			DaoName:        daoNameMapper.Convert(table),
		}
		for column, fieldType := range fieldTypes {
			for _, f := range e.Fields {
				if f.Column == column {
					f.FieldType = fieldType
					if strings.HasPrefix(f.FieldType, "[]") {
						if _, ok := supportedFieldTypes[strings.TrimLeft(f.FieldType, "[]")]; !ok {
							f.Valid = false
						}
					} else {
						if !strings.HasPrefix(f.FieldType, "*") {
							f.FieldType = "*" + f.FieldType
						}
						if _, ok := supportedFieldTypes[f.FieldType[1:]]; !ok {
							f.Valid = false
						}
					}
				}
			}
		}
		entities = append(entities, e)
	}

	if len(entities) == 0 {
		return
	}

	g.createOutPath()
	if g.c.GenDao {
		b := baseDao{
			DbType:  g.c.DbType,
			Package: g.c.Package,
		}
		err := g.createFile("base_dao.go", g.c.CoverBaseDao, g.d.getBaseDaoTemplate(), b)
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
		if g.c.GenDao {
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
	DbType         dbType
	Table          string
	EntityFileName string
	Package        string
	EntityName     string
	Fields         []*field
	Comment        string
	GenDao         bool
	DaoFileName    string
	DaoName        string
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
	// 需要生成的表，key为表名称，value为强制指定字段类型
	Tables Tables
	// 是否生成DAO
	GenDao bool
	// 覆盖BaseDao
	CoverBaseDao bool
}

type Tables map[string]FieldTypes

type FieldTypes map[string]string

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
