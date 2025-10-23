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

package gdao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	pkgErrors "github.com/pkg/errors"
)

var ctx_key_tx = P("")

type TxOption func(*txOption)

type txOption struct {
	db   *sql.DB
	opts *sql.TxOptions
	must bool
}

func Tx(ctx context.Context, do func(ctx context.Context) error, opts ...TxOption) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	o := &txOption{}
	for _, opt := range opts { // coverage-ignore
		opt(o)
	}

	tx := getTx(ctx)
	if tx == nil {
		var db *sql.DB
		if o.db == nil { // coverage-ignore
			db = cfg.DefaultDB
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
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = pkgErrors.WithStack(fmt.Errorf("%v", r))
			}
		} else {
			tx.Commit()
		}
	}()
	err = do(ctx)
	checkMust(o.must, err)
	return err
}

func SetTx(ctx context.Context, tx *sql.Tx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, ctx_key_tx, tx)
	return ctx
}

func WithDefaultTx(db *sql.DB, opts *sql.TxOptions) TxOption { // coverage-ignore
	return func(o *txOption) {
		o.db = db
		o.opts = opts
	}
}

func WithMust() TxOption {
	return func(o *txOption) {
		o.must = true
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
