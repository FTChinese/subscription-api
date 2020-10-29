package subs

import "github.com/FTChinese/subscription-api/pkg/client"

const StmtInsertOrderMeta = `
INSERT INTO premium.client
SET order_id = :order_id,
	client_type = :client_type,
	client_version = :client_version,
	user_ip = INET6_ATON(:user_ip),
	user_agent = :user_agent`

type OrderMeta struct {
	OrderID string `db:"order_id"`
	client.Client
}
