package client

const StmtInsertOrderClient = `
INSERT INTO premium.client
SET order_id = :order_id,
	client_type = :client_type,
	client_version = :client_version,
	user_ip = INET6_ATON(:user_ip),
	user_agent = :user_agent`
