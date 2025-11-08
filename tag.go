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
