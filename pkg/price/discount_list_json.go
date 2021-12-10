package price

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"sort"
)

type DiscountListJSON []Discount

func (l DiscountListJSON) Value() (driver.Value, error) {
	if len(l) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (l *DiscountListJSON) Scan(src interface{}) error {
	if src == nil {
		*l = DiscountListJSON{}
		return nil
	}
	switch s := src.(type) {
	case []byte:
		var tmp []Discount
		err := json.Unmarshal(s, &tmp)
		if err != nil {
			return err
		}
		*l = tmp
		return nil

	default:
		return errors.New("incompatible type to scan to DiscountListJSON")
	}
}

// FindApplicable select an offer from all Offers that a user eligible for.
// Filter criteria:
// * The offer validity period is in effect
// * User is qualified to enjoy
// * Lowest price/Highest discount rate.
// filters - all possible offers a membership currently enjoys, used to narrow down
// offers provided.
// If more than one offer is applicable, use the one with the largest price off.
func (l DiscountListJSON) FindApplicable(filters []OfferKind) Discount {
	// Filter all valid discount offers.
	var filtered = make([]Discount, 0)
	for _, v := range l {
		if v.IsValid() && v.Kind.ContainedBy(filters) {
			filtered = append(filtered, v)
		}
	}

	switch len(filtered) {
	case 0:
		return Discount{}
	case 1:
		return filtered[0]
	default:
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].PriceOff.Float64 > filtered[j].PriceOff.Float64
		})

		return filtered[0]
	}
}

func (l DiscountListJSON) FindValid(id string) (Discount, error) {
	for _, v := range l {
		if v.ID == id {
			if v.IsValid() {
				return v, nil
			} else {
				return Discount{}, errors.New("invalid offer selected")
			}
		}
	}

	return Discount{}, errors.New("the requested offer is not found")
}
