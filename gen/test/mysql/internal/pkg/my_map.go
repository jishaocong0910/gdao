package pkg

type MyMap map[string]any

func (m MyMap) GdaoValue() string { // coverage-ignore
	return ""
}

func (m MyMap) GdaoField(string) MyMap { // coverage-ignore
	return nil
}
