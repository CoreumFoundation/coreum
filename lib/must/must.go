package must

import "net/http"

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

// Any panics if err is not ni
func Any(_ interface{}, err error) {
	OK(err)
}

// HTTPRequest panics if err is not nil, v is returned otherwise
func HTTPRequest(v *http.Request, err error) *http.Request {
	OK(err)
	return v
}
