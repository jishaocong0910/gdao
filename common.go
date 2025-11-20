package gdao

func P[T any](t T) *T {
	return &t
}

func V[T any](t *T) T {
	var v T
	if t != nil {
		v = *t
	}
	return v
}

func checkMust(must bool, err error) { // coverage-ignore
	if must && err != nil {
		panic(err)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
