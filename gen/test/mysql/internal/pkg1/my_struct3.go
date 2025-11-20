package pkg

type MyStruct3 struct {
}

func (m MyStruct3) GdaoValue() string { // coverage-ignore
	return ""
}

func (m MyStruct3) GdaoField(string) *MyStruct3 { // coverage-ignore
	return nil
}
