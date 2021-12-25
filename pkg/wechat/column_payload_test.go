package wechat

import (
	"encoding/json"
	"testing"
)

func TestColumnPayload_Marshal(t *testing.T) {
	p := map[string]string{
		"foo": "bar",
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s", b)
}
