package legal

const colInsertLegal = `
author = :author,
title_en = :title_en,
title_cn = :title_cn,
summary = :summary,
body = :body,
keyword = :keyword
`

const StmtInsertLegal = `
INSERT INTO file_store.legal
SET title_hash = UNHEX(:title_hash),
` + colInsertLegal + `,
	created_utc = :created_utc
`

const StmtUpdateLegal = `
UPDATE file_store.legal
SET ` + colInsertLegal + `,
	updated_utc = :updated_utc
WHERE title_hash = UNHEX(:title_hash)
LIMIT 1
`

const StmtRetrieveLegal = `
SELECT LOWER(HEX(title_hash)) AS title_hash,
	author,
	title_en,
	title_cn,
	summary,
	body,
	keyword,
	created_utc,
	updated_utc
FROM file_store.legal
WHERE title_hash = UNHEX(?)
LIMIT 1
`

const StmtCount = `
SELECT COUNT(*) AS row_count
FROM file_store.legal
`

const StmtListLegal = `
SELECT LOWER(HEX(title_hash)) AS title_hash,
	title_cn,
	summary
FROM file_store.legal
ORDER BY id DESC
LIMIT ? OFFSET ?
`
