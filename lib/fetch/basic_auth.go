package fetch

type BasicAuth struct {
	Username string
	Password string
}

func (a BasicAuth) IsZero() bool {
	return a.Username == "" || a.Password == ""
}
