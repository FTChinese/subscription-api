# One of `api | sandbox | consumer`.
APP := api

version := `git describe --tags`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

executable := subscription-api
port := 8200
src_dir := .

ifeq ($(APP), sandbox)
	executable := subs_sandbox
	port := 8201
endif 

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit) -X main.port=$(port)"

ifeq ($(APP), consumer)
	executable := iap-kafka-consumer
	src_dir := ./cmd/iap-kafka-consumer/
	ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"
endif

build_dir := build
executable := $(build_dir)/$(executable)

goos := GOOS=linux GOARCH=amd64
go_version := go1.15

config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file)

build_executable := go build -o $(executable) $(ldflags) -tags production -v $(src_dir)

# Development
.PHONY: dev
dev :
	$(build_executable)

.PHONY: run
run :
	./$(executable) -sandbox=true

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
	rsync -v ./$(build_dir)/$(config_file) ucloud:/home/node/config

.PHONY: publish
publish :
	ssh ucloud "rm -f /home/node/go/bin/$(executable).bak"
	rsync -v $(executable) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(executable) ucloud:/home/node/go/bin/$(executable).bak"

.PHONY: restart
restart :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(executable).bak $(executable)"
	ssh ucloud supervisorctl restart $(executable)

.PHONY: clean
clean :
	go clean -x
	rm build/*
