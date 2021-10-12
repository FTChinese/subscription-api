//go:build !production
// +build !production

package test

import (
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/internal/pkg/input"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/addon"
	"github.com/FTChinese/subscription-api/pkg/apple"
	"github.com/FTChinese/subscription-api/pkg/reader"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/google/uuid"
	"github.com/guregu/null"
	"log"
	"net/http"
	"net/http/httptest"
)

const baseUrl = "http://localhost:8202"
const (
	urlMobileVrf    = baseUrl + "/auth/mobile/verification"
	urlMobileLink   = baseUrl + "/auth/mobile/link"
	urlMobileSignUp = baseUrl + "/auth/mobile/signup"
	urlAppleSubs    = baseUrl + "/apple/subs"
)

// ReqVerifySMSCode creates a request to verify sms code.
// To make it workable, you have to have a row in mobile_verify table.
func (r Repo) ReqVerifySMSCode(v ztsms.Verifier) *http.Request {

	r.MustSaveMobileVerifier(v)

	params := ztsms.VerifierParams{
		Mobile:      v.Mobile,
		Code:        v.Code,
		DeviceToken: null.String{},
	}

	return httptest.NewRequest(
		"POST",
		urlMobileVrf,
		faker.MustMarshalToReader(params))
}

func (r Repo) ReqVerifySMSForMobileEmail(v ztsms.Verifier) *http.Request {

	r.MustCreateUserInfo(
		NewPersona().
			WithMobile(v.Mobile).
			MobileOnlyAccount(),
	)

	r.MustSaveMobileVerifier(v)

	params := ztsms.VerifierParams{
		Mobile:      v.Mobile,
		Code:        v.Code,
		DeviceToken: null.String{},
	}

	return httptest.NewRequest(
		"POST",
		urlMobileVrf,
		faker.MustMarshalToReader(params))
}

func (r Repo) ReqMobileLinkNoProfile(a account.BaseAccount) *http.Request {
	repo := NewRepo()

	repo.MustCreateUserInfo(a)

	params := input.MobileLinkParams{
		EmailCredentials: input.EmailCredentials{
			Email:    a.Email,
			Password: a.Password,
		},
		DeviceToken: null.String{},
		Mobile:      faker.GenPhone(),
	}

	return httptest.NewRequest(
		"POST",
		urlMobileLink,
		faker.MustMarshalToReader(params))
}

func (r Repo) ReqMobileLinkWithProfile(a account.BaseAccount) *http.Request {
	repo := NewRepo()
	repo.MustCreateFtcAccount(a)

	mobile := a.Mobile.String
	if mobile == "" {
		mobile = faker.GenPhone()
	}

	params := input.MobileLinkParams{
		EmailCredentials: input.EmailCredentials{
			Email:    a.Email,
			Password: a.Password,
		},
		DeviceToken: null.String{},
		Mobile:      mobile,
	}

	return httptest.NewRequest(
		"POST",
		urlMobileLink,
		faker.MustMarshalToReader(params))
}

func (r Repo) ReqMobileLinkPhoneTaken(a account.BaseAccount) *http.Request {
	repo := NewRepo()
	repo.MustCreateFtcAccount(a)

	params := input.MobileLinkParams{
		EmailCredentials: input.EmailCredentials{
			Email:    a.Email,
			Password: a.Password,
		},
		DeviceToken: null.String{},
		Mobile:      faker.GenPhone(),
	}

	return httptest.NewRequest(
		"POST",
		urlMobileLink,
		faker.MustMarshalToReader(params))
}

func (r Repo) GenerateIAPUnlinkParams(hasAddOn bool) apple.LinkInput {
	defer r.logger.Sync()
	sugar := r.logger.Sugar()

	ftcID := uuid.New().String()
	iapID := faker.GenAppleSubID()

	builder := reader.NewMockMemberBuilderV2(enum.AccountKindFtc).
		WithFtcID(ftcID).
		WithPayMethod(enum.PayMethodApple).
		WithIapID(iapID)
	if hasAddOn {
		builder = builder.WithAddOn(addon.AddOn{
			Standard: 61,
			Premium:  0,
		})
	}

	m := builder.Build()
	sugar.Infof("%s", faker.MustMarshalIndent(m))

	iapSub := apple.NewMockSubsBuilder(ftcID).
		WithOriginalTxID(iapID).
		Build()
	sugar.Infof("%s", faker.MustMarshalIndent(iapSub))

	repo := NewRepo()
	repo.MustSaveMembership(m)
	repo.MustSaveIAPSubs(iapSub)

	return apple.LinkInput{
		FtcID:        ftcID,
		OriginalTxID: iapID,
		Force:        false,
	}
}

func AppleLinkReq(params apple.LinkInput) *http.Request {
	req := httptest.NewRequest(
		"POST",
		baseUrl+"/apple/link",
		faker.MustMarshalToReader(params))

	return req
}

func AppleUnlinkReq() *http.Request {

	p := NewPersona()

	repo := NewRepo()

	id := faker.GenAppleSubID()

	repo.MustCreateFtcAccount(p.EmailOnlyAccount())
	repo.MustSaveIAPSubs(
		NewIAPBuilder(id).
			Build())

	repo.MustSaveMembership(p.MemberBuilder().WithApple(id).Build())

	input := p.IAPLinkInput()

	log.Printf("%s", faker.MustMarshalIndent(input))

	req := httptest.NewRequest(
		"POST",
		baseUrl+"/apple/unlink",
		faker.MustMarshalToReader(input))

	return req
}

func AppleListSubsReq() *http.Request {
	p := NewPersona()

	NewRepo().MustSaveIAPSubs(p.IAPBuilder().Build())

	req := httptest.NewRequest(
		"GET",
		urlAppleSubs+"?page=1&per_page=10",
		nil)

	return req
}

func AppleSingleSubsReq() *http.Request {
	p := NewPersona()

	NewRepo().MustSaveIAPSubs(p.IAPBuilder().Build())

	req := httptest.NewRequest(
		"GET",
		urlAppleSubs+"/"+p.AppleSubID,
		nil)

	return req
}
