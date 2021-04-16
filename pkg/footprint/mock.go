// +build !production

package footprint

import (
	"github.com/FTChinese/go-rest/chrono"
	"github.com/FTChinese/go-rest/enum"
	"github.com/FTChinese/go-rest/rand"
	"github.com/FTChinese/subscription-api/faker"
	"github.com/brianvoe/gofakeit/v5"
	"github.com/google/uuid"
	"github.com/guregu/null"
)

func MockClient(ip string) Client {
	faker.SeedGoFake()

	if ip == "" {
		ip = gofakeit.IPv4Address()
	}

	return Client{
		Platform:  enum.Platform(rand.IntRange(1, 10)),
		Version:   null.StringFrom(faker.GenVersion()),
		UserIP:    null.StringFrom(ip),
		UserAgent: null.StringFrom(gofakeit.UserAgent()),
	}
}

func randomSource() Source {
	i := rand.IntRange(1, 5)

	return []Source{
		SourceNull,
		SourceLogin,
		SourceSignUp,
		SourceVerification,
		SourcePasswordReset,
	}[i]
}

type MockFootprintBuilder struct {
	ftcID      string
	ip         string
	client     Client
	authMethod enum.LoginMethod
	source     Source
}

func NewMockFootprintBuilder(ip string) MockFootprintBuilder {
	faker.SeedGoFake()

	if ip == "" {
		faker.SeedGoFake()
		ip = gofakeit.IPv4Address()
	}

	return MockFootprintBuilder{
		ftcID:      "",
		ip:         ip,
		client:     MockClient(ip),
		authMethod: enum.LoginMethodNull,
		source:     SourceNull,
	}
}

func (b MockFootprintBuilder) WithUserID(id string) MockFootprintBuilder {
	b.ftcID = id
	return b
}

func (b MockFootprintBuilder) WithAuthMethod(m enum.LoginMethod) MockFootprintBuilder {
	b.authMethod = m
	return b
}

func (b MockFootprintBuilder) WithSource(s Source) MockFootprintBuilder {
	b.source = s
	return b
}

func (b MockFootprintBuilder) Build() Footprint {
	ftcID := b.ftcID
	if ftcID == "" {
		ftcID = uuid.New().String()
	}

	authMethod := b.authMethod
	if authMethod == enum.LoginMethodNull {
		authMethod = enum.LoginMethod(rand.IntRange(1, 4))
	}

	source := b.source
	if source == "" {
		source = randomSource()
	}

	var deviceToken string
	if source == SourceSignUp {
		deviceToken = uuid.New().String()
	}

	return Footprint{
		FtcID:       ftcID,
		Client:      b.client,
		CreatedUTC:  chrono.TimeNow(),
		Source:      source,
		AuthMethod:  authMethod,
		DeviceToken: null.NewString(deviceToken, deviceToken != ""),
	}
}

func (b MockFootprintBuilder) BuildN(n int) []Footprint {
	var fs = make([]Footprint, 0)

	for i := 0; i < n; i++ {
		fs = append(fs, b.Build())
	}

	return fs
}
