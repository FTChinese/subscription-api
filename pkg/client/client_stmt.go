package client

const StmtInsertClient = `
INSERT INTO user_db.client_footprint
SET ftc_id 			= :ftc_id,
    platform 		= :platform,
    client_version 	= :client_version,
    user_ip 		= INET6_ATON(:user_ip),
    user_agent 		= :user_agent,
    created_utc 	= UTC_TIMESTAMP(),
    source 			= :source,
	auth_method 	= :auth_method,
	device_token 	= :device_token`

const StmtInsertOrderClient = `
INSERT INTO premium.client
SET order_id = :order_id,
	client_type = :client_type,
	client_version = :client_version,
	user_ip = INET6_ATON(:user_ip),
	user_agent = :user_agent`
