package collection

type StringSet map[string]interface{}

func NewStringSet() StringSet {
	return make(StringSet)
}

func (s StringSet) ToSlice() []string {
	var a []string
	for k := range s {
		a = append(a, k)
	}

	return a
}
