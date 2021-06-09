package controller

import (
	"github.com/FTChinese/subscription-api/test"
	"go.uber.org/zap/zaptest"
	"net/http/httptest"
	"testing"
)

func TestAuthRouter_LinkMobile(t *testing.T) {

	repo := test.NewRepoV2(zaptest.NewLogger(t))

	router := NewAuthRouter(test.SplitDB, test.Postman, zaptest.NewLogger(t))

	type args struct {
		w    *httptest.ResponseRecorder
		kind test.MobileLinkAccountKind
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{
			name: "Mobile row not exist",
			args: args{
				w:    httptest.NewRecorder(),
				kind: test.MobileLinkNoProfile,
			},
			wantStatus: 200,
		},
		{
			name: "Mobile row exist with empty phone",
			args: args{
				w:    httptest.NewRecorder(),
				kind: test.MobileLinkHasProfileNoPhone,
			},
			wantStatus: 200,
		},
		{
			name: "Mobile row exist with phone set",
			args: args{
				w:    httptest.NewRecorder(),
				kind: test.MobileLinkHasProfilePhoneSet,
			},
			wantStatus: 200,
		},
		{
			name: "Mobile row exist with phone taken",
			args: args{
				w:    httptest.NewRecorder(),
				kind: test.MobileLinkHasProfilePhoneTaken,
			},
			wantStatus: 422,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := repo.BuildMobileLinkReq(tt.args.kind)

			router.LinkMobile(tt.args.w, req)

			resp := tt.args.w.Result()

			if resp.StatusCode != tt.wantStatus {
				t.Logf("LinkMoble() want status %d, got %d", tt.wantStatus, resp.StatusCode)
				return
			}

			t.Logf("%s", test.GetRespBody(resp.Body))
		})
	}
}
