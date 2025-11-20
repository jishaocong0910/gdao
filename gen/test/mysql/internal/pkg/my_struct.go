package pkg

type MyStruct struct {
}

func (m MyStruct) GdaoValue() string {
	return ""
}

func (m MyStruct) GdaoField(string) *MyStruct {
	return &MyStruct{}
}
