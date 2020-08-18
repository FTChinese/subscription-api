package client

import (
	"fmt"
	"github.com/FTChinese/subscription-api/pkg/config"
)

const insertOrderClient = `
INSERT INTO %s.client
SET order_id = :order_id,
	client_type = :client_type,
	client_version = :client_version,
	user_ip = INET6_ATON(:user_ip),
	user_agent = :user_agent`

func StmtInsertOrderClient(dbName config.SubsDB) string {
	return fmt.Sprintf(insertOrderClient, dbName)
}
