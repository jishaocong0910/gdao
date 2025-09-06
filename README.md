# GDAO

GDAO是用于Golang的轻量级ORM框架，具有独特的数据库驱动兼容方案，用极少的API满足所有开发需求，提供了[代码生成器](#代码生成器)，设计特色如下：

- **SQL方言**。使用字符串而非SQL组装方法来构建SQL，最大限度兼容各种数据库方言，同时还提供了动态构建SQL方法。</br></br>
- **参数占位符** (reference : http://go-database-sql.org/prepared.html )。使用字符串构建SQL，因此不需要关注具体是哪种数据库，用户使用对应数据库驱动的参数占位符即可。有些数据库驱动的参数占位符是动态的，GDAO也提供了参数占位符的动态构建方法。</br></br>
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
    gdao.DEFAULT_DB = db // set a default db

    // create a dao
    userDao := gdao.NewDao[User](gdao.NewDaoReq{ColumnMapper: gdao.NewNameMapper().LowerSnakeCase()})

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
    affected, err := userDao.Exec(gdao.ExecReq[User]{
        Entities:       []*User{u},
        LastInsertIdAs: gdao.LAST_INSERT_ID_AS_FIRST_ID,
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("INSERT INTO user")
            cvs := b.ColumnValues(true)
            b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(column string, _ any) {
                b.Write(column)
            })
            b.EachColumnValues(cvs, b.SepFix(" VALUES(", ",", ")", false), func(_ string, value any) {
                b.Arg(value)
            })
        },
    })
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Println(affected, *u.Id) // auto increment key

    // query
    u2, _, err := userDao.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
        b.Write("SELECT ").WriteColumns().Write(" FROM user WHERE id=?", u.Id)
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
    userDao.Exec(gdao.ExecReq[User]{
        Entities: []*User{u3},
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("UPDATE user SET ")
            cvs := b.ColumnValues(true)
            b.EachColumnValues(cvs, b.Sep(","), func(column string, value any) {
                b.Write(column).Write("=?", value)
            })
            b.Write(" WHERE id=?", u.Id)
        },
    })

    // delete
    userDao.Exec(gdao.ExecReq[User]{
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("DELETE FROM user WHERE id=?", u.Id)
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

格式：`gdao="<values>"`，`<values>`有如下选项，多个时使用`;`拼接。
   
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
            <td width="252px"><code>Db *sql.DB</code></td>
            <td>可选，打开的<code>*sql.DB</code>变量。</td>
        </tr>
        <tr>
            <td><code>AllowInvalidField bool</code></td>
            <td>是否允许非法字段，如字段未导出、未使用指针等，默认为false，检测到非法字段将panic。</td>
        </tr>
        <tr>
            <td><code>ColumnMapper *NameMapper</code></td>
            <td>可选，指定默认的“实体->数据库“字段映射规则。若实体字段没有标签<code>gdao:"column=&lt;column_name&gt;"</code>，则使用此规则。使用<code>gdao.NewNameMapper</code>函数创建映射器，并指定映射方法，可链式调用指定多个按顺序处理。</td>
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
var UserDao = gdao.NewDao[User](gdao.NewDaoReq{
    ColumnMapper: gdao.NewNameMapper().LowerCamelCase().SubSuffix("_col"), // 转化小写下划线格式并去除后缀
})
```

## 推荐的代码风格

推荐每个实体具有独立的DAO，通过内嵌`*gdao.Dao`自定义所需查询。

*Example（MySQL驱动）*

```go
var UserDao = _UserDao{gdao.NewDao[User](gdao.NewDaoReq{})}
var AccountDao = _AccountDao{gdao.NewDao[Account](gdao.NewDaoReq{})}

type _UserDao struct {
    *gdao.Dao[User]
}

func (d _UserDao) GetById(id int32) (*User, error) {
    first, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b *gdao.Builder[User]) {
        b.Write("SELECT ").WriteColumns().Write(" FROM user WHERE id=?", id)
    }})
    if err != nil {
        return nil, err
    }
    return first, nil
}

func (d _UserDao) UpdateStatus(id int32, status int8) (int64, error) {
    affected, err := d.Exec(gdao.ExecReq[User]{BuildSql: func(b *gdao.Builder[User]) {
        b.Write("UPDATE user SET status=? WHERE id=?", status, id)
    }})
    if err != nil {
        return 0, err
    }
    return affected, nil
}

type _AccountDao struct {
    *gdao.Dao[Account]
}

func (d _AccountDao) GetByUserId(userId int32) (*Account, error) {
    first, _, err := d.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.Builder[Account]) {
        b.Write("SELECT ").WriteColumns().Write(" FROM user WHERE user_id=?", userId)
    }})
    if err != nil {
        return nil, err
    }
    return first, nil
}

func (d _AccountDao) ReduceBalance(id int32, balance int64) (bool, error) {
    a, _, err := d.Query(gdao.QueryReq[Account]{BuildSql: func(b *gdao.Builder[Account]) {
        b.Write("SELECT balance FROM user WHERE id=?", id)
    }})
    if err != nil {
        return false, err
    }

    oldBalance := *a.Balance
    newBalance := oldBalance - balance
    if newBalance < 0 {
        return false, nil
    }

    affected, err := d.Exec(gdao.ExecReq[Account]{BuildSql: func(b *gdao.Builder[Account]) {
        b.Write("UPDATE account SET balance=? WHERE id=? AND balance=?", newBalance, id, oldBalance)
    }})
    if err != nil {
        return false, err
    }
    return affected > 0, nil
}
```

# 设置DB

必须设置`*sql.DB`才可以执行SQL，有以下方式：

1. 全局默认DB

```go
gdao.DEFAULT_DB = db
```

2. 创建DAO时设置

```go
UserDao := gdao.NewDao[User](gdao.NewDaoReq{DB: db})
```

优先级：2 > 1

# DAO执行方法

`gdao.Dao`只有两个执行方法`Query`和`Exec`，它们的功能足以满足所有开发需求。

## Query

执行查询语句并将结果映射为实体，提了供获取自增ID的模式，见章节[获取自增ID](#获取自增ID)。

*参数*

| 字段                                  | 说明                              |
|-------------------------------------|---------------------------------|
| `Ctx context.Context`               | Context                         |
| `RowAs gdao.rowAs`                  | 指定当前为获取插入记录自增ID模式。              |
| `Entities []*T`                     | 实体参数，用于动态构建SQL，获取的自增ID会注入到这些实体。 |
| `BuildSql func(b *gdao.Builder[T])` | 动态构建SQL函数                       |

*Example（MySQL驱动）*

```go
func foo() error {
    user := &User{Status: gdao.Ptr[int8](1), Level: gdao.Ptr[int32](2)}

    _, list, err := UserDao.Query(gdao.QueryReq[User]{
        Entities: []*User{user},
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("SELECT ").WriteColumns().Write(" FROM user WHERE ")
            cvs := b.ColumnValues(true)
            b.EachColumnValues(cvs, b.Sep(" AND "), func(column string, value any) {
                b.Write(column).Write("=?").Arg(value)
            })
        }})

    if err != nil {
        return err
    }

    json, _ := json.Marshal(list)
    fmt.Println(json)
    return nil
}
```

## Exec

执行INSERT、UPDATE和DELETE语句并返回影响行数，提供了获取自增ID的模式，见章节[获取自增ID](#获取自增ID)。

*参数*

| 字段                                   | 说明                              |
|--------------------------------------|---------------------------------|
| `Ctx context.Context`                | Context                         |
| `LastInsertIdAs gdao.lastInsertIdAs` | 指定当前为获取插入记录自增ID模式。              |
| `Entities []*T`                      | 实体参数，用于动态构建SQL，获取的自增ID会注入到这些实体。 |
| `BuildSql func(b *gdao.Builder[T])`  | 动态构建SQL函数                       |

*Example（MySQL驱动）*

```go
func foo() {
    affected, err := UserDao.Exec(gdao.ExecReq[User]{
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("UPDATE user SET level=2 WHERE id=?", 1)
        }})

    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(affected)
    }
}
```

# 获取自增ID

`Query`和`Exec`方法分别提供了多种获取自增ID的模式，以适应不同数据库驱动在这方面的差异性。

## LastInsertIdAs参数

`Exec`方法用于支持`sql.Result#LastInsertId`方法的数据库驱动获取自增ID。尽管`sql.Result#LastInsertId`方法是Golang的标准，但不同数据库驱动实现的意义不同。`Exec`方法的`LastInsertIdAs`参数提供了以下模式，指定模式后会将`sql.Result#LastInsertId`方法的值注入到`Entities`参数中。


| 可选值                               | 说明                                                                    |
|-----------------------------------|-----------------------------------------------------------------------|
| `gdao.LAST_INSERT_ID_AS_FIRST_ID` | 将`sql.Result#LastInsertId`方法的值作为第一个插入纪录的自增ID，适配此模式的典型数据库：**MySQL**。   |
| `gdao.LAST_INSERT_ID_AS_LAST_ID`  | 将`sql.Result#LastInsertId`方法的值作为最后一个插入纪录的自增ID，适配此模式的典型数据库：**Sqlite**。 |

*Example（MySQL驱动）*

```go
func foo() {
    users := []*User{
        {Name: gdao.Ptr("Jack"), Phone: gdao.Ptr("12345"), Email: gdao.Ptr("jack@email.com")},
        {Name: gdao.Ptr("Nick"), Phone: gdao.Ptr("43422"), Email: gdao.Ptr("rose@email.com")},
    }

    affected, err := UserDao.Exec(gdao.ExecReq[User]{
        Entities:       users,
        LastInsertIdAs: gdao.LAST_INSERT_ID_AS_FIRST_ID,
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("INSERT INTO user(").WriteColumns().Write(") VALUES")
            b.EachEntity(b.Sep(","), func(_, _ int, entity *User) {
                cvs := b.ColumnValuesAt(entity, false)
                b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(_ int, _ string, value any) {
                    b.Write("?").Arg(value)
                })
            })
        }})

    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(affected)
        fmt.Println(*users[0].Id)
        fmt.Println(*users[1].Id)
    }
}
```

*Example（SQLite驱动）*

```go
func foo() {
    users := []*User{
        {Name: gdao.Ptr("Jack"), Phone: gdao.Ptr("12345"), Email: gdao.Ptr("jack@email.com")},
        {Name: gdao.Ptr("Nick"), Phone: gdao.Ptr("43422"), Email: gdao.Ptr("rose@email.com")},
    }

    affected, err := UserDao.Exec(gdao.ExecReq[User]{
        Entities:       users,
        LastInsertIdAs: gdao.LAST_INSERT_ID_AS_LAST_ID,
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("INSERT INTO user(").WriteColumns().Write(") VALUES")
            b.EachEntity(b.Sep(","), func(_, _ int, entity *User) {
                cvs := b.ColumnValuesAt(entity, false)
                b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(_ int, _ string, value any) {
                    b.Write("?").Arg(value)
                })
            })
        }})

    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(affected)
        fmt.Println(*users[0].Id)
        fmt.Println(*users[1].Id)
    }
}
```
## RowAs参数

`Query`方法提供了将查询结果作为自增ID的模式，用于不支持`sql.Result#LastInsertId`方法的数据库驱动获取自增ID。`Query`方法的`RowAs`参数提供了以下模式，指定模式后，`Query`方法不再返回查询结果，而是将查询结果注入到`Entities`参数中。

| 可选值                     | 说明                                           |
|-------------------------|----------------------------------------------|
| `gdao.ROW_AS_RETURNING` | 将查询结果作为每个实体的自增ID，适配此模式的典型数据库：**PostgreSQL**。 |
| `gdao.ROW_AS_LAST_ID`   | 将查询结果作为最后一个自增ID，适配此模式的典型数据库：**SQL Server**。  |

*Example（PostgreSQL驱动）*

```go
func foo() {
    users := []*User{
        {Name: gdao.Ptr("Jack"), Phone: gdao.Ptr("12345"), Email: gdao.Ptr("jack@email.com")},
        {Name: gdao.Ptr("Nick"), Phone: gdao.Ptr("43422"), Email: gdao.Ptr("nick@email.com")},
    }

    _, _, err := UserDao.Query(gdao.QueryReq[User]{
        Entities: users,
        RowAs:    gdao.ROW_AS_RETURNING,
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("INSERT INTO user(").WriteColumns().Write(") VALUES")
            b.EachEntity(b.Sep(","), func(_, _ int, entity *User) {
                cvs := b.ColumnValuesAt(entity, false)
                b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(_ string, value any) {
                    b.Write("?").Arg(value)
                })
            })
            b.Write(" RETURNING id")
        }})

    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(*users[0].Id)
        fmt.Println(*users[1].Id)
    }
}
```

*Example（SQL Server驱动）*

```go
func foo() {
    users := []*User{
        {Name: gdao.Ptr("Jack"), Phone: gdao.Ptr("12345"), Email: gdao.Ptr("jack@email.com")},
        {Name: gdao.Ptr("Nick"), Phone: gdao.Ptr("43422"), Email: gdao.Ptr("rose@email.com")},
    }

    _, _, err := UserDao.Query(gdao.QueryReq[User]{
        Entities: users,
        RowAs:    gdao.ROW_AS_LAST_ID,
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("INSERT INTO user (").WriteColumns().Write(") VALUES")
            b.EachEntity(b.Sep(","), func(_, _ int, entity *User) {
                cvs := b.ColumnValuesAt(entity, false)
                b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(_ string, value any) {
                    b.Write("?").Arg(value)
                })
            })
            b.Write("; select ID = convert(bigint, SCOPE_IDENTITY())")
        }})

    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(*users[0].Id)
        fmt.Println(*users[1].Id)
    }
}
```

# CountDao

`gdao.CountDao`专门用于聚合函数`count`的查询，它会将查询结果映射到`gdao.Count`结构体并且零值可用。SELECT语句字段列表只有一个字段时会自动映射，如果有多个字段则映射名称为“count”的字段。

*Example（MySQL驱动）*

```go
var CountDao = _CountDao{gdao.NewCountDao(gdao.NewCountDaoReq{})}

type _CountDao struct {
    *gdao.CountDao
}

func (d _CountDao) ExistUser(id string) (bool, error) {
    count, _, err := d.Count(gdao.CountReq{BuildSql: func(b *gdao.CountBuilder) {
        b.Write("SELECT count(*) FROM user WHERE id=?", id)
    }})
    return count.Bool(), err
}
```

# 动态构建SQL

`Query`和`Exec`方法具有的`Entities`和`BuildeSql`参数用于动态构建SQL。

`BuildeSql`是一个函数，其唯一参数`b *gdao.Bulider`用于拼接SQL和设置参数，并提供了许多动态构建SQL的方法，`Entities`将作为某些方法的数据来源。

*Example（MySQL驱动）*

```go
var UserDao = _UserDao{gdao.NewDao[User](gdao.NewDaoReq{})}

type _UserDao struct {
    *gdao.Dao[User]
}

// InsertBatch 批量插入数据
func (d _UserDao) InsertBatch(entities []*User) (int64, error) {
    affected, err := d.Exec(gdao.ExecReq[User]{
        Entities: entities,
        LastInsertIdAs: gdao.LAST_INSERT_ID_AS_FIRST_ID,
        BuildSql: func(b *gdao.Builder[User]) {
            // 如果请求参数是空的，则通过b.SetOk(false)设置为无SQL可执行，然后return
            if len(entities) == 0 {
                b.SetOk(false)
                return
            }
            // 开始拼接INSERT语句
            b.Write("INSERT INTO user(").WriteColumns().Write(") VALUES")
            // 遍历Entities的每个实体
            b.EachEntity(b.Sep(","), func(_, _ int, entity *User) {
                // 获取实体的“列名称-字段值“键值对列表
                cvs := b.ColumnValues(false)
                // 遍历这些键值对，指定开始、分隔和结束符号
                b.EachColumnValues(cvs, b.SepFix("(", ",", ")", false), func(columnName string, value any) {
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
        }})
    if err != nil {
        return 0, err
    }
    return affected, nil
}
```

## Builder的方法

| 方法             | 说明                                                                |
|----------------|-------------------------------------------------------------------|
| `Write`        | 拼接字符串并设置参数。                                                       |
| `WriteColumns` | 拼接列名称，使用逗号分隔，如果参数为空则拼接表的所有列名称。                                    |
| `SetArgs`      | 设置参数。                                                             |
| `Columns`      | 返回所有列名称，`onlyAssigned`参数指定是否过滤掉值为nil的字段，`ignoredColumns`参数指定忽略字段。 |
| `AutoColumns`  | 返回标签值有`gdao="auto"`的字段。                                           |
| `EntityAt`     | 返回`Entities`中指定索引的实体。                                             |
| `Entity`       | 相当于`EntityAt(0)`                                                  |
| `ColumnValue`  | 返回首个实体中指定列名称对应字段的值。                                               |
| `EachEntity`   | 遍历`Entities`，自动过滤nil元素，`handle`函数参数`n`为调用次数，从1开始。                 |
| `EachColumn`   | 遍历指定实体的“列名称-字段值”列表，`handle`函数参数`n`为调用次数。。                         |
| `Repeat`       | 循环指定次数，`handle`函数参数`n`为调用次数，从1开始，`i`为循环次数。                        |
| `Sep`          | 在“Each”开头的方法和`Repeat`方法中使用，拼接指定分隔符号                               |
| `SepFix`       | 在“Each”开头的方法和`Repeat`方法中使用，拼接指定开始、分隔和结束符号，可指定无元素时是否拼接开始、结束符号。     |
| `Pp`           | 返回带编号的占位符，编号从1开始，每次调用后递增1，适用于PostgreSQL、Oracle等驱动。                |
| `Sql`          | 返回拼接的字符串。                                                         |
| `Args`         | 返回所有设置的参数。                                                        |
| `SetError`     | 设置error，SQL将不执行，error将从执行方法（`Query`、`Exec`）的返回值返回。                |
| `Error`        | 返回已设置的error。                                                      |
| `SetOk`        | 设置SQL是否可执行，不设置默认为true，若已设置error此方法无效。                             |
| `Ok`           | 返回SQL是否可执行，若已设置error此方法返回false。                                   |

# 事务

## SetTx函数

`gdao.SetTx`函数可将`*sql.Tx`变量附加到`context.Context`变量中，调用`Query`和`Exec`方法时将`context.Context`变量作为`Ctx`参数，将会自动使用该`*sql.Tx`变量执行SQL。

*Example*

```go
func foo(ctx context.Context) error {
    tx, err := demo.UserDao.DB().Begin()
    if err != nil {
        return err
    }
    
    ctx = gdao.SetTx(ctx, tx)

    _, err = UserDao.Exec(gdao.ExecReq[User]{
        Ctx: ctx,
        BuildSql: func(b *gdao.Builder[User]) {
            b.Write("UPDATE user SET status=-1 WHERE user_id=?", 1)
        }})
    if err != nil {
        tx.Rollback()
        return err
    }

    _, err = AccountDao.Exec(gdao.ExecReq[Account]{
        Ctx: ctx,
        BuildSql: func(b *gdao.Builder[Account]) {
            b.Write("UPDATE account SET status=-1 WHERE user_id=?", 1)
        }})
    if err != nil {
        tx.Rollback()
        return err
    }

    tx.Commit()
}
```

## Tx函数

`gdao.Tx`函数用于便捷化开启事务。它的`do`函数参数中的`ctx`参数会自动设置`*sql.Tx`变量，从而保证SQL的执行处于事务中。`*sql.Tx`的创建逻辑是：优先使用`ctx`参数已有的`*sql.Tx`变量，若没有则使用`db`参数创建，若`db`参数为nil则使用`gdao.DEFAULT_DB`创建，如若创建失败会返回错误。

*参数*

| 字段                                               | 说明      |
|--------------------------------------------------|---------|
| `ctx context.Context`                            | Context |
| `do func(ctx context.Context, tx *sql.Tx) error` | 事务执行内容  |
| `opts gdao.TxOption`                             | 选项      |

*Example*

```go
func foo(c context.Context) {
    gdao.Tx(c, func(ctx context.Context) error {
        _, err := UserDao.Exec(gdao.ExecReq[User]{
            Ctx: ctx,
            BuildSql: func(b *gdao.Builder[User]) {
                b.Write("UPDATE user SET status=-1 WHERE user_id=?", 1)
            }})
        if err != nil {
            return err
        }

        _, err = AccountDao.Exec(gdao.ExecReq[Account]{
            Ctx: ctx,
            BuildSql: func(b *gdao.Builder[Account]) {
                b.Write("UPDATE account SET status=-1 WHERE user_id=?", 1)
            }})
        if err != nil {
            return err
        }
        return nil
    })
}
```

### 选项

#### WithDefaultTx

指定默认的`*sql.DB`或`*sql.TxOptions`开启事务，若不指定`*sql.DB`默认使用`gao.DEFAULT_DB`开启事务。

# 日志

通过`gdao.LogConf`函数配置日志。

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
            <td><code>printSqlLevel string</code></td>
            <td>打印SQL的日志级别，可选值："develop"、"info"。SQL执行失败会打印error级别日志，不受此配置影响。</td>
        </tr>
        <tr>
            <td><code>compressSql bool</code></td>
            <td>是否压缩SQL。</td>
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
    gen.GetGenerator(gen.Cfg{
        DbType:  gen.DB_MYSQL,
        Dsn:     "(dsn)",
        OutPath: "demo", // 生成文件相对路径，绝对路径为"os.Getwd()/OutPath"。
        TableCfg: gen.TableCfg{
            Tables: gen.Tables{"user", "account"},
        },
        DaoCfg: gen.DaoCfg{
            GenDao: true, // 是否生成DAO，false只生成实体。
        },
    }).Gen()
}
```

## 扩展生成代码

上文的例子生成的文件如下，这些文件都可以手动编辑扩展。**代码生成器再次执行时，只会覆盖实体文件，不会覆盖DAO文件**。

```
.
└─ demo
   ├─ base_dao.go    // 基础DAO文件
   ├─ account.go     // 实体文件
   ├─ account_dao.go // 实体DAO文件
   ├─ user.go        // 实体文件
   └─ user_dao.go    // 实体DAO文件
```

*扩展user_dao.go文件*

```go
// Code generated by https://github.com/jishaocong0910/gdao. YOU CAN EDIT FOR MORE.

package main

import (
    "context"
    "github.com/jishaocong0910/gdao"
)

var UserDao = _UserDao{newBaseDao[User](gdao.NewDaoReq{}, "user")}

type _UserDao struct {
    *baseDao[User]
}

// 扩展了一个查询方法
func (d _UserDao) QueryByStatus(ctx context.Context, status ...int) ([]*User, error) {
    _, list, err := d.Query(gdao.QueryReq[User]{
        Ctx: ctx,
        BuildSql: func(b *gdao.Builder[User]) {
            if len(status) == 0 {
                b.SetOk(false) //设置为false表示不执行SQL
                return
            }
            b.Write("SELECT * FROM user WHERE status IN")
            b.Repeat(len(status), b.SepFix("(", ",", ")", false), nil, func(n, i int) {
                b.Write(",", status[i])
            })
        },
    })
    return list, err
}
```

## 基础DAO

实体DAO内嵌了基础DAO结构体`baseDao`，提供了常用的单表操作能力，基础DAO不建议二次编辑。

*Example*

```go
package main

import (
    "github.com/jishaocong0910/gdao"
    "gdao-demo/demo"
)

func main() {
    // 将执行SQL如下：
    // UPDATE user SET email = 'some@email.com', status = 2 WHERE id = 1
    demo.UserDao.Update(demo.UpdateReq[demo.User]{
        Entity: &demo.User{
            Id:     gdao.Ptr[int32](1),
            Email:  gdao.Ptr("some@email.com"),
            Status: gdao.Ptr[int8](2),
        },
        WhereColumns: []string{"id"},
    })
}
```

### 内置方法

| 方法名称          | 说明     |
|---------------|--------|
| `Get`         | 查询单个记录 |
| `List`        | 查询记录列表 |
| `Insert`      | 插入单个记录 |
| `InsertBatch` | 批量插入记录 |
| `Update`      | 更新单个记录 |
| `UpdateBatch` | 批量更新记录 |
| `Delete`      | 删除记录   |




