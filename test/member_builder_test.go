package test

import (
	"testing"
)

func TestMemberBuilder_Build(t *testing.T) {
	m := NewPersona().MemberBuilder().Build()

	t.Logf("%+v", m)
}
