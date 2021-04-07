ENV := production

config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file_name)

version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"

app_name := subs-api-v2
build_dir := build
src_dir := .
go_version := go1.15

ifeq ($(ENV), sandbox)
	app_name := subs_sandbox
	src_dir := ./cmd/subs_sandbox/
endif

executable := $(build_dir)/$(app_name)
compile_exec := go build -o $(executable) $(ldflags) -tags production -v $(src_dir)

# Development
.PHONY: dev
dev :
	@echo "Build dev version $(version)"
	$(compile_exec)

.PHONY: run
run :
	$(executable) -sandbox=true

.PHONY: arm
arm :
	GOOS=linux GOARM=7 GOARCH=arm $(compile_exec)

.PHONY: install-go
install-go:
	gvm install $(go_version)
	gvm use $(go_version)

# For CI/CD
.PHONY: build
build : install-go
	@echo "Build production version $(version)"
	GOOS=linux GOARCH=amd64 $(compile_exec)

.PHONY: config
config :
	rsync -v tk11:/home/node/config/$(config_file_name) ./$(build_dir)
	rsync -v ./$(build_dir)/$(config_file_name) ucloud:/home/node/config
	#rsync -v ./configs/subs-api-v2.conf ucloud:/etc/supervisor/

.PHONY: publish
publish :
	ssh ucloud "rm -f /home/node/go/bin/$(app_name).bak"
	rsync -v ./$(executable) bj32:/home/node
	ssh bj32 "rsync -v /home/node/$(app_name) ucloud:/home/node/go/bin/$(app_name).bak"


.PHONY: restart
restart :
	ssh ucloud "cd /home/node/go/bin/ && \mv $(app_name).bak $(app_name)"
	#ssh ucloud supervisorctl update
	ssh ucloud supervisorctl restart $(app_name)

.PHONY: clean
clean :
	go clean -x
	rm -rf build/*
