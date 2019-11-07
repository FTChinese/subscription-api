build_dir := build
config_file := api.toml
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
	GOOS=linux GOARCH=amd64 /opt/server/go/bin/go build $(LDFLAGS) -o $(build_dir)/linux/$(BINARY) -v .

publish :
	scp -rp $(build_dir)/linux/$(BINARY) ucloud:/home/node/go/bin/

downconfig :
	rsync -v tk11:/home/node/config/$(config_file) ./$(build_dir)

upconfig :
	rsync -v ./$(build_dir)/$(config_file) ucloud:/home/node/config

restart :
	ssh ucloud supervisorctl restart $(BINARY)

# From local machine to production server
# Copy env varaible to server
config :
	rsync -v $(HOME)/config/$(config_file) tk11:/home/node/config

deploy : linux config
	rsync -v $(build_dir)/linux/$(BINARY) tk11:/home/node/go/bin/

clean :
	go clean -x
	rm build/*