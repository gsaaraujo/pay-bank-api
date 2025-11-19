package utils

import "github.com/google/uuid"

func IsValidUUID(input string) bool {
	if err := uuid.Validate(input); err != nil {
		return false
	}

	return input != uuid.Nil.String()
}
