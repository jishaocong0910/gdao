package pkg2

type MyStruct5 struct {
}

func (m MyStruct5) GdaoValue() string {
	return ""
}

func (m MyStruct5) GdaoField(string) *MyStruct5 {
	return &MyStruct5{}
}
