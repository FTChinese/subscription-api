// +build !production

package test

import (
	"bytes"
	"encoding/json"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/FTChinese/subscription-api/pkg/account"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/guregu/null"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
)

const baseUrl = "http://localhost:8202"

func BuildReqBody(v interface{}) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func GetRespBody(body io.Reader) []byte {
	b, err := ioutil.ReadAll(body)

	if err != nil {
		panic(err)
	}

	return b
}

func MustBuildReqBody(v interface{}) io.Reader {
	r, err := BuildReqBody(v)
	if err != nil {
		panic(err)
	}

	return r
}

type MobileLinkAccountKind int

const (
	MobileLinkNoProfile MobileLinkAccountKind = iota
	MobileLinkHasProfileNoPhone
	MobileLinkHasProfilePhoneSet
	MobileLinkHasProfilePhoneTaken
)

func (r Repo) GenerateMobileLinkParams(kind MobileLinkAccountKind) (string, pkg.MobileLinkParams) {
	defer r.logger.Sync()
	sugar := r.logger.Sugar()

	builder := account.NewMockFtcAccountBuilder(enum.AccountKindFtc)
	if kind == MobileLinkHasProfileNoPhone {
		builder = builder.WithMobile("")
	}
	baseAccount := builder.Build()

	sugar.Infof("Account to link mobile: %+v", baseAccount)

	r.MustCreateUserInfo(baseAccount)
	if kind != MobileLinkNoProfile {
		r.MustCreateProfile(baseAccount)
	}

	var phone string
	if kind == MobileLinkHasProfilePhoneSet {
		phone = baseAccount.Mobile.String
	} else {
		phone = faker.GenPhone()
	}

	v := ztsms.NewVerifier(phone, null.String{})
	sugar.Infof("Mobile verifier: %+v", v)
	r.MustSaveMobileVerifier(v)

	param := pkg.MobileLinkParams{
		EmailLoginParams: pkg.EmailLoginParams{
			Email:       baseAccount.Email,
			Password:    baseAccount.Password,
			DeviceToken: null.StringFrom(rand.String(36)),
		},
		Mobile: v.Mobile,
	}
	sugar.Infof("%s", faker.MustMarshalIndent(param))

	return baseAccount.FtcID, param
}

func (r Repo) BuildMobileLinkReq(kind MobileLinkAccountKind) *http.Request {
	_, params := r.GenerateMobileLinkParams(kind)

	return httptest.NewRequest(
		"POST",
		baseUrl+"/auth/mobile/link",
		MustBuildReqBody(params))
}

func AppleLinkReq() *http.Request {
	p := NewPersona()

	repo := NewRepo()

	repo.MustCreateFtcAccount(p.BaseAccount())
	repo.MustSaveIAPSubs(p.IAPSubs())

	input := p.IAPLinkInput()

	log.Printf("%s", faker.MustMarshalIndent(input))

	req := httptest.NewRequest(
		"POST",
		baseUrl+"/apple/link",
		MustBuildReqBody(input))

	return req
}

func AppleUnlinkReq() *http.Request {

	p := NewPersona().SetPayMethod(enum.PayMethodApple)

	repo := NewRepo()

	repo.MustCreateFtcAccount(p.BaseAccount())
	repo.MustSaveIAPSubs(p.IAPSubs())
	repo.MustSaveMembership(p.Membership())

	input := p.IAPLinkInput()

	log.Printf("%s", faker.MustMarshalIndent(input))

	req := httptest.NewRequest(
		"POST",
		baseUrl+"/apple/unlink",
		MustBuildReqBody(input))

	return req
}

func AppleListSubsReq() *http.Request {
	p := NewPersona()

	NewRepo().MustSaveIAPSubs(p.IAPSubs())

	req := httptest.NewRequest(
		"GET",
		baseUrl+"/apple/subs?page=1&per_page=10",
		nil)

	return req
}

func AppleSingleSubsReq() *http.Request {
	p := NewPersona()

	NewRepo().MustSaveIAPSubs(p.IAPSubs())

	req := httptest.NewRequest(
		"GET",
		baseUrl+"/apple/subs/"+p.AppleSubID,
		nil)

	return req
}
