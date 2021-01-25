# One of `api | sandbox | consumer | aliwx`.
APP := api

version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"

app_name := subscription-api-v2
src_dir := .

ifeq ($(APP), sandbox)
	app_name := subs_sandbox
	src_dir := ./cmd/subs_sandbox/
endif 

ifeq ($(APP), iap)
	app_name := iap-poller
	src_dir := ./cmd/iap-poller/
endif

ifeq ($(APP), aliwx)
	app_name := aliwx-poller
	src_dir := ./cmd/aliwx-poller/
endif

build_dir := build
executable := $(build_dir)/$(app_name)

goos := GOOS=linux GOARCH=amd64
go_version := go1.15

config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file_name)

build_executable := go build -o $(executable) $(ldflags) -tags production -v $(src_dir)

# Development
.PHONY: dev
dev :
	@echo "Build dev version $(version)"
	$(build_executable)

.PHONY: run
run :
	$(executable) -sandbox=true

# For CI/CD
.PHONY: build
build :
	@echo "Build production version $(version)"
	$(goos) $(build_executable)

.PHONY: install-go
install-go:
	gvm install $(go_version)
	gvm use $(go_version)

.PHONY: config
config :
	rsync -v tk11:/home/node/config/$(config_file_name) ./$(build_dir)
	rsync -v ./$(build_dir)/$(config_file_name) ucloud:/home/node/config

.PHONY: publish
publish :
	ssh ucloud "rm -f /home/node/go/bin/$(app_name).bak"
	rsync -v $(executable) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(app_name) ucloud:/home/node/go/bin/$(app_name).bak"

.PHONY: restart
restart :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(app_name).bak $(app_name)"
	ssh ucloud supervisorctl restart $(app_name)

test-deploy : build
	rsync -v $(executable) tk11:/home/node/go/bin

.PHONY: clean
clean :
	go clean -x
	rm build/*
