/*
Package access controls access right of a user or app to all API endpoints

When you are accessing the API directory from browser, add you access token to query parameter `access_key`;

When used by an app, set the token as the value of Bearer Token:

```
Authorization: Bearer <token>
```
*/
package access

import (
	"testing"

	"github.com/FTChinese/subscription-api/pkg/db"
	"github.com/patrickmn/go-cache"
)

func TestEnv_retrieveFromDB(t *testing.T) {
	gormDBs := db.MockGorm()

	type fields struct {
		dbs     db.ReadWriteMyDBs
		cache   *cache.Cache
		gormDBs db.MultiGormDBs
	}
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    OAuth
		wantErr bool
	}{
		{
			name: "retrive access token",
			fields: fields{
				gormDBs: gormDBs,
			},
			args: args{
				token: "9bcfafe3c1cc3883f16008452d2a66e8f4a320f3",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := Env{
				cache:   tt.fields.cache,
				gormDBs: tt.fields.gormDBs,
			}
			got, err := env.retrieveFromDB(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Env.retrieveFromDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("Env.retrieveFromDB() = %v, want %v", got, tt.want)
			// }

			t.Logf("%v", got)
		})
	}
}
