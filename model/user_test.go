package model

import (
	"testing"
)

func TestFindUser(t *testing.T) {
	u, err := devEnv.FindUser(mockUser.ID)

	if err != nil {
		t.Error(nil)
	}

	t.Log(u)
}

// func TestTemplate(t *testing.T) {
// 	tmpl, err := template.New("test").Parse(letter)

// 	if err != nil {
// 		t.Error(err)

// 		return
// 	}

// 	data := struct {
// 		User User
// 		Subs Subscription
// 	}{
// 		mockUser,
// 		mockSubs,
// 	}

// 	err = tmpl.Execute(os.Stdout, data)

// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func TestComposeEmail(t *testing.T) {
// 	p, err := ComposeEmail(mockUser, mockSubs)

// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Log(p)
// }
