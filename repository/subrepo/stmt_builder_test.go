package subrepo

import (
	"testing"
)

func Test_buildSelectMembership(t *testing.T) {
	t.Logf("Sandbox locked: %+v", buildSelectMembership(true, true))
	t.Logf("Production locked: %+v", buildSelectMembership(false, true))
	t.Logf("Sandobx no lock: %+v", buildSelectMembership(true, false))
	t.Logf("Production no lock: %+v", buildSelectMembership(false, false))
}
