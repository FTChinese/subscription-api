build_dir := build
BINARY := subscription-api

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}"

.PHONY: build run linux publish downconfig upconfig restart cofig upload deploy clean
# Development
build :
	go build $(LDFLAGS) -o $(build_dir)/$(BINARY) -v .

run :
	./$(build_dir)/${BINARY}

# For CI/CD
linux :
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(build_dir)/linux/$(BINARY) -v .

publish :
	rsync -v $(build_dir)/linux/$(BINARY) ucloud:/home/node/go/bin/

downconfig :
	rsync -v tk11:/home/node/config/api.toml ./build

upconfig :
	rsync -v ./api.toml ucloud:/home/node/config

restart :
	ssh ucloud supervisorctl restart $(BINARY)

# From local machine to production server
# Copy env varaible to server
config :
	rsync -v $(HOME)/config/api.toml tk11:/home/node/config

deploy : linux config
	rsync -v $(build_dir)/linux/$(BINARY) tk11:/home/node/go/bin/

clean :
	go clean -x
	rm build/*