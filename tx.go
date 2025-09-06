package gdao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	pkgErrors "github.com/pkg/errors"
)

var ctx_key_tx = Ptr("")

type TxOption func(*txOption)

type txOption struct {
	db   *sql.DB
	opts *sql.TxOptions
}

func Tx(ctx context.Context, do func(ctx context.Context) error, opts ...TxOption) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	o := &txOption{}
	for _, opt := range opts {
		opt(o)
	}

	tx := getTx(ctx)
	if tx == nil {
		var db *sql.DB
		if o.db == nil { // coverage-ignore
			db = DEFAULT_DB
		}
		if db == nil { // coverage-ignore
			return errors.New(`cannot begin a transaction, no available *sql.DB`)
		}
		tx, err = db.BeginTx(ctx, o.opts)
		if err != nil { // coverage-ignore
			return err
		}
		ctx = SetTx(ctx, tx)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else if r := recover(); r != nil {
			tx.Rollback()
			err = pkgErrors.WithStack(fmt.Errorf("%v", r))
		} else {
			tx.Commit()
		}
	}()
	err = do(ctx)
	return err
}

func SetTx(ctx context.Context, tx *sql.Tx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, ctx_key_tx, tx)
	return ctx
}

func WithDefaultTx(db *sql.DB, opts *sql.TxOptions) TxOption {
	return func(o *txOption) {
		o.db = db
		o.opts = opts
	}
}

func getTx(ctx context.Context) *sql.Tx {
	if ctx != nil {
		if tx, ok := ctx.Value(ctx_key_tx).(*sql.Tx); ok {
			return tx
		}
	}
	return nil
}
