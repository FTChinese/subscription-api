package controller

import (
	"github.com/google/uuid"
	"testing"
)

func TestIdempotencyKey(t *testing.T) {
	t.Log(uuid.New().String())
}
