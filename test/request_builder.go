package test

import (
	"bytes"
	"encoding/json"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/subscription-api/faker"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
)

const baseUrl = "http://localhost:8200"

func BuildReqBody(v interface{}) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

func MustBuildReqBody(v interface{}) io.Reader {
	r, err := BuildReqBody(v)
	if err != nil {
		panic(err)
	}

	return r
}

func AppleLinkReq() *http.Request {
	p := NewPersona()

	repo := NewRepo()

	repo.MustSaveAccount(p.FtcAccount())
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

	repo.MustSaveAccount(p.FtcAccount())
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
