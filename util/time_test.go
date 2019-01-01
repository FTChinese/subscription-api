package util

import (
	"testing"
)

func TestTimeMarshal(t *testing.T) {
	time := TimeNow()

	b, err := time.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Stringify result: %s\n", b)
}

func TestTimeMarshalZero(t *testing.T) {
	time := TimeZero()

	b, err := time.MarshalJSON()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Stringify null: %s\n", b)
}

func TestTimeUnmarshal(t *testing.T) {
	var time Time

	var data = []byte(`"2019-01-01T03:54:44Z"`)

	if err := time.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	t.Logf("Parsed: %+v\n", time)
}

func TestTimeUnmarshalEmpty(t *testing.T) {
	var time Time

	var data = []byte(`""`)

	if err := time.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	t.Logf("Parsed: %+v\n", time)
}

func TestTimeUnmarshalNull(t *testing.T) {

	var time Time
	var data = []byte(`null`)

	if err := time.UnmarshalJSON(data); err != nil {
		t.Error(err)
	}

	t.Logf("Parsed: %+v\n", time)
}

func TestTimeScan(t *testing.T) {
	var time Time

	if err := time.Scan([]byte("2019-01-01 03:54:44")); err != nil {
		t.Error(err)
	}

	t.Logf("Put SQL Datetime into Time: %+v\n", time)
}

func TestTimeValue(t *testing.T) {
	time := TimeNow()

	v, err := time.Value()

	if err != nil {
		t.Error(err)
	}

	t.Logf("Convert to SQL Datetime in UTC: %+v\n", v)
}
