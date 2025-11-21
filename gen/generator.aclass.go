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
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"github.com/jishaocong0910/gdao/internal"
	"golang.org/x/tools/imports"
	"io"
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
	getTableInfo(table string) ([]fieldTplParam, string, error)
	getBaseDaoTemplate() string
}

type generator__ struct {
	i               Generator_
	cfg             GenCfg
	db              *sql.DB
	dir             string
	entityDir       string
	entityPkgPath   string
	entityTpl       *template.Template
	daoTpl          *template.Template
	countDaoTpl     *template.Template
	entityTplParams []entityTplParam
	baseDaoTplParam baseDaoTplParam
}

func (this *generator__) generator_() { // coverage-ignore
}

func (this *generator__) Gen() {
	err := this.checkDir()
	if err != nil { // coverage-ignore
		log.Printf("%v", err)
		return
	}
	log.Println("start generating...")
	log.Printf("full output path: %s", this.dir)
	this.queryTplParams()
	this.genBaseDao()
	this.genEntity()
	log.Println("finish generating")
}

func (this *generator__) checkDir() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(wd, this.cfg.GoModPath, "go.mod"))
	if err != nil {
		return err
	}
	r := bufio.NewReader(file)
	var moduleName string
	for {
		bytes, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line := string(bytes)
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
	this.entityPkgPath = moduleName + "/" + this.cfg.OutPath + "/entity"

	goModPath, _ := filepath.Split(file.Name())
	this.dir = filepath.Join(goModPath, this.cfg.OutPath)
	this.entityDir = filepath.Join(this.dir, "entity")
	err = os.MkdirAll(this.dir, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(this.entityDir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (this *generator__) queryTplParams() {
	_, pkgName := filepath.Split(this.cfg.OutPath)
	if this.db != nil {
		for _, table := range this.cfg.TableCfg.Tables {
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
			entityName := entityNameMapper.Convert(table)
			e := entityTplParam{
				Table:      table,
				EntityName: entityName,
				Fields:     fields,
				Comment:    comment,
				Imports:    imports,
				dao: daoTplParam{
					Table:             table,
					PkgName:           pkgName,
					DaoName:           daoNameMapper.Convert(table),
					EntityName:        entityName,
					EntityPkgPath:     this.entityPkgPath,
					AllowInvalidField: this.cfg.DaoCfg.AllowInvalidField,
				},
			}
			this.entityTplParams = append(this.entityTplParams, e)
		}
	}
	this.baseDaoTplParam = baseDaoTplParam{
		PkgName: pkgName,
	}
}

func (this *generator__) ignoreFields(table string, fields []fieldTplParam) []fieldTplParam {
	ignoreColumns := this.cfg.TableCfg.Ignores[table]
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

func (this *generator__) mappingFields(table string, fields []fieldTplParam) ([]string, error) {
	mappings := this.cfg.TableCfg.Mappers[table]
	if mappings == nil {
		return nil, nil
	}
	pkgNameToPaths := make(map[string]string, len(mappings))

	for i := 0; i < len(fields); i++ {
		f := &fields[i]
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

func (this *generator__) genBaseDao() {
	baseDaoTpl := mustReturn(template.New("").Parse(this.i.getBaseDaoTemplate()))
	err := this.createFile(this.dir, "base_dao.go", this.cfg.DaoCfg.CoverBaseDao, baseDaoTpl, this.baseDaoTplParam)
	if err != nil { // coverage-ignore
		log.Printf("create base dao fail: %+v\n", err)
	} else {
		log.Println("create base dao success")
	}
	if this.cfg.DaoCfg.GenCountDao {
		for _, table := range this.cfg.TableCfg.Tables {
			if strings.EqualFold(table, "count") { // coverage-ignore
				log.Printf("create count dao fail because exists table named \"%s\"", table)
				return
			}
		}
		err = this.createFile(this.dir, "count_dao.go", this.cfg.DaoCfg.CoverCountDao, this.countDaoTpl, this.baseDaoTplParam)
		if err != nil { // coverage-ignore
			log.Printf("create count dao fail: %+v\n", err)
		} else {
			log.Println("create count dao success")
		}
	}
}

func (this *generator__) genEntity() {
	for _, e := range this.entityTplParams {
		err := this.createFile(this.entityDir, entityFileNameMapper.Convert(e.Table), true, this.entityTpl, e)
		if err != nil { // coverage-ignore
			log.Printf("create entity of table \"%s\" fail, error: %+v\n", e.Table, err)
		} else {
			log.Printf("create entity of table \"%s\" success\n", e.Table)
		}
		err = this.createFile(this.dir, daoFileNameMapper.Convert(e.Table), false, this.daoTpl, e.dao)
		if err != nil { // coverage-ignore
			log.Printf("create dao of table \"%s\" fail, error: %+v\n", e.Table, err)
		} else {
			log.Printf("create dao of table \"%s\" success\n", e.Table)
		}
	}
}

func (this *generator__) createFile(outPath, fileName string, cover bool, tpl *template.Template, param any) error {
	path := filepath.Join(outPath, fileName)
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

func extendGenerator_(i Generator_, cfg GenCfg) *generator__ {
	if cfg.OutPath == "" {
		cfg.OutPath = "dao"
	}

	var db *sql.DB
	if cfg.Dsn != "" {
		db = mustReturn(sql.Open(i.getDriverName(), cfg.Dsn))
	}

	entityTpl := mustReturn(template.New("").Parse(entityTpl))
	daoTpl := mustReturn(template.New("").Parse(daoTpl))
	countDaoTpl := mustReturn(template.New("").Parse(countDaoTpl))

	return &generator__{i: i, cfg: cfg, db: db, entityTpl: entityTpl, daoTpl: daoTpl, countDaoTpl: countDaoTpl}
}
