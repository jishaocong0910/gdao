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

package gdao

import (
	"reflect"
	"strconv"
	"strings"
)

type tag struct {
	column            string
	isAutoIncrement   bool
	autoIncrementStep int64
}

func parseTag(tf reflect.StructField) tag {
	t := tag{autoIncrementStep: 1}
	if gdaoTag, ok := tf.Tag.Lookup("gdao"); ok {
		params := strings.Split(gdaoTag, ";")
		for _, p := range params {
			kv := strings.Split(p, "=")
			if len(kv) == 1 {
				p = strings.TrimSpace(p)
				if p == "auto" {
					t.isAutoIncrement = true
				}
			}
			if len(kv) == 2 {
				k := strings.TrimSpace(kv[0])
				v := strings.TrimSpace(kv[1])
				switch k {
				case "column":
					t.column = v
				case "auto":
					t.isAutoIncrement = true
					i, err := strconv.ParseInt(v, 10, 64)
					if err == nil { // coverage-ignore
						t.autoIncrementStep = i
					}
				}
			}
		}
	}
	return t
}
