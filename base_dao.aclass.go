package gdao

import (
	"context"
	"database/sql"
	"errors"
)

type baseDao_ interface {
	baseDao_()
}

type baseDao__ struct {
	i  baseDao_
	db *sql.DB
}

func (this baseDao__) baseDao_() {}

func (this baseDao__) DB() *sql.DB {
	if this.db == nil { // coverage-ignore
		return global.DefaultDB
	}
	return this.db
}

func (this baseDao__) query(ctx context.Context, sql string, args []any) (rows *sql.Rows, columns []string, closeFunc func(), err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	prepare, err := this.createPrepare(ctx, sql)
	if err != nil { // coverage-ignore
		return nil, nil, nil, err
	}
	args = convertArgs(args)
	rows, err = prepare.QueryContext(ctx, args...)
	if err != nil { // coverage-ignore
		printWarn(ctx, prepare.Close())
		return nil, nil, nil, err
	}
	closeFunc = func() {
		printWarn(ctx, rows.Close())
		printWarn(ctx, prepare.Close())
	}
	columns, err = rows.Columns()
	if err != nil { // coverage-ignore
		closeFunc()
		return nil, nil, nil, err
	}
	return rows, columns, closeFunc, nil
}

func (this baseDao__) exec(ctx context.Context, sql string, args []any) (result sql.Result, affected int64, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	affected = int64(-1)
	prepare, err := this.createPrepare(ctx, sql)
	if err != nil { // coverage-ignore
		return nil, 0, err
	}
	defer func() {
		printWarn(ctx, prepare.Close())
	}()
	args = convertArgs(args)
	result, err = prepare.ExecContext(ctx, args...)
	if err != nil { // coverage-ignore
		return nil, 0, err
	}
	affected, err = result.RowsAffected()
	return
}

func (this baseDao__) createPrepare(ctx context.Context, _sql string) (*sql.Stmt, error) {
	if tx := getTx(ctx); tx != nil {
		return tx.PrepareContext(ctx, _sql)
	} else {
		db := this.DB()
		if db == nil { // coverage-ignore
			return nil, errors.New("no available *sql.DB variable")
		}
		return db.PrepareContext(ctx, _sql)
	}
}

func extendBaseDao(i baseDao_, db *sql.DB) *baseDao__ {
	return &baseDao__{i: i, db: db}
}
