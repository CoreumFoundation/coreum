package v3_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/CoreumFoundation/coreum/v2/app/apptesting"
	v3 "github.com/CoreumFoundation/coreum/v2/app/upgrade/v3"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestUpgrade() {
	s.Setup()

	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v3.Name, upgradeHeight)
}
