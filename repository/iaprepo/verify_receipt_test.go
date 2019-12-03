package iaprepo

import (
	"encoding/json"
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/guregu/null"
	"gitlab.com/ftchinese/subscription-api/models/apple"
	"gitlab.com/ftchinese/subscription-api/models/util"
	"gitlab.com/ftchinese/subscription-api/test"
	"testing"
	"time"
)

const mockReceiptToken = `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`

const mockResponse = `
{
  "status": 0,
  "environment": "Sandbox",
  "receipt": {
    "receipt_type": "ProductionSandbox",
    "adam_id": 0,
    "app_item_id": 0,
    "bundle_id": "com.ft.ftchinese.mobile",
    "application_version": "1",
    "download_id": 0,
    "version_external_identifier": 0,
    "receipt_creation_date": "2019-11-22 07:26:56 Etc/GMT",
    "receipt_creation_date_ms": "1574407616000",
    "receipt_creation_date_pst": "2019-11-21 23:26:56 America/Los_Angeles",
    "request_date": "2019-11-26 01:26:27 Etc/GMT",
    "request_date_ms": "1574731587005",
    "request_date_pst": "2019-11-25 17:26:27 America/Los_Angeles",
    "original_purchase_date": "2013-08-01 07:00:00 Etc/GMT",
    "original_purchase_date_ms": "1375340400000",
    "original_purchase_date_pst": "2013-08-01 00:00:00 America/Los_Angeles",
    "original_application_version": "1.0",
    "in_app": [
      {
        "quantity": "1",
        "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
        "transaction_id": "1000000595923656",
        "original_transaction_id": "1000000595923656",
        "purchase_date": "2019-11-22 07:21:55 Etc/GMT",
        "purchase_date_ms": "1574407315000",
        "purchase_date_pst": "2019-11-21 23:21:55 America/Los_Angeles",
        "original_purchase_date": "2019-11-22 07:21:13 Etc/GMT",
        "original_purchase_date_ms": "1574407273000",
        "original_purchase_date_pst": "2019-11-21 23:21:13 America/Los_Angeles",
        "expires_date": "2019-11-22 07:26:55 Etc/GMT",
        "expires_date_ms": "1574407615000",
        "expires_date_pst": "2019-11-21 23:26:55 America/Los_Angeles",
        "web_order_line_item_id": "1000000048450768",
        "is_trial_period": "false",
        "is_in_intro_offer_period": "false"
      },
      {
        "quantity": "1",
        "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
        "transaction_id": "1000000595926312",
        "original_transaction_id": "1000000595926312",
        "purchase_date": "2019-11-22 07:26:55 Etc/GMT",
        "purchase_date_ms": "1574407615000",
        "purchase_date_pst": "2019-11-21 23:26:55 America/Los_Angeles",
        "original_purchase_date": "2019-11-22 07:26:07 Etc/GMT",
        "original_purchase_date_ms": "1574407567000",
        "original_purchase_date_pst": "2019-11-21 23:26:07 America/Los_Angeles",
        "expires_date": "2019-11-22 07:31:55 Etc/GMT",
        "expires_date_ms": "1574407915000",
        "expires_date_pst": "2019-11-21 23:31:55 America/Los_Angeles",
        "web_order_line_item_id": "1000000048450931",
        "is_trial_period": "false",
        "is_in_intro_offer_period": "false"
      }
    ]
  },
  "latest_receipt_info": [
    {
      "quantity": "1",
      "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
      "transaction_id": "1000000595926312",
      "original_transaction_id": "1000000595926312",
      "purchase_date": "2019-11-22 07:26:55 Etc/GMT",
      "purchase_date_ms": "1574407615000",
      "purchase_date_pst": "2019-11-21 23:26:55 America/Los_Angeles",
      "original_purchase_date": "2019-11-22 07:26:07 Etc/GMT",
      "original_purchase_date_ms": "1574407567000",
      "original_purchase_date_pst": "2019-11-21 23:26:07 America/Los_Angeles",
      "expires_date": "2019-11-22 07:31:55 Etc/GMT",
      "expires_date_ms": "1574407915000",
      "expires_date_pst": "2019-11-21 23:31:55 America/Los_Angeles",
      "web_order_line_item_id": "1000000048450931",
      "is_trial_period": "false",
      "is_in_intro_offer_period": "false",
      "subscription_group_identifier": "20423285"
    },
    {
      "quantity": "1",
      "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
      "transaction_id": "1000000595951896",
      "original_transaction_id": "1000000595951896",
      "purchase_date": "2019-11-22 08:11:38 Etc/GMT",
      "purchase_date_ms": "1574410298000",
      "purchase_date_pst": "2019-11-22 00:11:38 America/Los_Angeles",
      "original_purchase_date": "2019-11-22 08:11:39 Etc/GMT",
      "original_purchase_date_ms": "1574410299000",
      "original_purchase_date_pst": "2019-11-22 00:11:39 America/Los_Angeles",
      "expires_date": "2019-11-22 08:16:38 Etc/GMT",
      "expires_date_ms": "1574410598000",
      "expires_date_pst": "2019-11-22 00:16:38 America/Los_Angeles",
      "web_order_line_item_id": "1000000048451078",
      "is_trial_period": "false",
      "is_in_intro_offer_period": "false",
      "subscription_group_identifier": "20423285"
    }
  ],
  "pending_renewal_info": [
    {
      "expiration_intent": "1",
      "auto_renew_product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
      "original_transaction_id": "1000000595951896",
      "is_in_billing_retry_period": "0",
      "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
      "auto_renew_status": "0"
    }
  ]
}`

func createMockResponse() *apple.VerificationResponseBody {
	var resp apple.VerificationResponseBody
	if err := json.Unmarshal([]byte(mockResponse), &resp); err != nil {
		panic(err)
	}

	return &resp
}

const mockReceipt = `
{
      "quantity": "1",
      "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
      "transaction_id": "1000000595951896",
      "original_transaction_id": "1000000595951896",
      "purchase_date": "2019-11-22 08:11:38 Etc/GMT",
      "purchase_date_ms": "1574410298000",
      "purchase_date_pst": "2019-11-22 00:11:38 America/Los_Angeles",
      "original_purchase_date": "2019-11-22 08:11:39 Etc/GMT",
      "original_purchase_date_ms": "1574410299000",
      "original_purchase_date_pst": "2019-11-22 00:11:39 America/Los_Angeles",
      "expires_date": "2019-11-22 08:16:38 Etc/GMT",
      "expires_date_ms": "1574410598000",
      "expires_date_pst": "2019-11-22 00:16:38 America/Los_Angeles",
      "web_order_line_item_id": "1000000048451078",
      "is_trial_period": "false",
      "is_in_intro_offer_period": "false",
      "subscription_group_identifier": "20423285"
}`

func createMockReceipt() apple.Transaction {
	var r apple.Transaction
	if err := json.Unmarshal([]byte(mockReceipt), &r); err != nil {
		panic(err)
	}

	return r
}

const mockPendingRenewal = `
{
  "expiration_intent": "1",
  "auto_renew_product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
  "original_transaction_id": "1000000595951896",
  "is_in_billing_retry_period": "0",
  "product_id": "com.ft.ftchinese.mobile.subscription.member.monthly",
  "auto_renew_status": "0"
}`

func createMockPendingRenewal() apple.PendingRenewal {
	var p apple.PendingRenewal
	if err := json.Unmarshal([]byte(mockPendingRenewal), &p); err != nil {
		panic(err)
	}

	return p
}

func TestIAPEnv_SaveVerificationSession(t *testing.T) {

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}

	type args struct {
		v apple.VerificationSessionSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Save Verification Session",
			args:    args{v: createMockResponse().SessionSchema()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveVerificationSession(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("SaveVerificationSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_SaveCustomerReceipt(t *testing.T) {
	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}

	type args struct {
		r apple.TransactionSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Save Customer Receipt",
			args:    args{r: createMockReceipt().Schema(apple.EnvSandbox)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveCustomerReceipt(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveCustomerReceipt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_SavePendingRenewal(t *testing.T) {

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}

	type args struct {
		p apple.PendingRenewalSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Save Pending Renewal",
			args:    args{p: createMockPendingRenewal().Schema(apple.EnvSandbox)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.SavePendingRenewal(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("SavePendingRenewal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_CreateSubscription(t *testing.T) {

	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}

	type args struct {
		s apple.Subscription
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Create Subscription",
			args: args{s: apple.Subscription{
				Environment:           apple.EnvSandbox,
				OriginalTransactionID: "1000000595951896",
				LastTransactionID:     "1000000595951896",
				ProductID:             "com.ft.ftchinese.mobile.subscription.member.monthly",
				PurchaseDateUTC:       chrono.TimeNow(),
				ExpiresDateUTC:        chrono.TimeFrom(time.Now().AddDate(0, 1, 0)),
				FtcID:                 null.String{},
				UnionID:               null.String{},
				Tier:                  enum.TierStandard,
				Cycle:                 enum.CycleMonth,
				AutoRenewal:           null.BoolFrom(true),
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.CreateSubscription(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("CreateSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_SaveReceiptToken(t *testing.T) {
	env := IAPEnv{
		c:  util.BuildConfig{},
		db: test.DB,
	}
	type args struct {
		r apple.ReceiptToken
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Receipt Token",
			args: args{r: apple.ReceiptToken{
				Environment:           apple.EnvSandbox,
				OriginalTransactionID: "1000000595951896",
				LatestReceipt:         mockReceiptToken,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveReceiptToken(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("SaveReceiptToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIAPEnv_VerifyReceipt(t *testing.T) {
	env := IAPEnv{
		c:  util.BuildConfig{Production: false, Sandbox: false},
		db: nil,
	}

	type args struct {
		r apple.VerificationRequestBody
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Verify Receipt",
			args: args{r: apple.VerificationRequestBody{
				ReceiptData: mockReceiptToken,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.VerifyReceipt(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyReceipt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("Verification response: %+v", got)
		})
	}
}

func TestIAPEnv_SaveVerificationFailure(t *testing.T) {
	env := IAPEnv{
		c:  util.BuildConfig{Production: false, Sandbox: false},
		db: test.DB,
	}

	type args struct {
		f apple.VerificationFailed
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Save Verification Failure",
			args: args{
				f: apple.VerificationFailed{
					Environment: apple.EnvNull,
					Status:      21199,
					Message:     null.String{},
					ReceiptData: mockReceiptToken,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.SaveVerificationFailure(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("SaveVerificationFailure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
