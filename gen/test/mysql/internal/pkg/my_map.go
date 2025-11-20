package pkg

type MyMap map[string]any

func (m MyMap) GdaoValue() string {
	return ""
}

func (m MyMap) GdaoField(string) MyMap {
	return make(MyMap)
}
