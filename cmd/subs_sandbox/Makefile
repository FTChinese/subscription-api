config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file_name)

version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"

app_name := subs_sandbox
go_version := go1.16

current_dir := $(shell pwd)
sys := $(shell uname -s)
hardware := $(shell uname -m)
build_dir := $(current_dir)/build
src_dir := $(current_dir)

default_exec := $(build_dir)/$(sys)/$(hardware)/$(app_name)
compile_default_exec := go build -o $(default_exec) $(ldflags) -tags production -v $(src_dir)

linux_x86_exec := $(build_dir)/linux/x86/$(app_name)
compile_linux_x86 := GOOS=linux GOARCH=amd64 go build -o $(linux_x86_exec) $(ldflags) -tags production -v $(src_dir)

linux_arm_exec := $(build_dir)/linux/arm/$(app_name)
compile_linux_arm := GOOS=linux GOARM=7 GOARCH=arm go build -o $(linux_arm_exec) $(ldflags) -tags production -v $(src_dir)

server_dir := /data/node/go/bin

.PHONY: build
build : outdir
	@echo "* Build from dir $(current_dir)"
	$(compile_default_exec)

.PHONY: outdir
outdir :
	mkdir -p $(build_dir)

.PHONY: devconfig
devconfig : outdir
	rsync $(local_config_file) $(build_dir)/$(config_file_name)

.PHONY: run
run :
	$(default_exec) -sandbox=true

.PHONY: amd64
amd64 :
	@echo "* Build production linux version $(version)"
	$(compile_linux_x86)

.PHONY: arm
arm :
	@echo "* Build production arm version $(version)"
	$(compile_linux_arm)

.PHONY: install-go
install-go:
	@echo "* Install go version $(go_version)"
	gvm install $(go_version)

.PHONY: config
config : outdir
	@echo "* Pulling config  file from server"
	rsync -v node@tk11:/home/node/config/$(config_file_name) $(build_dir)/$(config_file_name)

.PHONY: publish
publish :
	@echo "* Removing app binary on server side"
	ssh ucloud "rm -f $(server_dir)/$(app_name).bak"
	@echo "Uploading binary to server"
	rsync -v $(default_exec) ucloud:$(server_dir)/$(app_name).bak

.PHONY: restart
restart :
	# Rename xxx.bak to app name
	ssh ucloud "cd $(server_dir)/ && \mv $(app_name).bak $(app_name)"
	# Update supervistor
	#ssh ucloud supervisorctl update all
	ssh ucloud supervisorctl restart $(app_name)

.PHONY: clean
clean :
	go clean -x
	rm -rf build/*
