package controller

import (
	"github.com/FTChinese/subscription-api/faker"
	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/FTChinese/subscription-api/pkg/ztsms"
	"github.com/FTChinese/subscription-api/test"
	"github.com/guregu/null"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRouter_VerifySMSCode(t *testing.T) {

	repo := test.NewRepo()

	router := NewAuthRouter(
		db.MockMySQL(),
		test.Postman,
		zaptest.NewLogger(t),
	)

	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{
			name: "Existing mobile",
			args: args{
				w: httptest.NewRecorder(),
				req: repo.ReqVerifySMSCode(test.
					NewPersona().
					MobileVerifier()),
			},
			wantStatus: 200,
		},
		{
			name: "Email derived from mobile",
			args: args{
				w: httptest.NewRecorder(),
				req: repo.ReqVerifySMSForMobileEmail(
					ztsms.NewVerifier(faker.GenPhone(), null.String{}),
				),
			},
			wantStatus: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			router.VerifySMSCode(tt.args.w, tt.args.req)

			resp := tt.args.w.Result()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("VerifySMS() want status %d, got %d", tt.wantStatus, resp.StatusCode)
				t.Logf("%s", faker.MustReadBody(resp.Body))
				return
			}

			t.Logf("%s", faker.MustReadBody(resp.Body))
		})
	}
}

func TestAuthRouter_LinkMobile(t *testing.T) {

	repo := test.NewRepo()

	router := NewAuthRouter(test.SplitDB, test.Postman, zaptest.NewLogger(t))

	type args struct {
		w   *httptest.ResponseRecorder
		req *http.Request
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{
			name: "Profile row does not exist",
			args: args{
				w:   httptest.NewRecorder(),
				req: repo.ReqMobileLinkNoProfile(test.NewPersona().EmailOnlyAccount()),
			},
			wantStatus: 200,
		},
		{
			name: "Mobile row exist with empty phone",
			args: args{
				w: httptest.NewRecorder(),
				req: repo.ReqMobileLinkWithProfile(
					test.NewPersona().
						EmailOnlyAccount()),
			},
			wantStatus: 200,
		},
		{
			name: "Mobile row exist with phone set",
			args: args{
				w: httptest.NewRecorder(),
				req: repo.ReqMobileLinkWithProfile(
					test.NewPersona().
						EmailMobileAccount()),
			},
			wantStatus: 200,
		},
		{
			name: "Mobile row exist with phone taken",
			args: args{
				w: httptest.NewRecorder(),
				req: repo.ReqMobileLinkPhoneTaken(
					test.NewPersona().
						EmailMobileAccount()),
			},
			wantStatus: 422,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			router.MobileLinkExistingEmail(tt.args.w, tt.args.req)

			resp := tt.args.w.Result()

			if resp.StatusCode != tt.wantStatus {
				t.Logf("LinkMoble() want status %d, got %d", tt.wantStatus, resp.StatusCode)
				return
			}

			t.Logf("%s", faker.MustReadBody(resp.Body))
		})
	}
}
