# GDAO

GDAO是用于Golang的轻量级ORM框架，具有下列主要特色。

1. 查询数据并映射为实体
2. 插入数据后将自动生成key映射到实体
3. 兼容不支持`sql.Result#LastInsertId`方法的驱动获取自动生成key
4. 内置便利的自定义SQL工具

兼容所有数据库是GDAO的原则，它通过自定义SQL来执行，因此兼容所有数据库的方言和驱动的参数占位符。不同数据库驱动获取自动生成key的方式并不统一，GDAO对此做了兼容性设计。

[![Go Reference](https://pkg.go.dev/badge/github.com/jishaocong0910/gdao.svg)](https://pkg.go.dev/github.com/jishaocong0910/gdao)
[![Go Report Card](https://goreportcard.com/badge/github.com/jishaocong0910/gdao)](https://goreportcard.com/report/github.com/jishaocong0910/gdao)
![coverage](https://raw.githubusercontent.com/jishaocong0910/gdao/badges/.badges/main/coverage.svg)

reference : http://go-database-sql.org/prepared.html

> ### Parameter Placeholder Syntax
>
> The syntax for placeholder parameters in prepared statements is
> database-specific. For example, comparing MySQL, PostgreSQL, and Oracle:
>
>       MySQL               PostgreSQL            Oracle
>       =====               ==========            ======
>       WHERE col = ?       WHERE col = $1        WHERE col = :col
>       VALUES(?, ?, ?)     VALUES($1, $2, $3)    VALUES(:val1, :val2, :val3)

# 安装

```shell
go get github.com/jishaocong0910/gdao
```

# 用法与例子

*以MySQL驱动为例*

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
    db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/my_test?charset=utf8mb4,utf8&parseTime=True&loc=Local")
    if err != nil {
        log.Fatalln(err)
    }

    // create a dao
    userDao := gdao.MustNewDao[User](gdao.NewDaoReq{Db: db, Table: "user", ColumnMapper: gdao.NewNameMapper().LowerSnakeCase()})

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
        BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
            b.Write("INSERT INTO ").Write(b.Table()).Write("(").Write(b.Columns()).Write(") VALUES (")
            b.EachColumn(b.Separate("", ",", ""),
                func(i int, column string, value any) {
                    b.Write("?").AddArgs(value)
                })
            b.Write(")")
            return b.String(), b.Args()
        },
    }).Insert()
    if err != nil {
        log.Fatalln(err)
    }
    fmt.Println(affected, *u.Id) // auto increment key

    // query
    u2, _, err := userDao.Query(gdao.QueryReq[User]{BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        return "SELECT " + b.Columns() + " FROM " + b.Table() + " WHERE id=?", []any{u.Id}
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
        BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
            b.Write("UPDATE ").Write(b.Table()).Write(" SET ")
            b.EachAssignedColumn(b.Separate("", ",", ""), func(i int, column string, value any) {
                b.Write(column).Write("=?").AddArgs(value)
            })
            b.Write(" WHERE id=?").AddArgs(u2.Id)
            return b.String(), b.Args()
        },
    })

    // delete
    userDao.Mutation(gdao.MutationReq[User]{
        BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
            return "DELETE FROM " + b.Table() + " WHERE id=?", []any{u.Id}
        },
    })
}
```

# 实体声明

实体字段类型只支持如下类型的**指针**和切片，可多维切片（为了支持PostgreSQL）。

`bool` `string` `time.Time` `float32` `float64`

`int` `int8` `int16` `int32` `int64`

`uint` `uint8` `uint16` `uint32` `uint64`

字段标签为`gdao="<values>"`，`<values>`有如下选项，多个时使用`;`拼接。

| 标签值                    | 说明                                                                                                                                                                                |
|------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `column=<column_name>` | `<column_name> ::= 数据库字段名`<br/>指定对应的数据库字段。                                                                                                                                        |
| `auto[=<offset>]`      | `<offset> ::= 自增偏移量，可选，默认为1。`<br/>标记自增key字段，执行INSERT语句后，会将`sql.Result#LastInsertId`方法的值映射到该字段。因此只对支持`sql.Result#LastInsertId`方法的数据库驱动有效，例如MySQL、SQLite等，不支持的例如Oracle、PostgreSQL等。 |

*Example:*

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

## 代码生成器

内置简单的实体代码生成器。

*支持的数据库*

| 数据库        | 是否支持  |
|------------|-------|
| MySQL      | ✅支持   |
| PostgreSQL | ✅支持   |
| Oracle     | ✅支持   |
| SQLserver  | 🚧开发中 |
| SQLite     | 🚧开发中 |

*Example:*

```go
package main

import (
    "github.com/jishaocong0910/gdao"
    "github.com/jishaocong0910/gdao/gen"
)

func main() {
    gen.Generator(gen.Config{
        DbType:  gen.DbTypes.MYSQL,                                                                         // 数据库类型
        Dsn:     "root:12345678@tcp(localhost:3306)/my_test?charset=utf8mb4,utf8&parseTime=True&loc=Local", // 格式与数据库类型对应
        OutPath: "dao/gen/entity",                                                                          // 输出目录为：os.Getwd()/OutPath                                                                            // 生成路径为os.Getwd()+OutPath
        Package: "entity",                                                                                  // 指定包名，否则包名为生成路径的最后一个路径
        Tables: gen.Tables{ // key对应生成的表，value为指定字段映射的Go类型，无指定则按默认处理。
            "user": nil,
            "account": {
                "level": "*int8", // 只支持指针和切片，填写其他的不影响生成，但GDAO不会识别。
            },
        },
        FileNameMapper:   gdao.NewNameMapper().LowerSnakeCase(), // 指定表名->文件名的映射，默认为小写下划线格式
        EntityNameMapper: nil,                                   // 指定表名->实体名的映射，默认为大驼峰格式
        FieldNameMapper:  nil,                                   // 指定表字段名->实体字段名的映射，默认为大驼峰格式
    }).Gen()
}

```



# DAO声明

`gdao.NewDao`和`gdao.MustNewDao`函数用于创建指定实体的DAO。

| 参数                         | 说明                                                                                                                       |
|----------------------------|--------------------------------------------------------------------------------------------------------------------------|
| `Db *sql.DB`               | 必填，打开的`*sql.DB`变量。                                                                                                       |
| `Table string`             | 必填，对应的数据库表名。                                                                                                             |
| `ColumnMapper *NameMapper` | 默认的 实体->数据库 字段映射规则，若实体字段没有添加标签`gdao:"column=<column_name>"`，则使用此规则，通过`gdao.NewNameMapper`函数创建映射器，并指定映射方法，可链式调用指定多个按顺序处理。 |
| `ColumnCaseSensitive bool` | 数据库字段是否大小写敏感。                                                                                                            |

*Example:*

```go
type User struct {
    IdCol    *int64
    NameCol  *string
    PhoneCol *int8 `gdao:"column=mobile"`
}

// 映射数据库字段名为：id、name、mobile
userDao := gdao.MustNewDao[User](gdao.NewDaoReq{
    Db:           db,
    Table:        "user",
    ColumnMapper: gdao.NewNameMapper().LowerCamelCase().SubSuffix("_col"), // 转化小写下划线格式并去除后缀
})
```

## 推荐代码风格

推荐每个实体具有独立的DAO，通过内嵌`gdao.Dao`自定义所需查询。

*Example（MySQL驱动）:*

```go
var UserRepository = UserDao{gdao.MustNewDao[User](gdao.NewDaoReq{Db: db, Table: "user"})}
var AccountRepository = AccountDao{gdao.MustNewDao[Account](gdao.NewDaoReq{Db: db, Table: "account"})}

type UserDao struct {
    gdao.Dao[User]
}

func (d UserDao) GetById(id int32) *User {
    first, _, err := d.Query(gdao.QueryReq[User]{BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        return "SELECT " + b.Columns() + " FROM " + b.Table() + " WHERE id=?", []any{id}
    }})
    if err != nil {
        panic(err)
    }
    return first
}

func (d UserDao) UpdateStatus(id int32, status int8) int64 {
    affected, err := d.Mutation(gdao.MutationReq[User]{BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        return "UPDATE " + b.Table() + " SET status=? WHERE id=?", []any{status, id}
    }}).Exec()
    if err != nil {
        panic(err)
    }
    return affected
}

type AccountDao struct {
    gdao.Dao[Account]
}

func (d AccountDao) GetByUserId(userId int32) *Account {
    first, _, err := d.Query(gdao.QueryReq[Account]{BuildSql: func(b gdao.Builder[Account]) (sql string, args []any) {
        return "SELECT " + b.Columns() + " FROM " + b.Table() + " WHERE user_id=?", []any{userId}
    }})
    if err != nil {
        panic(err)
    }
    return first
}

func (d AccountDao) ReduceBalance(id int32, balance int64) bool {
    a, _, err := d.Query(gdao.QueryReq[Account]{BuildSql: func(b gdao.Builder[Account]) (sql string, args []any) {
        return "SELECT balance FROM " + b.Table() + " WHERE id=?", []any{balance, id}
    }})
    if err != nil {
        panic(err)
    }

    oldBalance := *a.Balance
    newBalance := oldBalance - balance
    if newBalance < 0 {
        return false
    }

    affected, err := d.Mutation(gdao.MutationReq[Account]{BuildSql: func(b gdao.Builder[Account]) (sql string, args []any) {
        return "UPDATE " + b.Table() + " SET balance=? WHERE id=? AND balance=?", []any{newBalance, id, oldBalance}
    }}).Exec()
    if err != nil {
        panic(err)
    }
    return affected > 0
}
```

# DAO执行方法

## RawQuery

执行查询并返回`*sql.Rows`值，由用户自行映射返回的数据。

*参数*

| 字段                    | 说明      |
|-----------------------|---------|
| `Ctx context.Context` | Context |
| `Tx *sql.Tx`          | 事务      |
| `Sql  string`         | SQL语句   |
| `Args []any`          | SQL参数   |

*Example（MySQL驱动）:*

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

*Example（MySQL驱动）:*

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

| 字段                                                          | 说明                     |
|-------------------------------------------------------------|------------------------|
| `Ctx context.Context`                                       | Context                |
| `Tx *sql.Tx`                                                | 事务                     |
| `Entities []*T`                                             | 实体参数，用于自定义SQL。         |
| `BuildSql func(b gdao.Builder[T]) (sql string, args []any)` | 自定义SQL函数，返回SQL和占位符对应参数 |

*Example（MySQL驱动）:*

```go
user := &User{Status: gdao.Ptr[int8](1), Level: gdao.Ptr[int32](2)}

_, list, err := userDao.Query(gdao.QueryReq[User]{
    Entities: []*User{user},
    BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        b.Write("SELECT ").Write(b.Columns()).
            Write(" FROM ").Write(b.Table()).
            Write(" WHERE ")
        b.EachAssignedColumn(b.Separate("", ",", ""),
            func(i int, column string, value any) {
                b.Write(column).Write("=?").AddArgs(value)
            })
        return b.String(), b.Args()
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

| 字段                                                          | 说明                                                                               |
|-------------------------------------------------------------|----------------------------------------------------------------------------------|
| `Ctx context.Context`                                       | Context                                                                          |
| `Tx *sql.Tx`                                                | 事务                                                                               |
| `Entities []*T`                                             | 实体参数，有两个作用：<br/>1. 用于自定义SQL<br/>2. 使用`Insert`或`Query`执行模式，会将自动生成key或返回结果映射回这些实体。 |
| `BuildSql func(b gdao.Builder[T]) (sql string, args []any)` | 自定义SQL函数，返回SQL和占位符对应参数                                                           |

### Exec模式

内部调用Golang的`sql.Stmt#ExecContext`方法执行，返回影响行数。**适合用于执行一般的UPDATE和DELETE语句**。

*Example（MySQL驱动）:*

```go
affected, err := userDao.Mutation(gdao.MutationReq[User]{
    BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        b.Write("UPDATE ").Write(b.Table()).Write(" SET level=2 WHERE id=?").AddArgs(1)
        return b.String(), b.Args()
    }}).Exec()
if err != nil {
    panic(err)
}
fmt.Println(affected)
```

### Insert模式

与Exec模式相同，但多了一个步骤：调用`sql.Result#LastInsertId`方法将自增key映射到实体参数中。**适合用于支持`sql.Result#LastInsertId`方法的数据库驱动执行INSERT语句，如果数据库驱动不支持，则映射自增key将返回error或不准确**。

*Example（MySQL驱动）:*

```go
users := []*User{
    {Name: gdao.Ptr("foo"), Age: gdao.Ptr(16), Phone: gdao.Ptr("12345")},
    {Name: gdao.Ptr("var"), Age: gdao.Ptr(22), Phone: gdao.Ptr("56789")},
}

affected, err := userDao.Mutation(gdao.MutationReq[User]{
    Entities: users,
    BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        b.Write("INSERT INTO ").Write(b.Table()).
            Write("(").Write(b.Columns()).Write(") VALUES")
        b.EachEntity(b.Separate("", ",", ""), func(i int) {
            b.EachColumnAt(i, b.Separate("(", ",", ")"), func(i int, column string, value any) {
                b.Write("?").AddArgs(value)
            })
        })
        return b.String(), b.Args()
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

*Example（PostgreSQL驱动）:*

```go
users := []*User{
    {Name: gdao.Ptr("foo"), Age: gdao.Ptr[int32](16), Phone: gdao.Ptr("12345")},
    {Name: gdao.Ptr("var"), Age: gdao.Ptr[int32](22), Phone: gdao.Ptr("56789")},
}

affected, err := userDao.Mutation(gdao.MutationReq[User]{
    Entities: users,
    BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        // PostgreSQL使用 INSERT...RETURNING... 语法来返回新增数据的自动生成key。
        b.Write("INSERT INTO ").Write(b.Table()).
            Write("(").Write(b.Columns()).Write(") VALUES")
        b.EachEntity(b.Separate("(", ",", ")"), func(i int) {
            b.EachColumnAt(i, b.Separate("", ",", ""), func(i int, column string, value any) {
                b.Write(b.ArgN("$")).AddArgs(value)
            })
        })
        b.Write(" RETURNING id")
        return b.String(), b.Args()
    }}).Query()

if err != nil {
    panic(err)
}

fmt.Println(affected)
fmt.Println(*users[0].Id)
fmt.Println(*users[1].Id)
```

*Example（SQLserver驱动）:*

```go
users := []*User{
    {Name: gdao.Ptr("foo"), Age: gdao.Ptr[int32](16), Phone: gdao.Ptr("12345")},
    {Name: gdao.Ptr("var"), Age: gdao.Ptr[int32](22), Phone: gdao.Ptr("56789")},
}

affected, err := userDao.Mutation(gdao.MutationReq[User]{
    Entities: users,
    BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
        // SQLserver驱动官方的返回自增key方式
        // reference: https://github.com/denisenkom/go-mssqldb/blob/master/lastinsertid_example_test.go
        b.Write("INSERT INTO ").Write(b.Table()).
            Write("(").Write(b.Columns()).Write(") VALUES")
        b.EachEntity(b.Separate("(", ",", ")"), func(i int) {
            b.EachColumnAt(i, b.Separate("", ",", ""), func(i int, column string, value any) {
                b.Write(b.ArgN(":")).AddArgs(value)
            })
        })
        b.Write(";")
        b.Write("select ID = convert(bigint, SCOPE_IDENTITY())")
        return b.String(), b.Args()
    }}).Query()

if err != nil {
    panic(err)
}

fmt.Println(affected)
fmt.Println(*users[0].Id)
fmt.Println(*users[1].Id)
```

# 自定义SQL工具

DAO的`Query`和`Mutation`方法具有的`BuildeSql`和`Entities`参数用于自定义SQL。

`BuildeSql`是一个函数，要求用户返回SQL和占位符对应参数。函数的`gdao.Builder`参数是自定义SQL的工具，它提供了许多方便自定义SQL的方法。

`Entities`将作为`gdao.Builder`某些方法的数据来源。

*gdao.Builder的方法*

| 方法                   | 说明                                                 |
|----------------------|----------------------------------------------------|
| `Table`              | 返回表名称                                              |
| `Columns`            | 返回所有数据库字段，以`,`拼接                                   |
| `Write`              | 拼接字符串                                              |
| `AddArgs`            | 添加占位符对应参数                                          |
| `ArgN`               | 返回带编号的占位符，编号从1开始，每次调用后递增1，适和用于PostgreSQL、Oracle等驱动 |
| `Entity`             | 返回`Entities`中第一个实体值。                               |
| `EntityAt`           | 返回`Entities`中指定索引的实体值                              |
| `EachEntity`         | 遍历`Entities`，其`handle`函数参数的`i`参数为实体参数索引            |
| `EachColumn`         | 遍历`Entities`中第一个实体的所有字段                            |
| `EachColumnAt`       | 遍历`Entities`中指定索引的实体的所有字段                          |
| `EachAssignedColumn` | 遍历`Entities`中第一个实体的所有不为nil的字段                      |
| `Separate`           | 用于所有Each开头的方法，指定开始、结束和分隔符号                         |
| `String`             | 返回最终拼接的字符串                                         |
| `Args`               | 返回所有占位符对应参数                                        |

*Example*

```go
users := []*User{
    {
        Name:    gdao.Ptr("foo"),
        Status:  gdao.Ptr[int8](3),
        Level:   gdao.Ptr[int32](10),
        Address: gdao.Ptr("home"),
        Phone:   gdao.Ptr("56789"),
    },
    {
        Name:    gdao.Ptr("bar"),
        Status:  gdao.Ptr[int8](2),
        Level:   gdao.Ptr[int32](2),
        Address: gdao.Ptr("addr"),
        Phone:   gdao.Ptr("2325325"),
    },
}
userDao.Mutation(gdao.MutationReq[User]{Entities: users, BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
    // 为了节省演示代码篇幅，以下每个代码块的注释为独立运行结果。
    {
        b.Write("Table: ").Write(b.Table()).Write(", Columns: ").Write(b.Columns())
        fmt.Println(b.String())
        // Output:
        // Table: user, Columns: id,name,age,address,phone,email,status,level,create_at
    }
    {
        // 演示PostgreSQL驱动的参数占位符。
        b.Write("UPDATE ").Write(b.Table()).Write(" SET").
            Write(" status=").Write(b.ArgN("$")).Write(",").
            Write(" level=").Write(b.ArgN("$")).
            Write(" WHERE id=").Write(b.ArgN("$")).AddArgs("3,10,1001")
        fmt.Println(b.String())
        fmt.Println(b.Args())
        // Output:
        // UPDATE user SET status=$1, level=$2 WHERE id=$3
        // [3 10 1001]
    }
    {
        fmt.Println(*b.Entity().Name, *b.Entity().Level)
        fmt.Println(*b.EntityAt(1).Name, *b.EntityAt(1).Level)
        b.EachEntity(b.Separate("(", ", ", ")"), func(i int) {
            b.Write(*b.EntityAt(i).Name).Write(" ").Write(*b.EntityAt(i).Phone)
        })
        fmt.Println(b.String())
        // Output:
        // foo 10
        // bar 2
        // (foo 56789, bar 2325325)
    }
    {
        b.EachEntity(b.Separate("[ ", "; ", " ]"), func(i int) {
            b.EachColumnAt(i, b.Separate("(", ",", ")"), func(i int, column string, value any) {
                b.Write(column)
            })
        })
        fmt.Println(b.String())
        // Output:
        // [ (id,name,age,address,phone,email,status,level,create_at); (id,name,age,address,phone,email,status,level,create_at) ]
    }
    {
        b.EachColumn(b.Separate("", ",", ""), func(i int, column string, value any) {
            b.Write(column)
            if value == nil {
                b.AddArgs("nil")
            } else {
                b.AddArgs(reflect.ValueOf(value).Elem().Interface())
            }
        })
        fmt.Println(b.String())
        fmt.Println(b.Args())
        // Output:
        // id,name,age,address,phone,email,status,level,create_at
        // [<nil> foo <nil> home 56789 <nil> 3 10 <nil>]
    }
    {
        b.EachAssignedColumn(b.Separate("", ",", ""), func(i int, column string, value any) {
            b.Write(column).AddArgs(reflect.ValueOf(value).Elem().Interface())
        })
        fmt.Println(b.String(), b.Args())
        // Output:
        // name,address,phone,status,level
        // [foo home 56789 3 10]
    }
    return "", nil
}})
```

# 事务

使用事务非常简单，每个DAO执行方法参数都有一个`tx`字段，只需创建`*sql.Tx`变量传入即可。

*Example*

```go
tx, err := userDao.Db().Begin()
if err != nil {
    panic(err)
}
defer tx.Rollback()

_, err = userDao.Mutation(gdao.MutationReq[User]{Tx: tx, BuildSql: func(b gdao.Builder[User]) (sql string, args []any) {
    return "UPDATE user SET status=-1 WHERE user_id=?", []any{1}
}}).Exec()
if err != nil {
    panic(err)
}
_, err = accountDao.Mutation(gdao.MutationReq[Account]{Tx: tx, BuildSql: func(b gdao.Builder[Account]) (sql string, args []any) {
    return "UPDATE account SET status=-1 WHERE user_id=?", []any{1}
}})
if err != nil {
    panic(err)
}

tx.Commit()
```

# 日志

| 配置                          | 说明                         |
|-----------------------------|----------------------------|
| `gdao.Log.Logger`           | 设置日志器，日志器须实现`gdao.Logger`。 |
| `gdao.Log.PrintSqlLogLevel` | 指定打印SQL的日志级别。              |

