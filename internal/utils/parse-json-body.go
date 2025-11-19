package utils

import (
	"encoding/json"
	"io"
)

func ParseJSONBody[T any](body io.ReadCloser) T {
	var v T

	defer func() {
		ThrowOnError(body.Close())
	}()

	data := GetOrThrow(io.ReadAll(body))
	ThrowOnError(json.Unmarshal(data, &v))

	return v
}
