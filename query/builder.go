package query

type Builder struct {
	sandbox bool
}

func (b Builder) MemberDB() string {
	if b.sandbox {
		return "sandbox"
	}

	return "premium"
}

func (b Builder) CmsTmp() string {
	if b.sandbox {
		return "sandbox"
	}

	return "cmstmp01"
}

func (b Builder) UserDB() string {
	if b.sandbox {
		return "sandbox"
	}

	return "user_db"
}
func NewBuilder(sandbox bool) Builder {
	return Builder{sandbox: sandbox}
}
