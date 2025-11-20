package pkg

type MySlice []string

func (m MySlice) GdaoValue() string { // coverage-ignore
	return ""
}

func (m MySlice) GdaoField(value string) MySlice { // coverage-ignore
	return nil
}
