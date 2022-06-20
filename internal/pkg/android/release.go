package android

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/render"
	"github.com/FTChinese/subscription-api/lib/validator"
	"github.com/FTChinese/subscription-api/pkg"
	"github.com/guregu/null"
	"strings"
)

// ReleaseInput contains the fields required to create or update a release.
// For creation, VersionName + VersionCode + ApkURL are required.
// For update, only ApkURL is required.
type ReleaseInput struct {
	// Required only when creating a new release. Ignore it when updating since the version name is
	// acquired from path parameter.
	VersionName string `json:"versionName" db:"version_name"`
	// Required only when creating a new release. This cannot be changed upon updating.
	VersionCode int64       `json:"versionCode" db:"version_code"`
	Body        null.String `json:"body" db:"body"`      // Optional
	ApkURL      string      `json:"apkUrl" db:"apk_url"` // Required.
}

func (r *ReleaseInput) ValidateUpdate() *render.ValidationError {
	r.Body.String = strings.TrimSpace(r.Body.String)
	r.ApkURL = strings.TrimSpace(r.ApkURL)

	return validator.New("apkUrl").
		Required().
		URL().
		Validate(r.ApkURL)
}

func (r *ReleaseInput) ValidateCreation() *render.ValidationError {
	r.VersionName = strings.TrimSpace(r.VersionName)

	ie := validator.New("versionName").
		Required().
		MaxLen(32).
		Validate(r.VersionName)
	if ie != nil {
		return ie
	}

	if r.VersionCode < 1 {
		return &render.ValidationError{
			Message: "version code must be larger than 0",
			Field:   "versionCode",
			Code:    render.CodeInvalid,
		}
	}

	return r.ValidateUpdate()
}

type Release struct {
	ReleaseInput
	CreatedAt chrono.Time `json:"createdAt" db:"created_utc"`
	UpdatedAt chrono.Time `json:"updatedAt" db:"updated_utc"`
}

func NewRelease(input ReleaseInput) Release {
	return Release{
		ReleaseInput: input,
		CreatedAt:    chrono.TimeNow(),
	}
}

func (r Release) Update(i ReleaseInput) Release {
	r.ReleaseInput = i
	r.UpdatedAt = chrono.TimeNow()

	return r
}

type ReleaseList struct {
	pkg.PagedList
	Data []Release `json:"data"`
}
