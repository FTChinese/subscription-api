package repository

import (
	gorest "github.com/FTChinese/go-rest"
	"github.com/FTChinese/subscription-api/internal/pkg/legal"
	"github.com/FTChinese/subscription-api/pkg/db"
	"go.uber.org/zap/zaptest"
	"testing"
)

func TestLegalRepo_Create(t *testing.T) {

	repo := NewLegalRepo(db.MockMySQL(), zaptest.NewLogger(t))

	type args struct {
		l legal.Legal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Create",
			args:    args{l: legal.NewMockLegal()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repo.Create(tt.args.l); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLegalRepo_Update(t *testing.T) {

	repo := NewLegalRepo(db.MockMySQL(), zaptest.NewLogger(t))

	l := legal.NewMockLegal()

	_ = repo.Create(l)

	type args struct {
		l legal.Legal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Update",
			args: args{
				l: l,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := repo.Update(tt.args.l); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLegalRepo_Retrieve(t *testing.T) {

	repo := NewLegalRepo(db.MockMySQL(), zaptest.NewLogger(t))
	l := legal.NewMockLegal()

	_ = repo.Create(l)

	type args struct {
		title string
	}
	tests := []struct {
		name    string
		args    args
		want    legal.Legal
		wantErr bool
	}{
		{
			name: "Retrieve",
			args: args{
				title: l.TitleEn,
			},
			want:    legal.Legal{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.Retrieve(tt.args.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrieve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("Retrieve() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}

func TestLegalRepo_ListLegal(t *testing.T) {
	repo := NewLegalRepo(db.MockMySQL(), zaptest.NewLogger(t))
	l := legal.NewMockLegal()

	_ = repo.Create(l)

	type args struct {
		p gorest.Pagination
	}
	tests := []struct {
		name    string
		args    args
		want    legal.List
		wantErr bool
	}{
		{
			name: "List",
			args: args{
				p: gorest.NewPagination(1, 20),
			},
			want:    legal.List{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := repo.ListLegal(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListLegal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("ListLegal() got = %v, want %v", got, tt.want)
			//}

			t.Logf("%v", got)
		})
	}
}
