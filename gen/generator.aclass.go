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
	"fmt"
	"golang.org/x/tools/imports"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator_ interface {
	generator_()
	Gen()
	getDriverName() string
	getTableInfo(table string) (bool, []*fieldTplParam, string)
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
	// 记录表名为“count”（不区分大小写）的表，因为表名为count时会和生成CountDao冲突。
	namedCountCiTable string
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
			if strings.EqualFold(table, "count") { // coverage-ignore
				this.namedCountCiTable = table
			}
			// 获取表信息
			exists, fields, comment := this.i.getTableInfo(table)
			if !exists { // coverage-ignore
				log.Printf("table \"%s\" is not exists", table)
				continue
			}
			// 过滤字段
			if columns, ok := this.c.TableCfg.IgnoreColumns[table]; ok {
				var filtered []*fieldTplParam
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
			if mappingTypes, ok := this.c.TableCfg.MappingTypes[table]; ok {
				for column, fieldType := range mappingTypes {
					for _, f := range fields {
						if f.Column == column {
							f.FieldType = fieldType
							if strings.HasPrefix(f.FieldType, "[]") {
								if _, ok := supportedFieldTypes[strings.TrimLeft(f.FieldType, "[]")]; !ok { // coverage-ignore
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
			}
			this.entityTplParams = append(this.entityTplParams, e)
		}
	}
}

func (this *generator__) createOutPath() {
	mustNoError(os.MkdirAll(this.c.OutPath, os.ModePerm))
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
			if this.namedCountCiTable != "" { // coverage-ignore
				log.Printf("create count dao fail because exists table named \"%s\"", this.namedCountCiTable)
			} else {
				err = this.createFile("count_dao.go", this.c.DaoCfg.CoverCountDao, this.countDaoTpl, b)
				if err != nil { // coverage-ignore
					log.Printf("create count dao fail: %+v\n", err)
				} else {
					log.Println("create count dao success")
				}
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
	content, err := imports.Process("", buf.Bytes(), nil)
	if err != nil { // coverage-ignore
		return err
	}
	err = os.WriteFile(path, content, 0644)
	if err != nil { // coverage-ignore
		return err
	}
	return nil
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

	db, err := sql.Open(i.getDriverName(), c.Dsn)
	if err != nil { // coverage-ignore
		err = fmt.Errorf("connect db fail, dsn: %s, error: %w", c.Dsn, err)
	}
	mustNoError(err)

	entityTpl := mustReturn(template.New("").Parse(entityTpl))
	daoTpl := mustReturn(template.New("").Parse(daoTpl))
	countDaoTpl := mustReturn(template.New("").Parse(countDaoTpl))

	return &generator__{i: i, c: c, db: db, entityTpl: entityTpl, daoTpl: daoTpl, countDaoTpl: countDaoTpl}
}
