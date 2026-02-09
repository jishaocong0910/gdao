# cozy-orm

cozy-orm是用于Golang的轻量级ORM框架，用极少的API满足所有开发需求，具有全面的数据库驱动兼容方案：

- **SQL方言**。使用字符串而非SQL组装方法来构建SQL，最大限度兼容各种数据库方言，同时还提供了动态构建SQL方法。
- **参数占位符** (reference : http://go-database-sql.org/prepared.html )。使用字符串构建SQL，因此不需要关注具体是哪种数据库，用户使用对应数据库驱动的参数占位符即可。有些数据库驱动的参数占位符是动态的，cozy-orm也提供了参数占位符的动态构建方法。
- **获取自动生成ID**。有些数据库驱动支持`sql.Result#LastInsertId`方法来获取自动生成ID，有些不支持此方法而是其他方式，cozy-orm对此做了兼容性设计。


[![Go Reference](https://pkg.go.dev/badge/github.com/jishaocong0910/orm.svg)](https://pkg.go.dev/github.com/jishaocong0910/cozy-orm)
[![Go Report Card](https://goreportcard.com/badge/github.com/jishaocong0910/cozy-orm)](https://goreportcard.com/report/github.com/jishaocong0910/cozy-orm)
![coverage](https://raw.githubusercontent.com/jishaocong0910/cozy-orm/badges/.badges/main/coverage.svg)

# 安装

```shell
go get github.com/jishaocong0910/cozy-orm
```

# 文档

详见Wiki：https://github.com/jishaocong0910/cozy-orm/wiki




