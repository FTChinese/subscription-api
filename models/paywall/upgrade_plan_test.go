package paywall

import "testing"

func TestGenerateUpgradeID(t *testing.T) {
	t.Logf("Upgrade id: %s", GenerateUpgradeID())
}
