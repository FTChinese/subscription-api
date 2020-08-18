package apple

import "strconv"

func MustParseBoolean(str string) bool {
	b, err := strconv.ParseBool(str)
	if err != nil {
		return false
	}

	return b
}

func MustParseInt64(str string) int64 {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}

	return i
}
