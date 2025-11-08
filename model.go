package gdao

import (
	"context"
	"database/sql"
)

type NewDaoReq struct {
	DB                *sql.DB
	AllowInvalidField bool
	ColumnMapper      *NameMapper
}

type NewCountDaoReq struct {
	DB *sql.DB
}

type QueryReq[T any] struct {
	Ctx         context.Context
	Must        bool
	SqlLogLevel SqlLogLevel
	Desc        string
	RowAs       RowAs
	Entities    []*T
	BuildSql    func(b *DaoSqlBuilder[T])
}

type ExecReq[T any] struct {
	Ctx            context.Context
	Must           bool
	SqlLogLevel    SqlLogLevel
	Desc           string
	LastInsertIdAs LastInsertIdAs
	Entities       []*T
	BuildSql       func(b *DaoSqlBuilder[T])
}

type CountReq struct {
	Ctx         context.Context
	Must        bool
	SqlLogLevel SqlLogLevel
	Desc        string
	BuildSql    func(b *CountBuilder)
}

type Count struct {
	Value *int64 `gdao:"column=count"`
}

func (c *Count) Int() int {
	if c == nil || c.Value == nil {
		return 0
	}
	return int(*c.Value)
}

func (c *Count) Int8() int8 {
	if c == nil || c.Value == nil {
		return 0
	}
	return int8(*c.Value)
}

func (c *Count) Int16() int16 {
	if c == nil || c.Value == nil {
		return 0
	}
	return int16(*c.Value)
}

func (c *Count) Int32() int32 {
	if c == nil || c.Value == nil {
		return 0
	}
	return int32(*c.Value)
}

func (c *Count) Int64() int64 {
	if c == nil || c.Value == nil {
		return 0
	}
	return *c.Value
}

func (c *Count) Bool() bool {
	if c == nil || c.Value == nil {
		return false
	}
	return *c.Value > 0
}

func (c *Count) IntPtr() *int {
	if c == nil {
		return nil
	}
	i := int(*c.Value)
	return &i
}

func (c *Count) Int8Ptr() *int8 {
	if c == nil {
		return nil
	}
	i := int8(*c.Value)
	return &i
}

func (c *Count) Int16Ptr() *int16 {
	if c == nil {
		return nil
	}
	i := int16(*c.Value)
	return &i
}

func (c *Count) Int32Ptr() *int32 {
	if c == nil {
		return nil
	}
	i := int32(*c.Value)
	return &i
}

func (c *Count) Int64Ptr() *int64 {
	if c == nil {
		return nil
	}
	return c.Value
}

func (c *Count) BoolPtr() *bool {
	if c == nil {
		return nil
	}
	b := *c.Value > 0
	return &b
}

type Cfg struct {
	DefaultDB      *sql.DB
	Logger         Logger
	SqlLogLevel    SqlLogLevel
	CompressSqlLog bool
}
