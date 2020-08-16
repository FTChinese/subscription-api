package product

type StripePlan struct {
	Edition
	livePlanID string
	testPlanID string
}

func (s StripePlan) GetID(live bool) string {
	if live {
		return s.livePlanID
	}

	return s.testPlanID
}
