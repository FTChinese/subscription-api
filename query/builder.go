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

func NewBuilder(sandbox bool) Builder {
	return Builder{sandbox: sandbox}
}
