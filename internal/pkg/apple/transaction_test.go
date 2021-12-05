package apple

import (
	"encoding/json"
	"testing"
)

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

func TestUnmarshalReceiptInfo(t *testing.T) {
	var r Transaction

	if err := json.Unmarshal([]byte(mockReceipt), &r); err != nil {
		t.Error(err)
	}

	t.Logf("A receipt: %+v", r)
}
