package test

import (
	"github.com/google/uuid"
	"testing"
)

func TestUUID(t *testing.T) {
	t.Logf("FTC ID: %s", uuid.New().String())
}

func TestGenerateNewUser(t *testing.T) {
	t.Logf("%+v", NewProfile())
}
