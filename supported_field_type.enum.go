package gdao

import o "github.com/jishaocong0910/go-object"

type supportedFieldType struct {
	*o.M_EnumValue
	Name string
}

type _supportedFieldType struct {
	*o.M_Enum[supportedFieldType]
	nameMap map[string]supportedFieldType
	BOOL, STRING, TIME, FLOAT32, FLOAT64,
	INT, INT8, INT16, INT32, INT64,
	UINT, UINT8, UINT16, UINT32, UINT64 supportedFieldType
}

func (this *_supportedFieldType) ContainsName(n string) bool {
	if _, ok := this.nameMap[n]; ok {
		return true
	}
	return false
}

var SupportedFieldTypes = func() _supportedFieldType {
	e := o.NewEnum[supportedFieldType](_supportedFieldType{
		BOOL:    supportedFieldType{Name: "bool"},
		STRING:  supportedFieldType{Name: "string"},
		TIME:    supportedFieldType{Name: "time.Time"},
		FLOAT32: supportedFieldType{Name: "float32"},
		FLOAT64: supportedFieldType{Name: "float64"},
		INT:     supportedFieldType{Name: "int"},
		INT8:    supportedFieldType{Name: "int8"},
		INT16:   supportedFieldType{Name: "int16"},
		INT32:   supportedFieldType{Name: "int32"},
		INT64:   supportedFieldType{Name: "int64"},
		UINT:    supportedFieldType{Name: "uint"},
		UINT8:   supportedFieldType{Name: "uint8"},
		UINT16:  supportedFieldType{Name: "uint16"},
		UINT32:  supportedFieldType{Name: "uint32"},
		UINT64:  supportedFieldType{Name: "uint64"},
	})
	e.nameMap = make(map[string]supportedFieldType, len(e.Values()))
	for _, value := range e.Values() {
		e.nameMap[value.Name] = value
	}
	return e
}()
