# One of `api | sandbox | consumer`.
APP := api

version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

app_name := subscription-api
port := 8200
src_dir := .

ifeq ($(APP), sandbox)
	app_name := subs_sandbox
	port := 8201
endif 

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit) -X main.port=$(port)"

ifeq ($(APP), consumer)
	app_name := iap-kafka-consumer
	src_dir := ./cmd/iap-kafka-consumer/
	ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"
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
	$(build_executable)

.PHONY: run
run :
	$(executable) -sandbox=true

# For CI/CD
.PHONY: build
build :
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

.PHONY: clean
clean :
	go clean -x
	rm build/*
