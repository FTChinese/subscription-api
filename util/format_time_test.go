package util

import "testing"

func TestParseWxTime(t *testing.T) {
	time, _ := ParseWxTime("20141030133525")

	t.Log(time)

	dt := ToSQLDatetimeUTC.FromTime(time)

	t.Log(dt)
}
