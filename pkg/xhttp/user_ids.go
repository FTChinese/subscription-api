package xhttp

import (
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/guregu/null"
	"net/http"
)

func GetUserIDs(h http.Header) ids.UserIDs {
	ftcID := h.Get(XUserID)
	unionID := h.Get(XUnionID)

	return ids.UserIDs{
		CompoundID: "",
		FtcID:      null.NewString(ftcID, ftcID != ""),
		UnionID:    null.NewString(unionID, unionID != ""),
	}.MustNormalize()
}
