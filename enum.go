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

type LastInsertIdAs struct {
	*e.EnumElem__
}

type _LastInsertIdAs struct {
	*e.Enum__[LastInsertIdAs]
	FIRST_ID,
	LAST_ID LastInsertIdAs
}

var LastInsertIdAs_ = e.NewEnum[LastInsertIdAs](_LastInsertIdAs{})

type RowAs struct {
	*e.EnumElem__
}

type _RowAs struct {
	*e.Enum__[RowAs]
	RETURNING,
	LAST_ID RowAs
}

var RowAs_ = e.NewEnum[RowAs](_RowAs{})
