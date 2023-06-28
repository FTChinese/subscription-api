package ids

import (
	"net/http"

	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"github.com/guregu/null"
)

func UserIDsFromHeader(h http.Header) UserIDs {
	ftcID := h.Get(xhttp.XUserID)
	unionID := h.Get(xhttp.XUnionID)

	return UserIDs{
		CompoundID: "",
		FtcID:      null.NewString(ftcID, ftcID != ""),
		UnionID:    null.NewString(unionID, unionID != ""),
	}.MustNormalize()
}
