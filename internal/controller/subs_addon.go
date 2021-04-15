package controller

import (
	"github.com/FTChinese/go-rest/render"
	"net/http"
)

func (router SubsRouter) ClaimAddOn(w http.ResponseWriter, req *http.Request) {
	readerIDs := getReaderIDs(req.Header)

	result, err := router.AddOnRepo.ClaimAddOn(readerIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	go func() {
		_ = router.SubsRepo.ArchiveMember(result.Snapshot)
	}()

	_ = render.New(w).OK(result.Membership)
}
