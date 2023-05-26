config_file_name := api.toml
local_config_file := $(HOME)/config/$(config_file_name)

version := `git describe --tags`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=${version} -X main.build=${build_time} -X main.commit=${commit}"

app_name := superyard
go_version := go1.18.1

sys := $(shell uname -s)
hardware := $(shell uname -m)
build_dir := build
src_dir := .

default_exec := $(build_dir)/$(sys)/$(hardware)/$(app_name)
compile_default_exec := go build -o $(default_exec) $(ldflags) -tags production -v $(src_dir)

linux_x86_exec := $(build_dir)/linux/x86/$(app_name)
compile_linux_x86 := GOOS=linux GOARCH=amd64 go build -o $(linux_x86_exec) $(ldflags) -tags production -v $(src_dir)

linux_arm_exec := $(build_dir)/linux/arm/$(app_name)
compile_linux_arm := GOOS=linux GOARM=7 GOARCH=arm go build -o $(linux_arm_exec) $(ldflags) -tags production -v $(src_dir)

.PHONY: build
build :
	#gvm use $(go_version)
	which go
	go version
	@echo "GOROOT=$(GOROOT)"
	@echo "GOPATH=$(GOPATH)"
	@echo "GOBIN=$(GOBIN)"
	@echo "GO111MODULEON=$(GO111MODULEON)"
	@echo "Build version $(version)"
	$(compile_default_exec)

devconfig :
	rsync -v $(local_config_file) $(build_dir)/$(config_file_name)

.PHONY: run
run :
	$(default_exec)

.PHONY: amd64
amd64 :
	@echo "Build production linux version $(version)"
	$(compile_linux_x86)

.PHONY: arm
arm :
	@echo "Build production arm version $(version)"
	$(compile_linux_arm)

.PHONY: install-go
install-go:
	@echo "Install go version $(go_version)"
	gvm install $(go_version)

.PHONY: config
config :
	mkdir -p ./$(build_dir)
	rsync -v node@tk11:/home/node/config/$(config_file_name) ./$(build_dir)/$(config_file_name)

.PHONY: publish
publish :
	rsync -v $(default_exec) /data/opt/server/API/go/bin/$(app_name)

.PHONY: restart
restart :
	supervisorctl restart superyard

.PHONY: clean
clean :
	go clean -x
	rm build/*

