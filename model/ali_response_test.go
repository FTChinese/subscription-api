package model

import "testing"

func TestSaveAliNoti(t *testing.T) {
	n := aliNoti()

	err := devEnv.SaveAliNotification(n)
	if err != nil {
		t.Error(err)
	}
}
