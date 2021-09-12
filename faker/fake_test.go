package faker

import "testing"

func TestGenAvatar(t *testing.T) {
	t.Logf("Avatar %s", GenAvatar())
}
