version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"

app_name := aliwx-poller
src_dir := .

sys := $(shell uname -s)
hardware := $(shell uname -m)

build_dir := ../../build

default_exec := $(build_dir)/$(sys)/$(hardware)/$(app_name)
compile_default_exec := go build -o $(default_exec) $(ldflags) -tags production -v $(src_dir)

linux_x86_exec := $(build_dir)/linux/x86/$(app_name)
compile_linux_x86 := GOOS=linux GOARCH=amd64 go build -o $(linux_x86_exec) $(ldflags) -tags production -v $(src_dir)

linux_arm_exec := $(build_dir)/linux/arm/$(app_name)
compile_linux_arm := GOOS=linux GOARM=7 GOARCH=arm go build -o $(linux_arm_exec) $(ldflags) -tags production -v $(src_dir)

# Development
.PHONY: build
build :
	@echo "Build dev version $(version)"
	$(compile_default_exec)

.PHONY: run
run :
	$(default_exec)

# For CI/CD
.PHONY: amd64
amd64 :
	@echo "Build production linux version $(version)"
	$(compile_linux_x86)

.PHONY: arm
arm :
	@echo "Build production arm version $(version)"
	$(compile_linux_arm)

.PHONY: publish
publish :
	rsync -v $(linux_x86_exec) tk11:/home/node/go/bin/

.PHONY: restart
restart :
	ssh tk11 supervisorctl restart $(app_name)

.PHONY: deploy
deploy : amd64 publish restart
	@echo "Deploy finished"

.PHONY: clean
clean :
	go clean -x
	rm ../..build/**
