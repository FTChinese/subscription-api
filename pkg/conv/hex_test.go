package conv

import (
	"encoding/json"
	"testing"
)

func TestDecodeHexString(t *testing.T) {
	inStr := `hello world`
	hash := NewMD5Sum(inStr).String()

	hb, err := DecodeHexString(hash)
	if err != nil {
		t.Error(err)
	}

	if hb.String() != `hello world` {
		t.Errorf("Execpted %s, got %s", inStr, hb.String())
	}

	t.Logf("Decoded hex: %b", hb)
}

func TestJSONUnmarshal(t *testing.T) {
	inStr := `hello world`
	hash := NewMD5Sum(inStr).String()

	jsonIn := []byte(`"` + hash + `"`)
	var target string
	err := json.Unmarshal(jsonIn, &target)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Parsed json: %s", target)
}

func TestHexBin_UnmarshalJSON(t *testing.T) {
	inStr := `hello world`
	expected := NewMD5Sum(inStr).String()
	t.Logf("Input hash %s", expected)

	jsonIn := []byte(`"` + expected + `"`)

	var got HexBin
	err := got.UnmarshalJSON(jsonIn)
	if err != nil {
		t.Error(err)
		return
	}

	if got.String() != expected {
		t.Errorf("expcted %s, got %s", expected, got)
	}

	t.Logf("Unmarshal result: %s", got)
}

func TestHexBin_MarshalJSON(t *testing.T) {
	inStr := `hello world`
	hash := NewMD5Sum(inStr).String()

	hb, err := DecodeHexString(hash)
	if err != nil {
		t.Error(err)
	}

	got, err := hb.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	expected := `"` + hash + `"`

	if string(got) != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}

	t.Logf("Marshal result: %s", got)
}
