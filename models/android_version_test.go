package models

import (
	"reflect"
	"testing"
)

func TestParseSemVer(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want SemVer
	}{
		{
			name: "Parse a semantic version",
			args: args{v: "3.2.0"},
			want: SemVer{
				Major: 3,
				Minor: 2,
				Patch: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSemVer(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSemVer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_Larger(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
	}
	type args struct {
		other SemVer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Larger",
			fields: fields{
				Major: 3,
				Minor: 2,
				Patch: 0,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: true,
		},
		{
			name: "Equal",
			fields: fields{
				Major: 3,
				Minor: 1,
				Patch: 3,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: false,
		},
		{
			name: "Smaller",
			fields: fields{
				Major: 3,
				Minor: 1,
				Patch: 2,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SemVer{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
			}
			if got := s.Larger(tt.args.other); got != tt.want {
				t.Errorf("Larger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_Compare(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
	}
	type args struct {
		other SemVer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Larger",
			fields: fields{
				Major: 3,
				Minor: 2,
				Patch: 0,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: 1,
		},
		{
			name: "Equal",
			fields: fields{
				Major: 3,
				Minor: 1,
				Patch: 3,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: 0,
		},
		{
			name: "Smaller",
			fields: fields{
				Major: 3,
				Minor: 1,
				Patch: 2,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SemVer{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
			}
			if got := s.Compare(tt.args.other); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSemVer_Smaller(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
	}
	type args struct {
		other SemVer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Larger",
			fields: fields{
				Major: 3,
				Minor: 2,
				Patch: 0,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: false,
		},
		{
			name: "Equal",
			fields: fields{
				Major: 3,
				Minor: 1,
				Patch: 3,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: false,
		},
		{
			name: "Smaller",
			fields: fields{
				Major: 3,
				Minor: 1,
				Patch: 2,
			},
			args: args{
				other: SemVer{
					Major: 3,
					Minor: 1,
					Patch: 3,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SemVer{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
			}
			if got := s.Smaller(tt.args.other); got != tt.want {
				t.Errorf("Smaller() = %v, want %v", got, tt.want)
			}
		})
	}
}
