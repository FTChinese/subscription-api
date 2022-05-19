package legal

const colInsertLegal = `
title = :title,
summary = :summary,
body = :body,
keyword = :keyword
`

const StmtInsertLegal = `
INSERT INTO file_store.legal
SET hash_id = UNHEX(:hash_id),
	author = :author,
` + colInsertLegal + `,
	created_utc = :created_utc
`

const StmtUpdateLegal = `
UPDATE file_store.legal
SET ` + colInsertLegal + `,
	updated_utc = :updated_utc
WHERE hash_id = UNHEX(:hash_id)
LIMIT 1
`

const StmtRetrieveLegal = `
SELECT LOWER(HEX(hash_id)) AS hash_id,
	active,
	author,
	title,
	summary,
	body,
	keyword,
	created_utc,
	updated_utc
FROM file_store.legal
WHERE hash_id = UNHEX(?)
LIMIT 1
`

const StmtUpdateStatus = `
UPDATE file_store.legal
SET active = :active,
	updated_utc = :updated_utc
WHERE hash_id = UNHEX(:hash_id)
LIMIT 1
`

const stmtCountAll = `
SELECT COUNT(*) AS row_count
FROM file_store.legal
`

const StmtListLegal = `
SELECT LOWER(HEX(hash_id)) AS hash_id,
	title,
	summary
FROM file_store.legal
ORDER BY auto_id DESC
LIMIT ? OFFSET ?
`
