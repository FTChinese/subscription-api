package api

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ClaimAddOn extends expiration time by transferring addon periods.
// This could be done either by client automatically, or by
// ftc staff manually.
// TODO: when claiming addon for an expired b2b, we
// revoke the linked licence automatically.
func (router FtcPayRouter) ClaimAddOn(w http.ResponseWriter, req *http.Request) {
	readerIDs := xhttp.UserIDsFromHeader(req.Header)

	result, err := router.AddOnRepo.ClaimAddOn(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !result.Versioned.IsZero() {
		go func() {
			_ = router.ReaderRepo.VersionMembership(result.Versioned)
		}()
	}

	_ = render.New(w).OK(result.Membership)
}
