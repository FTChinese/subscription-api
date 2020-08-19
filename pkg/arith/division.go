package arith

// Division returns the integer quotient and remainder of x/y
func Division(x, y float64) (int64, float64) {
	var q int64

	for x > y {
		q = q + 1
		x = x - y
	}

	return q, x
}
