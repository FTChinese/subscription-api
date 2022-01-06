package apprepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg/android"
)

func (env Env) Latest() (android.Release, error) {
	var r android.Release

	err := env.dbs.Read.Get(&r, android.StmtLatest)

	if err != nil {
		return r, err
	}

	return r, nil
}

func (env Env) Releases(p gorest.Pagination) ([]android.Release, error) {

	releases := make([]android.Release, 0)

	err := env.dbs.Read.Select(
		&releases,
		android.StmtList,
		p.Limit,
		p.Offset())

	if err != nil {

		return nil, err
	}

	return releases, nil
}

func (env Env) SingleRelease(versionName string) (android.Release, error) {
	var r android.Release

	err := env.dbs.Read.Get(&r, android.StmtRelease, versionName)

	if err != nil {
		return r, err
	}

	return r, nil
}
