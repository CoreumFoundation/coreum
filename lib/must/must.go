package must

// OK panics if err is not nil
func OK(err error) {
	if err != nil {
		panic(err)
	}
}

// String panics if err is not nil, v is returned otherwise
func String(v string, err error) string {
	OK(err)
	return v
}
