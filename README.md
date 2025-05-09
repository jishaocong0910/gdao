# GDAO

GDAO是用于Golang的轻量级ORM框架，独特的数据库驱动兼容方案，用极少的API满足所有开发需求，提供了[代码生成器](#代码生成器)，设计特色如下：

- **SQL方言**。使用字符串而非SQL组装方法来组装SQL，最大限度兼容各种数据库方言，同时还提供了动态构建SQL方法。</br></br>
- **参数占位符** (reference : http://go-database-sql.org/prepared.html )。使用字符串组装SQL，因此不需要关注具体是哪种数据库，用户使用对应数据库驱动的参数占位符即可。有些数据库驱动的参数占位符是动态的，GDAO也提供了参数占位符的动态构建方法。</br></br>
- **获取自动生成ID**。有些数据库驱动支持`sql.Result#LastInsertId`方法来获取自动生成ID，有些不支持此方法而是其他方式，GDAO对此做了兼容性设计。


[![Go Reference](https://pkg.go.dev/badge/github.com/jishaocong0910/gdao.svg)](https://pkg.go.dev/github.com/jishaocong0910/gdao)
[![Go Report Card](https://goreportcard.com/badge/github.com/jishaocong0910/gdao)](https://goreportcard.com/report/github.com/jishaocong0910/gdao)
![coverage](https://raw.githubusercontent.com/jishaocong0910/gdao/badges/.badges/main/coverage.svg)

# 安装

```shell
go get github.com/jishaocong0910/gdao
```

# 用法与例子

*Example（MySQL驱动）*

```go
package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
    "github.com/jishaocong0910/gdao"
)

type User struct {
    Id       *int32 `gdao:"auto"`
    Name     *string
    Age      *int32
    Address  *string
    Phone    *string
    Email    *string
    Status   *int8
    Level    *int32
    CreateAt *time.Time
}

func main() {
    // open a db
    db, err := sql.Open("mysql", "(dsn)")
    if err != nil {
        log.Fatalln(err)
    }
    gdao.Db = db // set a default db

    // create a dao
    userDao := gdao.NewDao[User](gdao.NewDaoReq{Table: "user", ColumnMapper: gdao.NewNameMapper().LowerSnakeCase()})

    // insert
    u := &User{
        Name:     gdao.Ptr("foo"),
        Age:      gdao.Ptr[int32](1),
        Address:  gdao.Ptr("bar"),
        Phone:    gdao.Ptr("1234"),
        Email:    gdao.Ptr("test@email.com"),
        Status:   gdao.Ptr[int8](1),
        Level:    gdao.Ptr[int32](0),
        CreateAt: gdao.Ptr(time.Now()),
    }
    affected, err := userDao.Mutation(gdao.MutationReq[User]{
        Entities: []*User{u},
        BuildSql: func(b *gdao.Builder[User]) (ok bool) {
            b.Write("INSERT INTO ").WriteTable()
            cvs, _ := b.ColumnValues(true)
            b.EachColumnValues(cvs, b.SepFix("(", ",", ")"), func(_ int, column string, _ any) {
                b.Write(column)
            })
            b.EachColumnValues(cvs, b.SepFix(" VALUES(", ",", ")"), func(_ int, _ string, value any) {
                b.Arg(value)
            })
            return true
        },
    }).Insert()
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Println(affected, *u.Id) // auto increment key

    // query
    u2, _, err := userDao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("SELECT ").WriteCommaColumns().Write(" FROM ").WriteTable().Write("WHERE id=?", u.Id)
        return true
    }})
    if err != nil {
        log.Fatalln(err)
    }
    j, _ := json.Marshal(u2)
    fmt.Println(string(j))

    // update
    u3 := &User{
        Email:  gdao.Ptr("example@email.com"),
        Status: gdao.Ptr[int8](2),
    }
    userDao.Mutation(gdao.MutationReq[User]{
        Entities: []*User{u3},
        BuildSql: func(b *gdao.Builder[User]) (ok bool) {
            b.Write("UPDATE ").WriteTable().Write(" SET ")
            cvs, _ := b.ColumnValues(true)
            b.EachColumnValues(cvs, b.Sep(","), func(_ int, column string, value any) {
                b.Write(column).Write("=?", value)
            })
            b.Write(" WHERE id=?", u.Id)
            return true
        },
    })

    // delete
    userDao.Mutation(gdao.MutationReq[User]{
        BuildSql: func(b *gdao.Builder[User]) (ok bool) {
            b.Write("DELETE FROM ").WriteTable().Write(" WHERE id=?", u.Id)
            return true
        },
    })
}
```

# 实体声明

实体字段类型只支持如下类型的**指针**和切片，可多维切片（为了支持PostgreSQL）。

`bool` `string` `time.Time` `float32` `float64`

`int` `int8` `int16` `int32` `int64`

`uint` `uint8` `uint16` `uint32` `uint64`

## 字段标签
提供了字段标签功能，格式：`gdao="<values>"`，`<values>`有如下选项，多个时使用`;`拼接。
   
<table>
    <thead>
        <th>标签值</th><th>说明</th>
    </thead>
    <tbody>
        <tr>
            <td width=216px><code>column=&lt;column_name&gt;</code></td>
            <td><code>&lt;column_name&gt;</code> ::= 数据库字段名<br/><br/>指定对应的数据库字段。</td>
        </tr>
        <tr>
            <td><code>auto[=&lt;step&gt;]</code></td>
            <td><code>&lt;step&gt;</code> ::= 自增偏移量，默认为1<br/><br/>用于标记自增ID字段。</td>
        </tr>
    </tbody>
</table>

*Example*

```go
type Account struct {
    Id       *int32     `gdao:"column=id;auto"`
    UserId   *int32     `gdao:"column=user_id"`
    Balance  *float64   `gdao:"column=balance"`
    Status   *int8      `gdao:"column=status"`
    CreateAt *time.Time `gdao:"column=create_at"`
    UpdateAt *time.Time `gdao:"column=update_at"`
}
```

# DAO声明

`gdao.NewDao`函数用于创建指定实体的DAO。

<table>
    <thead>
        <th>参数</th><th>说明</th>
    </thead>
    <tbody>
        <tr>
            <td><code>Table string</code></td>
            <td>必填，对应的数据库表名。</td>
        </tr>
        <tr>
            <td width="252px"><code>Db *sql.DB</code></td>
            <td>可选，打开的<code>*sql.DB</code>变量。</td>
        </tr>
        <tr>
            <td><code>ColumnMapper *NameMapper</code></td>
            <td>可选，指定“实体->数据库“字段映射规则。若实体字段没有添加标签<code>gdao:"column=<column_name>"</code>，则使用此规则。使用<code>gdao.NewNameMapper</code>函数创建映射器，并指定映射方法，可链式调用指定多个按顺序处理。</td>
        </tr>
    </tbody>
</table>

*Example*

```go
type User struct {
    IdCol    *int64
    NameCol  *string
    PhoneCol *int8 `gdao:"column=mobile"`
}

// 映射数据库字段名为：id、name、mobile
var userDao = gdao.NewDao[User](gdao.NewDaoReq{
    Table:        "user",
    ColumnMapper: gdao.NewNameMapper().LowerCamelCase().SubSuffix("_col"), // 转化小写下划线格式并去除后缀
})
```

## 推荐代码风格

推荐每个实体具有独立的DAO，通过内嵌`*gdao.Dao`自定义所需查询。

*Example（MySQL驱动）*

```go
var UserDao = userDao{gdao.NewDao[User](gdao.NewDaoReq{Table: "user"})}
var AccountDao = accountDao{gdao.NewDao[Account](gdao.NewDaoReq{Table: "account"})}

type userDao struct {
    *gdao.Dao[User]
}

func (d userDao) GetById(id int32) (*User, error) {
    first, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("SELECT ").WriteCommaColumns().Write(" FROM ").WriteTable().Write(" WHERE id=?", id)
        return true
    }})
    if err != nil {
        return nil, err
    }
    return first, nil
}

func (d userDao) UpdateStatus(id int32, status int8) (int64, error) {
    affected, err := d.Mutation(gdao.MutationReq[User]{BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("UPDATE").WriteTable().Write(" SET status=? WHERE id=?", status, id)
        return true
    }}).Exec()
    if err != nil {
        return 0, err
    }
    return affected, nil
}

type accountDao struct {
    *gdao.Dao[Account]
}

func (d accountDao) GetByUserId(userId int32) (*Account, error) {
    first, _, err := d.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.Builder[Account]) (ok bool) {
        b.Write("SELECT ").WriteCommaColumns().Write(" FROM ").WriteTable().Write(" WHERE user_id=?", userId)
        return true
    }})
    if err != nil {
        return nil, err
    }
    return first, nil
}

func (d accountDao) ReduceBalance(id int32, balance int64) (bool, error) {
    a, _, err := d.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.Builder[Account]) (ok bool) {
        b.Write("SELECT balance FROM ").WriteTable().Write(" WHERE id=?", id)
        return true
    }})
    if err != nil {
        return false, err
    }

    oldBalance := *a.Balance
    newBalance := oldBalance - balance
    if newBalance < 0 {
        return false, nil
    }

    affected, err := d.Mutation(gdao.MutationReq[Account]{BuildSql: func(b gdao.Builder[Account]) (ok bool) {
        b.Write("UPDATE ").WriteTable().Write(" SET balance=? WHERE id=? AND balance=?", newBalance, id, oldBalance)
        return true
    }}).Exec()
    if err != nil {
        panic(err)
    }
    return affected > 0, nil
}
```

# 设置DB

使用DAO之前必须设置`*sql.Db`，有以下方式：

* 创建DAO时设置

`userDao := gdao.NewDao[User](gdao.NewDaoReq{Table: "user", Db: db})`

* 创建DAO后设置或更改

`userDao.SetDb(db)`

* 设置默认DB，DAO未设置时默认使用此DB

`gdao.Db = db`

# DAO执行方法

## RawQuery

执行查询并返回`*sql.Rows`值。

*参数*

| 字段                    | 说明      |
|-----------------------|---------|
| `Ctx context.Context` | Context |
| `Tx *sql.Tx`          | 事务      |
| `Sql  string`         | SQL语句   |
| `Args []any`          | SQL参数   |

*Example（MySQL驱动）*

```go
rows, closeFunc, err := userDao.RawQuery(gdao.RawQueryReq{Sql: "SELECT count(*) FROM user WHERE id=?", Args: []any{1}})
if err != nil {
    panic(err)
}
defer closeFunc() // don't forget invoke the close function

var count int64
if rows.Next() {
    rows.Scan(&count)
}
fmt.Println(count)
```

## RawMutation

执行变更语句并返回`sql.Result`值。

*参数*

| 字段                    | 说明      |
|-----------------------|---------|
| `Ctx context.Context` | Context |
| `Tx *sql.Tx`          | 事务      |
| `Sql  string`         | SQL语句   |
| `Args []any`          | SQL参数   |

*Example（MySQL驱动）*

```go
result, err := userDao.RawMutation(gdao.RawMutationReq{
    Sql: "INSERT into" +
            " user(id,name,age,address,phone,email,status,level,create_at)" +
            " values (?,?,?,?,?,?,?,?,?)",
    Args: []any{nil, "foo", 1, "home", "123456", "test@mail.com", 1, 0, nil},
})
if err != nil {
    panic(err)
}

id, _ := result.LastInsertId()
affected, _ := result.RowsAffected()
fmt.Println(id, affected)
```

## Query

执行查询语句并将结果映射为实体。

*参数*

| 字段                                            | 说明              |
|-----------------------------------------------|-----------------|
| `Ctx context.Context`                         | Context         |
| `Tx *sql.Tx`                                  | 事务              |
| `Entities []*T`                               | 实体参数，用于动态构建SQL。 |
| `BuildSql func(b *gdao.Builder[T]) (ok bool)` | 动态构建SQL函数       |

*Example（MySQL驱动）*

```go  
user := &User{Status: gdao.Ptr[int8](1), Level: gdao.Ptr[int32](2)}

_, list, err := userDao.Query(gdao.QueryReq[User]{
    Entities: []*User{user},
    BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("SELECT ").WriteCommaColumns().Write(" FROM ").WriteTable().Write(" WHERE ")
        cvs, _ := b.ColumnValues(true)
        b.EachColumnValues(cvs, b.Sep(" AND "), func(_ int, column string, value any) {
            b.Write(column).Write("=?").Arg(value)
        })
        return true
    }})

if err != nil {
    panic(err)
}

json, _ := json.Marshal(list)
fmt.Println(json)
```

## Mutation

执行变更语句，该方法须链式调用`Exec`、`Insert`和`Query`方法来选择执行模式，它们用于不同场景。

*参数*

| 字段                                            | 说明                                                                                                |
|-----------------------------------------------|---------------------------------------------------------------------------------------------------|
| `Ctx context.Context`                         | Context                                                                                           |
| `Tx *sql.Tx`                                  | 事务                                                                                                |
| `Entities []*T`                               | 实体参数，有以下作用：<br/><ul><li>用于动态构建SQL。</li><li>使用`Insert`或`Query`执行模式，会将自动生成ID或返回结果映射回这些实体。</li></ul> |
| `BuildSql func(b *gdao.Builder[T]) (ok bool)` | 动态构建SQL函数                                                                                         |

### Exec模式

内部调用Golang的`sql.Stmt#ExecContext`方法执行，返回影响行数。**适合用于执行UPDATE和DELETE语句**。

*Example（MySQL驱动）*

```go
affected, err := userDao.Mutation(gdao.MutationReq[User]{
    BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("UPDATE ").WriteTable().Write(" SET level=2 WHERE id=?", 1)
        return true
    }}).Exec()
if err != nil {
    panic(err)
}
fmt.Println(affected)
```

### Insert模式

在Exec模式基础上多了一个步骤：调用`sql.Result#LastInsertId`方法将自增ID映射到实体参数中。**适合用于支持`sql.Result#LastInsertId`方法的数据库驱动执行INSERT语句，如果数据库驱动不支持，则映射自增ID将返回error或不准确**。

*Example（MySQL驱动）*

```go
affected, err := userDao.Mutation(gdao.MutationReq[User]{
    Entities: users,
    BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("INSERT INTO ").WriteTable().Write("(").WriteCommaColumns().Write(") VALUES")
        b.EachEntity(b.Sep(","), func(_ int, entity *User) {
            cvs, _ := b.ColumnValuesAt(entity, false)
            b.EachColumnValues(cvs,b.SepFix("(",",",")"), func(_ int, _ string, value any) {
                b.Write("?").Arg(value)
            })
        })
        return true
    }}).Insert()

if err != nil {
    panic(err)
}

fmt.Println(affected)
fmt.Println(*users[0].Id)
fmt.Println(*users[1].Id)
```

### Query模式

内部调用Golang的`sql.Stmt#QueryContext`方法执行，返回影响行数。`sql.Stmt#QueryContext`方法的返回结果`*sql.Rows`中的数据，会被映射到实体参数中。**适用于一些不支持`sql.Result#LastInsertId`方法的数据库驱动，执行INSERT语句后映射自动生成key**。

*Example（PostgreSQL驱动）*

```go
affected, err := userDao.Mutation(gdao.MutationReq[User]{
    Entities: users,
    BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("INSERT INTO ").WriteTable().Write("(").WriteCommaColumns().Write(") VALUES")
        b.EachEntity(b.Sep(","), func(_ int, entity *User) {
            cvs, _ := b.ColumnValuesAt(entity, false)
            b.EachColumnValues(cvs,b.SepFix("(",",",")"), func(_ int, _ string, value any) {
                b.Write("?").Arg(value)
            })
        })
        b.Write(" RETURNING id")
        return true
    }}).Query()

if err != nil {
    panic(err)
}

fmt.Println(affected)
fmt.Println(*users[0].Id)
fmt.Println(*users[1].Id)
```

*Example（SQLserver驱动）*

```go
affected, err := userDao.Mutation(gdao.MutationReq[User]{
    Entities: users,
    BuildSql: func(b *gdao.Builder[User]) (ok bool) {
        b.Write("INSERT INTO ").WriteTable().Write("(").WriteCommaColumns().Write(") VALUES")
        b.EachEntity(b.Sep(","), func(_ int, entity *User) {
            cvs, _ := b.ColumnValuesAt(entity, false)
            b.EachColumnValues(cvs,b.SepFix("(",",",")"), func(_ int, _ string, value any) {
                b.Write("?").Arg(value)
            })
        })
        b.Write("; select ID = convert(bigint, SCOPE_IDENTITY())")
        return true
    }}).Query()

if err != nil {
    panic(err)
}

fmt.Println(affected)
fmt.Println(*users[0].Id)
fmt.Println(*users[1].Id)
```

# 动态构建SQL

DAO的`Query`和`Mutation`方法具有的`Entities`和`BuildeSql`参数用于动态构建SQL。

`BuildeSql`是一个函数，其唯一参数`b *gdao.Bulider`用于拼接SQL和设置参数，并提供了许多动态构建SQL的方法，`Entities`将作为某些方法的数据来源。返回值`ok`用于标识SQL是否执行。

*Example（MySQL驱动）*

```go
// InsertBatch 批量插入数据
func (d userDao) InsertBatch(entities []*User) (int64, error) {
    affected, err := d.Mutation(gdao.MutationReq[User]{
        Entities: entities,
        BuildSql: func(b *gdao.Builder[User]) (ok bool) {
            // 如果请求参数是空的，则返回false表示不执行
            if len(entities) == 0 {
                return false
            }
            // 开始拼接INSERT语句
            b.Write("INSERT INTO ").WriteTable().Write("(").WriteCommaColumns().Write(") VALUES")
            // 遍历Entities的每个实体
            b.EachEntity(b.Sep(","), func(_ int, entity *User) {
                // 获取实体的“列名称-字段值“键值对列表
                cvs, _ := b.ColumnValues(false)
                // 遍历这些键值对
                b.EachColumnValues(cvs, b.SepFix("(", ",", ")"), func(index int, columnName string, value any) {
                    // 拼接参数占位符并设置参数
                    b.Write("?").Arg(value)
                })
            })
            // 最终会拼接成类似如下SQL（为了方便说明SQL已格式化）
            //
            // INSERT INTO
            //     user(id,name,age,address,phone,email,status,level,create_at)
            // VALUES
            //     (?,?,?,?,?,?,?,?,?),
            //     (?,?,?,?,?,?,?,?,?),
            //     ...
            //     (?,?,?,?,?,?,?,?,?)
            return true
        }}).Insert()
    if err != nil {
        return 0, err
    }
    return affected, nil
}
```

## gdao.Builder的方法

| 方法                  | 说明                                                                                                       |
|---------------------|----------------------------------------------------------------------------------------------------------|
| `Write`             | 拼接字符串并设置参数。                                                                                              |
| `Arg`               | 设置参数，`b.Write(foo,bar)`和`b.Write(foo).Arg(bar)`是等价的。                                                     |
| `WriteTable`        | 拼接表名称。                                                                                                   |
| `WriteCommaColumns` | 拼接列名称，使用逗号分隔，如果参数为空则拼接表的所有列名称。                                                                           |
| `Columns`           | 返回所有列名称。                                                                                                 |
| `AutoColumns`       | 返回标签值有`auto`的字段。                                                                                         |
| `Pp`                | 返回带编号的占位符，编号从1开始，每次调用后递增1，适用于PostgreSQL、Oracle等驱动。                                                       |
| `EntityAt`          | 返回`Entities`中指定索引的实体。                                                                                    |
| `Entity`            | 相当于`b.EntityAt(0)`                                                                                       |
| `ColumnValuesAt`    | 将实体转化为“列名称-字段值“键值对，并拆分为两个列表，第一个过滤掉列名称在`filterColumns`参数包含的，第二个则存放被过滤掉的，`onlyAssigned`参数表示两个都过滤掉值为nil的字段。 |
| `ColumnValues`      | 相当于`b.ColumnValuesAt(b.Entity())`                                                                        |
| `ColumnValue`       | 返回首个实体中指定列名称对应字段的值。                                                                                      |
| `EachColumnName`    | 遍历指定列名称，`filterColumns`参数指定过滤的列名称，`handle`函数参数`n`调用次数，从1开始，`i`为列名称索引。                                    |
| `EachEntity`        | 遍历`Entities`，过滤nil元素，`handle`函数参数`n`调用次数，从1开始，`i`为实体索引。                                                  |
| `EachColumnValues`  | 遍历“列名称-字段值”键值对列表。                                                                                        |
| `Repeat`            | 循环指定次数，`handle`函数参数`n`调用次数，从1开始，`i`为循环次数。                                                                |
| `Sep`               | 在Each开头的方法和Repeat方法中，拼接指定分隔符号                                                                            |
| `SepFix`            | 在Each开头的方法和Repeat方法中，拼接指定开始、分隔和结束符号。                                                                     |
| `Sql`               | 返回拼接的字符串。                                                                                                |
| `Args`              | 返回所有设置的参数。                                                                                               |

# 事务

使用事务非常简单，每个DAO执行方法参数都有一个`tx`字段，只需创建`*sql.Tx`变量传入即可。

*Example*

```go
tx, err := UserDao.Db().Begin()
if err != nil {
    panic(err)
}
defer tx.Rollback()

_, err = UserDao.Mutation(gdao.MutationReq[User]{Tx: tx, BuildSql: func(b *gdao.Builder[User]) (ok bool) {
    b.Write("UPDATE user SET status=-1 WHERE user_id=?", 1)
    return true
}}).Exec()
if err != nil {
    panic(err)
}

_, err = AccountDao.Mutation(gdao.MutationReq[Account]{Tx: tx, BuildSql: func(b *gdao.Builder[Account]) (ok bool) {
    b.Write("UPDATE account SET status=-1 WHERE user_id=?", 1)
    return true
}}).Exec()
if err != nil {
    panic(err)
}

tx.Commit()
```

# 日志

通过`gdao.LogConf`方法配置日志。

<table>
    <thead>
        <th>标签值</th>
        <th>说明</th>
    </thead>
    <tbody>
        <tr>
            <td width="210px"><code>log gdao.Logger</code></td>
            <td>设置日志器，日志器须实现<code>gdao.Logger</code>。</td>
        </tr>
        <tr>
            <td><code>level gdao.logLevel</code></td>
            <td>打印SQL的日志级别，使用常量<code>gdao.LOG_LEVEL_DEBUG</code>、<code>gdao.LOG_LEVEL_INFO</code>指定。</td>
        </tr>
    </tbody>
</table>

# 代码生成器

GDAO提供了常用数据库的实体和DAO代码生成器，**生成后的代码允许二次编辑**，方便扩展功能。

*数据库支持情况*

| 数据库        | 是否支持 |
|------------|------|
| MySQL      | ✅支持  |
| PostgreSQL | ✅支持  |
| Oracle     | ✅支持  |
| SQLserver  | ✅支持  |
| SQLite     | ✅支持  |

*Example*

```go
package main

import (
    "github.com/jishaocong0910/gdao/gen"
)

func main() {
    gen.GetGenerator(gen.Conf{
        DbType:  gen.DB_MYSQL,
        Dsn:     "(dsn)",
        OutPath: "dao", // 生成文件相对路径，绝对路径为"os.Getwd()/OutPath"。
        Tables: gen.Tables{ 
            // key为表名，value为强制映射字段类型，填写非GDAO支持类型的不会报错，但GDAO不会识别。
            "user":    nil,
            "account": nil,
        },
        GenDao: true, // 是否生成DAO，false只生成实体。
    }).Gen()
}
```

## 扩展生成代码

上文的例子生成的文件如下，这些文件都可以手动编辑扩展。**代码生成器再次执行时，只会覆盖实体文件，不会覆盖DAO文件**。

```
.
└─ dao
   ├─ base_dao.go    // 基础DAO文件
   ├─ account.go     // 实体文件
   ├─ account_dao.go // DAO文件
   ├─ user.go        // 实体文件
   └─ user_dao.go    // DAO文件
```

*生成的user_dao.go文件*

```go
// Code generated by https://github.com/jishaocong0910/gdao. YOU CAN EDIT FOR MORE.

package testdata

import "github.com/jishaocong0910/gdao"

var UserDao = _UserDao{newBaseDao[User](gdao.NewDaoReq{Table: "user"})}

type _UserDao struct {
    *BaseDao[User]
}
```


*扩展user_dao.go文件*

```go
// Code generated by https://github.com/jishaocong0910/gdao. YOU CAN EDIT FOR MORE.

package testdata

import (
    "context"
    "database/sql"

    "github.com/jishaocong0910/gdao"
)

var UserDao = _UserDao{newBaseDao[User](gdao.NewDaoReq{Table: "user"})}

type _UserDao struct {
    *BaseDao[User]
}

// 扩展了一个查询方法
func (d _UserDao) QueryByStatus(ctx context.Context, tx *sql.Tx, status ...int) ([]*User, error) {
    _, list, err := d.Query(gdao.QueryReq[User]{
        Ctx: ctx,
        Tx:  tx,
        BuildSql: func(b *gdao.Builder[User]) (ok bool) {
            if len(status) == 0 {
                return false
            }
            b.Write("SELECT * FROM user WHERE status IN")
            b.Repeat(len(status), b.SepFix("(", ",", ")"), nil, func(n, i int) {
                b.Write(",", status[i])
            })
            return true
        },
    })
    return list, err
}
```

# 基础DAO

生成的DAO内置基础的CURD方法，这些方法实现在`base_dao.go`文件中的`BaseDao`结构体中，不同数据库实现细节不同无法通用。

*Example*

```go
package main

import (
    "github.com/jishaocong0910/gdao"
    "demo/dao"
)

func main() {
    u := &dao.User{Id: gdao.Ptr[int64](1), Email: gdao.Ptr("foo"), Status: gdao.Ptr[int8](2)}
    u2 := &dao.User{Id: gdao.Ptr[int64](2), Email: gdao.Ptr("bar"), Status: gdao.Ptr[int8](0)}
    // 将执行SQL如下（为了方便说明SQL已格式化并填充参数）
    // UPDATE user SET
    //   email = CASE id
    //           WHEN 1 THEN 'foo'
    //           WHEN 2 THEN 'bar' END,
    //   status = CASE id
    //            WHEN 1 THEN 2
    //            WHEN 2 THEN 0 END
    // WHERE id IN (1, 2)
    dao.UserDao.UpdateBatch(nil, nil, []*dao.User{u, u2}, []string{"email", "status"}, nil, "id")
}
```

# BaseDao方法

## List

查询实体列表。

*参数*

| 字段                          | 说明                  |
|-----------------------------|---------------------|
| `ctx context.Context`       | Context             |
| `tx *sql.Tx`                | 事务                  |
| `entity *T`                 | 实体，不为nil的字段将作为查询条件。 |
| `selectColumns []string`    | 查询的字段列表             |
| `whereNullColumns []string` | `IS NULL`条件的字段列表    |

*Example*

```go
package main

import (
    "fmt"

    "github.com/jishaocong0910/gdao"
    "demo/dao"
)

func main() {
	cond := dao.And(dao.Eq("level", 10), dao.Or(dao.Eq("status", 1), dao.Eq("status", 2)))
	// 将执行SQL如下（为了方便说明SQL已格式化并填写参数）
	// SELECT id,name FROM user WHERE level=10 AND (status=1 OR status =2)
	list, err := dao.UserDao.List(nil, nil, []string{"id", "name"}, cond)
	if err != nil {
		fmt.Println(len(list))
	}
}
```

## Get

参数与`List`一致，返回单个实体。

## Insert

## InsertBatch

## Update

## UpdateBatch

## Delete





