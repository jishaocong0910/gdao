package pkg

type MyStruct3 struct {
}

func (m MyStruct3) GdaoValue() string {
	return ""
}

func (m MyStruct3) GdaoField(string) *MyStruct3 {
	return &MyStruct3{}
}
