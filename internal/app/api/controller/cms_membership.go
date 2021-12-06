package controller

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/xhttp"
	"net/http"
)

// CreateMembership creates a membership purchased via ali or wx.
// Request body:
// - tier: string;
// - cycle: string;
// - expireDate: string;
// - payMethod: string;
func (router CMSRouter) CreateMembership(w http.ResponseWriter, req *http.Request) {
	id, _ := xhttp.GetURLParam(req, "id").ToString()

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	ba, err := router.ReaderRepo.SearchUserByFtcOrWxID(id)
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

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	v, err := router.Repo.UpdateMembership(
		id,
		params)

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

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	m, err := router.Repo.DeleteMembership(id)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !m.IsZero() {
		go func() {
			_ = router.ReaderRepo.VersionMembership(m.Deleted(reader.Archiver{
				Name:   reader.ArchiveName(params.CreatedBy),
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
