package apprepo

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg/android"
	"github.com/FTChinese/subscription-api/pkg/db"
	"testing"
)

func TestEnv_CreateRelease(t *testing.T) {

	env := New(db.MockMySQL())

	type args struct {
		r android.Release
	}
	tests := []struct {
		name    string
		args    args
		want    android.Release
		wantErr bool
	}{
		{
			name: "Insert a new release",
			args: args{
				r: android.NewMockRelease(),
			},
			want:    android.Release{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := env.CreateRelease(tt.args.r)

			t.Logf("Save release %v", tt.args.r)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestEnv_UpdateRelease(t *testing.T) {

	env := New(db.MockMySQL())

	r := android.NewMockRelease()
	env.CreateRelease(r)

	type args struct {
		release android.Release
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update release",
			args: args{
				release: r,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.UpdateRelease(tt.args.release); (err != nil) != tt.wantErr {
				t.Errorf("UpdateRelease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnv_RetrieveRelease(t *testing.T) {
	env := New(db.MockMySQL())

	r := android.NewMockRelease()
	env.CreateRelease(r)

	type args struct {
		versionName string
	}
	tests := []struct {
		name    string
		args    args
		want    android.Release
		wantErr bool
	}{
		{
			name: "Retrieve release",
			args: args{
				versionName: r.VersionName,
			},
			want:    android.Release{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.RetrieveRelease(tt.args.versionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetrieveRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("RetrieveRelease() got = %v, want %v", got, tt.want)
			//}

			t.Logf("Got %v", got)
		})
	}
}

func TestEnv_ListReleases(t *testing.T) {
	env := New(db.MockMySQL())

	type args struct {
		p gorest.Pagination
	}
	tests := []struct {
		name    string
		args    args
		want    android.ReleaseList
		wantErr bool
	}{
		{
			name: "List releases",
			args: args{
				p: gorest.NewPagination(1, 20),
			},
			want:    android.ReleaseList{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := env.ListReleases(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListReleases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListReleases() got = %v, want %v", got, tt.want)
			//}

			t.Logf("Got %v", got)
		})
	}
}

func TestEnv_DeleteRelease(t *testing.T) {
	env := New(db.MockMySQL())

	r := android.NewMockRelease()
	env.CreateRelease(r)

	type args struct {
		versionName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Delete release",
			args: args{
				versionName: r.VersionName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := env.DeleteRelease(tt.args.versionName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteRelease() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
