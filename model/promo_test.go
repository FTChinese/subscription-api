package model

import "testing"

func TestGetPromo(t *testing.T) {
	p, err := devEnv.RetrievePromo()

	if err != nil {
		t.Error(err)
	}

	t.Log(p)
}
