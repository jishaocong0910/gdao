package pkg

type MyStruct4 struct {
}

func (m MyStruct4) GdaoValue() string {
	return ""
}

func (m MyStruct4) GdaoField(string) *MyStruct4 {
	return &MyStruct4{}
}
