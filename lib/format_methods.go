package lib

import (
	"github.com/FTChinese/go-rest/enum"
	"strings"
)

func formatMethods(methods []enum.PayMethod) string {
	l := len(methods)
	switch {
	case l == 0:
		return ""

	case l == 1:
		return methods[0].String()

	case l >= 2:
		var buf strings.Builder
		for i, v := range methods {
			if i == 0 {
				buf.WriteString(v.String())
				continue
			}
			if i == l-1 {
				buf.WriteString(" or " + v.String())
				continue
			}

			buf.WriteString(", " + v.String())
		}
		return buf.String()
	}

	return ""
}
