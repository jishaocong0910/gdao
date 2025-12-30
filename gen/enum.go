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

import e "github.com/jishaocong0910/enum"

type dbType struct {
	*e.EnumElem__
	driverName string
}

type _DbType struct {
	*e.Enum__[dbType]
	MYSQL,
	ORACLE,
	POSTGRES,
	SQLSERVER,
	SQLITE dbType
}

var DbType_ = e.NewEnum(_DbType{
	MYSQL:     dbType{driverName: "mysql"},
	ORACLE:    dbType{driverName: "oracle"},
	POSTGRES:  dbType{driverName: "postgres"},
	SQLSERVER: dbType{driverName: "mssql"},
	SQLITE:    dbType{driverName: "sqlite"},
})

type mappingType struct {
	*e.EnumElem__
}

type _mappingType struct {
	*e.Enum__[mappingType]
	base,
	slice,
	convert mappingType
}

var mappingType_ = e.NewEnum(_mappingType{})

type logicalDelMode struct {
	*e.EnumElem__
	code int
}

type _LogicalDelMode struct {
	*e.Enum__[logicalDelMode]
	SET_NULL,
	SET_ID logicalDelMode
}

var LogicalDelMode_ = e.NewEnum(_LogicalDelMode{
	SET_NULL: logicalDelMode{code: 1},
	SET_ID:   logicalDelMode{code: 2},
})
