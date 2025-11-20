package pkg

type MyStruct2 struct{}

func (m *MyStruct2) GdaoValue() string {
	return ""
}

func (m *MyStruct2) GdaoField(value string) *MyStruct2 {
	return nil
}
