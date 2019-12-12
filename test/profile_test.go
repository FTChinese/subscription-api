package test

import (
	"testing"
)

func TestNewProfile(t *testing.T) {
	t.Log(NewProfile().Email)
	t.Log(NewProfile().Email)
}
