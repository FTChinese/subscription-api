package test

import (
	"io/ioutil"
	"testing"
)

func TestAppleLinkReq(t *testing.T) {
	req := AppleLinkReq()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", body)
}

func TestAppleUnlinkReq(t *testing.T) {
	req := AppleUnlinkReq()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s", body)
}
