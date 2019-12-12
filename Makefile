build_dir := build
config_file := api.toml
BINARY := subscription-api

DEV_OUT := $(build_dir)/$(BINARY)
LINUX_OUT := $(build_dir)/linux/$(BINARY)

LOCAL_CONFIG_FILE := $(HOME)/config/$(config_file)

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD}"

BUILD_LINUX := GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(LINUX_OUT) -v .

.PHONY: local run linux config deploy build downconfig upconfig publish restart clean
# Development
local :
	go build $(LDFLAGS) -o $(DEV_OUT) -v .

# Run development build
run :
	./$(DEV_OUT)

# Cross compiling linux on for dev.
linux :
	$(BUILD_LINUX)

# From local machine to production server
# Copy env varaible to server
config :
	rsync -v $(LOCAL_CONFIG_FILE) tk11:/home/node/config

deploy : config linux
	rsync -v $(LINUX_OUT) tk11:/home/node/go/bin/
	ssh tk11 supervisorctl restart $(BINARY)

# For CI/CD
build :
	gvm install go1.13.4
	gvm use go1.13.4
	$(BUILD_LINUX)

downconfig :
	rsync -v tk11:/home/node/config/$(config_file) ./$(build_dir)

# Publish artifacts.
upconfig :
	rsync -v ./$(build_dir)/$(config_file) ucloud:/home/node/config

publish :
	scp -rp $(LINUX_OUT) ucloud:/home/node/go/bin/$(BINARY).bak

restart :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(BINARY).bak $(BINARY)"
	ssh ucloud supervisorctl restart $(BINARY)

clean :
	go clean -x
	rm build/*