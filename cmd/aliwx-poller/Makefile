version := `git tag -l --sort=-v:refname | head -n 1`
build_time := `date +%FT%T%z`
commit := `git log --max-count=1 --pretty=format:%aI_%h`

ldflags := -ldflags "-w -s -X main.version=$(version) -X main.build=$(build_time) -X main.commit=$(commit)"

app_name := aliwx-poller
src_dir := .

build_dir := ../../build
executable := $(build_dir)/$(app_name)
linux_exec := $(build_dir)/linux/$(app_name)

goos := GOOS=linux GOARCH=amd64

build_executable := go build -o $(executable) $(ldflags) -tags production -v $(src_dir)
build_linux := go build -o $(linux_exec) $(ldflags) -tags production -v $(src_dir)

# Development
.PHONY: dev
dev :
	@echo "Build dev version $(version)"
	$(build_executable)

.PHONY: run
run :
	$(executable)

# For CI/CD
.PHONY: build
build :
	@echo "Build production version $(version)"
	$(goos) $(build_linux)

.PHONY: publish
publish :
	rsync -v $(linux_exec) tk11:/home/node/go/bin/

.PHONY: restart
restart :
	ssh tk11 supervisorctl restart $(app_name)

.PHONY: deploy
deploy : build publish restart
	@echo "Deploy finished"

.PHONY: clean
clean :
	go clean -x
	rm build/*
