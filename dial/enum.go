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

package dial

import e "github.com/jishaocong0910/enum"

type dbType struct {
	e.EnumElem
	driverName string
}

type _DbType struct {
	e.Enum[dbType]
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
