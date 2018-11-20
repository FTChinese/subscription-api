package model

import (
	"testing"
)

func TestFindMember(t *testing.T) {
	m, err := devEnv.FindMember("e1a1f5c0-0e23-11e8-aa75-977ba2bcc6ae")

	if err != nil {
		t.Error(err)
	}

	t.Log(m)
}
