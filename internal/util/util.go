package util

func AssertNoError(err error) {
	if err == nil {
		return
	}
	panic(err)
}
