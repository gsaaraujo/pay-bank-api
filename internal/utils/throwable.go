package utils

func GetOrThrow[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}

func ThrowOnError(err error) {
	if err != nil {
		panic(err)
	}
}
