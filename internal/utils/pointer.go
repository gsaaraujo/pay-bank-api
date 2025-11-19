package utils

func NewPointer[T any](v T) *T {
	return &v
}
