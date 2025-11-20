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
	"bytes"
	"database/sql"
	"errors"
	"github.com/jishaocong0910/gdao/internal"
	"golang.org/x/tools/imports"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type Generator_ interface {
	generator_()
	Gen()
	getDriverName() string
	getTableInfo(table string) ([]*fieldTplParam, string, error)
	getBaseDaoTemplate() string
}

type generator__ struct {
	i               Generator_
	c               GenCfg
	db              *sql.DB
	entityTpl       *template.Template
	daoTpl          *template.Template
	countDaoTpl     *template.Template
	entityTplParams []entityTplParam
}

func (this *generator__) generator_() { // coverage-ignore
}

func (this *generator__) Gen() {
	log.Println("start generating...")
	log.Printf("full output path: %s", this.c.OutPath)
	this.queryEntityTplParams()
	this.createOutPath()
	this.genBaseDao()
	this.genEntity()
	log.Println("finish generating")
}

func (this *generator__) queryEntityTplParams() {
	if this.db != nil {
		for _, table := range this.c.TableCfg.Tables {
			// 获取表信息
			fields, comment, err := this.i.getTableInfo(table)
			if err != nil { // coverage-ignore
				log.Println(err.Error())
				continue
			}
			// 过滤字段
			fields = this.ignoreFields(table, fields)
			// 自定义映射
			imports, err := this.mappingFields(table, fields)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			// 创建实体模板参数
			e := entityTplParam{
				Table:             table,
				EntityFileName:    entityFileNameMapper.Convert(table),
				Package:           this.c.Package,
				EntityName:        entityNameMapper.Convert(table),
				Fields:            fields,
				Comment:           comment,
				GenDao:            this.c.DaoCfg.GenDao,
				DaoFileName:       daoFileNameMapper.Convert(table),
				DaoName:           daoNameMapper.Convert(table),
				AllowInvalidField: this.c.DaoCfg.AllowInvalidField,
				Imports:           imports,
			}
			this.entityTplParams = append(this.entityTplParams, e)
		}
	}
}

func (this *generator__) ignoreFields(table string, fields []*fieldTplParam) []*fieldTplParam {
	ignoreColumns := this.c.TableCfg.Ignores[table]
	if ignoreColumns != nil {
		var temp []*fieldTplParam
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

func (this *generator__) mappingFields(table string, fields []*fieldTplParam) ([]string, error) {
	mappings := this.c.TableCfg.Mappers[table]
	if mappings == nil {
		return nil, nil
	}
	pkgNameToPaths := make(map[string]string, len(mappings))

	for _, f := range fields {
		if m, ok := mappings[f.Column]; ok {
			switch m.mt.String() {
			case mappingType_.base.String():
				f.FieldType = "*" + reflect.TypeOf(m.t).String()
			case mappingType_.slice.String():
				f.FieldType = "[]" + reflect.TypeOf(m.t).String()
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
					pkgPath, pkgName, typeName := this.determineFieldType(ft, pkgNameToPaths)
					f.FieldType = pkgName + "." + typeName
					pkgNameToPaths[pkgName] = pkgPath
				} else { // coverage-ignore
					return nil, errors.New("the mapping of table \"" + table + "\"'s column \"" + f.Column + "\" is invalid implementing gdao.Convert")
				}
			}
		}
	}
	var imports []string
	for name, path := range pkgNameToPaths {
		if name != path[strings.LastIndex(path, "/")+1:] {
			imports = append(imports, name+" \""+path+"\"")
		} else {
			imports = append(imports, "\""+path+"\"")
		}
	}
	return imports, nil
}

func (this *generator__) determineFieldType(ft reflect.Type, pkgNameToPaths map[string]string) (string, string, string) {
	var pkgPath string
	if ft.Kind() == reflect.Pointer {
		pkgPath = ft.Elem().PkgPath()
	} else {
		pkgPath = ft.PkgPath()
	}
	arr := strings.SplitN(ft.String(), ".", 2)
	pkgName := arr[0]
	if pkgName[:1] == "*" {
		pkgName = pkgName[1:]
	}
	typeName := arr[1]
	pkgName = this.determinePkgName(pkgPath, pkgName, pkgNameToPaths, false)
	return pkgPath, pkgName, typeName
}

func (this *generator__) determinePkgName(pkgPath, pkgName string, pkgNameToPaths map[string]string, conflict bool) string {
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
			return this.determinePkgName(pkgPath, pkgName, pkgNameToPaths, true)
		}
	} else {
		return pkgName
	}
}

func (this *generator__) createOutPath() {
	must(os.MkdirAll(this.c.OutPath, os.ModePerm))
}

func (this *generator__) genBaseDao() {
	if this.c.DaoCfg.GenDao {
		b := baseDaoTplParam{
			Package: this.c.Package,
		}
		baseDaoTpl := mustReturn(template.New("").Parse(this.i.getBaseDaoTemplate()))
		err := this.createFile("base_dao.go", this.c.DaoCfg.CoverBaseDao, baseDaoTpl, b)
		if err != nil { // coverage-ignore
			log.Printf("create base dao fail: %+v\n", err)
		} else {
			log.Println("create base dao success")
		}
		if this.c.DaoCfg.GenCountDao {
			for _, table := range this.c.TableCfg.Tables {
				if strings.EqualFold(table, "count") { // coverage-ignore
					log.Printf("create count dao fail because exists table named \"%s\"", table)
					return
				}
			}
			err = this.createFile("count_dao.go", this.c.DaoCfg.CoverCountDao, this.countDaoTpl, b)
			if err != nil { // coverage-ignore
				log.Printf("create count dao fail: %+v\n", err)
			} else {
				log.Println("create count dao success")
			}
		}
	}
}

func (this *generator__) genEntity() {
	for _, e := range this.entityTplParams {
		err := this.createFile(e.EntityFileName, true, this.entityTpl, e)
		if err != nil { // coverage-ignore
			log.Printf("create entity of table \"%s\" fail, error: %+v\n", e.Table, err)
		} else {
			log.Printf("create entity of table \"%s\" success\n", e.Table)
		}
		if this.c.DaoCfg.GenDao {
			err := this.createFile(e.DaoFileName, false, this.daoTpl, e)
			if err != nil { // coverage-ignore
				log.Printf("create dao of table \"%s\" fail, error: %+v\n", e.Table, err)
			} else {
				log.Printf("create dao of table \"%s\" success\n", e.Table)
			}
		}
	}
}

func (this *generator__) createFile(fileName string, cover bool, tpl *template.Template, param any) error {
	path := filepath.Join(this.c.OutPath, fileName)
	if !cover {
		_, err := os.Stat(path)
		if err == nil { // coverage-ignore
			return errors.New("file already exists")
		}
	}
	var buf bytes.Buffer
	err := tpl.Execute(&buf, param)
	if err != nil { // coverage-ignore
		return err
	}
	content, importErr := imports.Process("", buf.Bytes(), nil)
	if importErr != nil { // coverage-ignore
		content = buf.Bytes()
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil { // coverage-ignore
		return err
	}
	return importErr
}

func extendGenerator_(i Generator_, c GenCfg) *generator__ {
	fullOutPath := mustReturn(os.Getwd())
	if c.OutPath != "" {
		fullOutPath = filepath.Join(fullOutPath, c.OutPath)
	}
	c.OutPath = fullOutPath

	if c.Package == "" { // coverage-ignore
		_, p := filepath.Split(c.OutPath)
		c.Package = p
	}

	var db *sql.DB
	if c.Dsn != "" {
		db = mustReturn(sql.Open(i.getDriverName(), c.Dsn))
	}

	entityTpl := mustReturn(template.New("").Parse(entityTpl))
	daoTpl := mustReturn(template.New("").Parse(daoTpl))
	countDaoTpl := mustReturn(template.New("").Parse(countDaoTpl))

	return &generator__{i: i, c: c, db: db, entityTpl: entityTpl, daoTpl: daoTpl, countDaoTpl: countDaoTpl}
}
