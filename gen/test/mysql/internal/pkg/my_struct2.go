package pkg

type MyStruct2 struct{}

func (m *MyStruct2) GdaoValue() string { // coverage-ignore
	return ""
}

func (m *MyStruct2) GdaoField(value string) *MyStruct2 { // coverage-ignore
	return nil
}
