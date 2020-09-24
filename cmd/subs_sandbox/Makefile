build_dir := ../../build
config_file := api.toml
BINARY := subs_sandbox

DEV_OUT := $(build_dir)/$(api_name)
LINUX_OUT := $(build_dir)/linux/$(api_name)

LOCAL_CONFIG_FILE := $(HOME)/config/$(config_file)

VERSION := `git describe --tags`
BUILD := `date +%FT%T%z`
COMMIT := `git log --max-count=1 --pretty=format:%aI_%h`

LDFLAGS := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD} -X main.commit=${COMMIT}"

BUILD_LINUX := GOOS=linux GOARCH=amd64 go build -o $(api_prod_out) $(ldflags) -tags production -v .

.PHONY: local run linux config deploy build downconfig upconfig publish restart clean
# Development
local :
	go build $(ldflags) -o $(api_dev_out) -v .

# Run development build
run :
	./$(api_dev_out) -sandbox

# Cross compiling linux on for dev.
linux :
	$(linux_api)

version :
	echo 'Build sandbox version'
	echo $(VERSION)

# From local machine to production server
# Copy env varaible to server
config :
	rsync -v $(LOCAL_CONFIG_FILE) tk11:/home/node/config

deploy : version
	rsync -v $(api_prod_out) tk11:/home/node/go/bin/
	ssh tk11 supervisorctl restart $(api_name)

# For CI/CD
build : version
	gvm install go1.15
	gvm use go1.15
	$(linux_api)

downconfig :
	rsync -v tk11:/home/node/config/$(config_file) ./$(build_dir)

# Publish artifacts.
upconfig :
	rsync -v ./$(build_dir)/$(config_file) ucloud:/home/node/config

publish :
	ssh ucloud "rm -f /home/node/go/bin/$(api_name).bak"
	rsync -v $(api_prod_out) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(api_name) ucloud:/home/node/go/bin/$(api_name).bak"
#	scp -rp $(LINUX_OUT) ucloud:/home/node/go/bin/$(BINARY).bak

restart :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(api_name).bak $(api_name)"
	ssh ucloud supervisorctl restart $(api_name)

clean :
	go clean -x
	rm build/*
