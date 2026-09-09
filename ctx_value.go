// Copyright 2024-present jishaocong0910
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

import "context"

var (
	cvTx = ctxValue[txInfo]{key: "github.com/jishaocong0910/cozy-orm:tx"}
)

type ctxValue[T any] struct {
	key string
}

func (c ctxValue[T]) get(ctx context.Context) (t *T) {
	if ctx != nil {
		t, _ = ctx.Value(c.key).(*T)
	}
	return
}

func (c ctxValue[T]) set(parent context.Context, val *T) (ctx context.Context) {
	if parent != nil {
		ctx = context.WithValue(parent, c.key, val)
	}
	return
}

func (c ctxValue[T]) clean(ctx context.Context) {
	if v := c.get(ctx); v != nil {
		var t T
		*v = t
	}
}
