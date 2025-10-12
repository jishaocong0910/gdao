package gdao

import e "github.com/jishaocong0910/enum"

type SqlLogLevel struct {
	*e.EnumElem__
}

type _SqlLogLevel struct {
	*e.Enum__[SqlLogLevel]
	OFF,
	DEBUG,
	INFO SqlLogLevel
}

var SqlLogLevel_ = e.NewEnum[SqlLogLevel](_SqlLogLevel{})
