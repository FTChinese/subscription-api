package controller

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/ids"
	"github.com/FTChinese/subscription-api/pkg/price"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// CreateMembership creates a membership purchased via ali or wx.
// Request body:
// - ftcId?: string;
// - unionId?: string;
// - tier: string;
// - cycle: string;
// - expireDate: string;
// - payMethod: string;
func (router CMSRouter) CreateMembership(w http.ResponseWriter, req *http.Request) {

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(true); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	ba, err := router.ReaderRepo.FindBaseAccount(ids.UserIDs{
		CompoundID: "",
		FtcID:      params.FtcID,
		UnionID:    params.UnionID,
	}.MustNormalize())

	// TODO: in the future client should present a drag-drop ui
	// so that user could directly select a price.
	paywall, err := router.PaywallRepo.LoadPaywall(router.Live)
	ftcPrice, _ := paywall.FindPriceByEdition(price.Edition{
		Tier:  params.Tier,
		Cycle: params.Cycle,
	})

	params.PriceID = ftcPrice.ID

	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	mmb, err := router.Repo.CreateMembership(ba, params)
	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}

	// TODO: send email to this user.

	_ = render.New(w).OK(mmb)
}

func (router CMSRouter) UpdateMembership(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()
	staffName := xhttp.GetStaffName(req.Header)

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(false); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	// TODO: in the future client should present a drag-drop ui
	// so that user could directly select a price.
	paywall, err := router.PaywallRepo.LoadPaywall(router.Live)
	ftcPrice, _ := paywall.FindPriceByEdition(price.Edition{
		Tier:  params.Tier,
		Cycle: params.Cycle,
	})

	params.PriceID = ftcPrice.ID

	v, err := router.Repo.UpdateMembership(
		id,
		params,
		staffName)

	if err != nil {
		var ve *render.ValidationError
		if errors.As(err, &ve) {
			_ = render.New(w).Unprocessable(ve)
			return
		}

		_ = render.New(w).DBError(err)
		return
	}
	go func() {
		// TODO: versioned by
		err := router.ReaderRepo.VersionMembership(v)
		if err != nil {

		}
	}()
	_ = render.New(w).OK(v.PostChange)
}

// DeleteMembership manually.
// Request body:
func (router CMSRouter) DeleteMembership(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()
	staffName := xhttp.GetStaffName(req.Header)

	m, err := router.Repo.DeleteMembership(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !m.IsZero() {
		go func() {
			_ = router.ReaderRepo.VersionMembership(m.Deleted(reader.Archiver{
				Name:   reader.ArchiveName(staffName),
				Action: reader.ArchiveActionDelete,
			}))
		}()
	}

	_ = render.New(w).NoContent()
}

// ListMemberSnapshots loads a list of membership change history.
// Pagination support by adding query parameter:
// page=<int>&per_page=<int>
func (router CMSRouter) ListMemberSnapshots(w http.ResponseWriter, req *http.Request) {

	id, _ := xhttp.GetURLParam(req, "id").ToString()

	ba, err := router.ReaderRepo.SearchUserByFtcOrWxID(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p := gorest.GetPagination(req)

	list, err := router.Repo.ListSnapshot(ba.CompoundIDs(), p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}
