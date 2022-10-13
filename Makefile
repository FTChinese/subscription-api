config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file_name)

app_name := subs-api-v6
go_version := go1.18

current_dir := $(shell pwd)
sys := $(shell uname -s)
hardware := $(shell uname -m)
src_dir := $(current_dir)
out_dir := $(current_dir)/out
build_dir := $(current_dir)/build

default_exec := $(out_dir)/$(sys)/$(hardware)/$(app_name)

linux_x86_exec := $(out_dir)/linux/x86/$(app_name)

linux_arm_exec := $(out_dir)/linux/arm/$(app_name)

server_dir := /data/node/go/bin

.PHONY: build
build : version
	go build -o $(default_exec) -tags production -v $(src_dir)

.PHONY: run
run :
	$(default_exec) -production=false -livemode=false

.PHONY: version
version :
	git describe --tags > build/version
	git log --max-count=1 --pretty=format:%aI_%h > build/commit
	date +%FT%T%z > build/build_time

.PHONY: amd64
amd64 :
	@echo "Build production linux version"
	GOOS=linux GOARCH=amd64 go build -o $(linux_x86_exec) -tags production -v $(src_dir)

.PHONY: arm
arm :
	@echo "Build production arm version"
	GOOS=linux GOARM=7 GOARCH=arm go build -o $(linux_arm_exec) -tags production -v $(src_dir)

.PHONY: install-go
install-go:
	@echo "* Install go version $(go_version)"
	gvm install $(go_version)

.PHONY: config
config : builddir
	@echo "* Pulling config  file from server"
	# Download configuration file
	rsync -v node@tk11:/home/node/config/$(config_file_name) $(build_dir)/$(config_file_name)

.PHONY: publish
publish :
	# Remove the .bak file
	ssh ucloud "rm -f $(server_dir)/$(app_name).bak"
	# Sync binary to the xxx.bak file
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

.PHONY: builddir
builddir :
	mkdir -p $(build_dir)

.PHONY: devenv
devenv : builddir
	rsync $(HOME)/config/env.dev.toml $(build_dir)/$(config_file_name)
	mkdir -p ./cmd/aliwx-poller/build
	rsync $(local_config_file) ./cmd/aliwx-poller/build/$(config_file_name)
	mkdir -p ./cmd/iap-poller/build
	rsync $(local_config_file) ./cmd/iap-poller/build/$(config_file_name)
	mkdir -p ./cmd/subs_sandbox/build
	rsync $(local_config_file) ./cmd/subs_sandbox/build/$(config_file_name)

.PHONY: dockerconfig
dockerenv : builddir
	rsync $(HOME)/config/env.docker.toml $(build_dir)/$(config_file_name)

.PHONY: network
network :
	docker network create my-api

.PHONY: mysql
mysql :
	docker run -d -p 3306:3306 --network my-api --network-alias mysql --name dev-mysql -v api-mysql-data:/var/lib/mysql -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:5.7 --default-time-zone='+00:00'

.PHONY : dockerbuild
dockerbuild :
	docker build -t subs-api .

.PHONY : dockerrun
dockerrun :
	docker run --name subs-api -p 8206:8206 --network my-api subs-api
