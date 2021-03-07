package price

type RenewalOffer struct {
	StartFrom int64
	Until     int64
	PriceOff  float64
	Percent   int64
}

var renewalOffers = []RenewalOffer{
	{
		StartFrom: 30,
		Until:     15,
		PriceOff:  40,
		Percent:   0,
	},
	{
		StartFrom: 15,
		Until:     1,
		PriceOff:  20,
		Percent:   0,
	},
}

func FindRenewalOffer(daysRemained int64) RenewalOffer {
	for _, v := range renewalOffers {
		if daysRemained >= v.Until && daysRemained <= v.StartFrom {
			return v
		}
	}

	return RenewalOffer{}
}
