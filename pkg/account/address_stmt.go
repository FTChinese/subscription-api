package account

const colsGetAddress = `
p.country AS country,
p.province AS province,
p.city AS city,
p.district AS district,
p.street AS street,
p.postcode AS postcode
`

const StmtLoadAddress = `
SELECT ` + colsGetAddress + `
FROM user_db.profile AS p
WHERE user_id = ?
LIMIT 1`

const colsSetAddress = `
country = :country,
province = :province,
city = :city,
district = :district,
street = :street,
postcode = :postcode,
updated_utc = CURRENT_TIMESTAMP
`

const StmtUpdateAddress = `
INSERT INTO user_db.profile
SET user_id = :ftc_id,` + colsSetAddress + `
ON DUPLICATE KEY UPDATE` + colsSetAddress
