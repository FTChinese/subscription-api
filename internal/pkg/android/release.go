package android

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/guregu/null"
)

type Release struct {
	VersionName string      `json:"versionName" db:"version_name"`
	VersionCode int64       `json:"versionCode" db:"version_code"`
	Body        null.String `json:"body" db:"body"`
	ApkURL      string      `json:"apkUrl" db:"apk_url"`
	CreatedAt   chrono.Time `json:"createdAt" db:"created_utc"`
	UpdatedAt   chrono.Time `json:"updatedAt" db:"updated_utc"`
}

const colAndroid = `
SELECT version_name,
	version_code,
	body,
	apk_url,
	created_utc,
	updated_utc
FROM file_store.android_release`

const StmtLatest = colAndroid + `
ORDER BY version_code DESC
LIMIT 1`

const StmtList = colAndroid + `
ORDER BY version_code DESC
LIMIT ? OFFSET ?`

const StmtRelease = colAndroid + `
WHERE version_name = ?
LIMIT 1`
