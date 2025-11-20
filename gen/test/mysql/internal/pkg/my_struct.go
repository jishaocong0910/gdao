package pkg

type MyStruct struct {
}

func (m MyStruct) GdaoValue() string { // coverage-ignore
	return ""
}

func (m MyStruct) GdaoField(string) *MyStruct { // coverage-ignore
	return nil
}
