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

# make默认执行第一个命令。
# 这个命令使用`go build`生成开发版的二进制文件，生成的二进制文件在`out`目录下。
# 该命令中的`-tags production`是告诉编译器忽略首行包含`//go:build !production`的文件，这些文件通常是测试用的mock数据。
.PHONY: build
build : version
	go build -o $(default_exec) -tags production -v $(src_dir)

# 在dev和sandbox模式运行编译的二进制。
.PHONY: run
run :
	$(default_exec) -production=false -livemode=false

# Create `build` direcetory under the project root, 
# since some tools cannot create this dir automatically.
.PHONY: builddir
builddir :
	mkdir -p $(build_dir)

# Copy configuration from your computer's `~/config` directory.
.PHONY: devenv
devenv : builddir
	rsync $(HOME)/config/env.dev.toml $(build_dir)/$(config_file_name)
	mkdir -p ./cmd/aliwx-poller/build
	rsync $(local_config_file) ./cmd/aliwx-poller/build/$(config_file_name)
	mkdir -p ./cmd/iap-poller/build
	rsync $(local_config_file) ./cmd/iap-poller/build/$(config_file_name)
	mkdir -p ./cmd/subs_sandbox/build
	rsync $(local_config_file) ./cmd/subs_sandbox/build/$(config_file_name)

# Generate build_time, commit, and version file
# under build directory.
.PHONY: version
version :
	git describe --tags > build/version
	git log --max-count=1 --pretty=format:%aI_%h > build/commit
	date +%FT%T%z > build/build_time

# Build binary for Linux on 64-bit CPU.
.PHONY: amd64
amd64 :
	@echo "Build production linux version"
	GOOS=linux GOARCH=amd64 go build -o $(linux_x86_exec) -tags production -v $(src_dir)

# Build binary for Linux on ARM.
.PHONY: arm
arm :
	@echo "Build production arm version"
	GOOS=linux GOARM=7 GOARCH=arm go build -o $(linux_arm_exec) -tags production -v $(src_dir)

# Install go using gvm.
# It seems not working on production server since there are problems with gvm installation.
.PHONY: install-go
install-go:
	@echo "* Install go version $(go_version)"
	gvm install $(go_version)

# Sync env file from tk11 server to build directory
# before building binary in Jenkins.
.PHONY: config
config : builddir
	@echo "* Pulling config  file from server"
	# Download configuration file
	rsync -v node@tk11:/home/node/config/$(config_file_name) $(build_dir)/$(config_file_name)

# After Jenkins build binary into out directory,
# use rsync to upload the binary to a production server on ucloud.
.PHONY: publish
publish :
	# Remove the .bak file
	ssh ucloud "rm -f $(server_dir)/$(app_name).bak"
	# Sync binary to the xxx.bak file
	rsync -v $(default_exec) ucloud:$(server_dir)/$(app_name).bak

# After uploaded binary to produciton server,
# use ssh to log into the server to execute
# superviser command to restart the new binary.
.PHONY: restart
restart :
	# Rename xxx.bak to app name
	ssh ucloud "cd $(server_dir)/ && \mv $(app_name).bak $(app_name)"
	# Update supervistor
	#ssh ucloud supervisorctl update all
	ssh ucloud supervisorctl restart $(app_name)

# Clear up intermidiate files.
.PHONY: clean
clean :
	go clean -x
	rm -rf build/*

# The following demonstrate how to build and run
# this app with docker. Not in production.
# Sync a env file tailored for docker.
.PHONY: dockerconfig
dockerenv : builddir
	rsync $(HOME)/config/env.docker.toml $(build_dir)/$(config_file_name)

# Create docker network
.PHONY: network
network :
	docker network create my-api

# Run a mysql instance in docker.
.PHONY: mysql
mysql :
	docker run -d -p 3306:3306 --network my-api --network-alias mysql --name dev-mysql -v api-mysql-data:/var/lib/mysql -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:5.7 --default-time-zone='+00:00'

# Build binary in docker.
.PHONY : dockerbuild
dockerbuild :
	docker build -t subs-api .

# Run containerized binary file.
.PHONY : dockerrun
dockerrun :
	docker run --name subs-api -p 8206:8206 --network my-api subs-api
