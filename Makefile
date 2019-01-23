build_dir := build
BINARY := subscription-api

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}"

.PHONY: build run linux deploy cofig clean
build :
	go build $(LDFLAGS) -o $(build_dir)/mac/$(BINARY) -v .

run :
	./$(build_dir)/mac/${BINARY}

sandbox :
	./$(build_dir)/mac/$(BINARY) -sandbox

deploy : linux
	rsync -v $(build_dir)/linux/$(BINARY) nodeserver:/home/node/go/bin/

# Copy env varaible to server
config :
	rsync -v $(HOME)/config/api.toml nodeserver:/home/node/config

linux : 
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(build_dir)/linux/$(BINARY) -v .

clean :
	go clean -x
	rm build/*