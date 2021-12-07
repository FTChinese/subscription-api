package controller

import (
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// ClaimAddOn extends expiration time by transferring addon periods.
// This could be done either by client automatically, or by
// ftc staff manually.
func (router SubsRouter) ClaimAddOn(w http.ResponseWriter, req *http.Request) {
	readerIDs := xhttp.GetUserIDs(req.Header)

	result, err := router.AddOnRepo.ClaimAddOn(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		_ = router.ReaderRepo.ArchiveMember(result.Snapshot)
	}()

	_ = render.New(w).OK(result.Membership)
}
