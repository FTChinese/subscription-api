package util

import "testing"

func TestDateMarshal(t *testing.T) {
	date := DateNow()

	b, err := date.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Stringify result: %s\n", b)
}

func TestDateMarshalZero(t *testing.T) {
	date := DateZero()

	b, err := date.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Stringify null: %s\n", b)
}

func TestDateUnmarshal(t *testing.T) {
	var date Date

	var data = []byte(`"2019-01-02"`)

	if err := date.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	t.Logf("Parsed: %+v\n", date)
}

func TestDateUnmarshalEmpty(t *testing.T) {
	var date Date

	var data = []byte(`""`)

	if err := date.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	t.Logf("arsed empty string: %+v\n", date)
}

func TestDateUnmarshalNull(t *testing.T) {
	var date Date
	var data = []byte(`null`)

	if err := date.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	t.Logf("Parsed null: %+v\n", date)
}

func TestDateScan(t *testing.T) {
	var date Date

	if err := date.Scan([]byte("2019-01-02")); err != nil {
		t.Error(err)
	}

	t.Logf("Retrieved SQL DATE into Go Date: %+v\n", date)
}

func TestDateValue(t *testing.T) {
	date := DateNow()

	v, err := date.Value()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Convert Go Date to SQL DATE in UTC: %s\n", v)
}
