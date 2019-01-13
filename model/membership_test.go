package model

import (
	"testing"
)

func TestMemberNotFound(t *testing.T) {
	m := NewMocker()

	subs, _ := m.CreateWxpaySubs()

	member, err := devEnv.findMember(subs)

	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v\n", member)
}
