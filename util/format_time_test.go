package util

import "testing"

func TestParseWxTime(t *testing.T) {
	time := ParseWxTime("20141030133525")

	t.Log(time)

	dt := SQLDatetimeUTC.FromTime(time)

	t.Log(dt)
}
