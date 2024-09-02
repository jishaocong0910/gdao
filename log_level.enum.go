package gdao

import o "github.com/jishaocong0910/go-object"

type logLevel struct {
	*o.M_EnumValue
}

var LogLevels = o.NewEnum[logLevel](struct {
	*o.M_Enum[logLevel]
	DEBUG, INFO logLevel
}{})
