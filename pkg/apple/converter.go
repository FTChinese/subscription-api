package apple

import (
	"github.com/guregu/null"
	"strconv"
)

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

func ParseOptionalInt(str string) null.Int {
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return null.Int{}
	}

	return null.NewInt(i, i != 0)
}
