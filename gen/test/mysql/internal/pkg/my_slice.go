package pkg

type MySlice []string

func (m MySlice) GdaoValue() string {
	return ""
}

func (m MySlice) GdaoField(value string) MySlice {
	return make(MySlice, 0)
}
