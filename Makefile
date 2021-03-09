config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file_name)

version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"

app_name := subs-api-v2

build_dir := build

go_version := go1.15

dev_executable := $(build_dir)/$(app_name)
build_dev := go build -o $(dev_executable) $(ldflags) -tags production -v .

linux_executable := $(build_dir)/linux/$(app_name)
build_linux := go build -o $(linux_executable) $(ldflags) -tags production -v .

# Development
.PHONY: dev
dev :
	@echo "Build dev version $(version)"
	$(build_dev)

.PHONY: run
run :
	$(dev_executable) -sandbox=true



.PHONY: arm
arm :
	GOOS=linux GOARM=7 GOARCH=arm $(build_linux)

.PHONY: install-go
install-go:
	gvm install $(go_version)
	gvm use $(go_version)

# For CI/CD
.PHONY: amd64
amd64 :
	@echo "Build production version $(version)"
	GOOS=linux GOARCH=amd64 $(build_linux)

.PHONY: config
config :
	rsync -v tk11:/home/node/config/$(config_file_name) ./$(build_dir)
	rsync -v ./$(build_dir)/$(config_file_name) ucloud:/home/node/config

.PHONY: publish
publish :
	ssh ucloud "rm -f /home/node/go/bin/$(app_name).bak"
	rsync -v $(linux_executable) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(app_name) ucloud:/home/node/go/bin/$(app_name).bak"
	rsync -v ./configs/subs-api-v2.conf ucloud:/etc/supervisor

.PHONY: restart
restart :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(app_name).bak $(app_name)"
	ssh ucloud supervisorctl restart $(app_name)

test-deploy : build
	rsync -v $(dev_executable) tk11:/home/node/go/bin

.PHONY: clean
clean :
	go clean -x
	rm -rf build/*
