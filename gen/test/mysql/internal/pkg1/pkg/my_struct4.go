package pkg

type MyStruct4 struct {
}

func (m MyStruct4) GdaoValue() string { // coverage-ignore
	return ""
}

func (m MyStruct4) GdaoField(string) *MyStruct4 { // coverage-ignore
	return nil
}
