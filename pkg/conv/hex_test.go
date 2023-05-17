package conv

import (
	"encoding/json"
	"testing"
)

func TestNewHexBin(t *testing.T) {
	hash := HashString(`hello world`)

	hb, err := NewHexBin(hash)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Hashed: %s", hb)
	t.Logf("Hashed bin: %b", hb)

	jsonOut, err := hb.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	t.Logf("JSON string: %s", jsonOut)

	jsonIn := []byte(`"` + hash + `"`)
	var target string
	err = json.Unmarshal(jsonIn, &target)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Parsed json: %s", target)

	var parsed HexBin
	err = parsed.UnmarshalJSON(jsonIn)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Parsed from json: %s", parsed)
	t.Logf("Pared as bin: %b", parsed)
}
