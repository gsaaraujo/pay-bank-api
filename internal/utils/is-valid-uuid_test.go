package utils_test

import (
	"testing"

	"github.com/gsaaraujo/pay-bank-api/internal/utils"
	"github.com/stretchr/testify/suite"
)

type IsValidUUIDSuite struct {
	suite.Suite
}

func (i *IsValidUUIDSuite) Test1() {
	i.Run("when uuid is valid, then return true", func() {
		i.True(utils.IsValidUUID("04d048fe-ca07-48bc-93a5-130440af41e0"))
	})
}

func (i *IsValidUUIDSuite) Test2() {
	i.Run("when uuid is invalid, then return false", func() {
		i.False(utils.IsValidUUID("00000000-0000-0000-0000-000000000000"))
	})
}

func TestIsValidUUID(t *testing.T) {
	suite.Run(t, new(IsValidUUIDSuite))
}
