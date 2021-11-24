package controller

import (
	"errors"
	"github.com/FTChinese/go-rest"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"net/http"
)

func (router AccountRouter) LoadMembership(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	m, err := router.userRepo.RetrieveMember(userIDs.CompoundID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(m)
}

// CreateMembership creates a membership purchased via ali or wx.
// Input: subs.FtcSubsCreationInput
// - ftcId?: string;
// - unionId?: string
// - tier: string;
// - cycle: string;
// - expireDate: string;
// - payMethod: string;
func (router AccountRouter) CreateMembership(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	ba, err := router.userRepo.FindBaseAccount(userIDs)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	mmb, err := router.userRepo.CreateMembership(ba, params)
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

func (router AccountRouter) UpdateMembership(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	if ve := params.Validate(); ve != nil {
		_ = render.New(w).Unprocessable(ve)
		return
	}

	v, err := router.userRepo.UpdateMembership(
		userIDs.CompoundID,
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
		err := router.userRepo.VersionMembership(v)
		if err != nil {

		}
	}()
	_ = render.New(w).OK(v.PostChange)
}

// DeleteMembership manually.
// Request body:
func (router AccountRouter) DeleteMembership(w http.ResponseWriter, req *http.Request) {
	userIDs := getReaderIDs(req.Header)

	var params input.MemberParams
	if err := gorest.ParseJSON(req.Body, &params); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	m, err := router.userRepo.DeleteMembership(userIDs.CompoundID)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	if !m.IsZero() {
		go func() {
			_ = router.userRepo.VersionMembership(m.Deleted(reader.Archiver{
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
func (router AccountRouter) ListMemberSnapshots(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		_ = render.New(w).BadRequest(err.Error())
		return
	}

	p := gorest.GetPagination(req)
	userIDs := getReaderIDs(req.Header)

	list, err := router.userRepo.ListSnapshot(userIDs, p)
	if err != nil {
		_ = render.New(w).DBError(err)
		return
	}

	_ = render.New(w).OK(list)
}
