build_dir := build
BINARY := subscription-api

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD} -X main.lastCommit=${COMMIT}"

SANDBOX_FLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD} -X main.lastCommit=${COMMIT} -X gitlab.com/ftchinese/subscription-api/model.memberTable=sandbox"

.PHONY: build run linux deploy cofig clean
build :
	go build $(LDFLAGS) -o $(build_dir)/mac/$(BINARY) -v .

run :
	./$(build_dir)/mac/${BINARY}

sandbox :
	go build $(SANDBOX_FLAGS) -o $(build_dir)/sandbox/$(BINARY) -v .

runsandbox :
	./$(build_dir)/sandbox/$(BINARY)

deploy : linux
	rsync -v $(build_dir)/linux/$(BINARY) nodeserver:/home/node/go/bin/

# Copy env varaible to server
config :
	rsync -v ../.env nodeserver:/home/node/go

linux : 
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(build_dir)/linux/$(BINARY) -v .

clean :
	go clean -x
	rm build/*