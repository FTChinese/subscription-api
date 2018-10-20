build_dir := build
BINARY := subscription-api

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}"

.PHONY: build linux deploy attack report lastcommit mkbuild clean
build :
	go build $(LDFLAGS) -o $(build_dir)/$(BINARY) -v .

deploy : linux
	rsync -v $(build_dir)/$(BINARY) ftaserver:/home/node/go/bin/

# Copy env varaible to server
config :
	rsync -v ../.env nodeserver:/home/node/go

linux : 
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(build_dir)/$(BINARY) -v .

mkbuild :
	mkdir -p build

clean :
	go clean -x
	rm build/*